// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// CompareInterfaces compares two interfaces for equality between the types
// contained within them. See CompareValues for details.
func CompareInterfaces(a, b interface{}) int {
	return CompareValues(reflect.ValueOf(a), reflect.ValueOf(b))
}

// CompareValues recursively compares two possibly compound values a and b for
// equality. It returns:
//
// a negative number (-1) if a is less than b.
// a zero (0) if a is equal to b.
// a positive number (1) if a is more than b.
//
// Comparison is done using magic as described below:
//
// Types are compared for reflect.Kind equality first and foremost. Comparison
// logic is taken as a and b's index in reflect.Kind enumeration.
//
// For structs, only published fields are enumerated and compared. Private
// fields do not affect comparison. A struct with less public fields returns a
// less result. Structs with equal number of fields are compared alphabetically
// ascending comparing field value kinds, names and finally values.
//
// Comparisons between two Arrays and/or slices return a less result for values
// with less dimensions. In equal dimensioned arrays or slices bytes are
// compared using bytes.Compare().
//
// Maps with less elements return a less result. Maps with equal number of
// keys are compared for same kinds, then key names and finally equal values.
// Comparison is done in ascending order after converting keys to strings/bytes.
//
// Pointer types are dereferenced do their values before comparison. Untyped
// pointers are compared by their address numerically.
//
// Complex numbers are compared as strings.
//
// Channel and func types are not supported, are ignored and will return 0.
//
// If an error occurs it is returned with a compare value that should be
// disregarded.
//
func CompareValues(a, b reflect.Value) int {
	// Compare kinds.
	if res := compareKind(a.Kind(), b.Kind()); res != 0 {
		return res
	}
	// Dereference pointers to values and compare pointer depth.
	apd, bpd := 0, 0
	for a.Kind() == reflect.Ptr {
		a = a.Elem()
		apd++
	}
	for b.Kind() == reflect.Ptr {
		b = b.Elem()
		bpd++
	}
	if apd > bpd {
		return 1
	}
	if bpd > apd {
		return -1
	}
	// Compare by Kind.
	switch a.Kind() {
	case reflect.Bool:
		if a.Bool() == b.Bool() {
			return 0
		}
		if a.Bool() {
			return 1
		}
		return -1
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if res := compareKind(a.Kind(), b.Kind()); res != 0 {
			return res
		}
		if a.Int() == b.Int() {
			return 0
		}
		if a.Int() > b.Int() {
			return 1
		}
		return -1
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Uintptr:
		if res := compareKind(a.Kind(), b.Kind()); res != 0 {
			return res
		}
		if a.Uint() == b.Uint() {
			return 0
		}
		if a.Uint() > b.Uint() {
			return 1
		}
		return -1
	case reflect.Float32, reflect.Float64:
		if res := compareKind(a.Kind(), b.Kind()); res != 0 {
			return res
		}
		if a.Float() == b.Float() {
			return 0
		}
		if a.Float() > b.Float() {
			return 1
		}
		return -1
	case reflect.Complex64, reflect.Complex128:
		if res := compareKind(a.Kind(), b.Kind()); res != 0 {
			return res
		}
		if fmt.Sprint(a.Complex()) == fmt.Sprint(b.Complex()) {
			return 0
		}
		if fmt.Sprint(a.Complex()) > fmt.Sprint(b.Complex()) {
			return 1
		}
		return -1
	case reflect.Array, reflect.Slice:
		if res := compareKind(a.Kind(), b.Kind()); res != 0 {
			return res
		}
		if a.Len() == b.Len() {
			for i := 0; i < a.Len(); i++ {
				if res := CompareValues(a.Index(i), b.Index(i)); res != 0 {
					return res
				}
			}
			return 0
		}
		if a.Len() > b.Len() {
			return 1
		}
		return -1
	case reflect.Map:
		// Compare lengths.
		if a.Len() > b.Len() {
			return 1
		}
		if a.Len() < b.Len() {
			return -1
		}
		// Compare keys.
		akeys := a.MapKeys()
		bkeys := b.MapKeys()
		sort.Slice(akeys, func(i, j int) bool {
			return akeys[i].String() < akeys[j].String()
		})
		sort.Slice(bkeys, func(i, j int) bool {
			return bkeys[i].String() < bkeys[j].String()
		})
		for i := 0; i < len(akeys); i++ {
			if res := compareKind(akeys[i].Kind(), bkeys[i].Kind()); res != 0 {
				return res
			}
			if res := strings.Compare(akeys[i].String(), bkeys[i].String()); res != 0 {
				return res
			}
		}
		// Compare values.
		for i := 0; i < len(akeys); i++ {
			aval := a.MapIndex(akeys[i])
			bval := b.MapIndex(bkeys[i])
			if a.Kind() != b.Kind() {
				if res := compareKind(aval.Kind(), bval.Kind()); res != 0 {
					return res
				}
				if res := CompareValues(aval, bval); res != 0 {
					return res
				}
			}
		}
	case reflect.String:
		return strings.Compare(a.String(), b.String())
	case reflect.Struct:
		// Enum public fields.
		aflds := make([]reflect.StructField, 0, a.NumField())
		bflds := make([]reflect.StructField, 0, b.NumField())
		for i := 0; i < a.NumField(); i++ {
			if !a.Field(i).CanSet() {
				continue
			}
			aflds = append(aflds, a.Type().Field(i))
		}
		for i := 0; i < b.NumField(); i++ {
			if !b.Field(i).CanSet() {
				continue
			}
			bflds = append(bflds, b.Type().Field(i))
		}
		// Compare by public field count.
		if len(aflds) > len(bflds) {
			return 1
		}
		if len(aflds) < len(bflds) {
			return -1
		}
		// Sort the fields and compare by kind and name.
		sort.Slice(aflds, func(i, j int) bool {
			return aflds[i].Name < aflds[j].Name
		})
		sort.Slice(bflds, func(i, j int) bool {
			return bflds[i].Name < bflds[j].Name
		})
		for i := 0; i < len(aflds); i++ {
			// Compare kind.
			if res := compareKind(aflds[i].Type.Kind(), bflds[i].Type.Kind()); res != 0 {
				return res
			}
			// Compare field name.
			if res := strings.Compare(aflds[i].Name, bflds[i].Name); res != 0 {
				return res
			}
			// Compare field value.
			if res := CompareValues(a.FieldByName(aflds[i].Name), b.FieldByName(bflds[i].Name)); res != 0 {
				return res
			}
		}
	case reflect.Interface:
		return CompareValues(reflect.ValueOf(a.Interface()), reflect.ValueOf(b.Interface()))
	case reflect.Ptr, reflect.UnsafePointer:
		if a.Pointer() == b.Pointer() {
			return 0
		}
		if a.Pointer() > b.Pointer() {
			return 1
		}
		return -1

	default:
		// Return 0 by default.
	}
	return 0
}

// compareKind compares a and b reflect.Kind as integer index in enumerarion and
//
// Returns a negative number (-1) if a is less than b.
// Returns a zero (0) if a is equal to b.
// Returns a positive number (1) if a is more than b.
func compareKind(a, b reflect.Kind) int {
	if int(a) > int(b) {
		return 1
	}
	if int(a) < int(b) {
		return -1
	}
	return 0
}
