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
	"reflect"

	"github.com/spf13/viper"
)

type Config struct {
	EnvPrefix  string
	ParsedCfg  any
	DefaultCfg any
	ConfigPath string
}

func setDefaults(v *viper.Viper, defaults interface{}) {
	val := reflect.ValueOf(defaults)
	typ := reflect.TypeOf(defaults)

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return
		}
		val = val.Elem()
		typ = typ.Elem()
	}

	if val.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Use the long or short tag if available
		fieldName := fieldType.Tag.Get("long")
		if fieldName == "" {
			fieldName = fieldType.Tag.Get("short")
		}

		// Skip fields without a long or short tag
		if fieldName == "" {
			continue
		}

		switch field.Kind() {
		case reflect.Struct:
			// Recursively handle nested structs
			setDefaults(v, field.Interface())
		case reflect.Ptr:
			// Handle pointer fields
			if !field.IsNil() {
				setDefaults(v, field.Interface())
			}
		default:
			// Set the default value
			if field.CanInterface() {
				v.SetDefault(fieldName, field.Interface())
			}
		}
	}
}
