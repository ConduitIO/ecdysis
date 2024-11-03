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
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Command interface {
	// Usage is the one-line usage message.
	// Recommended syntax is as follows:
	//   [ ] identifies an optional argument. Arguments that are not enclosed in brackets are required.
	//   ... indicates that you can specify multiple values for the previous argument.
	//   |   indicates mutually exclusive information. You can use the argument to the left of the separator or the
	//       argument to the right of the separator. You cannot use both arguments in a single use of the command.
	//   { } delimits a set of mutually exclusive arguments when one of the arguments is required. If the arguments are
	//       optional, they are enclosed in brackets ([ ]).
	// Example: add [-F file | -D dir]... [-f format] profile
	Usage() string
}

type CommandWithAliases interface {
	Command
	// Aliases is an array of aliases that can be used instead of the first word in Usage.
	Aliases() []string
}

type CommandWithArgs interface {
	Command
	// ParseArgs is meant to parse arguments after the command name.
	ParseArgs([]string) error
}

type CommandWithNoHeaders interface {
	Command
	HideHeaders(hide bool)
}

type CommandWithConfirmWithValue interface {
	Command
	// ValueToConfirm adds a prompt before the command is executed where the user is asked to write the exact value as
	// wantInput. If the user input matches the command will be executed, otherwise processing will be stopped.
	ValueToConfirm(ctx context.Context) (wantInput string)
}

type CommandWithPrompt interface {
	Command
	// Prompt adds a prompt before the command is executed where the user is asked to answer y/N to proceed
	Prompt() error

	// SkipPrompt will return logic around when to skip prompt (e.g.: when all flags and arguments are specified)
	SkipPrompt() bool

	// NotConfirmed indicates what to show in case user declines the answer
	NotConfirmed() (prompt string)
}

type CommandWithDocs interface {
	Command
	// Docs returns the documentation for the command.
	Docs() Docs
}

// Docs will be shown to the user when typing 'help' as well as in generated docs.
type Docs struct {
	// Short is the short description shown in the 'help' output.
	Short string
	// Long is the long message shown in the 'help <this-command>' output.
	Long string
	// Example is examples of how to use the command.
	Example string

	// Beta enabled will add (Beta) to the end of the short doc description
	Beta bool
}

type CommandWithDeprecated interface {
	Command
	Deprecated() string
}

type CommandWithExecute interface {
	Command
	// Execute is the actual work function. Most commands will implement this.
	Execute(ctx context.Context) error
}

type CommandWithFlags interface {
	Command
	// Flags returns the set of flags on this command.
	Flags() []Flag
}

type CommandWithHidden interface {
	Command
	// Hidden returns the desired hidden value for the command.
	Hidden() bool
}

type CommandWithLogger interface {
	Command
	// Logger provides the logger to the command.
	Logger(*slog.Logger)
}

// CommandWithoutEvent is to explicitly make a command not traceable via metrics.
type CommandWithoutEvent interface {
	Command
	Event() bool
}

type CommandWithSubCommands interface {
	Command
	// SubCommands defines subcommands of a command.
	SubCommands() []*cobra.Command
}

type CommandWithFeatureFlag interface {
	Command
	FeatureFlag() (string, error)
}

// BuildCobraCommand takes a Command and builds a *cobra.Command from it. It figures out if the command implements any
// other CommandWith* interfaces and configures the cobra command accordingly.
func BuildCobraCommand(c Command) *cobra.Command {
	cmd := &cobra.Command{
		Use: c.Usage(),
	}

	buildCommandWithAliases(cmd, c)
	buildCommandWithArgs(cmd, c)

	// buildCommandWithConfirmWithValue needs to go before buildCommandWithExecute to make sure there's a confirmation prompt
	// prior to execution.
	buildCommandWithConfirmWithValue(cmd, c)
	buildCommandWithConfirmWithoutValue(cmd, c)
	buildCommandWithExecute(cmd, c)

	buildCommandWithDocs(cmd, c)
	buildCommandWithFlags(cmd, c)
	buildCommandWithHidden(cmd, c)
	buildCommandWithDeprecated(cmd, c)
	buildCommandWithLogger(cmd, c)
	buildCommandWithNoHeaders(cmd, c)
	buildCommandWithSubCommands(cmd, c)

	return cmd
}

func buildCommandWithAliases(cmd *cobra.Command, c Command) {
	v, ok := c.(CommandWithAliases)
	if !ok {
		return
	}

	cmd.Aliases = v.Aliases()
}

func buildCommandWithArgs(cmd *cobra.Command, c Command) {
	v, ok := c.(CommandWithArgs)
	if !ok {
		return
	}

	old := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if old != nil {
			err := old(cmd, args)
			if err != nil {
				return err
			}
		}
		return v.ParseArgs(args)
	}
}

func buildCommandWithNoHeaders(cmd *cobra.Command, c Command) {
	v, ok := c.(CommandWithNoHeaders)
	if !ok {
		return
	}

	var noHeaders bool

	cmd.Flags().BoolVar(&noHeaders, "no-headers", false, "display output without headers")

	old := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if old != nil {
			err := old(cmd, args)
			if err != nil {
				return err
			}
		}

		v.HideHeaders(noHeaders)
		return nil
	}
}

func buildCommandWithConfirmWithValue(cmd *cobra.Command, c Command) {
	v, ok := c.(CommandWithConfirmWithValue)
	if !ok {
		return
	}

	var (
		force bool
		yolo  bool
	)
	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation")
	cmd.Flags().BoolVarP(&yolo, "yolo", "", false, "skip confirmation")
	err := cmd.Flags().MarkHidden("yolo")
	if err != nil {
		panic(fmt.Errorf("could not mark flag hidden: %w", err))
	}

	old := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if old != nil {
			err := old(cmd, args)
			if err != nil {
				return err
			}
		}

		// do not prompt for confirmation when --force (or --yolo 😜) is set
		if force || yolo {
			return nil
		}

		wantInput := v.ValueToConfirm(cmd.Context())

		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("To proceed, type %q or re-run this command with --force\n▸ ", wantInput)
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		if wantInput != strings.TrimRight(input, "\r\n") {
			return errors.New("action aborted")
		}

		return nil
	}
}

func buildCommandWithConfirmWithoutValue(cmd *cobra.Command, c Command) {
	v, ok := c.(CommandWithPrompt)
	if !ok {
		return
	}

	var skip bool
	cmd.Flags().BoolVarP(&skip, "yes", "y", false, "skip confirmation prompt")

	old := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if old != nil {
			err := old(cmd, args)
			if err != nil {
				return err
			}
		}

		// do not prompt for confirmation when --yes or when we explicitly want to skip prompt
		if skip || v.SkipPrompt() {
			return nil
		}

		e := v.Prompt()

		if e != nil {
			fmt.Println(v.NotConfirmed())
			os.Exit(1)
		}

		return nil
	}
}

func buildCommandWithDocs(cmd *cobra.Command, c Command) {
	v, ok := c.(CommandWithDocs)
	if !ok {
		return
	}

	docs := v.Docs()
	cmd.Long = docs.Long

	if docs.Beta {
		cmd.Short = fmt.Sprintf("%s (Beta)", docs.Short)
	} else {
		cmd.Short = docs.Short
	}

	cmd.Example = docs.Example
}

func buildCommandWithExecute(cmd *cobra.Command, c Command) {
	v, ok := c.(CommandWithExecute)
	if !ok {
		return
	}

	old := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if old != nil {
			err := old(cmd, args)
			if err != nil {
				return err
			}
		}
		err := v.Execute(cmd.Context())
		if err != nil && strings.Contains(err.Error(), "Unknown or invalid refresh token") {
			return fmt.Errorf("unknown or invalid refresh token, please run `meroxa login` again")
		}
		return err
	}
}

//nolint:funlen,gocyclo // this function has a big switch statement, can't get around that
func buildCommandWithFlags(cmd *cobra.Command, c Command) {
	v, ok := c.(CommandWithFlags)
	if !ok {
		return
	}

	for _, f := range v.Flags() {
		var flags *pflag.FlagSet
		if f.Persistent {
			flags = cmd.PersistentFlags()
		} else {
			flags = cmd.Flags()
		}

		if f.Required {
			f.Usage += " (required)"
		}

		switch val := f.Ptr.(type) {
		case *string:
			if f.Default == nil {
				f.Default = ""
			}
			flags.StringVarP(val, f.Long, f.Short, f.Default.(string), f.Usage)
		case *int:
			if f.Default == nil {
				f.Default = 0
			}
			flags.IntVarP(val, f.Long, f.Short, f.Default.(int), f.Usage)
		case *int8:
			if f.Default == nil {
				f.Default = int8(0)
			}
			flags.Int8VarP(val, f.Long, f.Short, f.Default.(int8), f.Usage)
		case *int16:
			if f.Default == nil {
				f.Default = int16(0)
			}
			flags.Int16VarP(val, f.Long, f.Short, f.Default.(int16), f.Usage)
		case *int32:
			if f.Default == nil {
				f.Default = int32(0)
			}
			flags.Int32VarP(val, f.Long, f.Short, f.Default.(int32), f.Usage)
		case *int64:
			if f.Default == nil {
				f.Default = int64(0)
			}
			flags.Int64VarP(val, f.Long, f.Short, f.Default.(int64), f.Usage)
		case *float32:
			if f.Default == nil {
				f.Default = float32(0)
			}
			flags.Float32VarP(val, f.Long, f.Short, f.Default.(float32), f.Usage)
		case *float64:
			if f.Default == nil {
				f.Default = float64(0)
			}
			flags.Float64VarP(val, f.Long, f.Short, f.Default.(float64), f.Usage)
		case *bool:
			if f.Default == nil {
				f.Default = false
			}
			flags.BoolVarP(val, f.Long, f.Short, f.Default.(bool), f.Usage)
		case *time.Duration:
			if f.Default == nil {
				f.Default = time.Duration(0)
			}
			flags.DurationVarP(val, f.Long, f.Short, f.Default.(time.Duration), f.Usage)
		case *[]bool:
			if f.Default == nil {
				f.Default = []bool(nil)
			}
			flags.BoolSliceVarP(val, f.Long, f.Short, f.Default.([]bool), f.Usage)
		case *[]float32:
			if f.Default == nil {
				f.Default = []float32(nil)
			}
			flags.Float32SliceVarP(val, f.Long, f.Short, f.Default.([]float32), f.Usage)
		case *[]float64:
			if f.Default == nil {
				f.Default = []float64(nil)
			}
			flags.Float64SliceVarP(val, f.Long, f.Short, f.Default.([]float64), f.Usage)
		case *[]int32:
			if f.Default == nil {
				f.Default = []int32(nil)
			}
			flags.Int32SliceVarP(val, f.Long, f.Short, f.Default.([]int32), f.Usage)
		case *[]int64:
			if f.Default == nil {
				f.Default = []int64(nil)
			}
			flags.Int64SliceVarP(val, f.Long, f.Short, f.Default.([]int64), f.Usage)
		case *[]int:
			if f.Default == nil {
				f.Default = []int(nil)
			}
			flags.IntSliceVarP(val, f.Long, f.Short, f.Default.([]int), f.Usage)
		case *[]string:
			if f.Default == nil {
				f.Default = []string(nil)
			}
			flags.StringSliceVarP(val, f.Long, f.Short, f.Default.([]string), f.Usage)
		default:
			panic(fmt.Errorf("unexpected flag value type: %T", val))
		}

		if f.Required {
			err := cobra.MarkFlagRequired(flags, f.Long)
			if err != nil {
				panic(fmt.Errorf("could not mark flag required: %w", err))
			}
		}

		if f.Hidden {
			err := flags.MarkHidden(f.Long)
			if err != nil {
				panic(fmt.Errorf("could not mark flag hidden: %w", err))
			}
		}
	}
}

func buildCommandWithHidden(cmd *cobra.Command, c Command) {
	v, ok := c.(CommandWithHidden)
	if !ok {
		return
	}

	cmd.Hidden = v.Hidden()
}

func buildCommandWithDeprecated(cmd *cobra.Command, c Command) {
	v, ok := c.(CommandWithDeprecated)
	if !ok {
		return
	}

	cmd.Hidden = true

	old := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if old != nil {
			err := old(cmd, args)
			if err != nil {
				return err
			}

			if cmd.Flags().Changed("json") {
				return nil
			}

			c := cmd.Name()
			if cmd.HasParent() {
				c = fmt.Sprintf("%s %s", cmd.Parent().Name(), c)
			}
			fmt.Printf("Command %q is deprecated, %s\n", c, v.Deprecated())
		}

		return nil
	}
}

func buildCommandWithLogger(cmd *cobra.Command, c Command) {
	v, ok := c.(CommandWithLogger)
	if !ok {
		return
	}

	old := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if old != nil {
			err := old(cmd, args)
			if err != nil {
				return err
			}
		}

		v.Logger(slog.Default())
		return nil
	}
}

func buildCommandWithSubCommands(cmd *cobra.Command, c Command) {
	v, ok := c.(CommandWithSubCommands)
	if !ok {
		return
	}

	for _, sub := range v.SubCommands() {
		cmd.AddCommand(sub)
	}
}
