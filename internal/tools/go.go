// Copyright (c) 2024 Fabio Massaioli
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/fbbdev/swiftgo/internal/quoted"
)

// GoOverrideKey is the environment variable that the end-user may set
// to override the location of the Go tool.
const GoOverrideKey = "SWIFTGO_GOTOOL"

// GoTool holds the path, arguments and configuration of the Go build tool.
type GoTool struct {
	Tool

	// Version holds the major, minor version and patch number of the tool binary.
	Version struct {
		String string
		Major  int
		Minor  int
		Patch  int
	}

	// Env holds the Go environment variables as reported by the 'go env' command.
	Env map[string]string
}

var goVersionRegex = regexp.MustCompile(`\bgo(\d+)\.(\d+)(?:\.(\d+))?\b`)

// LocateGoTool looks up the binary of the Go build tool
// according to the following heuristics:
//
//   - if the SWIFTGO_GOTOOL environment variable is set,
//     it is parsed as a command followed by a sequence of arguments
//     and used to look up the binary;
//   - if the GOROOT variable is set either in the program environment
//     or in the Go environment configuration,
//     LocateGoTool looks up for a binary at '$GOROOT/bin/go';
//   - otherwise, LocateGoTool looks for a binary named 'go'
//     in the directories named by the PATH environment variable.
//
// When the binary is found, tool.Env is populated with the Go environment
// as reported by the 'go env' command.
func LocateGoTool() (tool *GoTool, err error) {
	tool = &GoTool{
		Tool: Tool{
			Hint: "go",
		},
		Env: make(map[string]string, 64), // currently the Go environment comprises 40 variables
	}

	const descFmt = `Go tool "%v"`

	if override := os.Getenv(GoOverrideKey); override != "" {
		tool.Args, err = quoted.Split(override)
		if err != nil || len(tool.Args) < 1 {
			err = fmt.Errorf(GoOverrideKey+" environment variable could not be parsed: %w", err)
			return
		}

		tool.Desc = fmt.Sprintf(descFmt, override)
		tool.Hint = tool.Args[0]
		tool.Args = tool.Args[1:]
	} else {
		root := os.Getenv("GOROOT")
		if root == "" {
			root = queryEnvFile("GOROOT")
		}

		if root != "" {
			tool.Hint = filepath.Join(root, "bin", "go")
		}
	}

	if tool.Desc == "" {
		tool.Desc = fmt.Sprintf(descFmt, tool.Hint)
	}

	if err = tool.Locate(); err != nil {
		return
	}

	var version, env []byte

	// query version string and environment, then check for errors
	scope := "Go version"
	version, exitCode, err := tool.Output("version")
	if err == nil && exitCode == 0 {
		scope = "Go environment configuration"
		env, exitCode, err = tool.Output("env", "-json")
	}
	if err != nil {
		return
	} else if exitCode != 0 {
		err = fmt.Errorf("%v could not be retrieved: %w", scope, &ExitError{exitCode, &tool.Tool})
		return
	}

	tool.Version.String = string(version)

	// try parsing all fields of the version string, then check for errors
	errUnsupportedVersion := fmt.Errorf("Go version could not be retrieved: %v returned unsupported version string", tool.Desc)
	err = errUnsupportedVersion
	if match := goVersionRegex.FindSubmatch(version); match != nil {
		tool.Version.Major, err = strconv.Atoi(string(match[1]))
		if err == nil {
			tool.Version.Minor, err = strconv.Atoi(string(match[2]))
		}
		if err == nil && len(match[3]) > 0 {
			tool.Version.Patch, err = strconv.Atoi(string(match[3]))
		}
	}
	if err != nil {
		err = errUnsupportedVersion
		return
	}

	err = json.Unmarshal(env, &tool.Env)
	if err != nil {
		err = fmt.Errorf("Go environment configuration could not be retrieved: %w", err)
	}

	return
}
