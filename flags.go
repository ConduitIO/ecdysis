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
	// Default is the default value when the flag is not explicitly supplied. It should have the same type as the value
	// behind the pointer in field Ptr.
	Default interface{}
	// Ptr is a pointer to the value into which the flag will be parsed.
	Ptr interface{}
	// Hidden is used to mark the flag as hidden.
	Hidden bool
}

func BuildFlags(obj interface{}) []Flag {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		panic(fmt.Errorf("expected a pointer, got %s", v.Kind()))
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		panic(fmt.Errorf("expected a struct, got %s", v.Kind()))
	}
	t := v.Type()

	var err error
	flags := make([]Flag, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		flags[i], err = buildFlag(v.Field(i), t.Field(i))
		if err != nil {
			panic(err)
		}
	}
	return flags
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
