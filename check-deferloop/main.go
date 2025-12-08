// Copyright (C) 2021 Storj Labs, Inc.
// See LICENSE for copying information.

// check-deferloop finds defer being used inside a for loop.
package main

import (
	"github.com/storj/ci/check-deferloop/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(analyzer.Analyzer) }
