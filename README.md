# Ecdysis

[![License](https://img.shields.io/badge/license-Apache%202-blue)](https://github.com/ConduitIO/ecdysis/blob/main/LICENSE.md)
[![Test](https://github.com/ConduitIO/ecdysis/actions/workflows/test.yml/badge.svg)](https://github.com/ConduitIO/bwlimit/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/conduitio/ecdysis)](https://goreportcard.com/report/github.com/conduitio/ecdysis)
[![Go Reference](https://pkg.go.dev/badge/github.com/conduitio/ecdysis.svg)](https://pkg.go.dev/github.com/conduitio/ecdysis)

Ecdysis is a library for building CLI tools in Go. It is using
[spf13/cobra](https://github.com/spf13/cobra) under the hood and provides a novel
approach to building commands by declaring types with methods that define the
command's behavior.

## Quick Start

Install it using:

```sh
go get github.com/conduitio/ecdysis
```

To create a new command, define a struct that implements `ecdysis.Command` and
any other `ecdysis.CommandWih*` interfaces you need. The recommended pattern is
to list the interfaces that the command implements in a `var` block.

```go
type VersionCommand struct{}

var (
	_ ecdysis.CommandWithExecute = (*VersionCommand)(nil)
	_ ecdysis.CommandWithDocs    = (*VersionCommand)(nil)
)

func (*VersionCommand) Usage() string { return "version" }
func (*VersionCommand) Docs() ecdysis.Docs {
	return ecdysis.Docs{
		Short: "Print the version number of example-cli",
	}
}
func (*VersionCommand) Execute(context.Context) error {
	fmt.Println("example-cli v0.1.0")
	return nil
}
```

In the `main` function, call `ecdysis.New` and build a Cobra command that can
be executed like any other Cobra command.

```go
func main() {
	e := ecdysis.New()
	cmd := e.MustBuildCobraCommand(&VersionCommand{})
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
```

## Decorators

Decorators enable you to add functionality to commands and configure the resulting
Cobra command as you need. Ecdysis comes with a set of default decorators that
you can use to add flags, arguments, confirmation prompts, deprecation notices,
and other features to your commands. Check out the
[Go Reference](https://pkg.go.dev/github.com/conduitio/ecdysis) for a full list
of decorators.

You can implement your own decorators and use them to extend the functionality
of your commands.

For example, this is how you would add support for commands that log using Zerolog:

```go
type CommandWithZerolog interface {
	Command
	Zerolog(zerolog.Logger)
}

type CommandWithZerologDecorator struct{
	Logger zerolog.Logger
}

func (d CommandWithZerologDecorator) Decorate(_ *Ecdysis, _ *cobra.Command, c Command) error {
	v, ok := c.(CommandWithZerolog)
	if !ok {
		return nil
	}

	v.Logger(d.Logger)
	return nil
}
```

You need to supply the decorator to ecdysis when creating it.

```go
func main() {
	e := ecdysis.New(
		ecdysis.WithDecorators(
			&CommandWithZerologDecorator{Logger: zerolog.New(os.Stdout)},
		),
	)
	// build and execute command ...
}
```

### CommandWithConfig

Ecdysis provides an automatic way to parse a configuration file, environment variables, and flags using the [`viper`](https://github.com/spf13/viper) library. To use it, you need to implement the `CommandWithConfig` interface.

The order of precedence for configuration values is:

1. Default values (slices and maps are not currently supported)
2. Configuration file
3. Environment variables
4. Flags


> [!IMPORTANT]  
> For flags, it's important to set default values to ensure that the configuration will be correctly parsed. 
> Otherwise, they will be empty, and it will be considered as if the user set that intentionally.
> example: `flags.SetDefault("config.path", c.cfg.ConduitCfgPath)`

```go
var (
    _ ecdysis.CommandWithFlags   = (*RootCommand)(nil)
    _ ecdysis.CommandWithExecute = (*RootCommand)(nil)
    _ ecdysis.CommandWithConfig  = (*RootCommand)(nil)
)

type ConduitConfig struct {
    ConduitCfgPath string `long:"config.path" usage:"global conduit configuration file" default:"./conduit.yaml"`
    
    Connectors struct {
        Path string `long:"connectors.path" usage:"path to standalone connectors' directory"`
    }
    
    // ...
}

type RootFlags struct {
    ConduitConfig // you can embed any configuration, and it'll use the proper tags
}

type RootCommand struct {
    flags RootFlags
    cfg   ConduitConfig
}

func (c *RootCommand) Config() ecdysis.Config {
    return ecdysis.Config{
        EnvPrefix:     "CONDUIT",
        Parsed:        &c.Cfg,
        Path:          c.flags.ConduitCfgPath,
        DefaultValues: conduit.DefaultConfigWithBasePath(path),
    }
}

func (c *RootCommand) Execute(_ context.Context) error {
    // c.cfg is now populated with the right parsed configuration
    return nil
}

func (c *RootCommand) Flags() []ecdysis.Flag {
    flags := ecdysis.BuildFlags(&c.flags)
    
    // set a default value for each flag
    flags.SetDefault("config.path", c.cfg.ConduitCfgPath) 
    // ...
	
    return flags
}
```

### Fetching `cobra.Command` from `CommandWithExecute`

If you need to access the `cobra.Command` instance from a `CommandWithExecute` implementation, you can utilize
the `ecdysis.CobraCmdFromContext` function to fetch it from the context:

```go
func (c *RootCommand) Execute(ctx context.Context) error {
    if cmd := ecdysis.CobraCmdFromContext(ctx); cmd != nil {
        return cmd.Help()
    }
    return nil
}
```

## Flags

Ecdysis provides a way to define flags using field tags. Flags will be
automatically parsed and populated.

```go
type MyCommand struct {
	flags struct {
		Verbose bool   `long:"verbose" short:"v", usage:"enable verbose output" persistent:"true"`
		Config  string `long:"config" usage:"config file (default is $HOME/.example-cli.yaml)" persistent:"true"`
	}
}


func (c *MyCommand) Flags() []ecdysis.Flag {
	return ecdysis.BuildFlags(&c.flags)
}
```

A full list of supported tags:

- `long`: The long flag name
- `short`: The short flag name
- `required`: Whether the flag is required
- `persistent`: Whether the flag is persistent (i.e. available to subcommands)
- `usage`: The flag usage
- `hidden`: Whether the flag is hidden (i.e. not shown in help)

For a more example on how to use persistent flags in subcommands, see the
[example](./example).