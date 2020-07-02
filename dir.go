// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

var (
	// ErrNoConfigLoaded is returned by LoadConfig when no configs were loaded.
	ErrNoConfigLoaded = ErrConfig.Wrap("no configuration files loaded")

	// ErrProgramDirNotSupported is returned when trying to write or read from
	// a program directory on an platform that does not support it.
	//
	// It is returned by LoadProgramConfig on an OS that does not support
	// loading configs from program directory and internally.
	ErrProgramDirNotSupported = ErrConfig.Wrap("program directory configuration not supported on this os")
)

// Dir is a helper that represents a configuration directory in multiple
// locations defined by priorities, namely: System, User and Local/Executable
// level.
//
// A Dir takes a prefix which defines a subdirectory in either of configuration
// locations. If prefix is a path it is rooted at either configuration location
// being accessed.
//
type Dir struct {
	// prefix is the configuration prefix.
	prefix string
}

// NewDir returns a new Dir with the given prefix.
//
// Prefix represents the name of the directory to be read/written in any of
// locations Dir recognizes. It can be a directory name or a path in case of
// which it will be rooted at all configuration locations.
//
func NewDir(prefix string) *Dir { return &Dir{prefix} }

// LoadSystemConfig loads the config specified by name from the system config
// directory. See LoadConfig for details.
//
// If an error occurs it is returned.
func (d *Dir) LoadSystemConfig(name string, out interface{}) error {

	path, err := GetSystemConfigPath()

	if err != nil {
		return err
	}

	return ReadConfigFile(filepath.Join(path, d.prefix, name), out)
}

// LoadUserConfig loads the config specified by name from the user config
// directory. See LoadConfig for details.
//
// If an error occurs it is returned.
func (d *Dir) LoadUserConfig(name string, out interface{}) error {

	path, err := GetUserConfigPath()

	if err != nil {
		return err
	}

	return ReadConfigFile(filepath.Join(path, d.prefix, name), out)
}

// LoadProgramConfig loads the config specified by name from the program
// directory. See LoadConfig for details.
//
// Loading configuration from program directory is supported on windows only.
//
// If an error occurs it is returned.
func (d *Dir) LoadProgramConfig(name string, out interface{}) error {
	if runtime.GOOS != "windows" {
		return ErrProgramDirNotSupported
	}
	path := filepath.Join(GetProgramConfigPath(), name)
	return ReadConfigFile(path, out)
}

// LoadConfig searches for and loads configuration file specified by name in the
// following locations:
//
// program directory (windows only)
// user configuration directory
// system configuration directory
//
// File is read into out which must be a non-nil pointer to a variable
// compatible with config file being loaded.
//
// If override is not specified first found file in the order decribed above is
// loaded.
//
// If override is specified all found config files from all locations are
// loaded in reverse order described above with config files loaded later
// overriding any values loaded to out thus far.
//
// If a config file with the specified name is not found in any locations an
// ErrNoConfigLoaded is returned.
//
// name specifies the name of the configuration file including extension which
// selects the codec to use when reading the file and must be registered.
//
// An optional path in the specified name is rooted at the configuration
// directory being read and specifies a path to a file in a subdirectory of the
// configuration directory.
//
func (d *Dir) LoadConfig(name string, override bool, out interface{}) (err error) {

	if override {

		loaded := false

		if err = d.LoadSystemConfig(name, out); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		} else {
			loaded = true
		}

		if err = d.LoadUserConfig(name, out); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		} else {
			loaded = true
		}

		if err = d.LoadProgramConfig(name, out); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				if !errors.Is(err, ErrProgramDirNotSupported) {
					return err
				}
			}
		} else {
			loaded = true
		}

		if !loaded {
			return ErrNoConfigLoaded
		}

		return nil
	}

	if err = d.LoadProgramConfig(name, out); err != nil && !errors.Is(err, os.ErrNotExist) && !errors.Is(err, ErrProgramDirNotSupported) {
		return err
	}

	if err = d.LoadUserConfig(name, out); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if err = d.LoadSystemConfig(name, out); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

// enforceFilePath creates directories along the assumed path to a file
// specified by filename or returns an error.
func enforceFilePath(filename string) error {
	return os.MkdirAll(filepath.Dir(filename), os.ModePerm)
}

// SaveSystemConfig saves a configuration file defined by name to the system
// configuration subdirectory defined by Dir prefix. If name contains a path
// Subdirectories are created if they don't exist.
//
// Executable must have permission to write to system locations.
//
// If an error occurs it is returned.
func (d *Dir) SaveSystemConfig(name string, in interface{}) error {

	path, err := GetSystemConfigPath()

	if err != nil {
		return err
	}

	path = filepath.Join(path, d.prefix, name)
	if err := enforceFilePath(path); err != nil {
		return err
	}

	return WriteConfigFile(path, in)
}

// SaveUserConfig saves a configuration file defined by name to the user
// configuration subdirectory defined by Dir prefix. If name contains a path
// Subdirectories are created if they don't exist.
//
// If an error occurs it is returned.
func (d *Dir) SaveUserConfig(name string, in interface{}) error {

	path, err := GetUserConfigPath()

	if err != nil {
		return err
	}

	path = filepath.Join(path, d.prefix, name)
	if err := enforceFilePath(path); err != nil {
		return err
	}

	return WriteConfigFile(path, in)
}

// SaveProgramConfig saves a configuration file defined by name to the
// executable directory.
//
// Prefix is ignored when saving to program directory but if name
// contains a path it is respected and subdirectories are created inside the
// program directory.
//
// Saving to program directory is only supported on Windows.
//
// If an error occurs it is returned.
func (d *Dir) SaveProgramConfig(name string, in interface{}) error {

	if runtime.GOOS != "windows" {
		return ErrProgramDirNotSupported
	}

	path := filepath.Join(GetProgramConfigPath(), d.prefix, name)

	if err := enforceFilePath(path); err != nil {
		return nil
	}

	return WriteConfigFile(path, in)
}
