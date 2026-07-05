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
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/matryer/is"
)

type resultTestCmd struct{}

func (c *resultTestCmd) Usage() string { return "resultcmd" }

func (c *resultTestCmd) ExecuteResult(context.Context) (any, error) {
	return map[string]any{"name": "pipeline1", "status": "running"}, nil
}

func (c *resultTestCmd) Render(result any) string {
	m, _ := result.(map[string]any)
	return fmt.Sprintf("NAME=%s STATUS=%s\n", m["name"], m["status"])
}

func TestCommandWithResult_Human(t *testing.T) {
	is := is.New(t)

	cmd := New().MustBuildCobraCommand(&resultTestCmd{})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})

	is.NoErr(cmd.Execute())
	is.Equal(buf.String(), "NAME=pipeline1 STATUS=running\n")
}

func TestCommandWithResult_JSON(t *testing.T) {
	is := is.New(t)

	cmd := New().MustBuildCobraCommand(&resultTestCmd{})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--json"})

	is.NoErr(cmd.Execute())
	out := buf.String()
	is.True(strings.Contains(out, `"name": "pipeline1"`))
	is.True(strings.Contains(out, `"status": "running"`))
	is.True(!strings.Contains(out, "NAME=")) // human render is not used
}

// TestCommandWithResult_JSONFlagRegistered proves the framework adds the flag
// automatically — the command doesn't declare it.
func TestCommandWithResult_JSONFlagRegistered(t *testing.T) {
	is := is.New(t)
	cmd := New().MustBuildCobraCommand(&resultTestCmd{})
	is.True(cmd.Flags().Lookup("json") != nil)
}
