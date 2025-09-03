// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func main() {
	cmd := exec.CommandContext(context.Background(), "git", "ls-files", ".", "--others", "--exclude-standard")

	var stdout bytes.Buffer
	cmd.Stdout = io.MultiWriter(&stdout, os.Stdout)
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("Checking left-over files failed.", err)
		os.Exit(1)
	}

	leftover := skipEmpty(strings.Split(stdout.String(), "\n"))

	// Jenkins sometimes creates @tmp folders, files that we cannot do anything about.
	leftover = discard(leftover, func(file string) bool {
		return strings.Contains(file, "@tmp")
	})

	if len(leftover) != 0 {
		fmt.Println("Files left-over after running tests.")
		os.Exit(1)
	}
}

func discard(files []string, fn func(file string) bool) []string {
	result := files[:0]
	for _, file := range files {
		if fn(file) {
			continue
		}
		result = append(result, file)
	}
	return result
}

func skipEmpty(files []string) []string {
	return discard(files, func(file string) bool { return file == "" })
}
