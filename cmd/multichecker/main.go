package main

import (
	"strings"

	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"github.com/kisielk/errcheck/errcheck"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"

	"metricalert/cmd/staticlint"
)

func main() {
	// Добавляем стандартные анализаторы
	checks := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		structtag.Analyzer,
		nilfunc.Analyzer,
	}

	// Добавляем все SA-анализаторы, S-анализаторы и ST1008 staticcheck
	for _, v := range staticcheck.Analyzers {
		if v.Analyzer == nil {
			continue
		}

		if strings.HasPrefix(v.Analyzer.Name, "S") {
			checks = append(checks, v.Analyzer)
		}

		if v.Analyzer.Name == "ST1008" {
			checks = append(checks, v.Analyzer)
		}
	}

	// Добавляем дополнительные публичные анализаторы
	checks = append(checks,
		errcheck.Analyzer,
		bodyclose.Analyzer,
		ineffassign.Analyzer,
		staticlint.OsExitAnalyzer) // Добавляем анализатор staticlint

	// Запускаем multichecker
	multichecker.Main(checks...)
}
