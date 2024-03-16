// Copyright (c) 2024 Fabio Massaioli
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fbbdev/swiftgo/internal/tools"
	"github.com/fbbdev/swiftgo/internal/version"
)

const usage = `SwiftGo is a wrapper for the standard Go tool that adds support for embedded
Swift code when targeting darwin systems.

Usage:

    swiftgo <go command> [go arguments]

SwiftGo reads its configuration from the following environment variables:

    %-[1]*[2]s - Go tool binary
    %-[1]*[5]s   (default: '$GOROOT/bin/go', 'go')
    %-[1]*[3]s - Swift compiler binary
    %-[1]*[5]s   (default: 'xcrun swiftc', 'swiftc')
    %-[1]*[4]s - Swift compiler flags
    %-[1]*[5]s   (default: '-g -O')

If a package contains files with the extension '.swift.m' and the current build
context has GOOS=darwin, Swiftgo will compile them as Swift code instead of
Objective-C; otherwise, it will report an error.

SwiftGo finds all header files in the package with extension '.h' and makes
them available to Swift code as importable modules: for example, the directive

    import SwiftGo.HEADER

imports all declarations contained in 'HEADER.h' from the module directory,
if such a file is present. As a special case, the directive

    import SwiftGo.CgoExports

imports all declarations - if any - exported by Cgo files in the package.

A Swift to Objective-C bridging header will also be generated automatically
and included in the build, so that all Objective-C code may refer to Swift
classes marked with @objc/@objcMembers attributes.

Finally, SwiftGo scans each Go package for additional C flags specified
by '#cgo CFLAGS' directives and forwards them to the Swift compiler as follows:

    - if the '-mmacosx-version-min=<VERSION>' flag is present, it is used to
      select a target for the Swift compiler;
    - every additional header/framework search path is passed on as a
      module/framework search path.
`

const shortUsage = `
The Go build tool has been invoked through swiftgo, a wrapper that adds support
for embedded Swift code when targeting darwin systems.

For more about embedding Swift code, run '%s help swift'.
`

func main() {
	os.Exit(mainExitCode())
}

func mainExitCode() int {
	// NOTE: we never call os.Exit here to ensure deferred cleanup calls are executed

	// if no arguments were provided, add a fake first argument to make things simpler below
	if len(os.Args) < 1 {
		if path, err := os.Executable(); err == nil {
			os.Args = []string{path}
		} else {
			os.Args = []string{"swiftgo"}
		}
	}

	// default to help command
	// a short help message is displayed either when no command is given,
	// or when the help command is invoked with no topic, or when the topic is 'build'
	cmd := "help"
	shortHelp := true

	if len(os.Args) > 1 {
		cmd = os.Args[1]
		shortHelp = false
	}

	// if the help command has been invoked, we might want to display our own help messages
	if cmd == "help" {
		// prepend the help command to the topic to match the behavior of the Go tool
		// (i.e. no topic is not the same as empty topic)
		topic := strings.Join(os.Args[1:], " ")
		switch topic {
		case "help", "help build":
			shortHelp = true
		case "help swift":
			keyWidth := max(len(tools.GoOverrideKey), len(tools.SwiftcOverrideKey), len(tools.SwiftcFlagsKey))
			fmt.Printf(usage, keyWidth, tools.GoOverrideKey, tools.SwiftcOverrideKey, tools.SwiftcFlagsKey, "")
			return 0
		}
	}

	goTool, err := tools.LocateGoTool()
	if exitError := (*tools.ExitError)(nil); errors.As(err, &exitError) {
		return exitError.ExitCode
	} else if err != nil {
		fmt.Fprintln(os.Stderr, "swiftgo:", err)
		return 2
	}

	swiftgocc, err := tools.LocateSwiftGoCC()
	if err != nil {
		fmt.Fprintf(os.Stderr, "swiftgo: %v; reinstalling swiftgo might solve the issue\n", err)
		return 2
	}

	// swiftgocc's revision must be the same as our own
	revision := version.GetRevision("")
	if swiftgocc.Revision == "" || revision == "" || swiftgocc.Revision != revision {
		fmt.Fprintf(os.Stderr, "swiftgo: revision mismatch between driver binary and %v; reinstalling swiftgo might solve the issue\n", swiftgocc.Desc)
		return 2
	}

	// create a temporary build dir and schedule a cleanup operation
	// the prefix 'swiftgo-build' has been chosen by analogy with the Go tool's 'go-build'
	buildDir, err := os.MkdirTemp("", "swiftgo-build")
	if err != nil {
		fmt.Fprintln(os.Stderr, "swiftgo: could not create build directory: ", err)
		return 2
	}
	defer os.RemoveAll(buildDir)

	// setup the environment variables that swiftgocc expects
	err = os.Setenv(tools.SwiftGoCCBuildDirKey, buildDir)
	if err == nil {
		err = os.Setenv(tools.CCOverrideKey, goTool.Env["CC"])
	}
	if err == nil {
		err = os.Setenv("CC", "swiftgocc")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "swiftgo: environment configuration failed: ", err)
		return 2
	}

	// forward invocation to the Go tool
	exitCode, err := goTool.Run(os.Args[1:]...)
	if err != nil {
		fmt.Fprintln(os.Stderr, "swiftgo:", err)
		return 2
	}

	// if appropriate, append a short help message to the output of the Go tool
	if shortHelp {
		if exitCode == 0 {
			fmt.Fprintf(os.Stdout, shortUsage, os.Args[0])
		} else {
			fmt.Fprintf(os.Stderr, shortUsage, os.Args[0])
		}
	}

	// forward the exit code of the Go tool
	return exitCode
}
