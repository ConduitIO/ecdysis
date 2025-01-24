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
