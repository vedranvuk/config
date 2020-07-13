# Config

## Description

Config is a lightweight collection of configuration related functions aiding with marshaling, defaulting, sanitation and type system. It aims to be straightforward and portable first and foremost. It works with `struct` types as compositable configuration section containers and uses custom field tags to control properties of configuration values. 

Configuration marshaling functions use an user extensible Codec interface to provide multiple marshaling formats selected by extension at marshaling time. Formats supported out of the box are **json**, **xml** and **gob** and are imported by user as required.

A `Dir` helper type provides a configuration directory across system, user and program configuration locations in a multiplatform and unified manner.

An `Interface` helper type serves as a wrapper for `interface{}` values that unmarshals interfaces to types contained in them at marshaling time instead of generic `map[string]interface{}` type.

## License

MIT. 

See included LICENSE file.