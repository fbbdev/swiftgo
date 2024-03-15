// Copyright (c) 2024 Fabio Massaioli
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// Package tool simplifies the process of locating, configuring and running binary tools.
package tools

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
)

// An ExitError reports an unsuccessful exit by some tool
// while a Locate* function was gathering information.
type ExitError struct {
	ExitCode int
	Tool     *Tool
}

func (err *ExitError) Error() string {
	return fmt.Sprintf("%v returned non-zero exit code", err.Tool.Desc)
}

// Tool holds the path, arguments and configuration of a binary tool.
type Tool struct {
	// Desc is a user-readable description of the tool to be used in error messages.
	Desc string

	// Hint is the path that was used to look up the tool binary.
	Hint string

	// Path is the absolute path to the tool binary.
	Path string

	// Args holds command line argument that must be passed on whenever the tool is invoked.
	Args []string
}

// Locate looks up the path of the tool binary, if unknown, based on the content of the tool.Hint field.
func (tool *Tool) Locate() (err error) {
	if tool.Path == "" {
		tool.Path, err = exec.LookPath(tool.Hint)
		if err != nil {
			err = fmt.Errorf("%v not found: %w", tool.Desc, err)
		}
	}

	return
}

// ForwardMode stores a set of bitwise or-ed flags that determine
// which streams should be automatically forwarded when running the Go build tool.
type ForwardMode int

const (
	ForwardInput ForwardMode = 1 << iota
	ForwardOutput
	ForwardErrors

	Forward = ForwardInput | ForwardOutput | ForwardErrors

	CaptureInput  = ForwardOutput | ForwardErrors
	CaptureOutput = ForwardInput | ForwardErrors
	CaptureErrors = ForwardInput | ForwardOutput

	Capture = 0
)

// Command returns a Cmd struct instance configured to execute the tool with the given arguments.
// Any argument present in tool.Args is prepended to the given argument list.
// If the ForwardInput flag is present in the mode argument,
// the input stream is forwarded to os.Stdin.
// Similarly, if the ForwardOutput or ForwardErrors flags are present,
// output and error streams are forwarded respectively to os.Stdout and os.Stderr.
func (tool *Tool) Command(mode ForwardMode, args ...string) (cmd *exec.Cmd) {
	if len(tool.Args) > 0 {
		cmd = exec.Command(tool.Path, slices.Concat(tool.Args, args)...)
	} else {
		cmd = exec.Command(tool.Path, args...)
	}

	if mode&ForwardInput != 0 {
		cmd.Stdin = os.Stdin
	}

	if mode&ForwardOutput != 0 {
		cmd.Stdout = os.Stdout
	}

	if mode&ForwardErrors != 0 {
		cmd.Stderr = os.Stderr
	}

	return
}

// Run runs the tool with the given arguments and returns the exit code.
// No error is returned when the invocation succeded with non-zero exit code.
// Input, output and error streams are forwarded respectively
// to os.Stdin, os.Stdout and os.Stderr.
func (tool *Tool) Run(args ...string) (exitCode int, err error) {
	err = tool.Command(Forward, args...).Run()
	exitCode, err = tool.processInvocationError(err)
	return
}

// Output runs the Go tool with the given arguments
// and returns its standard output and the exit code.
// No error is returned when the invocation succeded with non-zero exit code.
// Input and error streams are forwarded to os.Stdin and os.Stderr.
func (tool *Tool) Output(args ...string) (output []byte, exitCode int, err error) {
	output, err = tool.Command(CaptureOutput, args...).Output()
	exitCode, err = tool.processInvocationError(err)
	return
}

// CombinedOutput runs the Go tool with the given arguments
// and returns its combined output and error streams as well as the exit code.
// No error is returned when the invocation succeded with non-zero exit code.
// The input stream is forwarded to os.Stdin.
func (tool *Tool) CombinedOutput(args ...string) (output []byte, exitCode int, err error) {
	output, err = tool.Command(ForwardInput, args...).CombinedOutput()
	exitCode, err = tool.processInvocationError(err)
	return
}

// processInvocationError transforms an error returned by exec.Cmd.Run
// into the format required by Tool methods.
func (tool *Tool) processInvocationError(invocationError error) (exitCode int, err error) {
	if exitError := (*exec.ExitError)(nil); errors.As(invocationError, &exitError) {
		return exitError.ExitCode(), nil
	} else if invocationError != nil {
		return 0, fmt.Errorf("%v invocation failed: %w", tool.Desc, err)
	}
	return
}
