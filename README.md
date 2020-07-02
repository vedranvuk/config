# Config

## Description

Config is a lightweight collection of configuration related functions aiding with marshaling, defaulting and sanitation. It aims to be straightforward and portable first and foremost.

Functions that manage configuration priorities across platforms use an user extensible Codec interface to provide multiple marshaling formats. Currently supported formats are **json**, **xml** and **gob**.

A `Dir` type provides a configuration directory across system, user and program configuration locations in an unified way.

Functions for defaulting and sanitizing fields inside configuration structs are provided.

## License

MIT. 

See included LICENSE file.