package noosexitmain

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "noosexitmain",
	Doc:  "reports calls to os.Exit, log.Fatal and panic outside main.main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			isMainFunc := pass.Pkg.Name() == "main" && fn.Name.Name == "main"

			ast.Inspect(fn, func(n ast.Node) bool {
				call, okCall := n.(*ast.CallExpr)
				if !okCall {
					return true
				}

				// panic(...)
				if ident, okIdent := call.Fun.(*ast.Ident); okIdent {
					if ident.Name == "panic" && !isMainFunc {
						pass.Reportf(call.Pos(), "do not call panic outside main.main")
					}
					return true
				}

				// os.Exit / log.Fatal*
				sel, okSel := call.Fun.(*ast.SelectorExpr)
				if !okSel {
					return true
				}

				pkgIdent, okPkg := sel.X.(*ast.Ident)
				if !okPkg {
					return true
				}

				switch pkgIdent.Name {
				case "os":
					if sel.Sel.Name == "Exit" && !isMainFunc {
						pass.Reportf(call.Pos(), "do not call os.Exit outside main.main")
					}
				case "log":
					if isLogFatal(sel.Sel.Name) && !isMainFunc {
						pass.Reportf(call.Pos(), "do not call log.Fatal outside main.main")
					}
				}

				return true
			})
		}
	}

	return nil, nil
}

func isLogFatal(name string) bool {
	return name == "Fatal" || name == "Fatalf" || name == "Fatalln"
}
