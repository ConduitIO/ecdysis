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
	"reflect"

	"github.com/spf13/cobra"
)

// Ecdysis is the main struct that holds all decorators and is used to build
// cobra.Command instances from Command instances.
type Ecdysis struct {
	// Decorators is a list of decorators that are applied to all commands.
	Decorators []Decorator
}

// Command is an interface that represents a command that can be decorated and
// converted to a cobra.Command instance.
type Command interface {
	Usage() string
}

// Decorator is an interface that can be used to decorate a cobra.Command
// instance.
type Decorator interface {
	Decorate(e *Ecdysis, cmd *cobra.Command, c Command) error
}

// New creates a new Ecdysis instance with the provided options. By default, it
// uses the DefaultDecorators. Options can be used to add or replace decorators.
func New(opts ...Option) *Ecdysis {
	e := &Ecdysis{
		Decorators: make([]Decorator, len(DefaultDecorators)),
	}

	// Make a copy of DefaultDecorators to prevent modifications to the original slice.
	copy(e.Decorators, DefaultDecorators)

	// Apply all options.
	for _, opt := range opts {
		opt(e)
	}

	return e
}

// BuildCobraCommand creates a new cobra.Command instance from the provided
// Command instance. It decorates the command with all registered decorators.
func (e *Ecdysis) BuildCobraCommand(c Command) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use: c.Usage(),
	}

	for _, d := range e.Decorators {
		if err := d.Decorate(e, cmd, c); err != nil {
			return nil, fmt.Errorf("failed to decorate command with %T: %w", d, err)
		}
	}

	return cmd, nil
}

// MustBuildCobraCommand creates a new cobra.Command instance from the provided
// Command instance. It decorates the command with all registered decorators. If
// an error occurs, it panics.
func (e *Ecdysis) MustBuildCobraCommand(c Command) *cobra.Command {
	cmd, err := e.BuildCobraCommand(c)
	if err != nil {
		panic(err)
	}
	return cmd
}

// Option is a function type that modifies an Ecdysis instance.
type Option func(*Ecdysis)

func getDecoratorType(d Decorator) reflect.Type {
	t := reflect.TypeOf(d)

	// If it's a pointer, get the underlying type.
	for t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t
}

// WithoutDefaultDecorators removes all default decorators.
func WithoutDefaultDecorators() Option {
	return func(e *Ecdysis) {
		e.Decorators = nil
	}
}

// WithDecorators adds or replaces a decorator of the same type.
func WithDecorators(decorators ...Decorator) Option {
	return func(e *Ecdysis) {
		for _, d := range decorators {
			newType := getDecoratorType(d)

			// Try to find and replace existing decorator of the same type.
			found := false
			for i, existing := range e.Decorators {
				if getDecoratorType(existing) == newType {
					e.Decorators[i] = d
					found = true
					break
				}
			}

			// If no existing decorator of this type was found, append the new one.
			if !found {
				e.Decorators = append(e.Decorators, d)
			}
		}
	}
}
