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
// It uses a type registry to allocate interface values of correct type when
// unmarshaling data to interface instead of generic map[string]interface{}.
// User still needs to assert the correct Value type when accessing it.
//
// All instances of Interface use a single type registry that relies on
// reflect.Type.String() to produce names of types contained in Value.
// See https://pkg.go.dev/reflect?tab=doc#Type for a gotcha if unfamiliar.
//
type Interface struct {

	// Type holds the name of the type contained in Value.
	// It should not be modified by user as it serves to store the name of type
	// contained in Value and gets overwritten on Interface marshalling.
	Type string

	// Value is the marshaled interface that get allocated on unmarshaling to
	// correct type stored in Type.
	Value interface{}
}

// InterfaceSetup takes a config struct and recursively preallocates the correct type
// contained in Interface.Value at any depth or returns an error.
//
// Specified config must be a pointer to a config struct and is modified by
// this function.
//
// Any Interface fields found inside config at any depth must have the Type
// field populated by a valid registered type.
//
// Returns a bool telling if any Interface types were detected and modified
// and if the config needs to be reloaded.
func InterfaceSetup(config interface{}) (bool, error) {

	if config == nil {
		return false, nil
	}

	v := reflect.Indirect(reflect.ValueOf(config))
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return false, ErrInvalidParam
	}

	return interfaceSetup(v)
}

// interfaceSetup is the implementation of InterfaceSetup.
func interfaceSetup(root reflect.Value) (updated bool, err error) {

	for i := 0; i < root.NumField(); i++ {

		if root.Field(i).Kind() == reflect.Struct {

			if root.Field(i).Type() == interfaceType {

				nvt, e := registry.GetType(root.Field(i).FieldByName("Type").String())
				if e != nil {
					err = e
					return
				}

				if nvt.Kind() == reflect.Ptr {
					root.Field(i).FieldByName("Value").Set(reflect.New(nvt.Elem()))
				} else {
					root.Field(i).FieldByName("Value").Set(reflect.New(nvt).Elem())
				}

				updated = true
			} else {
				if upd, e := interfaceSetup(root.Field(i)); e == nil {
					updated = upd
				} else {
					return false, e
				}
			}

		}

	}
	return
}

// interfaceRegister is used when marshaling a config struct and registers
// type contained in Interface.Value for later use at unmarshaling time.
func interfaceRegister(config interface{}) error {

	if config == nil {
		return nil
	}

	v := reflect.Indirect(reflect.ValueOf(config))
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return ErrInvalidParam
	}

	return traverseInterfaces(v)
}

func traverseInterfaces(root reflect.Value) error {

	for i := 0; i < root.NumField(); i++ {

		if root.Field(i).Kind() == reflect.Struct {

			if root.Field(i).Type() == interfaceType {

				ifacev := root.Field(i).FieldByName("Value")

				root.Field(i).FieldByName("Type").SetString(ifacev.Elem().Type().String())

				if err := registry.Register(ifacev.Elem().Interface()); err != nil {
					return err
				}

			} else {
				if err := traverseInterfaces(root.Field(i)); err != nil {
					return err
				}
			}

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
