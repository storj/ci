// Copyright (C) 2025 Storj Labs, Inc.
// See LICENSE for copying information.

// check-tx checks that transaction callbacks use the transaction parameter
// instead of accessing the database directly.
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	pathFlag = flag.String("path", ".", "Path to check (can be file or directory)")
	verbose  = flag.Bool("v", false, "Verbose output")
)

// dbMethods are database methods that should use tx instead of db in WithTx callbacks
var dbMethods = map[string]bool{
	"Exec":            true,
	"ExecContext":     true,
	"Query":           true,
	"QueryContext":    true,
	"QueryRow":        true,
	"QueryRowContext": true,
	"Prepare":         true,
	"PrepareContext":  true,
}

type visitor struct {
	fset     *token.FileSet
	issues   []Issue
	inWithTx bool
	txParam  string
	dbParam  string
}

// Issue represents a transaction usage violation found by the checker.
type Issue struct {
	File    string
	Line    int
	Column  int
	Message string
}

// Visit implements the ast.Visitor interface for the transaction usage checker.
func (v *visitor) Visit(node ast.Node) ast.Visitor {
	if n, ok := node.(*ast.CallExpr); ok {
		if v.isWithTxCall(n) {
			return v.visitWithTx(n)
		}
		if v.inWithTx && v.isDbMethodCall(n) {
			v.reportIssue(n, fmt.Sprintf("use transaction parameter '%s' instead of database '%s'", v.txParam, v.dbParam))
		}
	}
	return v
}

func (v *visitor) isWithTxCall(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if sel.Sel.Name == "WithTx" {
			// Handle txutil.WithTx and sqliteutil.WithTx
			if pkg, ok := sel.X.(*ast.Ident); ok {
				return pkg.Name == "txutil" || pkg.Name == "sqliteutil"
			}
			// Handle db.WithTx, adapter.WithTx, etc.
			return true
		}
	}
	return false
}

func (v *visitor) visitWithTx(call *ast.CallExpr) ast.Visitor {
	if len(call.Args) < 2 {
		return v
	}

	var callback ast.Expr
	var dbArg ast.Expr

	sel := call.Fun.(*ast.SelectorExpr)

	// Handle txutil.WithTx and sqliteutil.WithTx
	if pkg, ok := sel.X.(*ast.Ident); ok {
		if pkg.Name == "txutil" && len(call.Args) >= 4 {
			// txutil.WithTx(ctx, db, opts, fn)
			dbArg = call.Args[1]
			callback = call.Args[3]
		} else if pkg.Name == "sqliteutil" && len(call.Args) >= 3 {
			// sqliteutil.WithTx(ctx, db, fn)
			dbArg = call.Args[1]
			callback = call.Args[2]
		}
	} else {
		// Handle db.WithTx(ctx, opts, fn) or db.WithTx(ctx, fn) patterns
		if len(call.Args) >= 2 {
			// Find the database object from the selector (e.g., db in db.WithTx)
			dbArg = sel.X
			// Last argument is typically the callback
			callback = call.Args[len(call.Args)-1]
		}
	}

	if callback == nil {
		return v
	}

	// Extract database parameter name
	dbName := v.extractIdentifierName(dbArg)

	// Check if callback is a function literal
	if fn, ok := callback.(*ast.FuncLit); ok {
		if len(fn.Type.Params.List) >= 2 {
			// Extract tx parameter name (second parameter)
			txName := ""
			if len(fn.Type.Params.List[1].Names) > 0 {
				txName = fn.Type.Params.List[1].Names[0].Name
			}

			// Create new visitor for the callback
			childVisitor := &visitor{
				fset:     v.fset,
				issues:   v.issues,
				inWithTx: true,
				txParam:  txName,
				dbParam:  dbName,
			}

			ast.Walk(childVisitor, fn.Body)
			v.issues = childVisitor.issues
		}
	}

	return nil // Don't visit children of WithTx call
}

func (v *visitor) isDbMethodCall(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if dbMethods[sel.Sel.Name] {
			// Check if this method call starts with the database parameter
			return v.startsWithDbParam(sel.X)
		}
	}
	return false
}

func (v *visitor) startsWithDbParam(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		// Direct access: db.ExecContext
		return e.Name == v.dbParam
	case *ast.SelectorExpr:
		// Nested access: db.db.ExecContext, db.field.ExecContext, etc.
		return v.startsWithDbParam(e.X)
	case *ast.CallExpr:
		// Method call chains: db.GetDB().ExecContext
		if sel, ok := e.Fun.(*ast.SelectorExpr); ok {
			return v.startsWithDbParam(sel.X)
		}
	}
	return false
}

func (v *visitor) extractIdentifierName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		// Handle cases like db.ChooseAdapter(x)
		return v.extractIdentifierName(e.X)
	case *ast.CallExpr:
		// Handle method call chains
		if sel, ok := e.Fun.(*ast.SelectorExpr); ok {
			return v.extractIdentifierName(sel.X)
		}
	}
	return ""
}

func (v *visitor) reportIssue(node ast.Node, message string) {
	pos := v.fset.Position(node.Pos())
	v.issues = append(v.issues, Issue{
		File:    pos.Filename,
		Line:    pos.Line,
		Column:  pos.Column,
		Message: message,
	})
}

func checkFile(filename string) ([]Issue, error) {
	fset := token.NewFileSet()

	src, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	visitor := &visitor{fset: fset}
	ast.Walk(visitor, file)

	return visitor.issues, nil
}

func checkPath(path string) ([]Issue, error) {
	var allIssues []Issue

	err := filepath.Walk(path, func(filename string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(filename, ".go") {
			return nil
		}

		if strings.Contains(filename, "vendor/") {
			return nil
		}

		issues, err := checkFile(filename)
		if err != nil {
			if *verbose {
				log.Printf("Warning: could not parse %s: %v", filename, err)
			}
			return nil
		}

		allIssues = append(allIssues, issues...)
		return nil
	})

	return allIssues, err
}

func main() {
	flag.Parse()

	issues, err := checkPath(*pathFlag)
	if err != nil {
		log.Fatal(err)
	}

	if len(issues) == 0 {
		if *verbose {
			fmt.Println("No transaction usage issues found.")
		}
		return
	}

	for _, issue := range issues {
		fmt.Printf("%s:%d:%d: %s\n", issue.File, issue.Line, issue.Column, issue.Message)
	}

	os.Exit(1)
}
