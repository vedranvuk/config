// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"errors"
	"reflect"

	"github.com/vedranvuk/typeregistry"
)

var (
	// ErrInvalidType is returned when an Interface.Type value is invalid.
	ErrInvalidType = ErrConfig.Wrap("invalid type")
)

// Interface is a wrapper for marshalling interface values to and from abstract
// data formats such as JSON that do not store type information by design.
//
// It uses a type registry to allocate values of correct type into an interface
// prior to unmarshaling data into it to avoid unmarshaling to generic
// map[string]interface{} type for JSON, or similar for other packages.
// User still needs to assert the correct Value type when accessing it.
//
// Interfaces use a single config package-wide type registry that generates
// custom names for types contained in Value.
type Interface struct {
	// Type holds the name of the type contained in Value.
	// It should not be modified by user.
	// It is populated when marshaling the Value and read when unmarshaling it.
	Type string
	// Value is the value being wrapped.
	Value interface{}
}

// InitializeInterfaces takes a pointer to a possibly compound struct and
// preallocates Interface.Value fields at any depth with values of type
// registered under Interface.Type so that the data is unmarshaled into a
// proper type instead of a generic interface map.
//
// Config is modified by this function, possibly even in case of an error.
// InitializeInterfaces exclusively modifies contained Interface types.
//
// If config is not a pointer to a struct an ErrInvalidParam is returned.
// If an Interface with an unregistered Interface.Type value is found
// ErrInvalidType is returned that possibly wraps the cause of the error.
//
// On success returns a boolean specifying if one or more Interfaces were
// initialized and a nil error.
func InitializeInterfaces(config interface{}) (bool, error) {
	v := reflect.Indirect(reflect.ValueOf(config))
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return false, ErrInvalidParam
	}
	var modified bool
	return modified, initializeInterface(v, &modified)
}

// initializeInterface initializes an Interface.Value with a type named as
// Interface.Type.
func initializeInterface(fld reflect.Value, modified *bool) error {
	for fld.Kind() == reflect.Ptr {
		fld = reflect.Indirect(fld)
	}
	if fld.Kind() != reflect.Struct {
		return nil
	}
	if fld.Type() != interfaceType {
		return initializeInterfaces(fld, modified)
	}
	typename := fld.FieldByName("Type").String()
	if typename == "" {
		return nil
	}
	valuetype, err := registry.GetType(typename)
	if err != nil {
		return ErrInvalidType.WrapCause("", err)
	}
	if valuetype.Kind() == reflect.Ptr {
		fld.FieldByName("Value").Set(reflect.New(valuetype.Elem()))
	} else {
		fld.FieldByName("Value").Set(reflect.New(valuetype).Elem())
	}
	*modified = true
	return nil
}

// initializeInterfaces is the implementation of InitializeInterfaces.
func initializeInterfaces(root reflect.Value, modified *bool) error {
	var fld reflect.Value
	for i := 0; i < root.NumField(); i++ {
		fld = reflect.Indirect(root.Field(i))
		switch fld.Kind() {
		case reflect.Array, reflect.Slice:
			for i := 0; i < fld.Len(); i++ {
				if err := initializeInterface(fld.Index(i), modified); err != nil {
					return err
				}
			}
		case reflect.Map:
			iter := fld.MapRange()
			for iter.Next() {
				if err := initializeInterface(iter.Value(), modified); err != nil {
					return err
				}
			}
		case reflect.Struct:
			if err := initializeInterface(fld, modified); err != nil {
				return err
			}
		default:
			continue
		}
	}
	return nil
}

// RegisterInterfaces takes a pointer to a possibly compound struct and
// registers types of values in Interface.Value fields at any depth with config
// registry then writes their names into Interface.Type fields.
//
// Config is modified by this function, possibly even in case of an error.
// RegisterInterfaces exclusively modifies contained Interface types.
//
// Interface values with an empty Value field are skipped silently.
// Interface values with a non-empty Type field are skipped silently.
//
// If config is not a pointer to a struct an ErrInvalidParam is returned.
func RegisterInterfaces(config interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(config))
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return ErrInvalidParam
	}
	return registerInterface(v)
}

// registerInterface registers the type of an Interface.Value with the registry
// and sets Interface.Type to a name of the type.
func registerInterface(fld reflect.Value) error {
	fld = reflect.Indirect(fld)
	if fld.Kind() != reflect.Struct {
		return nil
	}
	if fld.Type() != interfaceType {
		return registerInterfaces(fld)
	}
	value := fld.FieldByName("Value")
	if !value.Elem().IsValid() {
		return nil
	}
	if fld.FieldByName("Type").String() != "" {
		return nil
	}
	typename := typeregistry.GetLongTypeName(value.Interface())
	fld.FieldByName("Type").SetString(typename)
	if err := registry.RegisterNamed(typename, value.Interface()); err != nil {
		// Skip duplicate registration errors;
		// Config could be loaded multiple times at runtime.
		if errors.Is(err, typeregistry.ErrDuplicateEntry) {
			return nil
		}
		return err
	}
	return nil
}

// registerInterfaces is the implementation of RegisterInterfaces.
func registerInterfaces(v reflect.Value) error {
	var fld reflect.Value
	for i := 0; i < v.NumField(); i++ {
		fld = reflect.Indirect(v.Field(i))
		switch fld.Kind() {
		case reflect.Array, reflect.Slice:
			for i := 0; i < fld.Len(); i++ {
				if err := registerInterface(fld.Index(i)); err != nil {
					return err
				}
			}
		case reflect.Map:
			iter := fld.MapRange()
			for iter.Next() {
				if err := registerInterface(fld.Index(i)); err != nil {
					return err
				}
			}
		case reflect.Struct:
			if err := registerInterface(fld); err != nil {
				return err
			}
		default:
			continue
		}
	}
	return nil
}

// RegisterType registers a type of specified value with the config registry.
func RegisterType(value interface{}) error {
	if err := registry.Register(value); err != nil {
		return err
	}
	return nil
}

// RegisteredTypeNames returns a slice of registered type names with config.
func RegisteredTypeNames() []string { return registry.RegisteredNames() }

var (
	// registry is the Interface type registry.
	registry = typeregistry.New()
	// interfaceType is the reflect.Type of Interface helper.
	interfaceType = reflect.TypeOf(Interface{})
)
