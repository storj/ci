// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
		if _, ierr := os.Stat(tempDir); ierr == nil {
			ierr := os.RemoveAll(tempDir)
			if ierr != nil {
				fmt.Fprintf(os.Stderr, "failed to delete temporary directory: %v\n", err)
			}
		}
	}()

	err = worktreeAdd(tempDir)
	if err != nil {
		return fmt.Errorf("failed to add worktree: %w", err)
	}

	defer func() {
		ierr := worktreeRemove(tempDir)
		if ierr != nil {
			fmt.Fprintf(os.Stderr, "failed to remove worktree: %v\n", ierr)
		}
	}()

	err = os.Chdir(tempDir)
	if err != nil {
		return fmt.Errorf("failed to change dir: %w", err)
	}

	err = tidy(tempDir)
	if err != nil {
		return fmt.Errorf("failed to tidy: %w", err)
	}

	err = diff()
	if err != nil {
		return fmt.Errorf("failed to diff: %w", err)
	}

	return err
}

func tidy(rootDir string) (err error) {
	modFiles, err := exec.Command("git", "ls-files", "go.mod", "**/go.mod").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "finding go.mod files failed: %v", err)
		return err
	}

	for _, modfile := range strings.Split(string(modFiles), "\n") {
		modfile = strings.TrimSpace(modfile)
		if modfile == "" {
			continue
		}

		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = filepath.Join(rootDir, filepath.Dir(modfile))
		cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr

		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("go mod tidy failed: %w", err)
		}
	}

	return nil
}

func worktreeAdd(dst string) (err error) {
	cmd := exec.Command("git", "worktree", "add", "--detach", dst)
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr

	return cmd.Run()
}

func worktreeRemove(dst string) (err error) {
	cmd := exec.Command("git", "worktree", "remove", "--force", dst)
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr

	return cmd.Run()
}

func diff() (err error) {
	cmd := exec.Command("git", "diff", "--exit-code", "go.mod", "go.sum", "**/go.mod", "**/go.sum")
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr

	return cmd.Run()
}
