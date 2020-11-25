// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Config struct field value sanitation and defaulting.

package config

import (
	"encoding"
	"reflect"
	"strings"

	"github.com/vedranvuk/errorex"
	"github.com/vedranvuk/reflectex"
)

var (
	// ErrWarning is returned when an error occurs during parsing of one
	// or more struct tags that carry a config key. ANy fields that produced
	// errors are stored in the returned error's Extra field. It is up to the
	// user to consider errors fatal or not.
	ErrWarning = ErrConfig.Wrap("warning")
	// ErrParse help.
	ErrParse = ErrConfig.Wrap("parse '%s'")
	// ErrNoTag is returned when a field in the config struct does not have a
	// Config tag defined or the tag has no defined keys.
	ErrNoTag = ErrConfig.WrapFormat("'%s' no config tag")
	// ErrInvalidTag is returned when an invalid tag is encountered on a field.
	ErrInvalidTag = ErrConfig.WrapFormat("'%s' invalid tag")
	// ErrNoRange help.
	ErrNoRange = ErrConfig.WrapFormat("'%s' no range defined")
	// ErrInvalidRange help.
	ErrInvalidRange = ErrConfig.WrapFormat("'%s' invalid range")
	// ErrNoDefault is returned when a field in the config struct does not have
	// a default value defined.
	ErrNoDefault = ErrConfig.WrapFormat("'%s' no default defined")
	// ErrInvalidDefault is returned when an invalid value was defined for
	// default field value.
	ErrInvalidDefault = ErrConfig.WrapFormat("'%s' invalid default")
	// ErrInvalidNil help.
	ErrInvalidNil = ErrConfig.WrapFormat("'%s' invalid default")
)

const (
	// ConfigTag is the name of the struct field tag read by this package.
	// It can contain multiple supported key=value pairs separated by ";".
	ConfigTag = "config"

	// NilKey is a tag that specifies the value for the field to be interpreted
	// as nil/empty for non-pointer field value types.
	NilKey = "nil"
	// RangeKey is a tag that defines range or set of values for the field.
	// Sets are sets of values delimited by ",", e.g. 1,2,3 foo,bar.
	// Ranges are min and max values separated by a ":", e.g. 0:100, 0: :100.
	// Just the ":" character is legal for a range value, and it inforces no
	// range.
	RangeKey = "range"
	// DefaultKey is a tag that defines the default value for the field.
	DefaultKey = "default"
)

// Sanitize takes a pointer to a config struct and recursively traverses
// possibly nested fields with config tags then applies the Default and Limit
// operations on those fields. For details see Default and Limit.
//
// Nested config structs are searched for in arrays, slices, maps of struct
// fields. Only non-compound typed fields and pointers to such fields can be
// defaulted, i.e. strings, bools, ints, pointers to such types and fields
// whose type implements a TextUnmarshaler.
//
// Returns ErrWarning of type (*errorex.ErrorEx) if any warnings occured with
// list of warnings retrievable via its' Extras method.
//
// Any other errors signify a no-op and a failure.
func Sanitize(config interface{}) error {
	return SanitizeValue(reflect.Indirect(reflect.ValueOf(config)))
}

// SanitizeValue is like Sanitize but takes a reflect value of config.
func SanitizeValue(v reflect.Value) error {
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return ErrInvalidParam
	}
	warnings := ErrWarning.Wrap("")
	traverse(v, "", true, true, false, true, nil, warnings)
	if len(warnings.Extras()) > 0 {
		return warnings
	}
	return nil
}

// Default takes a pointer to a config struct and recursively traverses
// possibly nested fields with config tags then sets their field values
// to values defined under default key if their values are zero of their type
// or match the value defined under nil key.
// e.g. nil for *int, 0 for int, or "" for strings.
//
// If reset is specified all fields with defined defaults are reset, regardless
// if they have already been initialized to non-nil values or their value has
// since been changed.
//
// If any errors or warnings occured it returns an ErrParseWarning of type
// *errorex.ErrorEx that contains all warnings is its' Extras field.
// It is returned under following conditions:
//
// If a field has no tag defined an ErrNoTag is appended to Extras.
// If a field has no default value an ErrNoDefault is appended to Extras.
// If a field has an incompatible/invalid default value defined an
// ErrInvalidDefault is appended to Extras.
//
// Any other errors signify a no-op and a failure.
func Default(config interface{}, reset bool) error {
	v := reflect.Indirect(reflect.ValueOf(config))
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return ErrInvalidParam
	}
	warnings := ErrWarning.Wrap("")
	traverse(v, "", true, false, reset, false, nil, warnings)
	if len(warnings.Extras()) > 0 {
		return warnings
	}
	return nil
}

// Limit takes a pointer to a config struct and recursively traverses
// possibly nested fields with config tags then if clamp was specified and
// their values are outside of defined range or set, sets their field values
// to values within the set or range, otherwise it just generates a warning.
// e.g. range=1;2;3 range=foo;bar range=0:100
//
// If clamp is specified fields with values outside of defined ranges are set
// to lowest or highest value defined, depending on boundary they exceed.
// If the value is not within the set a default operation is applied to the
// field.
//
// Limit supports enforcing sets and ranges and precognizes them as follows:
// Choices are strings separated by a ",", e.g.: 1,2,3 foo,bar,baz
// Ranges are two values separated by a ":", e.g.: 0: :100 0:100 or :
// Both range boundaries are optional, although that defeats the purpose.
// Supported kinds are String, Ints, Uints.
//
// If any errors or warnings occured it returns an ErrParseWarning of type
// *errorex.ErrorEx that contains all warnings is its' Extras field.
// It is returned under following conditions:
//
// If a field has no tag defined an ErrNoTag is appended to Extras.
// If a field has no range value an ErrNoRange is appended to Extras.
// If a field has an incompatible/invalid range value defined an
// ErrInvalidRange is appended to Extras.
//
// Any other errors signify a no-op and a failure.
func Limit(config interface{}, clamp bool) error {
	v := reflect.Indirect(reflect.ValueOf(config))
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return ErrInvalidParam
	}
	warnings := ErrWarning.Wrap("")
	traverse(v, "", false, true, false, clamp, nil, warnings)
	if len(warnings.Extras()) > 0 {
		return warnings
	}
	return nil
}

func traverse(v reflect.Value, name string, defaults, limits, reset, clamp bool, tags tagmap, warnings *errorex.ErrorEx) {
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			traverse(reflect.Indirect(v.Index(i)), name, defaults, limits, reset, clamp, tags, warnings)
		}
	case reflect.Map:
		for iter := v.MapRange(); iter.Next(); {
			traverse(reflect.Indirect(iter.Value()), name, defaults, limits, reset, clamp, tags, warnings)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			tag, ok := v.Type().Field(i).Tag.Lookup(ConfigTag)
			if !ok {
				warnings.Extra(ErrNoTag.WrapArgs(v.Type().Field(i).Name))
			}
			traverse(v.Field(i), v.Type().Field(i).Name, defaults, limits, reset, clamp, parseTagmap(tag), warnings)
		}
	case reflect.Interface:
		traverse(v.Elem(), name, defaults, limits, reset, clamp, tags, warnings)
		return
	case reflect.Ptr:
		if !v.IsZero() {
			traverse(v.Elem(), name, defaults, limits, reset, clamp, tags, warnings)
			return
		}
		fallthrough
	default:
		if !v.CanSet() {
			return
		}
		if defaults {
			setDefaults(v, name, tags, reset, warnings)
		}
		if limits {
			setLimits(v, name, tags, clamp, warnings)
		}
	}
}

func setDefaults(v reflect.Value, name string, tags tagmap, reset bool, warnings *errorex.ErrorEx) {
	var defval, nilval string
	var zero, ok bool = v.IsZero(), false
	defval, ok = tags[DefaultKey]
	if !ok {
		warnings.Extra(ErrNoDefault.WrapArgs(name))
		return
	}
	if nilval, ok = tags[NilKey]; ok {
		nv := reflect.New(v.Type())
		if err := reflectex.StringToValue(nilval, reflect.Indirect(nv)); err != nil {
			warnings.Extra(ErrInvalidNil.WrapCauseArgs(err, name))
			return
		}
		if reflectex.CompareValues(v, nv.Elem()) == 0 {
			zero = true
		}
	}
	if !zero && !reset {
		return
	}
	if tu, ok := v.Interface().(encoding.TextUnmarshaler); ok {
		if err := tu.UnmarshalText([]byte(defval)); err != nil {
			warnings.Extra(ErrInvalidDefault.WrapCauseArgs(err, name))
		}
		return
	}
	if err := reflectex.StringToValue(defval, v); err != nil {
		warnings.Extra(err)
	}
}

func setLimits(v reflect.Value, name string, tags tagmap, clamp bool, warnings *errorex.ErrorEx) {
	var rngval string
	var ok bool
	rngval, ok = tags[RangeKey]
	if !ok {
		warnings.Extra(ErrNoRange.WrapArgs(name))
		return
	}
	// Process choices.
	if strings.Contains(rngval, ",") {
		var vals []string = strings.Split(rngval, ",")
		var matched bool = false
		var cv reflect.Value = reflect.New(v.Type())
		for i := 0; i < len(vals); i++ {
			if err := reflectex.StringToValue(vals[i], reflect.Indirect(cv)); err != nil {
				warnings.Extra(ErrInvalidRange.WrapCauseArgs(err, name))
				return
			}
			if reflectex.CompareValues(v, cv.Elem()) == 0 {
				matched = true
				break
			}
		}
		if matched && !clamp {
			return
		}
		setDefaults(v, name, tags, true, warnings)
		return
	}
	// Process range.
	if strings.Contains(rngval, ":") {
		var vals []string = strings.Split(rngval, ":")
		var cv reflect.Value = reflect.New(v.Type())
		if len(vals) != 2 {
			warnings.Extra(ErrInvalidRange.WrapArgs(name))
			return
		}
		// Minimum.
		if vals[0] != "" {
			if err := reflectex.StringToValue(vals[0], reflect.Indirect(cv)); err != nil {
				warnings.Extra(ErrInvalidRange.WrapCauseArgs(err, name))
				return
			}
			if reflectex.CompareValues(v, cv.Elem()) < 0 {
				if clamp {
					v.Set(cv.Elem())
				} else {
					def, ok := tags[DefaultKey]
					if !ok {
						v.Set(reflect.Zero(v.Type()))
					}
					if err := reflectex.StringToValue(def, v); err != nil {
						warnings.Extra(ErrInvalidRange.WrapCauseArgs(err, name))
						return
					}
				}
			}
		}
		// Maximum.
		if vals[1] != "" {
			if err := reflectex.StringToValue(vals[1], reflect.Indirect(cv)); err != nil {
				warnings.Extra(ErrInvalidRange.WrapCauseArgs(err, name))
				return
			}
			if reflectex.CompareValues(v, cv.Elem()) > 0 {
				if clamp {
					v.Set(cv.Elem())
				} else {
					def, ok := tags[DefaultKey]
					if !ok {
						v.Set(reflect.Zero(v.Type()))
					}
					if err := reflectex.StringToValue(def, v); err != nil {
						warnings.Extra(ErrInvalidRange.WrapCauseArgs(err, name))
						return
					}
				}
			}
		}
	}
}

// tagmap maps tag keys to tag values.
type tagmap map[string]string

// parseTagmap returns a possibly empty tagmap parsed from the config tag
// key/value pairs.
func parseTagmap(tag string) tagmap {
	m := make(tagmap)
	for _, s := range strings.Split(tag, ";") {
		if s == "" {
			continue
		}
		kv := strings.Split(s, "=")
		if len(kv) != 2 {
			continue
		}
		m[kv[0]] = kv[1]
	}
	return m
}
