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
			fmt.Fprintf(os.Stderr, "failed to remove worktree: %v\n", err)
			err = ierr
		}
	}()

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
	cmd := exec.Command("git", "diff", "--exit-code", "go.mod", "go.sum")
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr

	return cmd.Run()
}
