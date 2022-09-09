// Copyright (C) 2021 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/zeebo/errs"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

var (
	ref = flag.String("ref", "HEAD", "which git ref to check")
)

var errDowngradesDetected = errors.New("downgrades detected")

func main() {
	flag.Parse()

	if err := run(*ref); errors.Is(err, errDowngradesDetected) {
		os.Exit(3)
	} else if err != nil {
		log.Fatalf("%+v", err)
	}
}

func run(ref string) error {
	olddir, err := ioutil.TempDir("", "check-downgrades-*")
	if err != nil {
		return errs.Wrap(err)
	}
	defer func() { err = errs.Combine(err, os.RemoveAll(olddir)) }()

	gitdirBytes, err := execute(".", "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return errs.Wrap(err)
	}
	// technically i think this doesn't work if there's a newline at the
	// end of the file path, but that is such an esoteric edge case that
	// i'm willing to ignore it. if this fails you due to that, do better.
	gitdir := strings.TrimRight(string(gitdirBytes), "\r\n")

	_, err = execute(gitdir, "git", "worktree", "add", "-f", olddir, ref+"^")
	if err != nil {
		return errs.Wrap(err)
	}
	defer func() {
		_, rerr := execute(gitdir, "git", "worktree", "remove", "-f", olddir)
		err = errs.Combine(err, rerr)
	}()

	allowlist, err := getAllowlist(gitdir, ref)
	if err != nil {
		return errs.Wrap(err)
	}

	var allProblems []string

	err = filepath.Walk(gitdir, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return errs.Wrap(err)
		}
		path, err = filepath.Rel(gitdir, path)
		if err != nil {
			return errs.Wrap(err)
		}
		if filepath.Base(path) != "go.mod" {
			return nil
		}

		fmt.Println("=== checking", path, "===")
		fmt.Println()

		problems, err := check(olddir, gitdir, path, allowlist)
		if err != nil {
			return errs.Wrap(err)
		}
		allProblems = append(allProblems, problems...)

		fmt.Println()

		return nil
	})
	if err != nil {
		return errs.Wrap(err)
	}

	if len(allProblems) > 0 {
		fmt.Println("=== PROBLEMS ===")
		fmt.Println()

		for _, problem := range allProblems {
			fmt.Println("\t", problem)
		}

		fmt.Println()
		return errDowngradesDetected
	}

	return nil
}

func check(olddir, newdir, modfile string, allowlist map[string]struct{}) (problems []string, err error) {
	oldModules, err := getModules(olddir, modfile)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	newModules, err := getModules(newdir, modfile)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	// get sorted list of paths
	var paths []string
	for path := range oldModules {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	tw := tabwriter.NewWriter(os.Stdout, 8, 4, 2, ' ', 0)
	defer func() { err = errs.Combine(err, tw.Flush()) }()

	var once sync.Once
	emit := func(key, kind, path, oldVersion, newVersion string) {
		once.Do(func() {
			fmt.Fprintf(tw, "key\tkind\tmodule\told version\tnew version\n")
			fmt.Fprintf(tw, "---\t----\t------\t-----------\t-----------\n")
		})
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", key, kind, path, oldVersion, newVersion)
	}

	// check for any unallowed downgrades
	for _, path := range paths {
		oldMod, oldOk := oldModules[path]
		newMod, newOk := newModules[path]

		if !oldOk && !newOk {
			continue // this should never happen
		} else if !oldOk && newOk {
			emit("+++", "add", path, "none", newMod.Version)
			continue
		} else if oldOk && !newOk {
			emit("---", "remove", path, oldMod.Version, "none")
			continue
		}

		switch semver.Compare(newMod.Version, oldMod.Version) {
		case 1: // upgrade
			emit("^^^", "upgrade", path, oldMod.Version, newMod.Version)
		case 0: // stable. don't print anything.
		case -1: // downgrade
			emit("vvv", "downgrade", path, oldMod.Version, newMod.Version)
			if _, ok := allowlist[path]; !ok {
				direct, err := directDependency(newdir, modfile, path)
				if err != nil {
					return nil, errs.Wrap(err)
				}
				if direct {
					problems = append(problems, fmt.Sprintf(
						"%s: %s was downgraded: if intended, add \"Downgrade: %s\" to commit message",
						modfile, path, path,
					))
				}
			}
		}
	}

	once.Do(func() { fmt.Println("No changes to module versions.") })

	return problems, errs.Wrap(tw.Flush())
}

func execute(dir, bin string, args ...string) ([]byte, error) {
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errs.New("%w: %s", err, out)
	}
	return out, nil
}

func foreachLine(data []byte, fn func(i int, line string) error) error {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for i := 0; scanner.Scan(); i++ {
		if err := fn(i, scanner.Text()); err != nil {
			return errs.Wrap(err)
		}
	}
	return errs.Wrap(scanner.Err())
}

func getModules(gitdir, modfile string) (map[string]module.Version, error) {
	// execute returns unclear errors, so pre-check if file exists
	if _, err := os.Stat(filepath.Join(gitdir, modfile)); os.IsNotExist(err) {
		return map[string]module.Version{}, nil
	}
	moddir, modfile := filepath.Split(filepath.Join(gitdir, modfile))
	data, err := execute(moddir, "go", "list", "-modfile", modfile, "-m", "all")
	if err != nil {
		return nil, errs.Wrap(err)
	}
	return parseModules(data)
}

func parseModules(data []byte) (map[string]module.Version, error) {
	out := make(map[string]module.Version)
	err := foreachLine(data, func(i int, line string) error {
		// skip the first entry
		if i == 0 {
			return nil
		}

		// strip off any replace suffix
		if index := strings.LastIndex(line, " => "); index != -1 {
			line = line[:index]
		}

		// find and split path and version by the last space
		index := strings.LastIndexByte(line, ' ')
		if index == -1 {
			return errs.New("invalid module line: %q", line)
		}
		path, version := line[:index], line[index+1:]

		// check for duplicates or invalid path/version
		if _, ok := out[path]; ok {
			return errs.New("duplicate module path: %q", line)
		} else if !semver.IsValid(version) {
			return errs.New("invalid module semver: %q", line)
		} else if err := module.CheckPath(path); err != nil {
			return errs.New("invalid module path: %q: %w", line, err)
		}

		out[path] = module.Version{
			Path:    path,
			Version: version,
		}
		return nil
	})
	return out, err
}

func getAllowlist(gitdir, ref string) (map[string]struct{}, error) {
	data, err := execute(gitdir, "git", "log", "-n", "1", "--format=%B", ref)
	if err != nil {
		return nil, errs.Wrap(err)
	}
	return parseAllowlist(data)
}

func parseAllowlist(data []byte) (map[string]struct{}, error) {
	out := make(map[string]struct{})
	err := foreachLine(data, func(_ int, line string) error {
		const header = "Downgrade: "
		if strings.HasPrefix(line, header) {
			out[line[len(header):]] = struct{}{}
		}
		return nil
	})
	return out, err
}

func directDependency(gitdir, modfile, path string) (bool, error) {
	// execute returns unclear errors, so pre-check if file exists
	if _, err := os.Stat(filepath.Join(gitdir, modfile)); os.IsNotExist(err) {
		return false, errs.New("module file missing")
	}
	moddir, modfile := filepath.Split(filepath.Join(gitdir, modfile))
	data, err := execute(moddir, "go", "mod", "why", "-modfile", modfile, "-m", path)
	if err != nil {
		return false, errs.Wrap(err)
	}
	return !bytes.Contains(data, []byte("main module does not need module")), nil
}
