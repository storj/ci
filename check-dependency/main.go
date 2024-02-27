// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Strings implements a semicolon delimited flag.
type Strings []string

// String returns the semi-colon delimited flag.
func (ss *Strings) String() string {
	return strings.Join(*ss, ";")
}

// Set sets the value.
func (ss *Strings) Set(v string) error {
	for _, v := range strings.Split(v, ";") {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		*ss = append(*ss, v)
	}
	return nil
}

var verbose bool

func main() {
	var ignore Strings
	var check Strings
	var includeTests bool

	flag.BoolVar(&verbose, "verbose", false, "print debug information")
	flag.BoolVar(&includeTests, "include-tests", false, "also check test packages")

	flag.Var(&ignore, "ignore", "ignore packages matching regular expression when listing")
	flag.Var(&check, "check", "succeeds when contains a package matching regular expression")

	flag.Parse()

	pkgNames := flag.Args()
	if len(pkgNames) == 0 {
		pkgNames = []string{"."}
	}

	roots, err := packages.Load(&packages.Config{
		Mode:  packages.NeedName | packages.NeedImports | packages.NeedDeps,
		Tests: includeTests,
	}, pkgNames...)
	if err != nil {
		panic(err)
	}

	var rxIgnore []*regexp.Regexp
	for _, rxstr := range ignore {
		rx, err := regexp.Compile(rxstr)
		if err != nil {
			panic(err)
		}
		rxIgnore = append(rxIgnore, rx)
	}

	var rxCheck []*regexp.Regexp
	for _, rxstr := range check {
		rx, err := regexp.Compile(rxstr)
		if err != nil {
			panic(err)
		}
		rxCheck = append(rxCheck, rx)
	}

	if verbose {
		fmt.Fprintln(os.Stderr, "loaded roots:", packagesToStrings(roots))
		fmt.Fprintln(os.Stderr, "ignore-rx:", ignore)
		fmt.Fprintln(os.Stderr, "check-rx:", check)
	}

	var exitCode int
	for _, root := range roots {
		if verbose {
			fmt.Fprintln(os.Stderr, "# ", root.PkgPath)
		}
		if target := matchesOne(root.PkgPath, rxIgnore); target != "" {
			if verbose {
				fmt.Fprintf(os.Stderr, "    skipping because it matched filter %q\n", target)
			}
			continue
		}

		if target := findPath(root, rxCheck); target != "" {
			fmt.Fprintln(os.Stderr, target)
			exitCode = 1
		}
	}

	os.Exit(exitCode)
}

func findPath(pkg *packages.Package, match []*regexp.Regexp) string {
	checked := map[string]struct{}{}

	var find func(int, *packages.Package) string
	find = func(ident int, p *packages.Package) string {
		if verbose {
			fmt.Fprintf(os.Stderr, "   %*s > %s\n", 4*ident, "", p.PkgPath)
		}
		if matched := matchesOne(p.PkgPath, match); matched != "" {
			return matched
		}

		checked[p.PkgPath] = struct{}{}

		for _, c := range p.Imports {
			if _, ok := checked[c.PkgPath]; ok {
				continue
			}

			if target := find(ident+1, c); target != "" {
				return p.PkgPath + " => " + target
			}
		}

		return ""
	}

	return find(0, pkg)
}

func matchesOne(s string, rxs []*regexp.Regexp) string {
	for _, rx := range rxs {
		if rx.MatchString(s) {
			return rx.String()
		}
	}
	return ""
}

func packagesToStrings(pkgs []*packages.Package) (rs []string) {
	for _, pkg := range pkgs {
		rs = append(rs, pkg.PkgPath)
	}
	return rs
}
