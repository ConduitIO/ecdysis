// Copyright © 2024 Meroxa, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ecdysis

import (
	"fmt"
	"reflect"
	"strconv"
)

// Flag describes a single command line flag.
type Flag struct {
	// Long name of the flag.
	Long string
	// Short name of the flag (one character).
	Short string
	// Usage is the description shown in the 'help' output.
	Usage string
	// Required is used to mark the flag as required.
	Required bool
	// Persistent is used to propagate the flag to subcommands.
	Persistent bool
	// Default is the default value when the flag is not explicitly supplied.
	// It should have the same type as the value behind the pointer in field Ptr.
	Default any
	// Ptr is a pointer to the value into which the flag will be parsed.
	Ptr any
	// Hidden is used to mark the flag as hidden.
	Hidden bool
}

type Flags []Flag

// GetFlag returns the flag with the given long name.
func (f Flags) GetFlag(long string) (Flag, bool) {
	for _, flag := range f {
		if flag.Long == long {
			return flag, true
		}
	}
	return Flag{}, false
}

// SetDefault sets the default value for the flag with the given long name.
func (f Flags) SetDefault(long string, val any) bool {
	for i, flag := range f {
		if flag.Long == long {
			flag.Default = val
			f[i] = flag
			return true
		}
	}
	return false
}

// BuildFlags creates a slice of Flags from a struct.
// It supports nested structs and will only generate flags if it finds a 'short' or 'long' tag.
func BuildFlags(obj any) Flags {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		panic(fmt.Errorf("expected a pointer, got %s", v.Kind()))
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		panic(fmt.Errorf("expected a struct, got %s", v.Kind()))
	}

	return buildFlagsRecursive(v)
}

func buildFlagsRecursive(v reflect.Value) Flags {
	t := v.Type()
	var flags Flags

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Only process fields with a 'short' or 'long' tag
		if hasTag(field.Tag, "short") || hasTag(field.Tag, "long") {
			flag, err := buildFlag(fieldValue, field)
			if err != nil {
				panic(err)
			}
			flags = append(flags, flag)
		} else if fieldValue.Kind() == reflect.Struct {
			// If the field is a struct, recurse into it
			embeddedFlags := buildFlagsRecursive(fieldValue)
			flags = append(flags, embeddedFlags...)
		}
	}
	return flags
}

func hasTag(tag reflect.StructTag, key string) bool {
	_, ok := tag.Lookup(key)
	return ok
}

func buildFlag(val reflect.Value, sf reflect.StructField) (Flag, error) {
	const (
		tagNameLong       = "long"
		tagNameShort      = "short"
		tagNameRequired   = "required"
		tagNamePersistent = "persistent"
		tagNameUsage      = "usage"
		tagNameHidden     = "hidden"
	)

	var (
		long       string
		short      string
		required   bool
		persistent bool
		usage      string
		hidden     bool
	)

	if v, ok := sf.Tag.Lookup(tagNameLong); ok {
		long = v
	}
	if v, ok := sf.Tag.Lookup(tagNameShort); ok {
		short = v
	}
	if v, ok := sf.Tag.Lookup(tagNameRequired); ok {
		var err error
		required, err = strconv.ParseBool(v)
		if err != nil {
			return Flag{}, fmt.Errorf("error parsing tag \"required\": %w", err)
		}
	}
	if v, ok := sf.Tag.Lookup(tagNamePersistent); ok {
		var err error
		persistent, err = strconv.ParseBool(v)
		if err != nil {
			return Flag{}, fmt.Errorf("error parsing tag \"persistent\": %w", err)
		}
	}
	if v, ok := sf.Tag.Lookup(tagNameUsage); ok {
		usage = v
	}
	if v, ok := sf.Tag.Lookup(tagNameHidden); ok {
		var err error
		hidden, err = strconv.ParseBool(v)
		if err != nil {
			return Flag{}, fmt.Errorf("error parsing tag \"hidden\": %w", err)
		}
	}

	return Flag{
		Long:       long,
		Short:      short,
		Usage:      usage,
		Required:   required,
		Persistent: persistent,
		Default:    nil,
		Ptr:        val.Addr().Interface(),
		Hidden:     hidden,
	}, nil
}
