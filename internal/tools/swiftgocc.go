// Copyright (c) 2024 Fabio Massaioli
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fbbdev/swiftgo/internal/version"
)

// SwiftGoCCBuildDirKey is the environment variable that is used internally
// to pass the global temporary build directory.
const SwiftGoCCBuildDirKey = "__SWIFTGO_PRIVATE_BUILDDIR"

// SwiftGoCC holds the path, arguments and configuration of our internal C compiler wrapper.
type SwiftGoCC struct {
	Tool

	// Revision holds the VCS revision from which the binary was compiled.
	Revision string
}

// LocateSwiftGoCC looks up the binary of our internal C compiler wrapper
// according to the following heuristics:
//
//   - first, LocateSwiftGoCC looks for a binary named 'swiftgocc'
//     in the same directory as the currently running binary;
//   - if the GOPATH variable is set either in the program environment
//     or in the Go environment configuration,
//     LocateSwiftGoCC looks up for a binary at '$GOPATH/bin/swiftgocc';
//   - otherwise, LocateSwiftGoCC looks for a binary named 'swiftgocc'
//     in the directories named by the PATH environment variable.
//
// When the binary is found, the tool.Revision field is set appropriately.
func LocateSwiftGoCC() (tool *SwiftGoCC, err error) {
	tool = &SwiftGoCC{
		Tool: Tool{
			Hint: "swiftgocc",
		},
	}

	const descFmt = `internal C compiler wrapper "%v"`

	if path := queryExecutableDir("swiftgocc"); path != "" {
		tool.Hint = path
		tool.Path = path
	} else {
		root := os.Getenv("GOPATH")
		if root == "" {
			root = queryEnvFile("GOPATH")
		}

		if root != "" {
			tool.Hint = filepath.Join(root, "bin", "swiftgocc")
		}
	}

	if tool.Desc == "" {
		tool.Desc = fmt.Sprintf(descFmt, tool.Hint)
	}

	if err = tool.Locate(); err != nil {
		return
	}

	tool.Revision = version.GetRevision(tool.Path)

	return
}
