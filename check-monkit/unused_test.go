// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func Test(t *testing.T) {
	t.Skip("run manually since analysistest does not support modules")
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, "a")
}
