// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

func main() {
	err := check()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "Repo is not tidy! Please run `go mod tidy` and check in the changes.")
		os.Exit(1)
	}
}

func check() (err error) {
	tempDir, err := ioutil.TempDir("", "check-mod-tidy-")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	defer func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to delete temporary directory: %v\n", err)
		}
	}()

	err = copyDir(".", tempDir)
	if err != nil {
		return fmt.Errorf("failed to copy dir: %w", err)
	}

	err = os.Chdir(tempDir)
	if err != nil {
		return fmt.Errorf("failed to change dir: %w", err)
	}

	err = checkout()
	if err != nil {
		return fmt.Errorf("failed to checkout: %w", err)
	}

	err = tidy()
	if err != nil {
		return fmt.Errorf("failed to tidy: %w", err)
	}

	err = diff()
	if err != nil {
		return fmt.Errorf("failed to diff: %w", err)
	}

	return err
}

func checkout() (err error) {
	cmd := exec.Command("git", "checkout", "go.mod", "go.sum")
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr

	return cmd.Run()
}

func tidy() (err error) {
	for repeat := 2; repeat > 0; repeat-- {
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr

		err = cmd.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "go mod tidy failed, retrying: %v", err)
			continue
		}

		break
	}

	return err
}

func copyDir(src, dst string) error {
	cmd := exec.Command("cp", "-a", src, dst)
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr

	return cmd.Run()
}

func diff() (err error) {
	cmd := exec.Command("git", "diff", "--exit-code", "go.mod", "go.sum")
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr

	return cmd.Run()
}
