// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

// check-monkit finds problems with using monkit.
package main

import (
	"github.com/storj/ci/check-monkit/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(analyzer.Analyzer) }
