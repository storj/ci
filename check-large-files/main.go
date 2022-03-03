// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var ignoreFile = map[string]struct{}{
	"package-lock.json": {},
	"icon.go":           {},
}

// Size constants.
const (
	fileSizeLimit = 650 * KB
	KB            = 1 << 10
)

func main() {
	cmd := exec.Command("git", "ls-files")
	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "running \"git ls-files\" failed:\n")
		fmt.Fprintf(os.Stderr, "\t%v\n", err)
		os.Exit(1)
	}

	var failed int
	committedFiles := strings.Split(string(out), "\n")
	for _, path := range committedFiles {
		if path == "" {
			continue
		}

		base := filepath.Base(path)
		if _, ok := ignoreFile[base]; ok {
			continue
		}
		if strings.Contains(base, ".dbx.") {
			continue
		}

		info, err := os.Stat(path)
		if err != nil {
			failed++
			fmt.Fprintf(os.Stderr, "failed to stat %q: %v\n", path, err)
			continue
		}

		size := info.Size()
		if size > fileSizeLimit {
			failed++
			fmt.Fprintf(os.Stderr, "%v (%vKB)\n", path, size/KB)
		}
	}

	if failed > 0 {
		fmt.Fprintf(os.Stderr, "some files were over size limit %v\n", fileSizeLimit)
		os.Exit(1)
	}
}
