// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func main() {
	cmd := exec.Command("git", "ls-files", ".", "--others", "--exclude-standard")

	var stdout bytes.Buffer
	cmd.Stdout = io.MultiWriter(&stdout, os.Stdout)
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("Checking left-over files failed.", err)
		os.Exit(1)
	}

	leftover := skipEmpty(strings.Split(stdout.String(), "\n"))

	if len(leftover) != 0 {
		fmt.Println("Files left-over after running tests.")
		os.Exit(1)
	}
}

var _ = ignorePrefix // we may need this in the future

func ignorePrefix(files []string, dir string) []string {
	result := files[:0]
	for _, file := range files {
		if file == "" {
			continue
		}
		if strings.HasPrefix(file, dir) {
			continue
		}
		result = append(result, file)
	}
	return result
}

func skipEmpty(files []string) []string {
	result := files[:0]
	for _, file := range files {
		if file == "" {
			continue
		}
		result = append(result, file)
	}
	return result
}
