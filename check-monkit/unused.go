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

func run(pass *analysis.Pass) (any, error) {
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

		start := ancestor(stack, 1)
		stop := ancestor(stack, 2)
		owner := ancestor(stack, 3)

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

		// mon.Task()(&ctx)(&err) outside of a defer stops the task immediately,
		// recording a zero duration and never seeing the returned error.
		if _, ok := owner.(*ast.DeferStmt); !ok && owner != nil {
			pass.Reportf(call.Lparen, "monitoring not deferred, stops immediately")
		}

		return true
	})
	return nil, nil
}

// ancestor returns the n-th ancestor of the node on top of the stack,
// skipping parentheses. ancestor(stack, 0) is the node itself.
// It returns nil when the stack does not go back that far.
func ancestor(stack []ast.Node, n int) ast.Node {
	for i := len(stack) - 1; i >= 0; i-- {
		if _, ok := stack[i].(*ast.ParenExpr); ok {
			continue
		}
		if n == 0 {
			return stack[i]
		}
		n--
	}
	return nil
}
