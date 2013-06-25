package findcallers_test

import (
	. "findcallers"
	"go/parser"
	"go/token"
	"testing"
)

const (
	TESTDATA = "./testdata/"
)

var setfunctests = []struct {
	tstFile string
	toFind  string
	out     string
	next    string
}{

	{"hello.go", "fmt.Print", "fmt.Print", "fmt.Print"},
	{"hello.go", "a", "a", "a"},
	{"hello.go", "ioutil.ReadFile", "ioutil.ReadFile", "ioutil.ReadFile"},
	{"simple.go", "f.Println", "f.Println", "fmt.Println"},
	{"simple.go", "fmt.Println", "f.Println", "fmt.Println"},
	{"simple.go", "bla", "bla", "foo.bla"},
	{"simple.go", "a", "a", "foo.a"},
	{"simple.go", "io.ReadFile", "io.ReadFile", "ioutil.ReadFile"},
}

// Verifies that
func TestSetFuncString(t *testing.T) {
	fset := token.NewFileSet()
	for _, tt := range setfunctests {
		filepath := TESTDATA + tt.tstFile
		v := new(FuncVisitor)
		v.OriginFind = tt.toFind
		filenode, err := parser.ParseFile(fset, filepath, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		s := v.SetFuncString(filenode)
		if s != tt.out {
			t.Errorf("v.SetFuncString(file=%q, toFind=%q) = <%s> want <%s>", tt.tstFile, tt.toFind, s, tt.out)
		}
		s = v.OriginFind
		if s != tt.next {
			t.Errorf("v.SetFuncString(file=%q, toFind=%q) nextFind = <%s> want <%s>", tt.tstFile, tt.toFind, s, tt.next)
		}
	}

}
