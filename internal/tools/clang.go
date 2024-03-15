// Copyright (c) 2024 Fabio Massaioli
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package tools

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fbbdev/swiftgo/internal/quoted"
)

// CCOverrideKey is the environment variable that the end-user may set
// to override the location of the C compiler.
const CCOverrideKey = "__SWIFTGO_PRIVATE_CC"

// CCompiler holds the path, arguments and configuration of the C compiler.
type CCompiler struct {
	Tool

	// IsClang is true if the located compiler appears to be Clang.
	IsClang bool
}

const ccClangTest = `#ifdef __clang__
#else
#error "Unsupported"
#endif`

// LocateCCompiler looks up the binary of the C compiler
// according to the following heuristics:
//
//   - if the __SWIFTGO_PRIVATE_CC environment variable is set,
//     it is parsed as a command followed by a sequence of arguments
//     and used to look up the binary;
//   - otherwise, LocateCCompiler looks for a binary named 'xcrun'
//     in the directories named by the PATH environment variable
//     and tries to obtain a path by running the command 'xcrun -f clang';
//   - finally, LocateCCompiler looks for a binary named 'clang'
//     in the directories named by the PATH environment variable.
//
// When the binary is found, the tool.IsClang field is set appropriately.
func LocateCCompiler() (tool *CCompiler, err error) {
	tool = &CCompiler{
		Tool: Tool{
			Hint: "clang",
		},

		IsClang: false,
	}

	const descFmt = `C compiler "%v"`

	if override := os.Getenv(CCOverrideKey); override != "" {
		tool.Args, err = quoted.Split(override)
		if err != nil || len(tool.Args) < 1 {
			err = fmt.Errorf(CCOverrideKey+" environment variable could not be parsed: %w", err)
			return
		}

		tool.Desc = fmt.Sprintf(descFmt, override)
		tool.Hint = tool.Args[0]
		tool.Args = tool.Args[1:]
	} else if xcrunPath := queryXCrun("clang"); xcrunPath != "" {
		tool.Hint = "xcrun clang"
		tool.Path = xcrunPath
	}

	if tool.Desc == "" {
		tool.Desc = fmt.Sprintf(descFmt, tool.Hint)
	}

	if err = tool.Locate(); err != nil {
		return
	}

	cmd := tool.Command(Capture, "-E", "-x", "c", "-", "-o", "-")
	cmd.Stdin = strings.NewReader(ccClangTest)
	exitCode, err := tool.processInvocationError(cmd.Run())
	if err != nil {
		return
	} else if exitCode == 0 {
		tool.IsClang = true
	}

	return
}

// NakedCommand returns a Cmd struct instance configured to execute the compiler
// with the given arguments, but with no default arguments, i.e. the tool.Args
// field is ignored.
// Otherwise, it works exactly as tool.Command.
func (tool *CCompiler) NakedCommand(mode ForwardMode, args ...string) (cmd *exec.Cmd) {
	defaultArgs := tool.Args
	tool.Args = nil
	defer func() {
		tool.Args = defaultArgs
	}()

	return tool.Command(mode, args...)
}
