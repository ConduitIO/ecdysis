package ecdysis

import (
	"fmt"
	"io"
	"os"
)

type Display interface {
	Stdout(string)
	Stderr(string)
}

type DefaultDisplay struct {
	stdout io.Writer
	stderr io.Writer
}

func NewDefaultDisplay() *DefaultDisplay {
	return &DefaultDisplay{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

// Stdout writes a message to the configured standard output.
func (d *DefaultDisplay) Stdout(msg string) {
	fmt.Fprint(d.stdout, msg)
}

// Stderr writes a message to the configured standard error.
func (d *DefaultDisplay) Stderr(msg string) {
	fmt.Fprint(d.stderr, msg)
}

// SetOutput allows overwriting the stdout and/or stderr for specific use cases (like testing).
func (d *DefaultDisplay) SetOutput(stdout, stderr io.Writer) {
	if stdout != nil {
		d.stdout = stdout
	}
	if stderr != nil {
		d.stderr = stderr
	}
}
