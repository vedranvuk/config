// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package codec defines a config codec interface.
package codec

import (
	"sync"

	"github.com/vedranvuk/errorex"
)

var (
	// ErrCodec is the base error of codec package.
	ErrCodec = errorex.New("codec")
	// ErrCodecNotRegistered is returned by Get if the requested Codec is not
	// registered.
	ErrCodecNotRegistered = ErrCodec.WrapFormat("codec '%s' not registered")
)

// Codec defines a configuration marshaling Codec interface.
type Codec interface {
	// Encode must encode interface to a byte slice or return an error.
	Encode(interface{}) ([]byte, error)
	// Decode must decode the byte slice to the interface or return an error.
	Decode([]byte, interface{}) error
}

// Register registers a Config codec under the specified name.
// It panics if the name is already registered.
func Register(name string, codec Codec) {
	regmu.Lock()
	if _, exists := registry[name]; exists {
		regmu.Unlock()
		panic("config codec registry: codec " + name + " already registered")
	}
	registry[name] = codec
	regmu.Unlock()
}

// Get returns the codec registered under specified name and a nil error,
// if found. Otherwise a nil filter and an error.
func Get(name string) (Codec, error) {
	regmu.Lock()
	filter, exists := registry[name]
	if !exists {
		regmu.Unlock()
		return nil, ErrCodecNotRegistered.WrapArgs(name)
	}
	regmu.Unlock()
	return filter, nil
}

var (
	// regmu is the codec registry mutex.
	regmu = sync.Mutex{}
	// registry is the codec registry.
	registry = map[string]Codec{}
)
