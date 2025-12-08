// Copyright (C) 2023 Storj Labs, Inc.
// See LICENSE for copying information.

package analyzer

import (
	"flag"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer verifies whether errs package is properly used.
var Analyzer = &analysis.Analyzer{
	Name: "callsize",
	Doc:  "check method/function calls where large number of bytes are passed",
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	FactTypes: []analysis.Fact{},
}

var maxParams = flag.Int64("max-args", 64, "maximum allowed argument size in bytes")
var maxResults = flag.Int64("max-results", 256, "maximum allowed results size in bytes")

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return
		}

		argsSize := int64(0)
		if fn.Recv != nil {
			for _, f := range fn.Recv.List {
				argsSize += typeSize(pass, f.Type)
			}
		}
		if fn.Type != nil && fn.Type.Params != nil {
			for _, f := range fn.Type.Params.List {
				argsSize += typeSize(pass, f.Type)
			}
		}

		resultSize := int64(0)
		if fn.Type != nil && fn.Type.Results != nil {
			for _, f := range fn.Type.Results.List {
				resultSize += typeSize(pass, f.Type)
			}
		}

		if argsSize > *maxParams || resultSize > *maxResults {
			pass.ReportRangef(fn, "%s too large (args %d bytes, result %d bytes)", fn.Name, argsSize, resultSize)
		}
	})

	return nil, nil
}

func typeSize(pass *analysis.Pass, t ast.Expr) int64 {
	tv, ok := pass.TypesInfo.Types[t]
	if !ok {
		panic(t)
	}
	// TODO: calculate things based on generic type variants
	// For now, let's assume that every generic argument is 8 bytes.
	if _, isGeneric := tv.Type.(*types.TypeParam); isGeneric {
		return 8
	}

	// TODO: should we assume that arguments use up a single register instead even when they are 1 byte?
	return pass.TypesSizes.Sizeof(tv.Type)
}
