package tools

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// loadEnvFile parses the Go environment configuration file
// and extracts the value associated to the given variable key, if present.
func queryEnvFile(key string) string {
	// current go env logic is to discard names that do not start with a capital latin letter
	if len(key) == 0 || key[0] < 'A' || 'Z' < key[0] {
		return ""
	}

	prefix := key + "="

	path, err := envFile()
	if path == "" || err != nil {
		return ""
	}

	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), prefix) {
			return scanner.Text()[len(prefix):]
		}
	}

	return ""
}

// envFile returns the name of the Go environment configuration file.
//
// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the GO_LICENSE file.
func envFile() (string, error) {
	if file := os.Getenv("GOENV"); file != "" {
		if file == "off" {
			return "", fmt.Errorf("GOENV=off")
		}
		return file, nil
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	if dir == "" {
		return "", fmt.Errorf("missing user-config dir")
	}

	return filepath.Join(dir, "go", "env"), nil
}

// queryXCrun invokes xcrun to obtain a path to the given tool.
func queryXCrun(tool string) string {
	output, err := exec.Command("xcrun", "-f", tool).Output()
	if err != nil {
		return ""
	}

	path, err := exec.LookPath(strings.TrimSpace(string(output)))
	if err != nil {
		return ""
	}

	return path
}

// queryExecutableDir looks for a tool relative to the directory
// where the current executable is stored.
func queryExecutableDir(tool string) string {
	self, err := os.Executable()
	if err != nil {
		return ""
	}

	path, err := exec.LookPath(filepath.Join(filepath.Dir(self), tool))
	if err != nil {
		return ""
	}

	return path
}
