// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

// check-monkit finds problems with using monkit.
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
	Name:     "monkitunused",
	Doc:      `check for unfinished calls to mon.Task()(ctx)(&err)`,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}
	inspect.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) (proceed bool) {
		if push {
			return true
		}

		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true // not a call statement
		}

		typ := pass.TypesInfo.TypeOf(call)
		if typ == nil || typ.String() != `github.com/spacemonkeygo/monkit/v3.Task` {
			return true
		}

		start := stack[len(stack)-2]
		stop := stack[len(stack)-3]

		if _, ok = start.(*ast.CallExpr); !ok {
			if _, ok := start.(*ast.AssignStmt); ok {
				return true
			}
			if _, ok := start.(*ast.ValueSpec); ok {
				return true
			}
			pass.Reportf(call.Lparen, "monitoring not started")
			return true
		}
		if _, ok = stop.(*ast.CallExpr); !ok {
			if _, ok := stop.(*ast.AssignStmt); ok {
				return true
			}
			pass.Reportf(call.Lparen, "monitoring not stopped")
			return true
		}

		return true
	})
	return nil, nil
}
