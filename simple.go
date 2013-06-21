package main

import (
	f "fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func a(val int) int {
	f.Println("Checking value")
	if val > 0 {
		return 1
	}
	return 0
}

func bla(val int) int {
	return -1
}

func main() {
	f.Println(a(1))
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "C:/Users/Manrich/AppData/Roaming/Sublime Text 2/Packages/GoFindCallers/simple.go", nil, 0)

	ast.Print(fset, file)

	for i := range file.Imports {
		f.Println(file.Imports[i].Name, file.Imports[i].Path.Value)
	}
}
