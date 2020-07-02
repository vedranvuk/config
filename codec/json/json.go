// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package json implements a JSON COnfog Codec.
package json

import (
	"encoding/json"

	"github.com/vedranvuk/config/codec"
)

// JSON is the JSON Config Codec.
type JSON struct{}

// Encode implements Codec.Encode.
func (j *JSON) Encode(config interface{}) ([]byte, error) {
	return json.Marshal(config)
}

// Decode implements Codec.Decode.
func (j *JSON) Decode(data []byte, config interface{}) error {
	return json.Unmarshal(data, config)
}

// init registers the Filter on package initialization in the filter registry.
func init() {
	codec.Register("json", &JSON{})
}
