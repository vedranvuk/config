// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func ExampleDefault() {
	type Example struct {
		Name   string   `config:"default=foo"`
		PName  *string  `config:"default=foo"`
		PPName **string `config:"default=foo"`
		Age    int      `config:"default=42"`
		PAge   *int     `config:"default=42"`
		Ping   int      `config:"nil=10;default=30"`
		Pong   int      `config:"nil=10;default=30"`
	}
	p := &Example{Ping: 10, Pong: 20}
	if err := Default(p, false); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Name:%s PName:%s, PPName:%s, Age:%d PAge:%d Ping:%d Pong:%d\n", p.Name, *p.PName, **p.PPName, p.Age, *p.PAge, p.Ping, p.Pong)
	// Output: Name:foo PName:foo, PPName:foo, Age:42 PAge:42 Ping:30 Pong:20
}

func ExampleLimit() {
	type Example struct {
		Name string `config:"range=foo,bar,baz;default=foobar"`
		Age  int    `config:"range=7:77;default=42"`
	}
	p := &Example{}
	if err := Limit(p, true); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Name:%s Age:%d\n", p.Name, p.Age)
	// Output: Name:foobar Age:7
}

func TestDefault(t *testing.T) {
	type Example struct {
		Name   string   `config:"default=foo"`
		PName  *string  `config:"default=foo"`
		PPName **string `config:"default=foo"`
	}
	p := Example{}
	if err := Default(&p, false); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%s %s %s", p.Name, *p.PName, **p.PPName) != "foo foo foo" {
		t.Fatal("Default failed.")
	}
}

// Child documentation
type sanitizeChild struct {
	Int    int      `config:"default=42"`
	PInt   *int     `config:"default=42"`
	PPInt  **int    `config:"default=42"`
	Bool   bool     `config:"default=true"`
	PBool  *bool    `config:"default=true"`
	PPBool **bool   `config:"default=true"`
	Str    string   `config:"default=foo"`
	PStr   *string  `config:"default=foo"`
	PPStr  **string `config:"default=foo"`
}

// String help.
func (sc *sanitizeChild) String() string {
	return fmt.Sprintf("Int: %d, PInt: %d, PPInt: %d, Bool: %t, PBool: %t, PPBool: %t, Str: %s, PStr: %s, PPStr: %s\n",
		sc.Int, *sc.PInt, **sc.PPInt, sc.Bool, *sc.PBool, **sc.PPBool, sc.Str, *sc.PStr, **sc.PPStr)
}

// Parent documentation
type sanitizeParent struct {
	Array  [1]sanitizeChild
	ArrayP [1]*sanitizeChild
	Slice  []sanitizeChild
	SliceP []*sanitizeChild
	Map    map[string]sanitizeChild
	MapP   map[string]*sanitizeChild
}

func TestRecursive(t *testing.T) {

	p := sanitizeParent{
		Array:  [1]sanitizeChild{sanitizeChild{}},
		ArrayP: [1]*sanitizeChild{&sanitizeChild{}},
		Slice:  []sanitizeChild{sanitizeChild{}},
		SliceP: []*sanitizeChild{&sanitizeChild{}},
		Map:    map[string]sanitizeChild{"1": sanitizeChild{}},
		MapP:   map[string]*sanitizeChild{"1": &sanitizeChild{}},
	}
	if err := Default(&p, false); err != nil && !errors.Is(err, ErrWarning) {
		t.Fatal(err)
	}
	show(p)
}

type interfaceContainer struct {
	Iface  Interface
	IfaceP *Interface
}

func TestSanitizeInterfaceValue(t *testing.T) {
	p1 := interfaceContainer{
		Iface:  Interface{Value: &sanitizeChild{}},
		IfaceP: &Interface{Value: &sanitizeChild{}},
	}
	if err := Default(&p1, false); err != nil && !errors.Is(err, ErrWarning) {
		t.Fatal(err)
	}
	p2 := interfaceContainer{
		Iface:  Interface{Value: &sanitizeChild{}},
		IfaceP: &Interface{Value: &sanitizeChild{}},
	}
	Default(p2, false)
	if !reflect.DeepEqual(p1, p2) {
		t.Fatal("TestSanitizeInterfaceValue failed.")
	}
}

func show(i interface{}) {
	if !testing.Verbose() {
		return
	}
	b, err := json.MarshalIndent(i, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
