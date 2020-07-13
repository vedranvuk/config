// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"reflect"

	"github.com/vedranvuk/typeregistry"
)

// Interface is a wrapper for marshalling interface values to abstract data
// formats such as JSON that do not store type information by design.
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

// InitializeInterfaceTypes takes a config struct and recursively preallocates
// values of types registered for that Interface into Value in order to avoid
// allocation of generic map[string]interface{} values for that interfaces when
// unmarshaling.
//
// Specified config must be a pointer to a config struct and is modified by
// this function, possibly even in case of an error.
//
// Any Interface fields found inside config at any depth must have the Type
// field populated by a valid registered type.
//
// Returns a bool telling if one or more Interface types were detected and
// modified or an error if one occurs.
//
// InitializeInterfaceTypes exclusively modifies contained Interface types.
func InitializeInterfaceTypes(config interface{}) (bool, error) {

	if config == nil {
		return false, nil
	}

	v := reflect.Indirect(reflect.ValueOf(config))
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return false, ErrInvalidParam
	}

	return initializeInterfaceTypes(v)
}

// initializeInterfaceTypes is the implementation of InterfaceSetup.
func initializeInterfaceTypes(root reflect.Value) (updated bool, err error) {

	var fld reflect.Value
	for i := 0; i < root.NumField(); i++ {
		fld = root.Field(i)

		if fld.Kind() != reflect.Struct {
			continue
		}

		if fld.Type() != interfaceType {
			if upd, e := initializeInterfaceTypes(fld); e == nil {
				if upd {
					updated = true
				}
			} else {
				return false, e
			}
			continue
		}

		nvt, e := registry.GetType(fld.FieldByName("Type").String())
		if e != nil {
			err = e
			return
		}

		if nvt.Kind() == reflect.Ptr {
			fld.FieldByName("Value").Set(reflect.New(nvt.Elem()))
		} else {
			fld.FieldByName("Value").Set(reflect.New(nvt).Elem())
		}

		updated = true
	}

	return
}

// RegisterInterfaceTypes is used when marshaling a config struct and registers
// type contained in Interface.Value for later use at unmarshaling time.
func RegisterInterfaceTypes(config interface{}) error {

	if config == nil {
		return nil
	}

	v := reflect.Indirect(reflect.ValueOf(config))
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return ErrInvalidParam
	}

	return registerInterfaceTypes(v)
}

// registerInterfaceTypes is the implementation of RegisterInterfaceTypes.
func registerInterfaceTypes(root reflect.Value) error {

	var fld reflect.Value
	for i := 0; i < root.NumField(); i++ {
		fld = root.Field(i)

		if fld.Kind() != reflect.Struct {
			continue
		}

		if fld.Type() != interfaceType {
			if err := registerInterfaceTypes(fld); err != nil {
				return err
			}
			continue
		}

		val := fld.FieldByName("Value")
		fld.FieldByName("Type").SetString(val.Elem().Type().String())

		if err := registry.Register(val.Elem().Interface()); err != nil {
			return err
		}
	}

	return nil
}

var (
	// registry is the Interface type registry.
	registry = typeregistry.New()
	// interfaceType is the reflect.Type of Interface helper.
	interfaceType = reflect.TypeOf(Interface{})
)
