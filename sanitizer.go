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
	// ErrInvalidParam is returned when an invalid parameter is passed to a
	// function.
	ErrInvalidParam = ErrConfig.Wrap("invalid parameter")
	// ErrParseWarning is returned when an error occurs during parsing of one
	// or more struct tags that carry a config key. ANy fields that produced
	// errors are stored in the returned error's Extra field. It is up to the
	// user to consider errors fatal or not.
	ErrParseWarning = ErrConfig.Wrap("warning")
	// ErrInvalidTag is returned when an invalid tag is encountered on a field.
	ErrInvalidTag = ErrConfig.WrapFormat("invalid tag for field '%s': '%s'")
	// ErrNoTag is returned when a field in the config struct does not have a
	// Config tag defined or the tag has no defined keys.
	ErrNoTag = ErrConfig.WrapFormat("no config tag defined on field '%s'")
	// ErrNoRange help.
	ErrNoRange = ErrConfig.WrapFormat("no range value defined for field '%s'")
	// ErrNoDefault is returned when a field in the config struct does not have
	// a default value defined.
	ErrNoDefault = ErrConfig.WrapFormat("no default value defined for field '%s'")
	// ErrInvalidDefault is returned when an invalid value was defined for
	// default field value.
	ErrInvalidDefault = ErrConfig.WrapFormat("invalid default value '%s' defined for field '%s'")
)

const (
	// ConfigTag is the name of the struct field tag read by this package.
	ConfigTag = "config"

	// NilKey is a tag that specifies the value for the field to be interpreted
	// as nil/empty for non-pointer field value types.
	NilKey = "nil"
	// RangeKey is a tag that defines the range of values for the field.
	RangeKey = "range"
	// DefaultKey is a tag that defines the default value for the field.
	DefaultKey = "default"
	// OmitEmptyKey is a LoadActionTag and SaveActionTag value that specifies
	// that the value should be ommitted when saving the field to configuration
	// file if the field value is nil and/or not modified in the field when
	// there is no coresponding value in the config file being loaded.
	OmitEmptyKey = "omitempty"
)

// SetDefaults takes a pointer to a config struct, traverses all of its fields
// at any level if there are embedded structs with config tags, sets their
// default field values according to tags and, if any errors or warnings
// occured returns an ErrParseWarning containing errors in its' Extra field.
//
// If all is specified all fields are reset to defaults, otherwise only nil
// fields or fields whose values equal nil value as defined in the field tag
// are reset.
//
// ErrParseWarning is returned if any of the following conditions occur:
//
// If a field has no tag defined an ErrNoTag is appended to Extras.
// If a field has no default defined an ErrNoDefault is appended to Extras.
// If a field has an incompatible/invalid default value defined an
// ErrInvalidDefault is appended to Extras.
//
// Any other errors signify a no-op and a failure.
//
func SetDefaults(config interface{}, all bool) error {
	v := reflect.Indirect(reflect.ValueOf(config))
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return ErrInvalidParam
	}
	warnings := ErrParseWarning.Wrap("")
	setDefaults(v, all, warnings)
	if len(warnings.Extras()) > 0 {
		return warnings
	}
	return nil
}

// setDefaults is the implementation of SetDefaults.
func setDefaults(v reflect.Value, all bool, warnings *errorex.ErrorEx) {
	for i := 0; i < v.NumField(); i++ {
		switch fld := v.Field(i); fld.Kind() {
		case reflect.Array, reflect.Slice:
			for i := 0; i < fld.Len(); i++ {
				setDefaults(fld.Index(i), all, warnings)
			}
		case reflect.Map:
			for iter := fld.MapRange(); iter.Next(); {
				setDefaults(iter.Value(), all, warnings)
			}
		case reflect.Struct:
			setDefaults(fld, all, warnings)
		default:
			if !fld.CanSet() {
				continue
			}
			if fld.Type() == interfaceType {
				continue
			}
			// Get config tag.
			tag, ok := v.Type().Field(i).Tag.Lookup(ConfigTag)
			if !ok {
				warnings.Extra(ErrNoTag.WrapArgs(v.Type().Field(i).Name))
				continue
			}
			tm := parseTagmap(tag)
			// Get default value.
			valdefault, okdefault := tm[DefaultKey]
			if !okdefault {
				warnings.Extra(ErrNoDefault.WrapArgs(v.Type().Field(i).Name))
				continue
			}
			// Try TextUnmarshaler first (time, etc..)
			if tu, ok := v.Field(i).Interface().(encoding.TextUnmarshaler); ok {
				if err := tu.UnmarshalText([]byte(valdefault)); err != nil {
					warnings.Extra(ErrInvalidDefault.WrapCauseArgs(err, valdefault, v.Type().Field(i).Name))
				}
				continue
			}
			// Check if current value is nil as defined in the tag, if defined.
			defaultnil := false
			if valnil, oknil := tm[NilKey]; oknil {
				nilv := reflect.New(v.Field(i).Type())
				if err := reflectex.StringToValue(valnil, reflect.Indirect(nilv)); err != nil {
					panic(err)
				}
				if reflectex.CompareValues(v.Field(i), nilv.Elem()) == 0 {
					defaultnil = true
				}
			} else {
				defaultnil = v.Field(i).IsZero()
			}
			// Set default value to field.
			if all || defaultnil {
				if err := reflectex.StringToValue(valdefault, v.Field(i)); err != nil {
					warnings.Extra(err)
				}
			}
		}
	}
}

// Sanitize enforces defined ranges on fields that exceed them at any depth of
// the specified config, which must be a pointer to a struct, by either
// clamping them to exceeded end value if clamp is true or by resetting them to
// default value if clamp is false and default value is specified or if that
// fails setting them to zero value of that field type.
//
// Sanitize supports enforcing choices and ranges.
//
// Choices (value1,value2,valueN):
//  a,b,c
//
// Ranges (min:max) (up to and including specified min or max):
//  0:      (min:+infinity)
//  :100    (-infinity:max)
//  0:100   (min:max)
//
// Supported kinds are String, Ints, Uints.
//
// If an error occurs it is returned.
//
func Sanitize(config interface{}, clamp bool) error {
	v := reflect.Indirect(reflect.ValueOf(config))
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return ErrInvalidParam
	}
	warnings := ErrParseWarning.Wrap("")
	sanitize(v, clamp, warnings)
	if len(warnings.Extras()) > 0 {
		return warnings
	}
	return nil
}

// sanitize is the implementation of Sanitize.
func sanitize(v reflect.Value, clamp bool, warnings *errorex.ErrorEx) {
	for i := 0; i < v.NumField(); i++ {
		switch fld := v.Field(i); fld.Kind() {
		case reflect.Array, reflect.Slice:
			for i := 0; i < fld.Len(); i++ {
				sanitize(fld.Index(i), clamp, warnings)
			}
		case reflect.Map:
			for iter := fld.MapRange(); iter.Next(); {
				sanitize(iter.Value(), clamp, warnings)
			}
		case reflect.Struct:
			sanitize(fld, clamp, warnings)
		default:
			if !fld.CanSet() {
				continue
			}
			if fld.Type() == interfaceType {
				continue
			}
			// Parse out the default value.
			tag, ok := v.Type().Field(i).Tag.Lookup(ConfigTag)
			if !ok {
				continue
			}
			tm := parseTagmap(tag)
			// Apply range.
			if err := processField(tm, v.Type().Field(i).Name, clamp, fld); err != nil {
				warnings.Extra(err)
				continue
			}
		}
	}
}

// processField applies the limitations on the specified value.
func processField(tags tagmap, name string, clamp bool, field reflect.Value) error {
	rng, ok := tags[RangeKey]
	if !ok {
		return ErrNoRange.WrapArgs(name)
	}
	// Process choices.
	if strings.Contains(rng, " ") {
		vals := strings.Split(rng, " ")
		i := 0
		for _, val := range vals {
			if val != "" {
				vals[i] = val
				i++
			}
		}
		vals = vals[:i]
		// Check if value matches any of choices.
		cv := reflect.New(field.Type())
		matched := false
		for i := 0; i < len(vals); i++ {
			if err := reflectex.StringToValue(vals[i], reflect.Indirect(cv)); err != nil {
				return err
			}
			if reflectex.CompareValues(field, cv.Elem()) == 0 {
				matched = true
			}
		}
		// No match, set default if exists.
		if !matched {
			def, ok := tags[DefaultKey]
			if !ok {
				field.Set(reflect.Zero(field.Type()))
				return nil
			}
			if err := reflectex.StringToValue(def, field); err != nil {
				return err
			}
		}
		return nil
	}
	// Process range.
	if strings.Contains(rng, ":") {
		cv := reflect.New(field.Type())
		vals := strings.Split(rng, ":")
		if len(vals) != 2 {
			return ErrInvalidTag
		}
		// Minimum
		if vals[0] != "" {
			if err := reflectex.StringToValue(vals[0], reflect.Indirect(cv)); err != nil {
				return err
			}
			if reflectex.CompareValues(field, cv.Elem()) < 0 {
				if clamp {
					field.Set(cv.Elem())
				} else {
					def, ok := tags[DefaultKey]
					if !ok {
						field.Set(reflect.Zero(field.Type()))
					}
					if err := reflectex.StringToValue(def, field); err != nil {
						return err
					}
				}
			}
		}
		// Maximum
		if vals[1] != "" {
			if err := reflectex.StringToValue(vals[1], reflect.Indirect(cv)); err != nil {
				return err
			}
			if reflectex.CompareValues(field, cv.Elem()) > 0 {
				if clamp {
					field.Set(cv.Elem())
				} else {
					def, ok := tags[DefaultKey]
					if !ok {
						field.Set(reflect.Zero(field.Type()))
					}
					if err := reflectex.StringToValue(def, field); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
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
