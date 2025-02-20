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
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

type Output interface {
	Stdout(any)
	Stderr(any)
}

type DefaultOutput struct {
	stdout io.Writer
	stderr io.Writer
}

func NewDefaultOutput(cmd *cobra.Command) *DefaultOutput {
	return &DefaultOutput{
		stdout: cmd.OutOrStdout(),
		stderr: cmd.OutOrStderr(),
	}
}

// Stdout writes a message to the configured standard output.
func (d *DefaultOutput) Stdout(msg any) {
	fmt.Fprint(d.stdout, msg)
}

// Stderr writes a message to the configured standard error.
func (d *DefaultOutput) Stderr(msg any) {
	fmt.Fprint(d.stderr, msg)
}

// Output allows overwriting the stdout and/or stderr for specific use cases (like testing).
func (d *DefaultOutput) Output(stdout, stderr io.Writer) {
	if stdout != nil {
		d.stdout = stdout
	}
	if stderr != nil {
		d.stderr = stderr
	}
}
