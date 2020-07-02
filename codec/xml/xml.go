// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package xml is the XML Config Codec.
package xml

import (
	"encoding/xml"

	"github.com/vedranvuk/config/codec"
)

// XML is the XML Config Codec.
type XML struct{}

// Encode implements Codec.Encode.
func (x *XML) Encode(config interface{}) ([]byte, error) {
	return xml.Marshal(config)
}

// Decode implements Codec.Decode.
func (x *XML) Decode(data []byte, config interface{}) error {
	return xml.Unmarshal(data, config)
}

// init registers the Filter on package initialization in the filter registry.
func init() {
	codec.Register("xml", &XML{})
}
