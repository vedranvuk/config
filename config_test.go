// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/vedranvuk/config/codec"

	_ "github.com/vedranvuk/config/codec/gob"
	_ "github.com/vedranvuk/config/codec/json"
	_ "github.com/vedranvuk/config/codec/xml"
)

func TestPaths(t *testing.T) {
	path, err := GetUserConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if testing.Verbose() {
		fmt.Println(path)
	}

	path, err = GetSystemConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if testing.Verbose() {
		fmt.Println(path)
	}
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

func readwriteconfig(codec string) error {
	type TestConfig struct {
		Name  string
		Age   int
		Truth *bool
	}
	filename := "testconfig." + codec
	t := true
	out := &TestConfig{"Foo", 42, &t}
	if err := WriteConfigFile(filename, out); err != nil {
		return err
	}
	defer os.Remove(filename)
	in := &TestConfig{}
	if err := ReadConfigFile(filename, in); err != nil {
		return err
	}
	if !reflect.DeepEqual(in, out) {
		return errors.New("TestReadWriteConfig failed: in and out not equal")
	}
	return nil
}
