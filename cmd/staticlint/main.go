package main

import (
	"golang.org/x/tools/go/analysis"

	"golang.org/x/tools/go/analysis/multichecker"

	// std passes
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"

	// staticcheck
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/stylecheck"

	// extra
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/copylock"

	// custom
	"github.com/ilyinon/go-musthave-metrics/cmd/staticlint/noosexitmain"
)

func main() {
	var analyzers = []*analysis.Analyzer{
		assign.Analyzer,
		bools.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,

		atomic.Analyzer,
		copylock.Analyzer,

		noosexitmain.Analyzer,
	}

	// SA analyzers
	for _, a := range staticcheck.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	// non-SA analyzers
	for _, a := range simple.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}
	for _, a := range stylecheck.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	multichecker.Main(analyzers...)
}
