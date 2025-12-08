// Copyright (C) 2021 Storj Labs, Inc.
// See LICENSE for copying information.

// Package analyzer finds defer being used inside a for loop.
package analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer implements unused task analysis pass.
var Analyzer = &analysis.Analyzer{
	Name:     "deferloop",
	Doc:      `check for defers inside a loop`,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	inspect.WithStack([]ast.Node{
		(*ast.DeferStmt)(nil),
	}, func(defern ast.Node, push bool, stack []ast.Node) (proceed bool) {
		if push {
			return true
		}

		// check if we are deferring and immediately returning
		if parent, ok := stack[len(stack)-2].(*ast.BlockStmt); ok {
			check := false
			for _, v := range parent.List {
				if v == defern {
					check = true
					continue
				}
				if check {
					if _, returned := v.(*ast.ReturnStmt); returned {
						return true
					}
				}
			}
		}

	check:
		for i := len(stack) - 1; i >= 0; i-- {
			n := stack[i]
			switch n.(type) {
			case *ast.ForStmt:
				pass.Reportf(defern.Pos(), "defer inside a loop")
				break check
			case *ast.ExprStmt, *ast.FuncLit:
				break check
			}
		}

		return true
	})
	return nil, nil
}
