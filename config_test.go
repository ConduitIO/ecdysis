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

type cookingConfig struct {
	HeatLevel int `long:"heat-level" usage:"sets the heat level" default:"5" mapstructure:"heat-level"`
}

func newCookingConfig() cookingConfig {
	return cookingConfig{HeatLevel: 5}
}

type cookCommand struct {
	flags cookingConfig
	Cfg   cookingConfig
}

func (c *cookCommand) Execute(ctx context.Context) error {
	fmt.Println("Cooking at heat level", c.Cfg.HeatLevel)

	return nil
}

func (c *cookCommand) Config() Config {
	return Config{
		EnvPrefix: "TestParseConfig_CookingConfig",
		Parsed:    &c.Cfg,
		DefaultValues: cookingConfig{
			HeatLevel: 5,
		},
		Path: "./test_parse_config_cooking_config.yaml",
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

func TestParseConfig(t *testing.T) {
	is := is.New(t)

	// t.Setenv("TESTPARSECONFIG_COOKINGCONFIG_HEAT_LEVEL", "33")
	originalArgs := os.Args
	os.Args = []string{originalArgs[0], "--heat-level=22"}
	defer func() {
		os.Args = originalArgs
	}()

	cookCmd := &cookCommand{}
	cookCobraCmd := New().MustBuildCobraCommand(cookCmd)
	is.NoErr(cookCobraCmd.Execute())
	is.Equal(cookCmd.Cfg, cookingConfig{HeatLevel: 33})
}
