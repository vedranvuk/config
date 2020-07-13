// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"os"
	"reflect"
	"testing"
)

type Root struct {
	Name  string
	Test2 Interface
	Test3 Interface
	Test4 Interface
}

type Child struct {
	Age int
}

func TestInterface(t *testing.T) {

	const TestFilename = "interface_test.json"

	val := 9001

	data := &Root{
		"Foo",
		Interface{
			Value: &Child{
				42,
			},
		},
		Interface{
			Value: 1337,
		},
		Interface{
			Value: &val,
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

	if !reflect.DeepEqual(data, newdata) {
		t.Fatal("TestInterface failed")
	}

}
