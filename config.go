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
	"errors"
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

// bindViperConfig parses the configuration (from cfg and cmd) into the viper instance.
func bindViperConfig(v *viper.Viper, cfg Config, cmd *cobra.Command) error {
	// Handle env variables
	v.SetEnvPrefix(cfg.EnvPrefix)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// cfg.EnvPrefix will contain the desired prefix for the environment variables.
	configPathEnvVar := fmt.Sprintf("%s_CONFIG_PATH", cfg.EnvPrefix)

	// Reads from that configuration if it's specified via environment variable.
	if os.Getenv(configPathEnvVar) != "" {
		cfg.Path = os.Getenv(configPathEnvVar)
	}

	v.SetConfigFile(cfg.Path)
	v.SetConfigType("yaml")

	// Handle config file
	if err := v.ReadInConfig(); err != nil {
		// we make the existence of the config file optional
		if !os.IsNotExist(err) {
			return fmt.Errorf("fatal error config file: %w", err)
		}
	}

	var errs []error

	// Handle flags
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if err := v.BindPFlag(f.Name, f); err != nil {
			errs = append(errs, err)
		}
	})

	if err := errors.Join(errs...); err != nil {
		return fmt.Errorf("error binding flags: %w", err)
	}
	return nil
}

// ParseConfig parses the configuration into cfg.Parsed using viper.
// This is useful for any decorator that needs to parse configuration based on the available flags.
func ParseConfig(cfg Config, cmd *cobra.Command) error {
	parsedType := reflect.TypeOf(cfg.Parsed)

	// Ensure Parsed is a pointer
	if parsedType.Kind() != reflect.Ptr {
		return fmt.Errorf("parsed must be a pointer")
	}

	if parsedType.Elem() != reflect.TypeOf(cfg.DefaultValues) {
		return fmt.Errorf("parsed and defaultValues must be the same type")
	}

	viper := viper.New()

	setDefaults(viper, cfg.DefaultValues)

	if err := bindViperConfig(viper, cfg, cmd); err != nil {
		return fmt.Errorf("error parsing config: %w", err)
	}

	if err := viper.Unmarshal(cfg.Parsed); err != nil {
		return fmt.Errorf("error unmarshalling config: %w", err)
	}

	return nil
}
