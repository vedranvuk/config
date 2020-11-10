// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"fmt"
)

func ExampleDefault() {
	type Example struct {
		Name  string  `config:"default=foo"`
		PName *string `config:"default=foo"`
		Age   int     `config:"default=42"`
		PAge  int     `config:"default=42"`
		Ping  int     `config:"nil=10;default=30"`
		Pong  int     `config:"nil=10;default=30"`
	}
	p := &Example{Ping: 10, Pong: 20}
	if err := Default(p, false); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Name:%s PName:%s, Age:%d PAge:%d Ping:%d Pong:%d\n", p.Name, *p.PName, p.Age, p.PAge, p.Ping, p.Pong)
	// Output: Name:foo PName:foo, Age:42 PAge:42 Ping:30 Pong:20
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
