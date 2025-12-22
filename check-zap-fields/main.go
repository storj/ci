// Copyright (C) 2026 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/singlechecker"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/types/typeutil"
)

func main() { singlechecker.Main(Analyzer) }

// Analyzer verifies that zap field names are valid.
// Valid field names must only contain lowercase ASCII letters, numbers, and underscores (only in the middle).
var Analyzer = &analysis.Analyzer{
	Name: "zapfields",
	Doc:  "check that zap logger field names only contain lowercase ASCII letters, numbers, and underscores (only in the middle)",
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

func run(pass *analysis.Pass) (any, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	// Build a map of files with ignore-file directives
	ignoredFiles := buildIgnoredFilesMap(pass)

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)

		fn := typeutil.StaticCallee(pass.TypesInfo, call)
		if fn == nil {
			return
		}

		if isZapFieldFunction(fn) {
			checkZapFieldName(pass, call, fn, ignoredFiles)
		}
	})

	return nil, nil
}

// isZapFieldFunction checks if the function is a zap field function.
func isZapFieldFunction(fn *types.Func) bool {
	// Use FullName to distinguish between top-level functions and methods
	// Zap field functions are top-level functions like "go.uber.org/zap.String"
	// Logger methods would be like "(*go.uber.org/zap.Logger).Error"
	switch fn.FullName() {
	// Basic types
	case "go.uber.org/zap.String",
		"go.uber.org/zap.Strings",
		"go.uber.org/zap.Bool",
		"go.uber.org/zap.Bools",
		"go.uber.org/zap.Int",
		"go.uber.org/zap.Int64",
		"go.uber.org/zap.Int32",
		"go.uber.org/zap.Int16",
		"go.uber.org/zap.Int8",
		"go.uber.org/zap.Uint",
		"go.uber.org/zap.Uint64",
		"go.uber.org/zap.Uint32",
		"go.uber.org/zap.Uint16",
		"go.uber.org/zap.Uint8",
		"go.uber.org/zap.Uintptr",
		"go.uber.org/zap.Float64",
		"go.uber.org/zap.Float32",
		"go.uber.org/zap.Complex128",
		"go.uber.org/zap.Complex64":
		return true

	// Time and duration
	case "go.uber.org/zap.Duration",
		"go.uber.org/zap.Durations",
		"go.uber.org/zap.Time",
		"go.uber.org/zap.Times":
		return true

	// Error types (note: zap.Error is excluded as it only takes an error, not a field name)
	case "go.uber.org/zap.Errors",
		"go.uber.org/zap.NamedError":
		return true

	// Other types
	case "go.uber.org/zap.Any",
		"go.uber.org/zap.Reflect",
		"go.uber.org/zap.Stringer",
		"go.uber.org/zap.ByteString",
		"go.uber.org/zap.ByteStrings",
		"go.uber.org/zap.Binary",
		"go.uber.org/zap.Namespace":
		return true

	// Inline types
	case "go.uber.org/zap.Inline",
		"go.uber.org/zap.Object",
		"go.uber.org/zap.Array":
		return true

	// Stack trace
	case "go.uber.org/zap.Stack",
		"go.uber.org/zap.StackSkip":
		return true
	}

	return false
}

// checkZapFieldName checks if the first argument (field name) is valid.
// Valid field names must:
//   - Not empty
//   - Not contain spaces
//   - Not contain uppercase letters
//   - Only contain ASCII letters, numbers, and underscores
//   - Underscores must be in the middle (not at start or end)
func checkZapFieldName(pass *analysis.Pass, call *ast.CallExpr, fn *types.Func, ignoredFiles map[string]struct{}) {
	if len(call.Args) == 0 {
		return
	}

	firstArg := call.Args[0]

	// Check if it's a basic string literal
	if lit, ok := firstArg.(*ast.BasicLit); ok {
		if lit.Kind == token.STRING {
			// Check if this line should be ignored
			pos := pass.Fset.Position(lit.Pos())
			if _, ok := ignoredFiles[pos.Filename]; ok {
				return
			}

			checkNameLiteral(pass, lit, fn)
		}
	}
}

// checkNameLiteral checks if name matches the ^[a-z0-9]+(_[a-z0-9]+)*$  regular expression and
// reports it when not.
//
// When possible it suggests an auto-fix.
//
// It only acts only if name is basic literal string.
func checkNameLiteral(pass *analysis.Pass, name *ast.BasicLit, zapFn *types.Func) {
	if name.Kind != token.STRING {
		return
	}

	// Remove the enclosing double quotes.
	nval, err := strconv.Unquote(name.Value)
	if err != nil {
		panic("BUG: the `strconv.Unquote` should have received an `ast.BasicLit.Value` of `kind == STRING`, so the value must have been quoted")
	}

	sanitized, valid := sanitizeString(nval)
	if valid {
		return
	}

	if hasIgnoreDirective(pass, name.Pos()) {
		// Don't do anything because this field's name is ignored.
		return
	}

	if sanitized == "" {
		if nval == "" {
			nval = "<empty>"
		}

		pass.Report(analysis.Diagnostic{
			Message: fmt.Sprintf(
				"zap.%s field name %s doesn't match %s regular expression (NOT auto-fixable)",
				zapFn.Name(), nval, rxValidFieldName.String()),
			Pos: name.Pos(),
		})
		return
	}

	// The name is compliant.
	if sanitized == nval {
		return
	}

	pass.Report(analysis.Diagnostic{
		Message: fmt.Sprintf(
			"zap.%s field name %s doesn't match %s regular expression (auto-fixable)",
			zapFn.Name(), nval, rxValidFieldName.String()),
		Pos: name.Pos(),
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message: "Replace unsupported chars by underscores and convert from camelCase to snake_case",
				TextEdits: []analysis.TextEdit{
					// Add one to start and subtract one to end to keep the double quote of the string literal.
					{Pos: name.Pos() + 1, End: name.End() - 1, NewText: []byte(sanitized)},
				},
			},
		},
	})
}

func buildIgnoredFilesMap(pass *analysis.Pass) map[string]struct{} {
	ignoredFiles := make(map[string]struct{})

	for _, file := range pass.Files {
		pos := pass.Fset.Position(file.Pos())
		// Check the first comment after the package keyword.
		pkgLine := pass.Fset.Position(file.Package).Line
	FileComments:
		for _, cg := range file.Comments {
			cgPos := pass.Fset.Position(cg.Pos())
			cgEnd := pass.Fset.Position(cg.End())

			if cgPos.Line < pkgLine {
				// Skip comment groups before package keyword.
				continue
			}

			// Ignore directive should be the first 50 lines after the package keyword to be considered.
			if cgEnd.Line > pkgLine+51 {
				break
			}

			for _, comment := range cg.List {
				if isIgnoreDirective(comment, ignoreFile) {
					ignoredFiles[pos.Filename] = struct{}{}
					break FileComments
				}
			}
		}
	}
	return ignoredFiles
}

// hasIgnoreDirective checks if there's an ignore directive for the given position.
// Checks both the same line and the line immediately before.
func hasIgnoreDirective(pass *analysis.Pass, pos token.Pos) bool {
	position := pass.Fset.Position(pos)
	targetLine := position.Line

	// Find the file containing this position
	var file *ast.File
	for _, f := range pass.Files {
		fPos := pass.Fset.Position(f.Pos())
		if fPos.Filename == position.Filename {
			file = f
			break
		}
	}

	if file == nil {
		return false
	}

	// Check all comments in the file
	for _, cg := range file.Comments {
		for _, comment := range cg.List {
			commentPos := pass.Fset.Position(comment.Pos())
			// Check if comment is on the same line or the line before
			if commentPos.Line == targetLine || commentPos.Line == targetLine-1 {
				if isIgnoreDirective(comment, ignoreLine) {
					return true
				}
			}
		}
	}

	return false
}

const (
	ignoreLine ignoreDirective = "//zapfields:ignore"
	ignoreFile ignoreDirective = "//zapfields:ignore-file"
)

type ignoreDirective string

func (i ignoreDirective) String() string {
	return string(i)
}

// isIgnoreDirective checks if a comment is an ignore directive.
// Only supports: //zapfields:ignore with no space between the directive and //.
func isIgnoreDirective(c *ast.Comment, directive ignoreDirective) bool {
	comment := strings.TrimSpace(c.Text)
	// Do not consider `/*` comments because they are trickier to find out if the ignore directive is
	// in the last line of the comment.
	if strings.HasPrefix(comment, "/*") {
		return false
	}

	// Any characters after the directive separated by a space are fine
	return comment == directive.String() ||
		strings.HasPrefix(comment, directive.String()+" ")
}

var (
	// All valid Zap Logger field names should match this.
	rxValidFieldName = regexp.MustCompile(`^[a-z0-9]+(_[a-z0-9]+)*$`)

	// All unsupported field name characters for sanitization.
	rxSanitizeUnsupported = regexp.MustCompile(`[^a-zA-Z0-9_]+`)
)

// sanitizeString replace all the unsupported characters for Zap Logger field names by underscores,
// suppressing any leading and trailing underscore, and replacing uppercase letters by lowercase
// separating them by underscores (a.k.a. camelCase to snake_case) considering acronyms (e.g.
// DNSResolution becomes dns_resolution).
//
// NOTE according to our conventions, Zap Logger field names must match the ^[a-z0-9]+(_[a-z0-9]+)*$
// regular expression
//
// It returns the same input and valid as true, otherwise, the sanitized input and false.
func sanitizeString(input string) (sanitized string, valid bool) {
	if rxValidFieldName.MatchString(input) {
		return input, true
	}

	// 1. Replace unsupported characters with underscores.
	input = rxSanitizeUnsupported.ReplaceAllString(input, "_")

	var (
		runes  = []rune(input)
		result strings.Builder
		prev   rune
	)

	for i, curr := range runes {
		if curr == '_' && prev == '_' {
			continue
		}

		// If it's not the first character and it's uppercase
		if i > 0 && unicode.IsUpper(curr) {
			// Add an underscore if the previous char was lowercase
			// OR if the next char is lowercase (handling acronyms like DNSResolution --> dns_resolution)
			if unicode.IsLower(prev) || (i+1 < len(runes) && unicode.IsLower(runes[i+1])) {
				if prev != '_' {
					result.WriteRune('_')
				}
			}
		}
		result.WriteRune(unicode.ToLower(curr))
		prev = curr
	}

	// 2. Drop possible leading and trailing underscores.
	return strings.Trim(result.String(), "_"), false
}
