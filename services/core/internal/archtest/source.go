package archtest

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
)

// CheckSource parses one Go file belonging to package pkg and returns every boundary violation found.
func CheckSource(pkg, filename string, src []byte) ([]error, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", filename, err)
	}

	var violations []error
	timeAlias := ""
	for _, spec := range file.Imports {
		imp, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return nil, fmt.Errorf("parse import in %s: %w", filename, err)
		}
		if err := CheckImport(pkg, imp); err != nil {
			violations = append(violations, err)
		}
		if imp == "time" {
			timeAlias = "time"
			if spec.Name != nil {
				timeAlias = spec.Name.Name
			}
		}
	}

	if loc, ok := locate(pkg); ok && isBusinessModule(loc.module) && loc.layer == "domain" && timeAlias != "" {
		ast.Inspect(file, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Now" {
				return true
			}
			if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == timeAlias {
				violations = append(violations, fmt.Errorf("%s: domain must use the Clock port instead of time.Now", fset.Position(sel.Pos())))
			}
			return true
		})
	}

	return violations, nil
}
