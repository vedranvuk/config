// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"fmt"
	"os"
	"testing"
)

type Root struct {
	PStruct Interface
	PInt    Interface
	Int     Interface
}

type Child struct {
	Age int
}

func TestInterface(t *testing.T) {

	const TestFilename = "interface_test.json"

	val := 9001

	data := &Root{
		Interface{
			Value: &Child{
				42,
			},
		},
		Interface{
			Value: &val,
		},
		Interface{
			Value: 1337,
		},
	}

	defer func() {
		os.Remove(TestFilename)
	}()

	if err := WriteConfigFile(TestFilename, data); err != nil {
		t.Fatal(err)
	}
	newdata := &Root{}

	if err := ReadConfigFile(TestFilename, newdata); err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%#v\n", data)
	fmt.Printf("%#v\n", newdata)

	/*
		if !reflect.DeepEqual(data, newdata) {
			t.Fatal("TestInterface failed")
		}
	*/

}
