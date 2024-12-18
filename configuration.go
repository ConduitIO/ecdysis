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

// setDefaults sets the default values for the configuration. slices and maps are not supported.
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

		fieldName := fieldType.Tag.Get("long")
		if fieldName == "" {
			fieldName = fieldType.Tag.Get("short")
		}

		if fieldName == "" {
			continue
		}

		switch field.Kind() { //nolint:exhaustive // no need to handle all cases
		case reflect.Struct:
			setDefaults(v, field.Interface())
		case reflect.Ptr:
			if !field.IsNil() {
				setDefaults(v, field.Interface())
			}
		default:
			if field.CanInterface() {
				v.SetDefault(fieldName, field.Interface())
			}
		}
	}
}
