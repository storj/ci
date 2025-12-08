// Copyright (C) 2025 Storj Labs, Inc.
// See LICENSE for copying information.

// check-all runs multiple static analysis checks using multichecker.
package main

import (
	errsAnalyzer "github.com/storj/ci/check-errs/analyzer"
	monkitAnalyzer "github.com/storj/ci/check-monkit/analyzer"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(
		// callsizeAnalyzer.Analyzer,
		// deferloopAnalyzer.Analyzer,
		errsAnalyzer.Analyzer,
		monkitAnalyzer.Analyzer,
	)
}
