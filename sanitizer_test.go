// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"errors"
	"reflect"
	"testing"
)

func TestSetDefaults(t *testing.T) {

	type Test struct {
		Name   string `config:"default=foo"`
		Age    int    `config:"nil=-1;default=42"`
		Height int    `config:"default=190"`
		Weight int    `config:"default=80"`
		Income int    `config:"default=50000"`
	}

	test := &Test{"", -1, 0, 90, 0}

	if err := SetDefaults(test, false); err != nil && !errors.Is(err, ErrParseWarning) {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(test, &Test{"foo", 42, 190, 90, 50000}) {
		t.Fatal("TestSetDefaults failed.")
	}
}

func TestSetDefaultsAll(t *testing.T) {

	type Test struct {
		Name   string `config:"default=foo"`
		Age    int    `config:"nil=-1;default=42"`
		Height int    `config:"default=190"`
		Weight int    `config:"default=80"`
		Income int    `config:"default=50000"`
	}

	test := &Test{"INVALID", 64, 105, 25, 100}

	if err := SetDefaults(test, true); err != nil && !errors.Is(err, ErrParseWarning) {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(test, &Test{"foo", 42, 190, 80, 50000}) {
		t.Fatal("TestSetDefaultsAll failed.")
	}
}

func TestSanitize(t *testing.T) {

	type Test struct {
		Name     string `config:"range=foo bar baz;default=bar"`
		Surname  string `config:"range=doe dope donut"`
		CarBrand string `config:"range=tesla bmw lada"`
		Age      int    `config:"range=0:150"`
		Height   int    `config:"range=:200"`
		Weight   int    `config:"range=10:"`
	}

	test := &Test{"INVALID", "dope", "INVALID", 160, 210, 5}

	if err := Sanitize(test, true); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(test, &Test{"bar", "dope", "", 150, 200, 10}) {
		t.Fatal("TestSanitize failed.")
	}
}

func BenchmarkSetDefaults(b *testing.B) {

	b.StopTimer()

	type Test struct {
		Name   string `config:"default=foo"`
		Age    int    `config:"nil=-1;default=42"`
		Height int    `config:"default=190"`
		Weight int    `config:"default=80"`
		Income int    `config:"default=50000"`
	}

	test := &Test{"", -1, 0, 90, 0}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		SetDefaults(test, false)
	}
}
func BenchmarkSetDefaultsAll(b *testing.B) {

	b.StopTimer()

	type Test struct {
		Name   string `config:"default=foo"`
		Age    int    `config:"nil=-1;default=42"`
		Height int    `config:"default=190"`
		Weight int    `config:"default=80"`
		Income int    `config:"default=50000"`
	}

	test := &Test{"", -1, 0, 90, 0}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		SetDefaults(test, true)
	}
}

func BenchmarkSanitize(b *testing.B) {

	b.StopTimer()

	type Test struct {
		Name     string `config:"range=foo bar baz;default=bar"`
		Surname  string `config:"range=doe dope donut"`
		CarBrand string `config:"range=tesla bmw lada"`
		Age      int    `config:"range=0:150"`
		Height   int    `config:"range=:200"`
		Weight   int    `config:"range=10:"`
	}

	test := &Test{"INVALID", "blah", "INVALID", 160, 210, 5}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		Sanitize(test, true)
	}
}

func BenchmarkSanitizeNoClamp(b *testing.B) {

	b.StopTimer()

	type Test struct {
		Name     string `config:"range=foo bar baz;default=bar"`
		Surname  string `config:"range=doe dope donut"`
		CarBrand string `config:"range=tesla bmw lada"`
		Age      int    `config:"range=0:150"`
		Height   int    `config:"range=:200"`
		Weight   int    `config:"range=10:"`
	}

	test := &Test{"INVALID", "blah", "INVALID", 160, 210, 5}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		Sanitize(test, false)
	}
}
