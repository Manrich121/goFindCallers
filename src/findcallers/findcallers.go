package findcallers

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Struct that inherits the Visit method needed by ast.Walk
// toFind is the function name to be found
// poslist is a slice of type token.Pos used to store Positions within files
type FuncVisitor struct {
	OriginFind string
	toFind     string
	poslist    []token.Pos
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
	// if v.toFind == "" {
	// 	v.toFind = v.OriginFind
	// }
	return v.findAndMatch(fun, v.toFind)
}

// findAndMatch is the implementation of the find method
// It takes the current function node and toFind string and
// return type bool, true if the find was a match
func (v *FuncVisitor) findAndMatch(fun ast.Node, toFind string) bool {
	switch a := fun.(type) {
	// If at the deepest node find and Match
	case *ast.Ident:
		return strings.EqualFold(a.String(), toFind)
	// If at selector expression split on '.' and call findAndMatch on each part
	case *ast.SelectorExpr:
		exprSel := strings.Split(v.toFind, ".")
		if v.findAndMatch(a.X, exprSel[0]) {
			return v.findAndMatch(a.Sel, exprSel[1])
		}
	}
	// return false if node not ast.Indent or ast.SelectorExpr
	return false
}

// ParseDirectory recursively walk through the path and parses each file using parser.ParseFile
// as well as calls findAndMatch
// It takes fset, the starting filepath and an ast.Vistor
func (v *FuncVisitor) ParseDirectory(fset *token.FileSet, path string) (w *FuncVisitor, first error) {
	fd, err := os.Open(path)
	if err != nil {
		return v, err
	}
	defer fd.Close()
	fileList, err := fd.Readdir(-1)
	if err != nil {
		return v, err
	}
	for _, f := range fileList {
		filepath := filepath.Join(path, f.Name())
		if f.IsDir() {
			v, err := v.ParseDirectory(fset, filepath)
			if err != nil {
				return v, err
			}
		} else {
			// Only parse .go-files
			if strings.HasSuffix(f.Name(), ".go") {
				filenode, err := parser.ParseFile(fset, filepath, nil, 0)
				if err != nil {
					return v, err
				}
				v.SetFuncString(filenode)
				//Walk and find function
				ast.Walk(v, filenode)
			}
		}
	}
	return v, nil
}

// Checks current file and its package and import information to determine
// what the search string should be change to
func (v *FuncVisitor) SetFuncString(file *ast.File) string {
	// Check if selector Expression
	if strings.Contains(v.OriginFind, ".") {
		// If exprSel split
		exprSel := strings.Split(v.OriginFind, ".")
		// if import rename != nil
		for i := range file.Imports {
			curImport := file.Imports[i]
			if curImport.Name != nil {
				// If import rename == Expr import name
				selc := strings.Split(strings.Trim(curImport.Path.Value, "\""), "/")
				if strings.EqualFold(exprSel[0], selc[len(selc)-1]) {
					v.toFind = curImport.Name.String() + "." + exprSel[1]
					return v.toFind
				} else {
					// If original import name == Expr
					if strings.EqualFold(exprSel[0], curImport.Name.String()) {
						v.toFind = v.OriginFind
						v.OriginFind = selc[len(selc)-1] + "." + exprSel[1]
						return v.toFind
					}
				}
			}
		}
	} else {
		// If the current file is a non-main package and the toFind string its function
		if file.Scope.Objects[v.OriginFind] != nil && !strings.EqualFold(file.Name.Name, "main") {
			v.toFind = v.OriginFind
			v.OriginFind = file.Name.Name + "." + v.OriginFind
			return v.toFind
		}
	}
	v.toFind = v.OriginFind
	return v.toFind
}

// Function builds a output string based on the FuncVistor's poslist, relative to fset
// format: 	filename\n comma separated line-positions\n
// returns "NotFound" if the poslist is empty, thus contains no find results
func (v *FuncVisitor) BuildOutput(fset *token.FileSet) string {
	// Map with filepath as key and string array of lines
	posoutput := make(map[string][]string)
	OutputString := ""

	if len(v.poslist) > 0 {
		for n := range v.poslist {
			cur := v.poslist[n]
			if cur.IsValid() {
				posoutput[fset.Position(cur).Filename] = append(posoutput[fset.Position(cur).Filename], strconv.Itoa(fset.Position(cur).Line))
			}
		}
		// For each key=filepath append string to output string
		for filekey, _ := range posoutput {
			OutputString = OutputString + filekey + "\n" + strings.Join(posoutput[filekey], ",") + "\n"
		}
		return OutputString
	} else {
		// return flag NotFound to indicate that the function was not found
		return "NotFound"
	}
}
