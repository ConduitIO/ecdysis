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
	"os"
)

type Output interface {
	Stdout(string)
	Stderr(string)
}

type DefaultOutput struct {
	stdout io.Writer
	stderr io.Writer
}

func NewDefaultOutput() *DefaultOutput {
	return &DefaultOutput{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

// Stdout writes a message to the configured standard output.
func (d *DefaultOutput) Stdout(msg string) {
	fmt.Fprint(d.stdout, msg)
}

// Stderr writes a message to the configured standard error.
func (d *DefaultOutput) Stderr(msg string) {
	fmt.Fprint(d.stderr, msg)
}

// SetOutput allows overwriting the stdout and/or stderr for specific use cases (like testing).
func (d *DefaultOutput) SetOutput(stdout, stderr io.Writer) {
	if stdout != nil {
		d.stdout = stdout
	}
	if stderr != nil {
		d.stderr = stderr
	}
}
