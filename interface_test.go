// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"reflect"
	"testing"
)

func TestInterface(t *testing.T) {
	type Container struct {
		I Interface
	}
	type Data struct {
		Name string
		Age  int
	}
	data := Data{"foo", 42}
	out := &Container{I: Interface{Value: data}}
	if err := RegisterInterfaces(out); err != nil {
		t.Fatal(err)
	}
	in := &Container{I: Interface{Type: out.I.Type}}
	modified, err := InitializeInterfaces(in)
	if err != nil {
		t.Fatal(err)
	}
	if !modified {
		t.Fatal("Failed initializing an Interface")
	}
	if reflect.TypeOf(out.I.Value) != reflect.TypeOf(in.I.Value) {
		t.Fatal("Interface failed.")
	}
}

func TestInterfaceP(t *testing.T) {
	type Container struct {
		I Interface
	}
	type Data struct {
		Name string
		Age  int
	}
	data := &Data{"foo", 42}
	out := &Container{I: Interface{Value: data}}
	if err := RegisterInterfaces(out); err != nil {
		t.Fatal(err)
	}
	in := &Container{I: Interface{Type: out.I.Type}}
	modified, err := InitializeInterfaces(in)
	if err != nil {
		t.Fatal(err)
	}
	if !modified {
		t.Fatal("Failed initializing an Interface")
	}
	if reflect.TypeOf(out.I.Value) != reflect.TypeOf(in.I.Value) {
		t.Fatal("Interface failed.")
	}
}

func TestPInterface(t *testing.T) {
	type Container struct {
		I *Interface
	}
	type Data struct {
		Name string
		Age  int
	}
	data := Data{"foo", 42}
	out := &Container{I: &Interface{Value: data}}
	if err := RegisterInterfaces(out); err != nil {
		t.Fatal(err)
	}
	in := &Container{I: &Interface{Type: out.I.Type}}
	modified, err := InitializeInterfaces(in)
	if err != nil {
		t.Fatal(err)
	}
	if !modified {
		t.Fatal("Failed initializing an Interface")
	}
	if reflect.ValueOf(out.I.Value).Type().String() != reflect.ValueOf(in.I.Value).Type().String() {
		t.Fatal("Interface failed.")
	}
}

func TestPInterfaceP(t *testing.T) {
	type Container struct {
		I *Interface
	}
	type Data struct {
		Name string
		Age  int
	}
	data := &Data{"foo", 42}
	out := &Container{I: &Interface{Value: data}}
	if err := RegisterInterfaces(out); err != nil {
		t.Fatal(err)
	}
	in := &Container{I: &Interface{Type: out.I.Type}}
	modified, err := InitializeInterfaces(in)
	if err != nil {
		t.Fatal(err)
	}
	if !modified {
		t.Fatal("Failed initializing an Interface")
	}
	if reflect.ValueOf(out.I.Value).Type().String() != reflect.ValueOf(in.I.Value).Type().String() {
		t.Fatal("Interface failed.")
	}
}
