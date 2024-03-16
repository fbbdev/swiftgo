// Copyright (c) 2024 Fabio Massaioli
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/fbbdev/swiftgo/internal/tools"
)

type Config struct {
	Target struct {
		Arch string
		OS   string
	}

	Package   string
	InPackage bool

	BuildDir string

	CCompiler     *tools.CCompiler
	SwiftCompiler *tools.SwiftCompiler
}

func main() {
	// if no arguments were provided, add a fake first argument to make things simpler below
	if len(os.Args) < 1 {
		if path, err := os.Executable(); err == nil {
			os.Args = []string{path}
		} else {
			os.Args = []string{"swiftgo"}
		}
	}

	var config Config

	config.Target.Arch = os.Getenv("GOARCH")
	if config.Target.Arch == "" {
		config.Target.Arch = runtime.GOARCH
	}

	config.Target.OS = os.Getenv("GOOS")
	if config.Target.OS == "" {
		config.Target.OS = runtime.GOOS
	}

	config.Package, config.InPackage = os.LookupEnv("TOOLEXEC_IMPORTPATH")
	config.BuildDir = os.Getenv(tools.SwiftGoCCBuildDirKey)

	// private environment variables should always be present
	if _, haveCompiler := os.LookupEnv(tools.CCOverrideKey); config.BuildDir == "" || !haveCompiler {
		fmt.Fprintln(os.Stderr, "swiftgocc: environment variables are not configured correctly; ensure swiftgocc is only invoked through the swiftgo driver")
		os.Exit(1)
	}

	var err error
	config.CCompiler, err = tools.LocateCCompiler()
	if err != nil {
		fmt.Fprintln(os.Stderr, "swiftgo:", err)
		os.Exit(1)
	}

	// stub: forward everything to C compiler
	exitCode, err := config.CCompiler.Run(os.Args[1:]...)
	if err != nil {
		fmt.Fprintln(os.Stderr, "swiftgo:", err)
		os.Exit(1)
	}

	// forward C compiler exit code
	os.Exit(exitCode)
}
