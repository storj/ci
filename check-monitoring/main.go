// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io/ioutil"
	"log"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

var (
	monkitPaths = map[string]struct{}{
		"gopkg.in/spacemonkeygo/monkit.v2":   {},
		"github.com/spacemonkeygo/monkit/v3": {},
	}
)

func main() {
	output := flag.String("out", "", "output lock file")
	flag.Parse()

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedCompiledGoFiles |
			packages.NeedSyntax |
			packages.NeedName |
			packages.NeedTypes |
			packages.NeedTypesInfo,
	}, flag.Args()...)
	if err != nil {
		log.Fatalf("error while loading packages: %s", err)
	}

	var lockedFnNames []string
	for _, pkg := range pkgs {
		lockedFnNames = append(lockedFnNames, findLockedFnNames(pkg)...)
	}
	sortedNames := sortAndUnique(lockedFnNames)

	outputStr := strings.Join(sortedNames, "\n")
	if *output != "" {
		if err := ioutil.WriteFile(*output, []byte(outputStr+"\n"), 0644); err != nil {
			log.Fatalf("error while writing to file %q: %s", *output, err)
		}
	} else {
		fmt.Println(outputStr)
	}
}

func findLockedFnNames(pkg *packages.Package) []string {
	type posWithScope struct {
		Scope string
		Pos   token.Pos
	}
	type fnWithScope struct {
		Scope string
		Func  *ast.FuncDecl
	}
	var (
		lockedTasksPos []posWithScope
		lockedTaskFns  []fnWithScope
		lockedFnInfos  []string
	)

	// Collect locked comments and what line they are on.
	for _, file := range pkg.Syntax {
		lockedLines := make(map[int]struct{})
		for _, group := range file.Comments {
			for _, comment := range group.List {
				if comment.Text == "//locked" || comment.Text == "//mon:locked" {
					commentLine := pkg.Fset.Position(comment.Pos()).Line
					lockedLines[commentLine] = struct{}{}
				}
			}
		}
		if len(lockedLines) == 0 {
			continue
		}

		// Find calls to monkit functions we're interested in that are on the
		// same line as a "locked" comment and keep track of their position.
		// NB: always return true to walk entire node tree.
		ast.Inspect(file, func(node ast.Node) bool {
			if node == nil {
				return true
			}
			if !isMonkitCall(pkg, node) {
				return true
			}

			// Ensure call line matches a "locked" comment line.
			callLine := pkg.Fset.Position(node.End()).Line
			if _, ok := lockedLines[callLine]; !ok {
				return true
			}

			// We are already checking to ensure that these type assertions are valid in `isMonkitCall`.
			sel := node.(*ast.CallExpr).Fun.(*ast.SelectorExpr)

			scope := extractScopeName(pkg, sel)

			// Track `mon.Task` calls.
			if sel.Sel.Name == "Task" {
				lockedTasksPos = append(lockedTasksPos, posWithScope{Scope: scope, Pos: node.End()})
				return true
			}

			// Track other monkit calls that have one or more argument (e.g. monkit.FloatVal, etc.)
			// and transform the first arg to the representative string.
			if len(node.(*ast.CallExpr).Args) < 1 {
				return true
			}
			argLiteral, ok := node.(*ast.CallExpr).Args[0].(*ast.BasicLit)
			if !ok {
				return true
			}
			if argLiteral.Kind == token.STRING {
				lockedFnInfo := scope + "." + argLiteral.Value + " " + sel.Sel.Name
				lockedFnInfos = append(lockedFnInfos, lockedFnInfo)
			}
			return true
		})

		// Track all function declarations containing locked `mon.Task` calls.
		ast.Inspect(file, func(node ast.Node) bool {
			fn, ok := node.(*ast.FuncDecl)
			if !ok {
				return true
			}
			for _, locked := range lockedTasksPos {
				if fn.Pos() < locked.Pos && locked.Pos < fn.End() {
					lockedTaskFns = append(lockedTaskFns, fnWithScope{Scope: locked.Scope, Func: fn})
				}
			}
			return true
		})

	}

	// Transform the ast.FuncDecls containing locked `mon.Task` calls to representative string.
	for _, sfn := range lockedTaskFns {
		fn := sfn.Func
		object := pkg.TypesInfo.Defs[fn.Name]

		var receiver string
		if fn.Recv != nil {
			typ := fn.Recv.List[0].Type
			if star, ok := typ.(*ast.StarExpr); ok {
				receiver = ".*"
				typ = star.X
			} else {
				receiver = "."
			}
			recvObj := pkg.TypesInfo.Uses[typ.(*ast.Ident)]
			receiver += recvObj.Name()
		}

		lockedFnInfo := sfn.Scope + receiver + "." + object.Name() + " Task"
		lockedFnInfos = append(lockedFnInfos, lockedFnInfo)

	}
	return lockedFnInfos
}

// isMonkitCall returns whether the node is a call to a function in the monkit package.
func isMonkitCall(pkg *packages.Package, in ast.Node) bool {
	call, ok := in.(*ast.CallExpr)
	if !ok {
		return false
	}
	fun, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := fun.X.(*ast.Ident)
	if !ok {
		return false
	}

	tvar, ok := pkg.TypesInfo.Uses[ident].(*types.Var)
	if !ok {
		return false
	}
	tptr, ok := tvar.Type().(*types.Pointer)
	if !ok {
		return false
	}
	named, ok := tptr.Elem().(*types.Named)
	if !ok {
		return false
	}

	importPath := named.Obj().Pkg().Path()
	_, match := monkitPaths[importPath]
	return match
}

func extractScopeName(pkg *packages.Package, sel *ast.SelectorExpr) (scopeName string) {
	defer func() {
		if scopeName == "" {
			scopeName = pkg.PkgPath
		}
	}()

	recvIdent, ok := sel.X.(*ast.Ident)
	if !ok {
		return ""
	}
	if recvIdent.Obj == nil || recvIdent.Obj.Decl == nil {
		return ""
	}
	valueSpec, ok := recvIdent.Obj.Decl.(*ast.ValueSpec)
	if !ok {
		return ""
	}

	for _, value := range valueSpec.Values {
		call, ok := value.(*ast.CallExpr)
		if !ok {
			continue
		}

		selExpr, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		if !isIdent(selExpr.X, "monkit") || selExpr.Sel.Name != "ScopeNamed" {
			continue
		}
		if len(call.Args) != 1 {
			continue
		}

		name, ok := call.Args[0].(*ast.BasicLit)
		if !ok {
			continue
		}

		return name.Value[1 : len(name.Value)-1]
	}
	return ""
}

func isIdent(x ast.Expr, name string) bool {
	ident, ok := x.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == name
}

func sortAndUnique(input []string) (unique []string) {
	set := make(map[string]struct{})
	for _, item := range input {
		if _, ok := set[item]; ok {
			continue
		} else {
			set[item] = struct{}{}
		}
	}
	for item := range set {
		unique = append(unique, item)
	}
	sort.Strings(unique)
	return unique
}
