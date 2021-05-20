// Copyright (C) 2021 Storj Labs, Inc.
// See LICENSE for copying information.

// check-deferloop finds defer being used inside a for loop.
package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/singlechecker"
	"golang.org/x/tools/go/ast/inspector"
)

func main() { singlechecker.Main(Analyzer) }

// Analyzer implements unused task analysis pass.
var Analyzer = &analysis.Analyzer{
	Name:     "deferloop",
	Doc:      `check for defers inside a loop`,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	inspect.Nodes([]ast.Node{
		(*ast.ForStmt)(nil),
	}, func(n ast.Node, push bool) (proceed bool) {
		if push {
			return true
		}

		ast.Inspect(n, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.DeferStmt:
				pass.Reportf(n.Pos(), "defer inside a loop")
				return false
			case *ast.ExprStmt, *ast.FuncLit:
				return false
			}
			return true
		})

		return true
	})
	return nil, nil
}
