// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package gob implements a GOB Config Codec.
package gob

import (
	"bytes"
	"encoding/gob"
	"errors"

	"github.com/vedranvuk/config/codec"
)

// GOB is the GOB Config Codec.
type GOB struct {
	buf *bytes.Buffer
	enc *gob.Encoder
	dec *gob.Decoder
}

// Encode implements Codec.Encode.
func (g *GOB) Encode(config interface{}) ([]byte, error) {

	if config == nil {
		return nil, errors.New("cannot encode nil value")
	}

	defer g.buf.Reset()

	if err := g.enc.Encode(config); err != nil {
		return nil, err
	}

	return g.buf.Bytes(), nil
}

// Decode implements Codec.Decode.
func (g *GOB) Decode(data []byte, config interface{}) error {

	defer g.buf.Reset()

	if _, err := g.buf.Write(data); err != nil {
		return err
	}

	if err := g.dec.Decode(config); err != nil {
		return err
	}

	return nil
}

// init registers the Filter on package initialization in the filter registry.
func init() {
	c := &GOB{
		buf: bytes.NewBuffer(nil),
	}
	c.enc = gob.NewEncoder(c.buf)
	c.dec = gob.NewDecoder(c.buf)

	codec.Register("gob", c)
}
