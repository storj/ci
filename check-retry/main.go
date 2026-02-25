// Copyright (C) 2026 Storj Labs, Inc.
// See LICENSE for copying information.

// check-retry checks that variables modified inside retry callbacks are
// properly reset at the beginning of the callback body.
//
// When a function like WithRetry or ReadWriteTransaction retries the callback,
// variables from the outer scope retain values from previous attempts. This
// can cause inflated counters, duplicate entries, or stale results.
//
// # Usage
//
//	check-retry ./...
//
// # Bad
//
//	rows := []int{}
//	WithRetry(func() {
//	    rows = append(rows, 123)
//	})
//
// # Good
//
//	var rows []int
//	WithRetry(func() {
//	    rows = []int{}
//	    rows = append(rows, 123)
//	})
package main

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"
	"sync"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/singlechecker"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/types/typeutil"
)

func main() { singlechecker.Main(Analyzer) }

// Analyzer checks that outer variables modified inside retry callbacks are reset.
var Analyzer = &analysis.Analyzer{
	Name: "checkretry",
	Doc:  "check that variables modified inside retry callbacks are reset at the beginning of the callback",
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

var extraFuncs string

func init() {
	Analyzer.Flags.StringVar(&extraFuncs, "funcs", "", "comma-separated list of additional retry function names to check")
}

// defaultRetryFuncNames lists the built-in function/method names that indicate a retrying pattern.
var defaultRetryFuncNames = []string{
	"withRetries",
	"WithRetry",
	"WithTx",
	"ReadWriteTransaction",
	"ReadWriteTransactionWithOptions",
	"ReadTransaction",
}

var (
	mergeOnce      sync.Once
	retryFuncNames map[string]bool
)

// mergedRetryFuncNames returns the combined set of default and user-specified retry function names.
func mergedRetryFuncNames() map[string]bool {
	mergeOnce.Do(func() {
		retryFuncNames = make(map[string]bool, len(defaultRetryFuncNames))
		for _, name := range defaultRetryFuncNames {
			retryFuncNames[name] = true
		}
		if extraFuncs != "" {
			for _, name := range strings.Split(extraFuncs, ",") {
				name = strings.TrimSpace(name)
				if name != "" {
					retryFuncNames[name] = true
				}
			}
		}
	})
	return retryFuncNames
}

func run(pass *analysis.Pass) (any, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		if !isRetryCall(pass, call) {
			return
		}

		// Find the callback argument (a function literal) among all arguments.
		for _, arg := range call.Args {
			fn, ok := arg.(*ast.FuncLit)
			if !ok {
				continue
			}
			checkCallback(pass, fn)
			break
		}
	})

	return nil, nil
}

// isRetryCall checks whether call is a call to a known retry function.
func isRetryCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	names := mergedRetryFuncNames()

	// Check via type info for static callees.
	if fn := typeutil.StaticCallee(pass.TypesInfo, call); fn != nil {
		name := fn.Name()
		if names[name] {
			return true
		}
		// Also check by full name suffix for methods.
		full := fn.FullName()
		for n := range names {
			if strings.HasSuffix(full, "."+n) {
				return true
			}
		}
	}

	// Fallback: check the AST for unresolved or dynamic calls.
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		return names[fun.Name]
	case *ast.SelectorExpr:
		return names[fun.Sel.Name]
	}
	return false
}

// checkCallback inspects a retry callback for outer-scope variables that are
// modified but not reset at the top of the callback body.
func checkCallback(pass *analysis.Pass, fn *ast.FuncLit) {
	if fn.Body == nil || len(fn.Body.List) == 0 {
		return
	}

	// Collect the set of variables declared as callback parameters.
	paramVars := collectParamVars(pass, fn)

	// Find all outer-scope variables that are modified inside the callback.
	modified := findModifiedOuterVars(pass, fn, paramVars)
	if len(modified) == 0 {
		return
	}

	// Find variables that are reset at the top of the callback.
	reset := findResetVars(pass, fn)

	// Report any modified-but-not-reset variables.
	for obj, positions := range modified {
		if reset[obj] {
			continue
		}
		for _, pos := range positions {
			if hasNolintDirective(pass, pos) {
				continue
			}
			pass.Reportf(pos, "variable %q is modified inside retry callback but not reset at the top of the callback", obj.Name())
		}
	}
}

// collectParamVars returns the set of variables declared as parameters of fn.
func collectParamVars(pass *analysis.Pass, fn *ast.FuncLit) map[types.Object]bool {
	params := map[types.Object]bool{}
	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			for _, name := range field.Names {
				if obj := pass.TypesInfo.ObjectOf(name); obj != nil {
					params[obj] = true
				}
			}
		}
	}
	if fn.Type.Results != nil {
		for _, field := range fn.Type.Results.List {
			for _, name := range field.Names {
				if obj := pass.TypesInfo.ObjectOf(name); obj != nil {
					params[obj] = true
				}
			}
		}
	}
	return params
}

// findModifiedOuterVars walks the callback body and returns outer-scope variables
// that are assigned or accumulated. It maps each variable object to the positions
// where modifications occur.
func findModifiedOuterVars(pass *analysis.Pass, fn *ast.FuncLit, paramVars map[types.Object]bool) map[types.Object][]token.Pos {
	modified := map[types.Object][]token.Pos{}

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.FuncLit:
			// Skip the root callback node itself (we're already inside it).
			// Do descend into nested function literals — closures can
			// still modify variables from the retry callback's outer scope,
			// and those modifications need to be reset too.
			return true
		case *ast.AssignStmt:
			for _, lhs := range n.Lhs {
				if obj := outerVar(pass, fn, lhs, paramVars); obj != nil {
					modified[obj] = append(modified[obj], lhs.Pos())
				}
			}
		case *ast.IncDecStmt:
			if obj := outerVar(pass, fn, n.X, paramVars); obj != nil {
				modified[obj] = append(modified[obj], n.X.Pos())
			}
		case *ast.RangeStmt:
			// "for outerKey, outerVal = range items {}" assigns to outer vars.
			if n.Tok == token.ASSIGN {
				if n.Key != nil {
					if obj := outerVar(pass, fn, n.Key, paramVars); obj != nil {
						modified[obj] = append(modified[obj], n.Key.Pos())
					}
				}
				if n.Value != nil {
					if obj := outerVar(pass, fn, n.Value, paramVars); obj != nil {
						modified[obj] = append(modified[obj], n.Value.Pos())
					}
				}
			}
		case *ast.ExprStmt:
			// "delete(outerMap, key)" mutates the map.
			if call, ok := n.X.(*ast.CallExpr); ok {
				if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "delete" && len(call.Args) == 2 {
					if obj := outerVar(pass, fn, call.Args[0], paramVars); obj != nil {
						modified[obj] = append(modified[obj], call.Args[0].Pos())
					}
				}
			}
		}
		return true
	})

	return modified
}

// outerVar checks if expr refers to a variable from an outer scope (not declared
// inside the callback and not a parameter). Returns the variable's types.Object
// or nil.
func outerVar(pass *analysis.Pass, fn *ast.FuncLit, expr ast.Expr, paramVars map[types.Object]bool) types.Object {
	// Unwrap index/selector expressions: m[k] = v → check m, s.field = v → check s.
	expr = unwrapIndexAndSelector(expr)

	ident, ok := expr.(*ast.Ident)
	if !ok {
		return nil
	}

	obj := pass.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return nil
	}

	// Skip parameters.
	if paramVars[obj] {
		return nil
	}

	// Skip blank identifiers.
	if ident.Name == "_" {
		return nil
	}

	v, ok := obj.(*types.Var)
	if !ok {
		return nil
	}

	// Check if the variable was declared inside the callback body.
	if fn.Body.Pos() <= v.Pos() && v.Pos() < fn.Body.End() {
		return nil
	}

	return obj
}

// findResetVars scans the top-level statements of the callback body for
// plain "=" assignments that reset outer variables.
//
// It skips over local variable declarations (var/short-var) and stops at the
// first statement that is neither a declaration nor a plain "=" assignment,
// since resets should happen before other logic.
//
// Self-referential assignments (where the variable appears on the RHS) are NOT
// considered resets, e.g. "rows = append(rows, 123)".
func findResetVars(pass *analysis.Pass, fn *ast.FuncLit) map[types.Object]bool {
	reset := map[types.Object]bool{}

	for _, stmt := range fn.Body.List {
		switch s := stmt.(type) {
		case *ast.DeclStmt:
			// Local variable declarations (var err error) are fine, skip over them.
			continue
		case *ast.AssignStmt:
			if s.Tok == token.DEFINE {
				// Short variable declarations (:=) are local, skip over them.
				continue
			}
			if s.Tok != token.ASSIGN {
				// Compound assignments (+=, etc.) are not resets, stop scanning.
				return reset
			}

			for _, lhs := range s.Lhs {
				ident, ok := lhs.(*ast.Ident)
				if !ok {
					continue
				}
				obj := pass.TypesInfo.ObjectOf(ident)
				if obj == nil {
					continue
				}
				// Check that the RHS does not reference the same variable.
				// For multi-value returns (e.g. "info, err = fn()"), len(Rhs)==1,
				// so always check the single RHS expression.
				if selfRef(pass, s, obj) {
					continue
				}
				reset[obj] = true
			}
		case *ast.ExprStmt:
			// Handle "clear(m)" as a reset for m.
			if call, ok := s.X.(*ast.CallExpr); ok {
				if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "clear" && len(call.Args) == 1 {
					if arg, ok := call.Args[0].(*ast.Ident); ok {
						if obj := pass.TypesInfo.ObjectOf(arg); obj != nil {
							reset[obj] = true
						}
					}
				}
			}
			continue
		case *ast.IfStmt:
			// Handle "if x, err = f(); err != nil { ... }" — the init
			// assignment counts as a reset for the assigned variables.
			assign, ok := s.Init.(*ast.AssignStmt)
			if !ok || assign.Tok != token.ASSIGN {
				return reset
			}
			for _, lhs := range assign.Lhs {
				ident, ok := lhs.(*ast.Ident)
				if !ok {
					continue
				}
				obj := pass.TypesInfo.ObjectOf(ident)
				if obj == nil {
					continue
				}
				if selfRef(pass, assign, obj) {
					continue
				}
				reset[obj] = true
			}
			return reset
		default:
			return reset
		}
	}

	return reset
}

// unwrapIndexAndSelector strips *ast.IndexExpr, *ast.SelectorExpr, and
// *ast.StarExpr wrappers so that m[k] yields m, s.field yields s,
// and *ptr yields ptr.
func unwrapIndexAndSelector(expr ast.Expr) ast.Expr {
	for {
		switch e := expr.(type) {
		case *ast.IndexExpr:
			expr = e.X
		case *ast.SelectorExpr:
			expr = e.X
		case *ast.StarExpr:
			expr = e.X
		default:
			return expr
		}
	}
}

// hasNolintDirective checks if there's a "//check-retry:ignore" comment on the
// same line or the line immediately before pos.
func hasNolintDirective(pass *analysis.Pass, pos token.Pos) bool {
	position := pass.Fset.Position(pos)
	targetLine := position.Line

	var file *ast.File
	for _, f := range pass.Files {
		if pass.Fset.Position(f.Pos()).Filename == position.Filename {
			file = f
			break
		}
	}
	if file == nil {
		return false
	}

	for _, cg := range file.Comments {
		for _, c := range cg.List {
			commentLine := pass.Fset.Position(c.Pos()).Line
			if commentLine == targetLine || commentLine == targetLine-1 {
				text := strings.TrimSpace(c.Text)
				if text == "//check-retry:ignore" || strings.HasPrefix(text, "//check-retry:ignore ") {
					return true
				}
			}
		}
	}

	return false
}

// selfRef reports whether the RHS of assign references obj.
func selfRef(pass *analysis.Pass, assign *ast.AssignStmt, obj types.Object) bool {
	for _, rhs := range assign.Rhs {
		if exprReferences(pass, rhs, obj) {
			return true
		}
	}
	return false
}

// exprReferences reports whether expr contains a reference to obj.
// It does not descend into nested function literals, because a reference
// inside a closure argument (e.g. CollectRow(..., func() { err = ... }))
// is not a self-reference that would cause accumulation.
func exprReferences(pass *analysis.Pass, expr ast.Expr, obj types.Object) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if found {
			return false
		}
		if _, ok := n.(*ast.FuncLit); ok {
			return false
		}
		if ident, ok := n.(*ast.Ident); ok {
			if pass.TypesInfo.ObjectOf(ident) == obj {
				found = true
				return false
			}
		}
		return true
	})
	return found
}
