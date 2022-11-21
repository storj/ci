// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

// check-cross-compile checks whether whether the program can be easily cross-compiled.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	compiler := strings.Join([]string{"go"}, ",")
	platform := strings.Join([]string{
		"linux/amd64",
		"linux/386",
		"linux/arm64",
		"linux/arm",
		"windows/amd64",
		"windows/386",
		"windows/arm64",
		"darwin/amd64",
		"darwin/arm64",
	}, ",")
	parallel := 4

	tags := ""

	flag.StringVar(&tags, "tags", tags, "tags for building")
	flag.StringVar(&compiler, "compiler", compiler, "comma separated list of compilers to test")
	flag.StringVar(&platform, "platform", platform, "comma separated list of platforms to test")
	flag.IntVar(&parallel, "parallel", parallel, "concurrent compilations")

	flag.Parse()

	packages := flag.Args()

	compilers := csv(compiler)
	platforms := csv(platform)

	results := make([]chan result, len(compilers)*len(platforms))

	lim := newLimiter(parallel)
	for i, compiler := range compilers {
		i, compiler := i, compiler
		for k, platform := range platforms {
			k, platform := k, platform
			ri := i*len(platforms) + k
			results[ri] = make(chan result, 1)
			lim.Go(func() {
				results[ri] <- tryCompile(tags, compiler, platform, packages)
			})
		}
	}
	lim.Wait()

	exit := 0
	for _, r := range results {
		r := <-r
		switch {
		case r.skip:
			fmt.Println("#", r.compiler, r.platform, "SKIPPED", r.err)
			fmt.Println(r.output)
			fmt.Println()

		case r.err == nil:
			fmt.Println("#", r.compiler, r.platform, "SUCCESS")

		default:
			fmt.Println("#", r.compiler, r.platform, "FAILED", r.err)
			fmt.Println(r.output)
			fmt.Println()
			exit = 1
		}
	}
	os.Exit(exit)
}

func csv(vs string) []string {
	xs := []string{}
	for _, v := range strings.Split(vs, ",") {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		xs = append(xs, v)
	}
	return xs
}

type result struct {
	compiler string
	platform string
	output   string
	skip     bool
	err      error
}

func tryCompile(tags, compiler, platform string, packages []string) result {
	r := result{compiler: compiler, platform: platform}

	if compiler[0] == '?' {
		_, err := exec.LookPath(compiler)
		if err != nil {
			r.skip = true
			r.output = "compiler missing"
			return r
		}
		compiler = compiler[1:]
	}

	goos, goarch, found := strings.Cut(platform, "/")
	if !found {
		panic("invalid platform")
	}

	args := []string{"build"}
	if tags != "" {
		args = append(args, "-tags", tags)
	}
	args = append(args, "errors") // add a non-binary package to prevent creating binaries
	args = append(args, packages...)

	cmd := exec.Command(compiler, args...)
	cmd.Env = append(os.Environ(),
		"GOOS="+goos,
		"GOARCH="+goarch,
	)

	data, err := cmd.CombinedOutput()
	r.output = strings.TrimSpace(string(data))
	if strings.Contains(r.output, "unsupported GOOS/GOARCH") {
		r.skip = true
	}
	r.err = err

	return r
}

type limiter struct{ limit chan struct{} }

func newLimiter(n int) *limiter { return &limiter{limit: make(chan struct{}, n)} }

func (lim *limiter) Go(fn func()) {
	lim.limit <- struct{}{}
	go func() {
		defer func() { <-lim.limit }()
		fn()
	}()
}

func (lim *limiter) Wait() {
	for i := 0; i < cap(lim.limit); i++ {
		lim.limit <- struct{}{}
	}
}
