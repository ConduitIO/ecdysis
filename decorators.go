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
	"reflect"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var DefaultDecorators = []Decorator{
	CommandWithLoggerDecorator{},
	CommandWithAliasesDecorator{},
	CommandWithFlagsDecorator{},

	// Parsing Config need to be after Flags to make sure the flags are parsed.
	CommandWithParsingConfigDecorator{},

	CommandWithDocsDecorator{},
	CommandWithHiddenDecorator{},
	CommandWithSubCommandsDecorator{},
	CommandWithDeprecatedDecorator{},
	CommandWithArgsDecorator{},

	// Confirm and Prompt need to go before Execute to make sure there's a
	// confirmation prompt prior to execution.
	CommandWithConfirmDecorator{},
	CommandWithPromptDecorator{},

	CommandWithExecuteDecorator{},
}

// -- LOGGER -------------------------------------------------------------------

// CommandWithLogger can be implemented by a command to get a logger.
type CommandWithLogger interface {
	Command
	// Logger provides the logger to the command.
	Logger(*slog.Logger)
}

// CommandWithLoggerDecorator is a decorator that provides a logger to the command.
// If the Logger field is not set, the default slog logger will be provided.
type CommandWithLoggerDecorator struct {
	Logger *slog.Logger
}

// Decorate provides the logger to the command.
func (d CommandWithLoggerDecorator) Decorate(_ *Ecdysis, _ *cobra.Command, c Command) error {
	v, ok := c.(CommandWithLogger)
	if !ok {
		return nil
	}

	if d.Logger == nil {
		v.Logger(slog.Default())
	} else {
		v.Logger(d.Logger)
	}

	return nil
}

// -- ALIASES ------------------------------------------------------------------

// CommandWithAliases can be implemented by a command to provide aliases.
type CommandWithAliases interface {
	Command
	// Aliases is a slice of aliases that can be used instead of the first word
	// in Usage.
	Aliases() []string
}

// CommandWithAliasesDecorator is a decorator that sets the command aliases.
type CommandWithAliasesDecorator struct{}

// Decorate sets the command aliases.
func (CommandWithAliasesDecorator) Decorate(_ *Ecdysis, cmd *cobra.Command, c Command) error {
	v, ok := c.(CommandWithAliases)
	if !ok {
		return nil
	}

	cmd.Aliases = v.Aliases()
	return nil
}

// -- FLAGS --------------------------------------------------------------------

// CommandWithFlags can be implemented by a command to provide flags.
type CommandWithFlags interface {
	Command
	// Flags returns the set of flags on this command.
	Flags() []Flag
}

// CommandWithFlagsDecorator is a decorator that sets the command flags.
type CommandWithFlagsDecorator struct{}

// Decorate sets the command flags.
//
//nolint:funlen,gocyclo,gocognit,forcetypeassert // this function has a big switch statement, can't get around that
func (CommandWithFlagsDecorator) Decorate(_ *Ecdysis, cmd *cobra.Command, c Command) error {
	v, ok := c.(CommandWithFlags)
	if !ok {
		return nil
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
			return fmt.Errorf("unexpected flag value type: %T", val)
		}

		if f.Required {
			err := cobra.MarkFlagRequired(flags, f.Long)
			if err != nil {
				return fmt.Errorf("could not mark flag required: %w", err)
			}
		}

		if f.Hidden {
			err := flags.MarkHidden(f.Long)
			if err != nil {
				return fmt.Errorf("could not mark flag hidden: %w", err)
			}
		}
	}

	return nil
}

// -- PARSING CONFIGURATION --------------------------------------------------------------------

// CommandWithParsingConfig can be implemented by a command to parsing configuration.
type CommandWithParsingConfig interface {
	Command

	Config() UserConfig
}

// CommandWithParsingConfigDecorator is a decorator that sets the command flags.
type CommandWithParsingConfigDecorator struct{}

// Decorate parses the configuration based on flags.
func (CommandWithParsingConfigDecorator) Decorate(_ *Ecdysis, cmd *cobra.Command, c Command) error {
	v, ok := c.(CommandWithParsingConfig)
	if !ok {
		return nil
	}

	old := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if old != nil {
			err := old(cmd, args)
			if err != nil {
				return err
			}
		}

		usrCfg := v.Config()

		// Ensure ParsedConfig is a pointer
		if reflect.ValueOf(usrCfg.ParsedConfig).Kind() != reflect.Ptr {
			return fmt.Errorf("ParsedConfig must be a pointer")
		}

		viper := viper.New()

		// set default values
		setDefaults(viper, usrCfg.DefaultConfig)

		// Set environment variable handling
		viper.SetEnvPrefix(usrCfg.Prefix)
		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

		// Load configuration file
		viper.SetConfigFile(usrCfg.ConfigPath)
		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("fatal error config file: %w", err)
		}

		// Bind flags to Viper
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if err := viper.BindPFlag(f.Name, f); err != nil {
				fmt.Printf("error binding flag: %v\n", err)
			}
		})
		// Unmarshal the configuration into the ParsedConfig
		if err := viper.Unmarshal(usrCfg.ParsedConfig); err != nil {
			return fmt.Errorf("error unmarshalling config: %w", err)
		}
		return nil
	}
	return nil
}

// -- DOCS ---------------------------------------------------------------------

// CommandWithDocs can be implemented by a command to provide documentation.
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
}

// CommandWithDocsDecorator is a decorator that sets the command documentation.
type CommandWithDocsDecorator struct{}

// Decorate sets the command documentation.
func (CommandWithDocsDecorator) Decorate(_ *Ecdysis, cmd *cobra.Command, c Command) error {
	v, ok := c.(CommandWithDocs)
	if !ok {
		return nil
	}

	docs := v.Docs()
	cmd.Long = docs.Long
	cmd.Short = docs.Short
	cmd.Example = docs.Example

	return nil
}

// -- HIDDEN -------------------------------------------------------------------

// CommandWithHidden can be implemented by a command to hide it from the help.
type CommandWithHidden interface {
	Command
	// Hidden returns the desired hidden value for the command.
	Hidden() bool
}

// CommandWithHiddenDecorator is a decorator that sets the command hidden value.
type CommandWithHiddenDecorator struct{}

// Decorate sets the command hidden value.
func (CommandWithHiddenDecorator) Decorate(_ *Ecdysis, cmd *cobra.Command, c Command) error {
	v, ok := c.(CommandWithHidden)
	if !ok {
		return nil
	}

	cmd.Hidden = v.Hidden()
	return nil
}

// -- SUB COMMANDS -------------------------------------------------------------

// CommandWithSubCommands can be implemented by a command to provide subcommands.
type CommandWithSubCommands interface {
	Command
	// SubCommands defines subcommands of a command.
	SubCommands() []Command
}

// CommandWithSubCommandsDecorator is a decorator that sets the command subcommands.
type CommandWithSubCommandsDecorator struct{}

// Decorate sets the command subcommands.
func (CommandWithSubCommandsDecorator) Decorate(e *Ecdysis, cmd *cobra.Command, c Command) error {
	v, ok := c.(CommandWithSubCommands)
	if !ok {
		return nil
	}

	for _, sub := range v.SubCommands() {
		subCmd, err := e.BuildCobraCommand(sub)
		if err != nil {
			return fmt.Errorf("failed to build subcommand %q: %w", sub.Usage(), err)
		}
		cmd.AddCommand(subCmd)
	}
	return nil
}

// -- DEPRECATED ---------------------------------------------------------------

// CommandWithDeprecated can be implemented by a command to mark it as deprecated
// and print a message when it is used.
type CommandWithDeprecated interface {
	Command
	// Deprecated returns a message that will be printed when the command is used.
	Deprecated() string
}

// CommandWithDeprecatedDecorator is a decorator that deprecates the command.
type CommandWithDeprecatedDecorator struct{}

// Decorate deprecates the command.
func (CommandWithDeprecatedDecorator) Decorate(_ *Ecdysis, cmd *cobra.Command, c Command) error {
	v, ok := c.(CommandWithDeprecated)
	if !ok {
		return nil
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

	return nil
}

// -- ARGS ---------------------------------------------------------------------

// CommandWithArgs can be implemented by a command to parse arguments.
type CommandWithArgs interface {
	Command
	// Args is meant to parse arguments after the command name.
	Args([]string) error
}

// CommandWithArgsDecorator is a decorator that provides the command arguments.
type CommandWithArgsDecorator struct{}

// Decorate provides the command arguments.
func (CommandWithArgsDecorator) Decorate(_ *Ecdysis, cmd *cobra.Command, c Command) error {
	v, ok := c.(CommandWithArgs)
	if !ok {
		return nil
	}

	old := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if old != nil {
			err := old(cmd, args)
			if err != nil {
				return err
			}
		}
		return v.Args(args)
	}
	return nil
}

// -- CONFIRM ------------------------------------------------------------------

// CommandWithConfirm can be implemented by a command to require confirmation
// before execution. The user will be prompted to enter a specific value.
// If the value matches, the command will be executed, otherwise it will be
// aborted.
type CommandWithConfirm interface {
	Command
	// ValueToConfirm adds a prompt before the command is executed where the
	// user is asked to write the exact value as wantInput. If the user input
	// matches the command will be executed, otherwise processing will be
	// aborted.
	ValueToConfirm(context.Context) (wantInput string)
}

// CommandWithConfirmDecorator is a decorator that sets up a confirmation prompt
// before executing the command.
type CommandWithConfirmDecorator struct{}

// Decorate sets up a confirmation prompt before executing the command.
func (CommandWithConfirmDecorator) Decorate(_ *Ecdysis, cmd *cobra.Command, c Command) error {
	v, ok := c.(CommandWithConfirm)
	if !ok {
		return nil
	}

	var (
		force bool
		yolo  bool
	)
	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation prompt")
	cmd.Flags().BoolVarP(&yolo, "yolo", "", false, "skip confirmation prompt")
	err := cmd.Flags().MarkHidden("yolo")
	if err != nil {
		return fmt.Errorf("could not mark flag hidden: %w", err)
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
			return fmt.Errorf("failed to read user input: %w", err)
		}

		if wantInput != strings.TrimRight(input, "\r\n") {
			return errors.New("action aborted")
		}

		return nil
	}

	return nil
}

// -- PROMPT -------------------------------------------------------------------

// CommandWithPrompt can be implemented by a command to require confirmation
// before execution. The user will be prompted to answer Y/N to proceed.
type CommandWithPrompt interface {
	Command

	// Prompt adds a prompt before the command is executed where the user is
	// asked to answer Y/N to proceed. It returns the message to be printed and
	// a boolean indicating if the prompt was successfully processed.
	Prompt() (message string, ok bool)
	// SkipPrompt will return logic around when to skip prompt (e.g.: when all
	// flags and arguments are specified).
	SkipPrompt() bool
}

// CommandWithPromptDecorator is a decorator that sets up a confirmation prompt
// before executing the command.
type CommandWithPromptDecorator struct{}

// Decorate sets up a confirmation prompt before executing the command.
func (CommandWithPromptDecorator) Decorate(_ *Ecdysis, cmd *cobra.Command, c Command) error {
	v, ok := c.(CommandWithPrompt)
	if !ok {
		return nil
	}

	var (
		force bool
		yolo  bool
	)
	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation prompt")
	cmd.Flags().BoolVarP(&yolo, "yolo", "", false, "skip confirmation prompt")
	err := cmd.Flags().MarkHidden("yolo")
	if err != nil {
		return fmt.Errorf("could not mark flag hidden: %w", err)
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
		if force || yolo || v.SkipPrompt() {
			return nil
		}

		msg, ok := v.Prompt()

		if !ok {
			fmt.Println(msg)
			os.Exit(1)
		}

		return nil
	}

	return nil
}

// -- EXECUTE ------------------------------------------------------------------

// CommandWithExecute can be implemented by a command to provide an execution
// function.
type CommandWithExecute interface {
	Command
	// Execute is the actual work function. Most commands will implement this.
	Execute(ctx context.Context) error
}

// CommandWithExecuteDecorator is a decorator that sets the command execution.
type CommandWithExecuteDecorator struct{}

// Decorate sets the command execution.
func (CommandWithExecuteDecorator) Decorate(_ *Ecdysis, cmd *cobra.Command, c Command) error {
	v, ok := c.(CommandWithExecute)
	if !ok {
		return nil
	}

	old := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if old != nil {
			err := old(cmd, args)
			if err != nil {
				return err
			}
		}
		return v.Execute(cmd.Context())
	}

	return nil
}
