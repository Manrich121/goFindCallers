package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Struct that inherits the Visit method needed by ast.Walk
// toFind is the function name to be found
// poslist is a slice of type token.Pos used to store Positions within files
type FuncVisitor struct {
	orginalFind string
	toFind      string
	poslist     []token.Pos
}

// Visit interface used by as.Walk to traverse the AST
// FuncVistor is define above
func (v *FuncVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch t := node.(type) {
	case *ast.CallExpr:
		if v.find(t.Fun) {
			v.poslist = append(v.poslist, node.Pos())
		}
	}
	return v
}

// Find funcVisitor also has a find() method used as wrapper to locate the function within a give ast.Node
// return type is a bool to determine if the function was found at the current node
func (v *FuncVisitor) find(fun ast.Node) bool {
	return v.findAndMatch(fun, v.toFind)
}

// findAndMatch is the implementation of the find method
// It takes the current function node and toFind string and
// return type bool, true if the find was a match
func (v *FuncVisitor) findAndMatch(fun ast.Node, toFind string) bool {
	switch a := fun.(type) {
	// If at the deepest node find and Match
	case *ast.Ident:
		if strings.EqualFold(a.String(), toFind) {
			return true
		}
	// If at selector expression split on '.' and call findAndMatch on each part
	case *ast.SelectorExpr:
		exprSel := strings.Split(v.toFind, ".")
		if v.findAndMatch(a.X, exprSel[0]) {
			if v.findAndMatch(a.Sel, exprSel[1]) {
				return true
			}
		}
	}
	// return false if node not ast.Indent or ast.SelectorExpr
	return false
}

// parseDirectory recursively walk through the path and parses each file using parser.ParseFile
// as well as calls findAndMatch
// It takes fset, the starting filepath and an ast.Vistor
func (v *FuncVisitor) parseDirectory(fset *token.FileSet, path string) (first error) {
	fd, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fd.Close()
	fileList, err := fd.Readdir(-1)
	if err != nil {
		return err
	}
	for _, f := range fileList {
		filepath := filepath.Join(path, f.Name())
		if f.IsDir() {
			v.parseDirectory(fset, filepath)
		} else {
			// Only parse .go-files
			if strings.HasSuffix(f.Name(), ".go") {
				filenode, err := parser.ParseFile(fset, filepath, nil, 0)
				if err != nil {
					return err
				}
				v.toFind, _ = checkExprSel(filenode, v.orginalFind)
				//Walk and find function
				ast.Walk(v, filenode)
			}
		}
	}
	return nil
}

func getFunctionString(file *ast.File, toFind string) (string, bool) {

	if file.Scope.Objects[toFind] != nil && !strings.EqualFold(file.Name.Name, "main") {
		return (file.Name.Name + "." + toFind), false
	}

	return checkExprSel(file, toFind)
}

func checkExprSel(file *ast.File, toFind string) (string, bool) {
	exprSel := strings.Split(toFind, ".")
	if len(exprSel) > 1 {
		for i := range file.Imports {
			curImport := file.Imports[i]
			if curImport.Name != nil && strings.EqualFold(exprSel[0], strings.Trim(curImport.Path.Value, "\"")) {
				return (curImport.Name.String() + "." + exprSel[1]), false
			} else {
				if curImport.Name != nil && strings.EqualFold(exprSel[0], curImport.Name.String()) {
					return (strings.Trim(curImport.Path.Value, "\"") + "." + exprSel[1]), true
				}
			}
			// f.Println(file.Imports[i].Name, file.Imports[i].Path.Value)
		}
	}
	return toFind, false
}

func main() {
	// visitor used to ast.Walk
	visitor := new(FuncVisitor)
	reader := bufio.NewReader(os.Stdin)

	// Read in function name to find and starting file path to search from stdin
	line, _, err := reader.ReadLine()
	if err != nil && err != io.EOF {
		panic(err)
	}

	// Format: functoFind=filepath
	splitInput := strings.Split(string(line), "=")

	filepath := splitInput[1]

	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	filenode, err := parser.ParseFile(fset, filepath, nil, 0)
	if err != nil {
		panic(err)
	}

	visitor.orginalFind = splitInput[0]
	visitor.toFind = splitInput[0]
	// walk through first file
	ast.Walk(visitor, filenode)

	tf, change := getFunctionString(filenode, splitInput[0])
	visitor.toFind = tf
	if change {
		visitor.orginalFind = tf
	}

	// Find, open and parse Gopath
	gopath := os.Getenv("GOPATH")
	err = visitor.parseDirectory(fset, gopath)
	if err != nil {
		panic(err)
	}

	// Map with filepath as key and string array of lines
	posoutput := make(map[string][]string)
	OutputString := ""

	if len(visitor.poslist) > 0 {
		for n := range visitor.poslist {
			cur := visitor.poslist[n]
			if cur.IsValid() {
				posoutput[fset.Position(cur).Filename] = append(posoutput[fset.Position(cur).Filename], strconv.Itoa(fset.Position(cur).Line))
			}
		}
		// For each key=filepath append string to output string
		for filekey, _ := range posoutput {
			OutputString = OutputString + filekey + "\n" + strings.Join(posoutput[filekey], ",") + "\n"
		}
		fmt.Print(OutputString)
	} else {
		// Print flag NotFound to indicate that the function was not found
		fmt.Printf("NotFound")
	}

}
