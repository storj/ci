// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/singlechecker"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/types/typeutil"
)

func main() { singlechecker.Main(Analyzer) }

// Analyzer verifies whether errs package is properly used.
var Analyzer = &analysis.Analyzer{
	Name: "errs",
	Doc:  "check for proper usage of errs package",
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	FactTypes: []analysis.Fact{},
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)

		fn := typeutil.StaticCallee(pass.TypesInfo, call)
		if fn != nil {
			handleStaticCall(pass, call, fn)
			return // not a static call
		}
		if isErrsClassCast(pass, call) {
			handleErrsClassCast(pass, call)
			return
		}
	})

	return nil, nil
}

func isErrsClassCast(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	{ // check that the package name matches
		qualifier, ok := sel.X.(*ast.Ident)
		if !ok {
			return false
		}

		obj, ok := pass.TypesInfo.Uses[qualifier] // qualified identifier?
		if !ok {
			return false
		}

		pkgname, ok := obj.(*types.PkgName)
		if !ok {
			return false
		}

		if pkgname.Imported().Path() != "github.com/zeebo/errs" {
			return false
		}
	}

	return sel.Sel.Name == "Class"
}

func handleStaticCall(pass *analysis.Pass, call *ast.CallExpr, fn *types.Func) {
	switch fn.FullName() {
	case "github.com/zeebo/errs.Combine":
		if len(call.Args) == 0 {
			pass.Reportf(call.Lparen, "errs.Combine() can be simplified to nil")
		}
		if len(call.Args) == 1 && call.Ellipsis == token.NoPos {
			pass.Reportf(call.Lparen, "errs.Combine(x) can be simplified to x")
		}

	case "(*github.com/zeebo/errs.Class).New", "github.com/zeebo/errs.New":
		if len(call.Args) == 0 {
			return
		}
		// Disallow things like Error.New(err.Error())

		switch arg := call.Args[0].(type) {
		case *ast.BasicLit: // allow string constants
		case *ast.Ident: // allow string variables
		default:
			// allow "alpha" + "beta" + "gamma"
			if IsConcatString(arg) {
				return
			}

			pass.Reportf(call.Lparen, fn.FullName()+" with non-obvious format string")
		}
	}
}

func handleErrsClassCast(pass *analysis.Pass, call *ast.CallExpr) {
	if len(call.Args) == 0 {
		return
	}

	// Disallow things like errs.Class(fmt.Sprintf("xyz"))
	switch arg := call.Args[0].(type) {
	case *ast.BasicLit: // allow string constants
	default:
		// allow "alpha" + "beta" + "gamma"
		if IsConcatString(arg) {
			return
		}

		pass.Reportf(call.Lparen, "errs.Class(x), where x is not a constant")
	}
}

// IsConcatString returns whether arg is a basic string expression.
func IsConcatString(arg ast.Expr) bool {
	switch arg := arg.(type) {
	case *ast.BasicLit:
		return arg.Kind == token.STRING
	case *ast.BinaryExpr:
		return arg.Op == token.ADD && IsConcatString(arg.X) && IsConcatString(arg.Y)
	default:
		return false
	}
}
