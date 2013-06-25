package findcallers_test

import (
	. "findcallers"
	"go/parser"
	"go/token"
	"testing"
)

const (
	TESTPATH = "./testdata/"
)

var setfunctests = []struct {
	tstFile string
	toFind  string
	out     string
	after   string
}{

	{"hello.go", "fmt.Println", "fmt.Println", "fmt.Println"},
	{"hello.go", "a", "a", "a"},
	{"hello.go", "ioutil.ReadFile", "ioutil.ReadFile", "ioutil.ReadFile"},

	// Import renamed
	{"simple.go", "f.Println", "f.Println", "fmt.Println"},
	{"simple.go", "fmt.Println", "f.Println", "fmt.Println"},
	{"simple.go", "Bla", "Bla", "foo.Bla"},
	{"simple.go", "a", "a", "foo.a"},
	{"simple.go", "foo.B", "B", "foo.B"},
	{"simple.go", "io.ReadFile", "io.ReadFile", "ioutil.ReadFile"},
}

// Verifies SetFuncString called on a findcallers.FuncVisitor
func TestSetFuncString(t *testing.T) {
	fset := token.NewFileSet()
	for _, tt := range setfunctests {
		v := new(FuncVisitor)
		filepath := TESTPATH + tt.tstFile
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
		if s != tt.after {
			t.Errorf("v.SetFuncString(file=%q, toFind=%q) afterFind = <%s> want <%s>", tt.tstFile, tt.toFind, s, tt.after)
		}
	}
}

var buildOutputtests = []struct {
	toFind string
	out    string
}{
	{"a", "testdata\\hello.go\n" +
		"14\n" +
		"testdata\\simple.go\n" +
		"25\n"},
	{"fmt.Println", "testdata\\hello.go\n" +
		"15\n" +
		"testdata\\simple.go\n" +
		"9,25\n"},
	{"panic", "testdata\\hello.go\n" +
		"18\n" +
		"testdata\\simple.go\n" +
		"28\n"},
	{"foo.B", "testdata\\hello.go\n" +
		"20\n" +
		"testdata\\simple.go\n" +
		"30\n"},
	{"foo","NotFound"},
}

// Test the output string generated after parsing the TESTPATH
func TestBuildOutput(t *testing.T) {
	for _, tt := range buildOutputtests {
		v := new(FuncVisitor)
		fset := token.NewFileSet()
		v.OriginFind = tt.toFind
		err := v.ParseDirectory(fset, TESTPATH)
		if err != nil {
			t.Fatal(err)
		}

		s := v.BuildOutput(fset)
		if s != tt.out {
			t.Errorf("v.BuildOutput(path=%q, toFind=%q) = <%s> want <%s>", TESTPATH, tt.toFind, s, tt.out)
		}
	}

}
