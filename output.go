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
