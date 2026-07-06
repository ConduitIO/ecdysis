// Copyright © 2025 Meroxa, Inc.
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
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/matryer/is"
)

var testConfigPath = "./test_parse_config_cooking_config.yaml"

type cookingConfig struct {
	HeatLevel int `long:"heat-level" usage:"sets the heat level" default:"5" mapstructure:"heat-level"`
}

func newCookingConfig() cookingConfig {
	return cookingConfig{HeatLevel: 5}
}

type cookCommand struct {
	Cfg cookingConfig
}

func (c *cookCommand) Execute(context.Context) error {
	return nil
}

func (c *cookCommand) Config() Config {
	return Config{
		EnvPrefix: "TestParseConfig_CookingConfig",
		Parsed:    &c.Cfg,
		DefaultValues: cookingConfig{
			HeatLevel: 5,
		},
		Path: testConfigPath,
	}
}

func (c *cookCommand) Flags() []Flag {
	flags := BuildFlags(&c.Cfg)

	c.Cfg = newCookingConfig()
	flags.SetDefault("heat-level", c.Cfg.HeatLevel)
	return flags
}

func (c *cookCommand) Usage() string {
	return "cook something"
}

func (c *cookCommand) Docs() Docs {
	return Docs{
		Short:   "cook short description",
		Long:    "cook long description",
		Example: "cook --heat-level 10",
	}
}

// nestedConfig mirrors the shape that triggers the collision: a scalar flag whose
// name equals a nested-struct config key (here, `pipelines` vs the `Pipelines`
// struct with `pipelines.*` keys).
type nestedConfig struct {
	Pipelines struct {
		Path string `long:"pipelines.path" usage:"pipelines path" default:"" mapstructure:"path"`
	} `mapstructure:"pipelines"`
}

type nestedCommand struct {
	Cfg           nestedConfig
	pipelines     string // value of the excluded --pipelines alias flag
	excludedFlags []string
}

func (c *nestedCommand) Execute(context.Context) error { return nil }
func (c *nestedCommand) Usage() string                 { return "nested" }

func (c *nestedCommand) Config() Config {
	return Config{
		EnvPrefix:     "TestExcludedFlags",
		Parsed:        &c.Cfg,
		DefaultValues: nestedConfig{},
		Path:          "./does-not-exist.yaml",
		ExcludedFlags: c.excludedFlags,
	}
}

func (c *nestedCommand) Flags() []Flag {
	flags := BuildFlags(&c.Cfg)
	// --pipelines is a scalar flag whose name collides with the Pipelines struct.
	flags = append(flags, Flag{Long: "pipelines", Ptr: &c.pipelines, Usage: "alias for pipelines.path"})
	return flags
}

func TestParseConfig_ExcludedFlags(t *testing.T) {
	is := is.New(t)

	// Without exclusion: binding the scalar `pipelines` flag into the Pipelines
	// struct fails.
	notExcluded := &nestedCommand{}
	cmd := New().MustBuildCobraCommand(notExcluded)
	cmd.SetArgs([]string{"--pipelines=/foo"})
	is.True(cmd.Execute() != nil) // struct-key collision

	// With exclusion: no collision, and the flag is still parsed by cobra.
	excluded := &nestedCommand{excludedFlags: []string{"pipelines"}}
	cmd = New().MustBuildCobraCommand(excluded)
	cmd.SetArgs([]string{"--pipelines=/foo"})
	is.NoErr(cmd.Execute())
	is.Equal(excluded.pipelines, "/foo")      // cobra parsed it
	is.Equal(excluded.Cfg.Pipelines.Path, "") // not bound into the config struct
}

func TestParseConfig_NameWithDash_EnvVar(t *testing.T) {
	is := is.New(t)

	t.Setenv("TESTPARSECONFIG_COOKINGCONFIG_HEAT_LEVEL", "33")

	cookCmd := &cookCommand{}
	cookCobraCmd := New().MustBuildCobraCommand(cookCmd)
	is.NoErr(cookCobraCmd.Execute())
	is.Equal(cookCmd.Cfg, cookingConfig{HeatLevel: 33})
}

func TestParseConfig_NameWithDash_Flag(t *testing.T) {
	is := is.New(t)

	originalArgs := os.Args
	os.Args = []string{originalArgs[0], "--heat-level=22"}
	defer func() {
		os.Args = originalArgs
	}()

	cookCmd := &cookCommand{}
	cookCobraCmd := New().MustBuildCobraCommand(cookCmd)
	is.NoErr(cookCobraCmd.Execute())
	is.Equal(cookCmd.Cfg, cookingConfig{HeatLevel: 22})
}

func TestParseConfig_NameWithDash_File(t *testing.T) {
	is := is.New(t)

	cookCmd := &cookCommand{}
	cookCobraCmd := New().MustBuildCobraCommand(cookCmd)
	is.NoErr(cookCobraCmd.Execute())
	is.Equal(cookCmd.Cfg, cookingConfig{HeatLevel: 11})
}

func TestParseConfig_NameWithDash_Default(t *testing.T) {
	is := is.New(t)
	cfgFile, err := os.CreateTemp("", "test_parse_config_cooking_config_empty.yaml")
	is.NoErr(err)
	defer os.Remove(cfgFile.Name())

	testConfigPath = cfgFile.Name()

	cookCmd := &cookCommand{}
	cookCobraCmd := New().MustBuildCobraCommand(cookCmd)
	is.NoErr(cookCobraCmd.Execute())
	is.Equal(cookCmd.Cfg, cookingConfig{HeatLevel: 5})
}

// customCookCommand is a command that uses a custom environment prefix for testing.
type customCookCommand struct {
	cookCommand
}

// Config returns a configuration with a custom environment prefix.
func (c *customCookCommand) Config() Config {
	cfg := c.cookCommand.Config()
	cfg.EnvPrefix = "MY_ENV"
	return cfg
}

func TestParseConfig_CustomConfigPath(t *testing.T) {
	is := is.New(t)

	// Create a temporary config file with a unique heat-level value
	customConfigFile, err := os.CreateTemp("", "test_custom_config_*.yaml")
	is.NoErr(err)
	defer os.Remove(customConfigFile.Name())

	// Write custom config to the temporary file
	customHeatLevel := 42
	_, err = fmt.Fprintf(customConfigFile, "heat-level: %d\n", customHeatLevel)
	is.NoErr(err)
	err = customConfigFile.Close()
	is.NoErr(err)

	// Set MY_ENV_CONFIG_PATH environment variable to the temporary file path
	t.Setenv("MY_ENV_CONFIG_PATH", customConfigFile.Name())

	// Execute the command and verify the config was loaded from the custom path
	customCmd := &customCookCommand{}
	cookCobraCmd := New().MustBuildCobraCommand(customCmd)
	is.NoErr(cookCobraCmd.Execute())
	is.Equal(customCmd.Cfg, cookingConfig{HeatLevel: customHeatLevel})
}
