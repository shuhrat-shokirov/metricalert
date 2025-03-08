//go:build exclude

/*
Package staticlint предоставляет пользовательские анализаторы для multichecker.

Этот пакет включает анализатор `OsExitAnalyzer`, который запрещает использование `os.Exit`
в `main`-функции пакета `main`, обеспечивая правильное управление выходом из программы.

📌 Использование

Запустите анализатор с помощью `multichecker`:

	```sh
	multichecker -c=1 -all -tags=staticlint ./...
	```
*/
package staticlint

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// OsExitAnalyzer проверяет использование os.Exit в функции main пакета main.
var OsExitAnalyzer = &analysis.Analyzer{
	Name: "osexitlint",
	Doc:  "Запрещает использование os.Exit в main-функции пакета main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	osExitFunc := func(x *ast.ExprStmt) {
		// проверяем, что выражение представляет собой вызов функции
		// и что функция - это os.Exit
		if call, ok := x.X.(*ast.CallExpr); ok {
			if isOsExit(call) {
				pass.Reportf(x.Pos(), "использование os.Exit в main-функции")
			}
		}
	}

	for _, file := range pass.Files {
		if !isMain(pass, file) {
			continue
		}

		// функцией ast.Inspect проходим по всем узлам AST
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.ExprStmt: // выражение
				osExitFunc(x)
			default:
			}
			return true
		})
	}
	return nil, nil //nolint:nilnil,nolintlint,gocritic
}

// isMain проверяет, что файл - это main-функция пакета main.
func isMain(pass *analysis.Pass, file *ast.File) bool {
	// Получаем путь к файлу
	pos := pass.Fset.Position(file.Pos())

	if !strings.Contains(pos.Filename, "/cmd/") {
		return false
	}

	if !strings.HasSuffix(pos.Filename, "main.go") {
		return false
	}

	return true
}

// isOsExit проверяет, что вызов функции - это os.Exit.
func isOsExit(call *ast.CallExpr) bool {
	// Проверяем, является ли вызов селектором (например, os.Exit)
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		// Проверяем, что селектор принадлежит идентификатору os
		if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "os" && sel.Sel.Name == "Exit" {
			return true
		}
	}
	return false
}
