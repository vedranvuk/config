// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDirShallow(t *testing.T) {

	type Config struct {
		Name string
		Age  int
	}

	configdir := "configtest"
	configname := "config.json"

	dir, err := NewDir(configdir)
	if err != nil {
		t.Fatal(err)
	}
	out := &Config{"Foo", 42}
	in := &Config{}

	defer func() {
		path, err := GetUserConfigPath()
		if err != nil {
			return
		}
		os.RemoveAll(filepath.Join(path, configdir))
	}()

	if err := dir.SaveUserConfig(configname, out); err != nil {
		t.Fatal(err)
	}

	if err := dir.LoadConfig(configname, true, in); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(in, out) {
		t.Fatal("fail")
	}
}

func TestDirDeep(t *testing.T) {

	type Config struct {
		Name string
		Age  int
	}

	configdir := "configtest/child1/child2/child3"
	configname := "deep1/deep2/deep3/config.xml"

	dir, err := NewDir(configdir)
	if err != nil {
		t.Fatal(err)
	}
	out := &Config{"Foo", 42}
	in := &Config{}

	defer func() {
		path, err := GetUserConfigPath()
		if err != nil {
			return
		}
		os.RemoveAll(filepath.Join(path, "configtest"))
	}()

	if err := dir.SaveUserConfig(configname, out); err != nil {
		t.Fatal(err)
	}

	if err := dir.LoadConfig(configname, true, in); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(in, out) {
		t.Fatal("fail")
	}
}
