package io

import (
	"bytes"
	"io"
	"os"
)

// InputOutput holds the reader and writers used during execution
type InputOutput struct {
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
	ReadStdin StdinReader
}

// Standard creates InputOutput for use from a terminal
func Standard() InputOutput {
	return withStdinReader(&InputOutput{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
}

// Headless creates InputOutput for use from a terminal that can't accept input
func Headless() InputOutput {
	return withStdinReader(&InputOutput{
		Stdin:  nil,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
}

// Buffered creates a buffered InputOutput object
func Buffered(stdin io.Reader) (io InputOutput, stdout *bytes.Buffer, stderr *bytes.Buffer) {
	var bufOut bytes.Buffer
	var bufErr bytes.Buffer
	return withStdinReader(&InputOutput{
		Stdin:  stdin,
		Stdout: &bufOut,
		Stderr: &bufErr,
	}), &bufOut, &bufErr
}

// BufferedCombined creates a buffered InputOutput object
func BufferedCombined(stdin io.Reader) (io InputOutput, combined *bytes.Buffer) {
	var buf bytes.Buffer
	return withStdinReader(&InputOutput{
		Stdin:  stdin,
		Stdout: &buf,
		Stderr: &buf,
	}), &buf
}

func withStdinReader(io *InputOutput) InputOutput {
	io.ReadStdin = NewStdinReader(*io)
	return *io
}
