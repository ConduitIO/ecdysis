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
	"os"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	EnvPrefix     string
	Parsed        any
	DefaultValues any
	Path          string
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

// parseConfig parses the configuration (from cfg and cmd) into the viper instance.
func parseConfig(v *viper.Viper, cfg Config, cmd *cobra.Command) error {
	// Handle env variables
	v.SetEnvPrefix(cfg.EnvPrefix)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Handle config file
	v.SetConfigFile(cfg.Path)
	if err := v.ReadInConfig(); err != nil {
		// we make the existence of the config file optional
		if !os.IsNotExist(err) {
			return fmt.Errorf("fatal error config file: %w", err)
		}
	}

	var errors []error

	// Handle flags
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if err := v.BindPFlag(f.Name, f); err != nil {
			errors = append(errors, err)
		}
	})

	if len(errors) > 0 {
		var errStrs []string
		for _, err := range errors {
			errStrs = append(errStrs, err.Error())
		}
		return fmt.Errorf("error binding flags: %s", strings.Join(errStrs, "; "))
	}
	return nil
}
