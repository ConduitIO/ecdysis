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
	"bytes"
	"testing"

	"github.com/matryer/is"
	"github.com/spf13/cobra"
)

func TestNewDefaultOutput(t *testing.T) {
	is := is.New(t)
	cmd := &cobra.Command{}
	out := NewDefaultOutput(cmd)

	is.Equal(out.stdout, cmd.OutOrStdout())
	is.Equal(out.stderr, cmd.OutOrStderr())
}

func TestStdout(t *testing.T) {
	is := is.New(t)

	var buf bytes.Buffer
	out := &DefaultOutput{stdout: &buf}

	message := "hello stdout"
	out.Stdout(message)

	is.Equal(buf.String(), message)
}

func TestStderr(t *testing.T) {
	is := is.New(t)

	var buf bytes.Buffer
	out := &DefaultOutput{stderr: &buf}

	message := "hello stderr"
	out.Stderr(message)

	is.Equal(buf.String(), message)
}

func TestOutput(t *testing.T) {
	is := is.New(t)

	var stdoutBuf, stderrBuf bytes.Buffer
	out := &DefaultOutput{}

	out.Output(&stdoutBuf, &stderrBuf)

	stdoutMsg := "stdout test"
	stderrMsg := "stderr test"
	out.Stdout(stdoutMsg)
	out.Stderr(stderrMsg)

	is.Equal(stdoutMsg, stdoutBuf.String())
	is.Equal(stderrMsg, stderrBuf.String())
}
