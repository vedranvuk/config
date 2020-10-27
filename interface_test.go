// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"os"
	"testing"
)

type Root struct {
	Nil      Interface
	PPStruct Interface
	PStruct  Interface
	PInt     Interface
	Int      Interface
	IntP     *Interface
	IntPN    *Interface
}

type Child struct {
	Age int
}

func TestReadWrite(t *testing.T) {

	const TestFilename = "interface_test.json"

	val := int(9001)

	p := &Child{69}
	pp := &p

	data := &Root{
		Interface{
			Value: nil,
		},

		Interface{
			Value: pp,
		},
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
		&Interface{
			Value: 1337,
		},
		nil,
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

	a := data.PStruct.Value.(*Child)
	b := newdata.PStruct.Value.(*Child)
	if a.Age != b.Age {
		t.Fatal("TestInterface failed.")
	}
	if *data.PInt.Value.(*int) != *newdata.PInt.Value.(*int) {
		t.Fatal("TestInterface failed.")
	}
	if data.Int.Value.(int) != int(newdata.Int.Value.(float64)) {
		t.Fatal("TestInterface failed.")
	}
}
