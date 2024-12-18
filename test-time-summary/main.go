// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

// test-time-summary finds the slowest test times from a tests .json output.
package main

import (
	"bytes"
	"cmp"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"text/tabwriter"

	"github.com/mfridman/tparse/parse"
)

func main() {
	count := flag.Int("n", 50, "number of slowest tests to print")
	flag.Parse()

	paths := []string{}
	for _, arg := range flag.Args() {
		paths = append(paths, must(filepath.Glob(arg))...)
	}

	timing := map[string]float64{}
	for _, file := range paths {
		data := must(os.ReadFile(file))
		summary := must(parse.Process(bytes.NewReader(data)))

		for _, pkg := range summary.Packages {
			for _, test := range pkg.Tests {
				if test.Status() != parse.ActionPass {
					continue
				}
				name := pkg.Summary.Package + "\t" + test.Name

				if old, ok := timing[name]; ok {
					timing[name] = min(old, test.Elapsed())
				} else {
					timing[name] = test.Elapsed()
				}
			}
		}
	}

	type Time struct {
		Name string
		Min  float64
	}
	var sorted []Time

	for test, min := range timing {
		sorted = append(sorted, Time{
			Name: test,
			Min:  min,
		})
	}

	slices.SortFunc(sorted, func(a, b Time) int {
		return cmp.Compare(b.Min, a.Min)
	})

	if len(sorted) > *count {
		sorted = sorted[:*count]
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)
	for _, test := range sorted {
		_, _ = fmt.Fprintf(w, "%v\t%v\n", test.Min, test.Name)
	}
	_ = w.Flush()
}

func must[T any](x T, err error) T {
	if err != nil {
		panic(err)
	}
	return x
}
