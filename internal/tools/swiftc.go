// Copyright (c) 2024 Fabio Massaioli
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/fbbdev/swiftgo/internal/quoted"
)

// SwiftcOverrideKey is the environment variable that the end-user may set
// to override the location of the Swift compiler.
const SwiftcOverrideKey = "SWIFTGO_SWIFTC"

// SwiftcFlagsKey is the environment variable that the end-user may set
// to pass additional flags to the Swift compiler.
const SwiftcFlagsKey = "SWIFTGO_SWIFTFLAGS"

// SwiftCompiler holds the path, arguments and configuration of the Swift compiler.
type SwiftCompiler struct {
	Tool

	// Version holds the major, minor version and patch number of the compiler binary.
	Version struct {
		String string
		Major  int
		Minor  int
		Patch  int
	}

	// LinkerFlags holds the flags that must be passed on to the linker
	// when linking swift modules compiled with the current configuration.
	LinkerFlags []string
}

var swiftcVersionRegex = regexp.MustCompile(`\bversion (\d+)\.(\d+)(?:\.(\d+))?\b`)

// LocateSwiftCompiler looks up the binary of the Swift compiler
// according to the following heuristics:
//
//   - if the SWIFTGO_SWIFTC environment variable is set,
//     it is parsed as a command followed by a sequence of arguments
//     and used to look up the binary;
//   - otherwise, LocateSwiftCompiler looks for a binary named 'xcrun'
//     in the directories named by the PATH environment variable
//     and tries to obtain a path by running the command 'xcrun -f swiftc';
//   - finally, LocateSwiftCompiler looks for a binary named 'swiftc'
//     in the directories named by the PATH environment variable.
//
// When the binary is found, tool.LinkerFlags is populated
// with the appropriate linker flags for the current configuration,
// as specified by the flags parameter.
func LocateSwiftCompiler(flags []string) (tool *SwiftCompiler, err error) {
	tool = &SwiftCompiler{
		Tool: Tool{
			Hint: "swiftc",
		},

		LinkerFlags: []string{
			"-lobjc",
			"-Wl,-no_objc_category_merging",
		},
	}

	userFlags, err := quoted.Split(os.Getenv(SwiftcFlagsKey))
	if err != nil {
		err = fmt.Errorf(SwiftcFlagsKey+" environment variable could not be parsed: %w", err)
		return
	}

	const descFmt = `Swift compiler "%v"`

	if override := os.Getenv(SwiftcOverrideKey); override != "" {
		tool.Args, err = quoted.Split(override)
		if err != nil || len(tool.Args) < 1 {
			err = fmt.Errorf(SwiftcOverrideKey+" environment variable could not be parsed: %w", err)
			return
		}

		tool.Desc = fmt.Sprintf(descFmt, override)
		tool.Hint = tool.Args[0]
		tool.Args = tool.Args[1:]
	} else if xcrunPath := queryXCrun("swiftc"); xcrunPath != "" {
		tool.Hint = "xcrun swiftc"
		tool.Path = xcrunPath
	}

	if tool.Desc == "" {
		tool.Desc = fmt.Sprintf(descFmt, tool.Hint)
	}

	if err = tool.Locate(); err != nil {
		return
	}

	tool.Args = append(tool.Args, userFlags...)
	tool.Args = append(tool.Args, flags...)

	data, exitCode, err := tool.Output("-print-target-info")
	if err != nil {
		return
	} else if exitCode != 0 {
		err = fmt.Errorf("Swift target information could not be retrieved: %w", &ExitError{exitCode, &tool.Tool})
		return
	}

	var info targetInfo
	err = json.Unmarshal(data, &info)
	if err != nil {
		err = fmt.Errorf("Swift target information could not be retrieved: %v returned invalid or unsupported data", tool.Desc)
		return
	}

	tool.Version.String = info.CompilerVersion

	// try parsing all fields of the version string, then check for errors
	errUnsupportedVersion := fmt.Errorf("Swift version could not be retrieved: %v returned unsupported version string", tool.Desc)
	err = errUnsupportedVersion
	if match := swiftcVersionRegex.FindStringSubmatch(info.CompilerVersion); match != nil {
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

	for _, lib := range info.Target.CompatibilityLibraries {
		if path := info.FindLibrary(lib.LibraryName); path != "" {
			tool.LinkerFlags = append(tool.LinkerFlags, "-force_load", path)
		} else {
			tool.LinkerFlags = append(tool.LinkerFlags, "-l"+lib.LibraryName)
		}
	}

	for _, path := range info.Paths.RuntimeLibraryPaths {
		tool.LinkerFlags = append(tool.LinkerFlags, "-L"+path)
	}

	if info.Target.LibrariesRequireRPath {
		for _, path := range info.Paths.RuntimeLibraryPaths {
			tool.LinkerFlags = append(tool.LinkerFlags, "-rpath", path)
		}
	}

	return
}

// targetInfo is used to unmarshal JSON data returned by the Swift compiler option '-print-target-info'.
type targetInfo struct {
	CompilerVersion string `json:"compilerVersion"`

	Target struct {
		CompatibilityLibraries []struct {
			LibraryName string `json:"libraryName"`
		} `json:"compatibilityLibraries"`

		LibrariesRequireRPath bool `json:"librariesRequireRPath"`
	} `json:"target"`

	Paths struct {
		SDKPath             string   `json:"sdkPath"`
		RuntimeLibraryPaths []string `json:"runtimeLibraryPaths"`
	} `json:"paths"`
}

// FindLibrary attempts to find an absolute path for a given library name.
func (info *targetInfo) FindLibrary(name string) string {
	if info.Paths.SDKPath == "" {
		output, err := exec.Command("xcrun", "--show-sdk-path").Output()
		if err == nil {
			info.Paths.SDKPath = string(output)
		}
	}

	if filepath.IsAbs(name) {
		if _, err := os.Stat(name); err == nil {
			return name
		}
	}

	name += ".a"

	for _, searchPath := range info.Paths.RuntimeLibraryPaths {
		libPath := filepath.Join(searchPath, name)
		if _, err := os.Stat(name); err == nil {
			return libPath
		}

		if info.Paths.SDKPath != "" {
			libPath = filepath.Join(info.Paths.SDKPath, libPath)
			if _, err := os.Stat(name); err == nil {
				return libPath
			}
		}
	}

	return ""
}
