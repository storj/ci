// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	cmd := exec.Command("git", "ls-files", ".", "--others")

	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Checking left-over files failed.")
		fmt.Println(err)
		os.Exit(1)
	}

	leftover := strings.Split(strings.TrimSpace(string(out)), "\n")
	leftover = ignorePrefix(leftover, ".build")

	// there's no easy way to modify npm to use tmp folders
	leftover = ignorePrefix(leftover, "web/satellite/node_modules/")
	leftover = ignorePrefix(leftover, "web/satellite/coverage/")
	leftover = ignorePrefix(leftover, "web/satellite/dist/")
	leftover = ignorePrefix(leftover, "web/satellite/package-lock.json")

	leftover = ignorePrefix(leftover, "web/storagenode/node_modules/")
	leftover = ignorePrefix(leftover, "web/storagenode/coverage/")
	leftover = ignorePrefix(leftover, "web/storagenode/dist/")
	leftover = ignorePrefix(leftover, "web/storagenode/package-lock.json")

	if len(leftover) != 0 {
		fmt.Println("Files left-over after running tests:")
		for _, file := range leftover {
			fmt.Println(file)
		}
		os.Exit(1)
	}
}

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
