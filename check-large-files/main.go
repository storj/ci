// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

var ignoreFolder = map[string]bool{
	".build":       true,
	".git":         true,
	"node_modules": true,
	"coverage":     true,
	"dist":         true,
	"dbx":          true,
}

// Size constants
const (
	KB = 1 << 10
)

func main() {
	const fileSizeLimit = 650 * KB

	var failed int

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return nil
		}
		if info.IsDir() && ignoreFolder[info.Name()] {
			return filepath.SkipDir
		}

		size := info.Size()
		if size > fileSizeLimit {
			failed++
			fmt.Printf("%v (%vKB)\n", path, size/KB)
		}

		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	if failed > 0 {
		fmt.Printf("some files were over size limit %v\n", fileSizeLimit)
		os.Exit(1)
	}
}
