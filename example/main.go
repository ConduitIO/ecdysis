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

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/conduitio/ecdysis"
)

func main() {
	e := ecdysis.New()
	cmd := e.MustBuildCobraCommand(&RootCommand{})
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type RootFlags struct {
	Config  string `long:"config" usage:"config file (default is $HOME/.example-cli.yaml)" persistent:"true"`
	Author  string `long:"author" short:"a" usage:"author name for copyright attribution" persistent:"true"`
	License string `long:"license" short:"l" usage:"name of license for the project" persistent:"true"`
	Viper   bool   `long:"viper" usage:"use Viper for configuration" persistent:"true"`
}

type RootCommand struct {
	flags RootFlags
}

var (
	_ ecdysis.CommandWithFlags       = (*RootCommand)(nil)
	_ ecdysis.CommandWithDocs        = (*RootCommand)(nil)
	_ ecdysis.CommandWithSubCommands = (*RootCommand)(nil)
)

func (c *RootCommand) Usage() string { return "example-cli" }
func (c *RootCommand) Flags() []ecdysis.Flag {
	flags := ecdysis.BuildFlags(&c.flags)
	flags.SetDefault("author", "YOUR NAME")
	flags.SetDefault("viper", true)
	return flags
}

func (c *RootCommand) Docs() ecdysis.Docs {
	return ecdysis.Docs{
		Short: "An example CLI for ecdysis based Applications",
		Long: `Example CLI showcases the power of ecdysis.
This application is an example made using ecdysis,
a wrapper around Cobra that allows you to declare
commands as Go types.`,
	}
}

func (c *RootCommand) SubCommands() []ecdysis.Command {
	return []ecdysis.Command{
		// inject root flags in sub-command
		&AddCommand{rootFlags: &c.flags},
		&VersionCommand{},
	}
}

type AddCommand struct {
	rootFlags *RootFlags
}

var _ ecdysis.CommandWithExecute = (*AddCommand)(nil)

func (c *AddCommand) Usage() string { return "add" }
func (c *AddCommand) Execute(context.Context) error {
	fmt.Printf("root flags: %#v\n", c.rootFlags)
	return nil
}

type VersionCommand struct{}

var (
	_ ecdysis.CommandWithExecute = (*VersionCommand)(nil)
	_ ecdysis.CommandWithDocs    = (*VersionCommand)(nil)
)

func (c *VersionCommand) Usage() string { return "version" }
func (c *VersionCommand) Docs() ecdysis.Docs {
	return ecdysis.Docs{
		Short: "Print the version number of example-cli",
	}
}

func (c *VersionCommand) Execute(context.Context) error {
	fmt.Println("example-cli v0.1.0")
	return nil
}
