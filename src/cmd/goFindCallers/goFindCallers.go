package main

import (
	"bufio"
	"findcallers"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	// Read in function name to find and first file path to search from stdin
	inline, _, err := reader.ReadLine()
	if err != nil && err != io.EOF {
		panic(err)
	}

	// Format: funcToFind=filepath
	find_Path := strings.Split(string(inline), "=")
	// Split firstfile and gopath variables
	filepath := strings.Split(find_Path[1], string(os.PathListSeparator))
	gopath := filepath[1:]

	// visitor used to ast.Walk
	visitor := findcallers.NewFuncVisitor(find_Path[0])

	// Build the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	filenode, err := parser.ParseFile(fset, filepath[0], nil, 0)
	if err != nil {
		panic(err)
	}

	// Check current file to get new search func string
	visitor.SetFuncString(filenode)
	// walk through first file
	ast.Walk(visitor, filenode)

	// Open and parse files in Gopath
	for p := range gopath {
		err = visitor.ParseDirectory(fset, gopath[p])
		if err != nil {
			panic(err)
		}
	}

	//Output on stdout
	fmt.Print(visitor.BuildOutput(fset))
}
