// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package config implements configuration related helpers and primitives.
//
// It works with optionally embedded struct types as configuration containers
// and relies on struct tags for functionality.
//
// Config reads the "config" tag from struct fields. It accepts multiple
// key/value pairs which must be separated by ";".
package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/vedranvuk/config/codec"
	"github.com/vedranvuk/errorex"
)

var (
	// ErrConfig is the base error of config package.
	ErrConfig = errorex.New("config")
	// ErrInvalidParam is returned when an invalid parameter is passed to a
	// function.
	ErrInvalidParam = ErrConfig.Wrap("invalid parameter")

	// ErrUnsupportedOS is returned by GetSystemConfigPath and
	// GetUserConfigPath on an unsupported OS.
	ErrUnsupportedOS = ErrConfig.WrapFormat("unsupported OS '%s'")
)

// WriteConfigFile writes config to a file specified by filename.
// Codec is selected from filename extension and must be registered.
// WriteConfigFile registers all Interface types in config at any depth.
// If an error occurs it is returned.
func WriteConfigFile(filename string, config interface{}) error {
	if err := RegisterInterfaces(config); err != nil {
		return err
	}
	c, err := codec.Get(ext(filename))
	if err != nil {
		return err
	}
	data, err := c.Encode(config)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

// ReadConfigFile reads a configuration file specified by filename into
// config which must be a non-nil pointer to a value compatible with
// configuration being read.
//
// Codec is selected from extension and must be registered by importing it in
// the program or package.
//
// ReadConfigFile unmarshals the loaded stream twice if any Interface structs
// are detected at any level in config whose Type field is not empty. First run
// reads Interface.Type fields and initializes Interface.Value to zero values of
// that type then unmarshals the config again to fill ethe Value fields.
//
// Types must be registered with the registry in order for Interfaces to be
// initialized properly. Types are registered automatically when using
// WriteConfigFile and can be manually registered using RegisterType.
//
// If an error occurs it is returned.
func ReadConfigFile(filename string, config interface{}) error {
	c, err := codec.Get(ext(filename))
	if err != nil {
		return err
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	if err := c.Decode(data, config); err != nil {
		return err
	}
	needsreload, err := InitializeInterfaces(config)
	if err != nil {
		return err
	}
	if !needsreload {
		return nil
	}
	return c.Decode(data, config)
}

// ext is a helper that extracts the extension from the filename, without the
// dot. If no extension is found in filename, an empty string is returned.
func ext(filename string) (s string) {
	s = filepath.Ext(filename)
	if len(s) == 0 {
		return
	}
	if s[0] == '.' {
		s = s[1:]
	}
	return
}

// GetSystemConfigPath returns the path to the base system configuration
// directory that depends on the running OS and is defined as follows:
//
// darwin:             "/private/etc"
// unix, linux et al:  "/etc"
// windows:            "%ALLUSERSPROFILE%"
//
// If an unsupported OS is detected returns empty path and ErrUnsupportedOS.
func GetSystemConfigPath() (path string, err error) {
	switch runtime.GOOS {
	case "darwin":
		path = "/private/etc"
	case "aix", "android", "dragonfly", "freebsd", "illumos", "linux", "netbsd",
		"openbsd", "plan9", "solaris":
		path = "/etc"
	case "windows":
		path = os.ExpandEnv("$ALLUSERSPROFILE")
	case "js":
		fallthrough
	default:
		return "", ErrUnsupportedOS.WrapArgs(runtime.GOOS)
	}
	if err != nil {
		return "", err
	}
	return
}

// GetUserConfigPath returns the path to the base user configuration
// directory that depends on the running OS and is defined as follows:
//
// darwin:             "$HOME"
// unix, linux et al:  "$HOME"
// windows:            "%USERPROFILE%"
//
// If an unsupported OS is detected returns empty path and ErrUnsupportedOS.
func GetUserConfigPath() (path string, err error) {
	switch runtime.GOOS {
	case "darwin":
		path = filepath.Join(os.ExpandEnv("$HOME"), ".config")
	case "aix", "android", "dragonfly", "freebsd", "illumos", "linux", "netbsd",
		"openbsd", "plan9", "solaris":
		path = filepath.Join(os.ExpandEnv("$HOME"), ".config")
	case "windows":
		path = os.ExpandEnv("$USERPROFILE")
	case "js":
		fallthrough
	default:
		return "", ErrUnsupportedOS.WrapArgs(runtime.GOOS)
	}
	if err != nil {
		return "", err
	}
	return
}

// GetProgramConfigPath returns path to the directory of the executable.
func GetProgramConfigPath() string {
	return filepath.Dir(os.Args[0])
}
