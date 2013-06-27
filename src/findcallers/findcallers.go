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
	pkgPath  string
	nextFind string
	toFind   string
	poslist  []token.Pos
}

func NewFuncVisitor(toFind string) *FuncVisitor {
	v := new(FuncVisitor)
	v.nextFind = toFind
	return v
}

func (v *FuncVisitor) NextFind() string {
	return v.nextFind
}

func (v *FuncVisitor) ToFind() string {
	return v.toFind
}

func (v *FuncVisitor) PkgPath() string {
	return v.pkgPath
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
		return strings.EqualFold(a.String(), toFind)
	// If at selector expression split on '.' and call findAndMatch on each part
	case *ast.SelectorExpr:
		exprSel := strings.Split(v.toFind, ".")
		if len(exprSel) > 1 {
			if v.findAndMatch(a.X, exprSel[0]) {
				return v.findAndMatch(a.Sel, exprSel[1])
			}
		}
	}
	// return false if node is not ast.Indent or ast.SelectorExpr
	return false
}

// ParseDirectory recursively walk through the path and parses each file using parser.ParseFile
// as well as calls findAndMatch
// It takes fset, the starting filepath and an ast.Vistor
func (v *FuncVisitor) ParseDirectory(fset *token.FileSet, p string) (first error) {
	fd, err := os.Open(p)
	if err != nil {
		return err
	}
	defer fd.Close()
	fileList, err := fd.Readdir(-1)
	if err != nil {
		return err
	}
	for _, f := range fileList {
		fpath := filepath.Join(p, f.Name())
		if f.IsDir() {
			err := v.ParseDirectory(fset, fpath)
			if err != nil {
				return err
			}
		} else {
			// Only parse .go-files
			if filepath.Ext(f.Name()) == ".go" {
				filenode, err := parser.ParseFile(fset, fpath, nil, 0)
				if err != nil {
					return err
				}

				v.SetFuncString(filenode)
				//Walk and find function
				ast.Walk(v, filenode)
			}
		}
	}
	return nil
}

func (v *FuncVisitor) pkgMatch(file *ast.File, fpath string) bool {

	for _, i := range file.Imports {

		if v.pkgPath == unquote(i.Path.Value) {
			return true
		}
		if strings.Contains(unquote(i.Path.Value), strings.Replace(filepath.Dir(fpath), "\\", "/", -1)) {
			return true
		}
	}
	return false
}

func (v *FuncVisitor) SetPkgPath(file *ast.File, fpath string, gopath []string) error {
	fpath, err := filepath.Abs(fpath)
	if err != nil {
		return err
	}
	fsplit := strings.SplitAfterN(filepath.Dir(fpath), "\\src\\", 2)
	fpath = strings.Replace(fsplit[len(fsplit)-1], "\\", "/", -1)
	if strings.Contains(v.nextFind, ".") {
		// If exprSel split
		exprSel := strings.Split(v.nextFind, ".")
		//Pkg Name match and in scope
		if strings.EqualFold(file.Name.Name, exprSel[0]) && file.Scope.Objects[exprSel[1]] != nil {
			v.pkgPath = fpath
			return nil
		} else {
			for i := range file.Imports {
				curImport := file.Imports[i]
				_, selc := filepath.Split(unquote(curImport.Path.Value))
				// Import name match OR selector match
				if curImport.Name != nil && (curImport.Name.String() == exprSel[0]) || exprSel[0] == selc {
					v.pkgPath = unquote(curImport.Path.Value)
					return nil
				} else {
					for _, p := range gopath {
						curPath := filepath.Clean(p + "\\src\\" + unquote(curImport.Path.Value))
						_, err := os.Stat(curPath)
						if !os.IsNotExist(err) {
							fset := token.NewFileSet()
							pkgs, err := parser.ParseDir(fset, curPath, isFile, parser.PackageClauseOnly)
							if err != nil {
								return err
							}
							for _, i := range pkgs {
								if i.Name == exprSel[0] {
									v.pkgPath = unquote(curImport.Path.Value)
									return nil
								}
							}
						}
					}
				}
			}
		}
	} else {
		if file.Scope.Objects[v.nextFind] != nil && !strings.EqualFold(file.Name.Name, "main") {
			v.pkgPath = fpath
			return nil
		}
	}
	return nil
}

// Checks current file and its package and import information to determine
// what the search string should be change to
func (v *FuncVisitor) SetFuncString(file *ast.File) {
	// Check if selector Expression
	if strings.Contains(v.nextFind, ".") {
		// If exprSel split
		exprSel := strings.Split(v.nextFind, ".")
		// if import rename != nil
		for i := range file.Imports {
			curImport := file.Imports[i]
			if curImport.Name != nil {
				// If import rename == Expr import name
				_, selc := filepath.Split(unquote(curImport.Path.Value))
				if strings.EqualFold(exprSel[0], selc) {
					v.toFind = curImport.Name.String() + "." + exprSel[1]
					return
				} else {
					// If original import name == Expr
					if strings.EqualFold(exprSel[0], curImport.Name.String()) {
						v.toFind = v.nextFind
						v.nextFind = selc + "." + exprSel[1]
						return
					}
				}
			}
		}
		if strings.EqualFold(file.Name.Name, exprSel[0]) && file.Scope.Objects[exprSel[1]] != nil {
			v.toFind = exprSel[1]
			return
		}
	} else {
		// If the current file is a non-main package and the toFind string its function
		if file.Scope.Objects[v.nextFind] != nil && !strings.EqualFold(file.Name.Name, "main") {
			v.toFind = v.nextFind
			v.nextFind = file.Name.Name + "." + v.nextFind
			return
		}
	}
	v.toFind = v.nextFind
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

func unquote(s string) string {
	return strings.Trim(s, "\"")
}

func isFile(f os.FileInfo) bool {
	return !f.IsDir()
}
