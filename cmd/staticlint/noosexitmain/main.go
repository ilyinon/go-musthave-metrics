package noosexitmain

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "noosexitmain",
	Doc:  "forbid direct os.Exit call inside main.main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// только main пакеты
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	pkgPath := pass.Pkg.Path()

	if pkgPath != "github.com/ilyinon/go-musthave-metrics/cmd/server" &&
		pkgPath != "github.com/ilyinon/go-musthave-metrics/cmd/agent" {
		return nil, nil
	}

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" {
				continue
			}

			ast.Inspect(fn, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				ident, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}

				if ident.Name == "os" && sel.Sel.Name == "Exit" {
					pass.Reportf(call.Pos(), "do not call os.Exit in main")
				}

				return true
			})
		}
	}

	return nil, nil
}
