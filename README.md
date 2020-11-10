# Config

## Status

It's being added to periodically, as needed, for a personal project. All milestones are tagged and bagged so take what you need. API may change, but not by much at this point.

## Description

Config is a collection of helpers for dealing with configurations. It helps with marshaling, defaulting, sanitation and configuration locations in a file system.

Config works with `struct` types as configuration containers. They allow for grouping, compositing, tagging and fast access from code so they are the logical choice to focus on. 

Config package consists of following helpers:

* [Codecs](##Codecs)
* [Dir](##Dir)
* [Interface](##Interface)
* [Sanitizer](##Sanitizer)
* [Utilities](##Utilities)

## Codecs

Config defines an extensible `Codec` interface:

```go
// Codec defines a configuration marshaling Codec interface.
type Codec interface {
	// Encode must encode interface to a byte slice or return an error.
	Encode(interface{}) ([]byte, error)
	// Decode must decode the byte slice to the interface or return an error.
	Decode([]byte, interface{}) error
}
```

and implements three codecs: **gob**, **json** and **xml**.

Codecs when included by user as needed register themselves with the config package and are used by the package opaquely.

```go
include (
	_ "github.com/vedranvuk/config/codec/gob"
	_ "github.com/vedranvuk/config/codec/json"
	_ "github.com/vedranvuk/config/codec/xml"

	_ "github.com/someone/config/codec/yaml"
```

[Utilities](#Utilities) from the package use the codecs to read or write configurations simply by specifying extension.

## Dir

Dir maintains a configuration subdirectory in multiple locations on a filesystem and represents them as **program**, **user** and **system**  and  configurations with their priorities being in order of mention. It provides methods for automatically loading a configuration file by priority and selectively. It uses codecs to select marshaling format.

The locations of **user** and **system** directories depends on the platform. See [Utilities](#Utilities) for how those locations are determined.

The **program** location is only available for writing on Windows OS.

### Example

```go
func DirExample() {

	type Params struct {
		TLS bool
		Version string
	}

	dir := NewDir("MyApp")

	// Saves tlsparams.json to a standard user configuration location under
	// "MyApp/params/tlsparams.json".
	if err := dir.SaveUserConfig("params/tlsparams.json", &Params{true, "1.0"}); err != nil {
		log.Fatal(err)
	}

	// Loads MyApp/params/tlsparams.json from any of configuration locations,
	// if found, in order of priority, overriding values as found.
	params := &Params{}
	if err := dir.LoadConfig("params/tlsparams.json", true, params); err != nil {
		log.Fatal(err)
	}

	_ = params
}
```
Dir api consists of the following:

```
LoadSystemConfig(name string, out interface{}) error
LoadUserConfig(name string, out interface{}) error
LoadProgramConfig(name string, out interface{}) error
LoadConfig(name string, override bool, out interface{}) (err error)
SaveSystemConfig(name string, in interface{}) error
SaveUserConfig(name string, in interface{}) error
SaveProgramConfig(name string, in interface{}) error
```

## Interface

Interface is a wrapper for marshaling interface values to and from abstract data formats such as JSON that do not store type information. The most notable example would be unmarshaling a JSON object to an empty interface; the JSON package unmarshals it as `map[string]interface{}` which is usually incompatible with types it was unmarshaled from.

The solution is to pre-allocate correct types in the interface it is being unmarshaled into. So instead of an `interface{}` you would use `config.Interface`.

```go
type Interface struct {

	// Type holds the name of the type contained in Value.
	// It should not be modified by user as it serves to store the name of type
	// contained in Value and gets overwritten on Interface marshalling.
	Type string

	// Value is the value being wrapped. It's type is registered with type
	// registry and type name stored to Type when marshaling and preallocated
	// from type registry by addressing it using Type when unmarshaling.
	Value interface{}
}
```

Interface uses a type registry to store type information and provides functions to manually register types either by their type name of under a custom name.

It is most useful when saving configurations that may differ by type in a single configuration field.

### Example

```go
package main

import (
	"fmt"

	"github.com/vedranvuk/config"
	_ "github.com/vedranvuk/config/codec/xml"
)

type EngineXConfig struct {
	Name string
}

type EngineYConfig struct {
	Length int
}

type Config struct {
	Engine string
	Config *config.Interface
}

func main() {
	myConfig := &Config{
		Engine: "EngineX",
		Config: &config.Interface{
			Value: &EngineXConfig{"TheExcellent"},
		},
	}
	config.WriteConfigFile("engine.xml", myConfig)

	myNewConfig := &Config{}
	config.ReadConfigFile("engine.xml", myNewConfig)

	enginecfg, ok := myNewConfig.Config.Value.(*EngineXConfig)
	if !ok {
		panic("oh no!")
	}

	// Output: &main.EngineXConfig{Name:"TheExcellent"}
	fmt.Printf("%#v\n", enginecfg)
}
```

## Sanitizer

Sanitizer functions use struct field tags to set defaults, enforce value
ranges or values from a set on non-compound type fields and fields implementing 
TextUnmarshaler.

Keys are parsed from a `"config"` struct tag.

There are few utility functions provided but the main two are Default and Limit.

### Default

```Go
// Default takes a pointer to a config struct and recursively traverses
// possibly nested fields with config tags then sets their field values
// to values defined under default key if their values are zero of their type
// or match the value defined under nil key.
// e.g. nil for *int, 0 for int, or "" for strings.
//
// If reset is specified all fields with defined defaults are reset, regardless
// if they have already been initialized to non-nil values or their value has
// since been changed.
//
// If any errors or warnings occured it returns an ErrParseWarning of type
// *errorex.ErrorEx that contains all warnings is its' Extras field.
// It is returned under following conditions:
//
// If a field has no tag defined an ErrNoTag is appended to Extras.
// If a field has no default value an ErrNoDefault is appended to Extras.
// If a field has an incompatible/invalid default value defined an
// ErrInvalidDefault is appended to Extras.
//
// Any other errors signify a no-op and a failure.
Default(config interface{}, reset bool) error
```

#### Example

```Go
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
	fmt.Printf("%#v\n", err.Error())
}
fmt.Printf("Name:%s PName:%s, Age:%d PAge:%d Ping:%d Pong:%d\n", p.Name, *p.PName, p.Age, p.PAge, p.Ping, p.Pong)
// Output: Name:foo PName:foo, Age:42 PAge:42 Ping:30 Pong:20
```

### Limit

```Go
// Limit takes a pointer to a config struct and recursively traverses
// possibly nested fields with config tags then if clamp was specified and
// their values are outside of defined range or set, sets their field values
// to values within the set or range, otherwise it just generates a warning.
// e.g. range=1;2;3 range=foo;bar range=0:100
//
// If clamp is specified fields with values outside of defined ranges are set
// to lowest or highest value defined, depending on boundary they exceed.
// If the value is not within the set a default operation is applied to the
// field.
//
// Limit supports enforcing sets and ranges and precognizes them as follows:
// Choices are strings separated by a ",", e.g.: 1,2,3 foo,bar,baz
// Ranges are two values separated by a ":", e.g.: 0: :100 0:100 or :
// Both range boundaries are optional, although that defeats the purpose.
// Supported kinds are String, Ints, Uints.
//
// If any errors or warnings occured it returns an ErrParseWarning of type
// *errorex.ErrorEx that contains all warnings is its' Extras field.
// It is returned under following conditions:
//
// If a field has no tag defined an ErrNoTag is appended to Extras.
// If a field has no range value an ErrNoRange is appended to Extras.
// If a field has an incompatible/invalid range value defined an
// ErrInvalidRange is appended to Extras.
//
// Any other errors signify a no-op and a failure.
Limit(config interface{}, clamp bool) error
```

#### Example 

```Go
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

```

## Utilities

Utility functions make use of shared `config` functionality.

The API consists of the following:

```go
// WriteConfigFile writes specified config to a file specified by filename.
//
// Codec is selected from extension and must be registered.
//
// WriteConfigFile registers all found config.Interface types at any depth with
// the type registry.
//
// If an error occurs it is returned.
WriteConfigFile(filename string, config interface{}) error

// ReadConfigFile reads a configuration file specified by filename into
// config which must be a non-nil pointer to a value compatible with config
// being read.
//
// Codec is selected from extension and must be registered.
//
// If an error occurs it is returned.
//
// ReadConfigFile decodes the loaded stream twice if any Interface structs are
// detected at any level in config. This is required to replace returned
// map[string]interface{} vars in contained interfaces with adequate types at
// marshaling time.
//
// Types must be registered with the registry in order for Interfaces to be
// loaded properly. If an instance of a type being read was not written to file
// prior to this call using WriteConfigFile the type of the value specified by
// config must have been registered manually using RegisterType or
// RegisterTypeByName.
ReadConfigFile(filename string, config interface{}) error

// GetSystemConfigPath returns the path to the configuration directory named as
// the defined prefix under a system-wide configuration directory that depends
// on the underlying operating system and is defined as follows:
//
// darwin:             "/private/etc"
// linux, unix, et al: "/etc"
// windows:            "%ALLUSERSPROFILE%"
//
GetSystemConfigPath() (path string, err error)

// GetUserConfigPath returns the path to the configuration directory named as
// the defined prefix under a user configuration directory that depends
// on the underlying operating system and is defined as follows:
//
// darwin:             "$HOME"
// linux, unix, et al: "$HOME"
// windows:            "%USERPROFILE%"
//
GetUserConfigPath() (path string, err error)

// GetProgramConfigPath returns path to the directory of the executable.
GetProgramConfigPath() string
```

## License

MIT. 

See included LICENSE file.