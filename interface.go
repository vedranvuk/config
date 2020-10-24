// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"errors"
	"reflect"

	"github.com/vedranvuk/typeregistry"
)

// Interface is a wrapper for marshalling interface values to and from abstract
// data formats such as JSON that do not store type information by design.
//
// It uses a type registry to allocate interface values of correct type prior
// to unmarshaling data to interface to avoid unmarshaling to generic
// map[string]interface{} type.
// User still needs to assert the correct Value type when accessing it.
//
// All instances of Interface use a single type registry that relies on
// reflect.Type.String() to produce names of types contained in Value.
// See https://pkg.go.dev/reflect?tab=doc#Type for a gotcha if unfamiliar.
//
// Any caveats that apply to used marshaling format apply to Interface as well.
// For example, if using JSON its' rules still apply; if Value holds a struct
// value during registration instead of a pointer to a struct Codec will
// unmarshal map[string]interface{} into Value even with preallocated value
// of correct type prior to unmarshaling.
//
type Interface struct {

	// Type holds the name of the type contained in Value.
	// It should not be modified by user as it serves to store the name of type
	// contained in Value and gets overwritten on Interface marshalling.
	Type string

	// Value is the value being wrapped. It's type is registered with type
	// registry and type name stored to Type when marshaling and preallocated
	// from type registry by addressing it using Type when unmarshaling.
	Value interface{}
}

// RegisterType pre-registers a type of specified value with the registry.
func RegisterType(value interface{}) error {
	if err := registry.Register(value); err != nil {
		return err
	}
	return nil
}

// RegisterTypeByName pre-registers a type of specified value with the registry
// by name.
func RegisterTypeByName(name string, value interface{}) error {
	if err := registry.RegisterNamed(name, value); err != nil {
		return err
	}
	return nil
}

// InitializeInterfaces takes a pointer to a config struct and recursively
// preallocates values of types registered for that Interface into Value in
// order to avoid allocation of generic map[string]interface{} values for those
// interfaces when unmarshaling.
//
// Specified config must be a pointer to a config struct and is modified by
// this function, possibly even in case of an error.
//
// Any Interface fields found inside config at any depth must have the Type
// field populated by a valid registered type or an error will occur.
//
// Returns a bool telling if one or more Interface types were detected and
// modified or an error if one occurs.
//
// InitializeInterfaces exclusively modifies contained Interface types.
func InitializeInterfaces(config interface{}) (bool, error) {

	if config == nil {
		return false, nil
	}

	v := reflect.Indirect(reflect.ValueOf(config))
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return false, ErrInvalidParam
	}

	updated := false
	return updated, initializeInterfaces(v, &updated)
}

// initializeInterfaces is the implementation of InterfaceSetup.
func initializeInterfaces(root reflect.Value, updated *bool) error {

	var fld reflect.Value
	for i := 0; i < root.NumField(); i++ {
		fld = reflect.Indirect(root.Field(i))

		switch fld.Kind() {
		case reflect.Array, reflect.Slice:
			for i := 0; i < fld.Len(); i++ {
				if err := initializeInterfaceField(fld.Index(i), updated); err != nil {
					return err
				}
			}
		case reflect.Map:
			iter := fld.MapRange()
			for iter.Next() {
				if err := initializeInterfaceField(iter.Value(), updated); err != nil {
					return err
				}
			}
		case reflect.Struct:
			if err := initializeInterfaceField(fld, updated); err != nil {
				return err
			}

		default:
			continue
		}

	}

	return nil
}

// initializeInterfaceField initializes an Interface type in a config struct
// field.
func initializeInterfaceField(fld reflect.Value, updated *bool) error {

	for fld.Kind() == reflect.Ptr {
		fld = reflect.Indirect(fld)
	}

	if fld.Kind() != reflect.Struct {
		return nil
	}

	if fld.Type() != interfaceType {
		if err := initializeInterfaces(fld, updated); err != nil {
			return err
		}
		return nil
	}

	tn := fld.FieldByName("Type").String()
	if tn == "" {
		return nil
	}
	nvt, err := registry.GetType(tn)
	if err != nil {
		return err
	}

	if nvt.Kind() == reflect.Ptr {
		fld.FieldByName("Value").Set(reflect.New(nvt.Elem()))
	} else {
		fld.FieldByName("Value").Set(reflect.New(nvt).Elem())
	}

	*updated = true

	return nil
}

// RegisterInterfaces takes a pointer to a config struct and recursively
// registers types of Values with type registry found in any Interface at any
// level and populate Type field of Interfaces with detected type.
//
// This prepares a config struct for marshaling so registered types can be
// restored to Value interfaces on unmarshaling.
//
// Specified config must be a pointer to a config struct and is modified by
// this function, possibly even in case of an error.
//
// Returns an error if one occurs.
func RegisterInterfaces(config interface{}) error {

	if config == nil {
		return nil
	}

	v := reflect.Indirect(reflect.ValueOf(config))
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return ErrInvalidParam
	}

	return registerInterfaces(v)
}

// registerInterfaces is the implementation of RegisterInterfaceTypes.
func registerInterfaces(root reflect.Value) error {

	var fld reflect.Value
	for i := 0; i < root.NumField(); i++ {
		fld = reflect.Indirect(root.Field(i))

		switch fld.Kind() {
		case reflect.Array, reflect.Slice:
			for i := 0; i < fld.Len(); i++ {
				if err := registerInterfaceValue(fld.Index(i)); err != nil {
					return err
				}
			}
		case reflect.Map:
			iter := fld.MapRange()
			for iter.Next() {
				if err := registerInterfaceValue(fld.Index(i)); err != nil {
					return err
				}
			}
		case reflect.Struct:
			if err := registerInterfaceValue(fld); err != nil {
				return err
			}
		default:
			continue
		}
	}

	return nil
}

// registerInterfaceValue registers a type in a config struct field.
func registerInterfaceValue(fld reflect.Value) error {

	if fld.Type() != interfaceType {
		return registerInterfaces(fld)
	}

	val := fld.FieldByName("Value")
	if !val.Elem().IsValid() {
		return nil
	}

	if fld.FieldByName("Type").String() != "" {
		return nil
	}

	typename := typeregistry.GetLongTypeName(val.Interface())
	fld.FieldByName("Type").SetString(typename)

	if err := registry.RegisterNamed(typename, val.Interface()); err != nil {
		// Skip duplicate registration errors;
		// Config could be loaded multiple times at runtime.
		if errors.Is(err, typeregistry.ErrDuplicateEntry) {
			return nil
		}
		return err
	}

	return nil
}

var (
	// registry is the Interface type registry.
	registry = typeregistry.New()
	// interfaceType is the reflect.Type of Interface helper.
	interfaceType = reflect.TypeOf(Interface{})
)
