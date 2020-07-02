// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/vedranvuk/config/codec"

	_ "github.com/vedranvuk/config/codec/gob"
	_ "github.com/vedranvuk/config/codec/json"
	_ "github.com/vedranvuk/config/codec/xml"
)

func readwriteconfig(codec string) error {

	type Config struct {
		Name string
		Age  int
	}

	fn := "testconfig." + codec

	out := &Config{"Foo", 42}

	if err := WriteConfigFile(fn, out); err != nil {
		return err
	}
	defer os.Remove(fn)

	in := &Config{}

	if err := ReadConfigFile(fn, in); err != nil {
		return err
	}

	if !reflect.DeepEqual(in, out) {
		return errors.New("TestReadWriteConfig failed: in and out not equal")
	}

	return nil
}

func TestReadWriteConfigFile(t *testing.T) {

	if err := readwriteconfig("json"); err != nil {
		t.Fatal(err)
	}
	if err := readwriteconfig("xml"); err != nil {
		t.Fatal(err)
	}
	if err := readwriteconfig("gob"); err != nil {
		t.Fatal(err)
	}
	if err := readwriteconfig("INVALIDCODEC"); err != nil {
		if !errors.Is(err, codec.ErrCodecNotRegistered) {
			t.Fatal(err)
		}
	}
}
