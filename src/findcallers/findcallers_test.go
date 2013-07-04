package findcallers_test

import (
	. "findcallers"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

const (
	TESTPATH = "./testdata/"
	GOPATH   = "C:\\Users\\Manrich\\AppData\\Roaming\\Sublime Text 2\\Packages\\goFindCallers\\"
)

var setfunctests = []struct {
	tstFile string
	toFind  string
	out     string
	next    string
}{

	{"hello.go", "fmt.Println", "fmt.Println", "fmt.Println"},
	{"hello.go", "a", "a", "a"},
	{"hello.go", "ioutil.ReadFile", "ioutil.ReadFile", "ioutil.ReadFile"},

	// Import renamed
	{"foo/simple.go", "f.Println", "f.Println", "fmt.Println"},
	{"foo/simple.go", "fmt.Println", "f.Println", "fmt.Println"},
	{"foo/simple.go", "Bla", "Bla", "foo.Bla"},
	{"foo/simple.go", "a", "a", "a"},
	{"foo/simple.go", "foo.B", "B", "foo.B"},
	{"foo/simple.go", "io.ReadFile", "io.ReadFile", "ioutil.ReadFile"},

	// pgkpath and pkgname mismatch
	{"pakpak/mainpak.go", "pak.Pubpak", "pak.Pubpak", "pak.Pubpak"},
}

// Verifies SetFuncString called on a findcallers.FuncVisitor
func TestSetFuncString(t *testing.T) {
	fset := token.NewFileSet()
	for _, tt := range setfunctests {
		v := NewFuncVisitor(tt.toFind)
		filepath := TESTPATH + tt.tstFile
		filenode, err := parser.ParseFile(fset, filepath, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		v.SetFuncString(filenode)
		s := v.ToFind()
		if s != tt.out {
			t.Errorf("v.SetFuncString(file=%q, toFind=%q) = <%s> want <%s>", tt.tstFile, tt.toFind, s, tt.out)
		}
		s = v.NextFind()
		if s != tt.next {
			t.Errorf("v.SetFuncString(file=%q, toFind=%q) nextFind = <%s> want <%s>", tt.tstFile, tt.toFind, s, tt.next)
		}
	}
}

var pkgpathtests = []struct {
	tstFile string
	toFind  string
	out     string
}{
	// Normal imports
	{"hello.go", "fmt.Println", "fmt"},
	{"hello.go", "foo.B", "findcallers/testdata/foo"},
	{"hello.go", "ioutil.ReadFile", "io/ioutil"},
	{"hello.go", "panic", ""},

	// Import rename
	{"foo/simple.go", "f.Println", "fmt"},
	{"foo/simple.go", "io.ReadFile", "io/ioutil"},
	{"foo/simple.go", "Bla", "findcallers/testdata/foo"},

	// Package paths
	{"pakpak/pakkie/pakkie.go", "Pak", "findcallers/testdata/pakpak/pakkie"},
	{"pakpak/pak/pak.go", "Pubpak", "findcallers/testdata/pakpak/pak"},
	{"pakpak/mainpak.go", "pakkie.Pak", "findcallers/testdata/pakpak/pakkie"},

	// Package name and path dont match
	{"pakpak/mainpak.go", "foo_pak.Pubpak", "findcallers/testdata/pakpak/pak"},
}

func TestPkgPath(t *testing.T) {
	for _, tt := range pkgpathtests {
		v := NewFuncVisitor(tt.toFind)
		fset := token.NewFileSet()
		filepath := TESTPATH + tt.tstFile
		filenode, err := parser.ParseFile(fset, filepath, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		err = v.SetPkgPath(filenode, filepath, strings.Split(GOPATH, ";"))
		if err != nil {
			t.Fatal(err)
		}
		s := v.PkgPath()
		if s != tt.out {
			t.Errorf("v.PgkPath(path=%q, toFind=%q) = <%s> want <%s>", filepath, tt.toFind, s, tt.out)
		}
	}
}

var buildOutputtests = []struct {
	toFind string
	out    string
}{
	{"a", GOPATH + "src\\findcallers\\testdata\\foo\\simple.go\n" +
		"25\n"},
	{"fmt.Println", GOPATH + "src\\findcallers\\testdata\\foo\\simple.go\n" +
		"9,25\n" +
		GOPATH + "src\\findcallers\\testdata\\hello.go\n" +
		"15\n"},
	{"panic", GOPATH + "src\\findcallers\\testdata\\foo\\simple.go\n" +
		"28\n" +
		GOPATH + "src\\findcallers\\testdata\\hello.go\n" +
		"18\n"},
	{"Bla", GOPATH + "src\\findcallers\\testdata\\foo\\dot.go\n" +
		"12\n"},
	{"foo.B", GOPATH + "src\\findcallers\\testdata\\foo\\simple.go\n" +
		"30\n" +
		GOPATH + "src\\findcallers\\testdata\\hello.go\n" +
		"20\n"},
	{"foo", "NotFound"},
}

// Test the output string generated after parsing the TESTPATH
func TestBuildOutput(t *testing.T) {

	for _, tt := range buildOutputtests {
		fset := token.NewFileSet()
		v := NewFuncVisitor(tt.toFind)

		filepath := TESTPATH + "foo/simple.go"
		filenode, err := parser.ParseFile(fset, filepath, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		v.SetFuncString(filenode)
		err = v.SetPkgPath(filenode, filepath, strings.Split(GOPATH, ";"))
		if err != nil {
			t.Fatal(err)
		}
		err = v.ParseDirectory(fset, TESTPATH)
		if err != nil {
			t.Fatal(err)
		}
		s := v.BuildOutput(fset)
		if s != tt.out {
			t.Errorf("v.BuildOutput(path=%q, toFind=%q) = <%s> want <%s>", TESTPATH, tt.toFind, s, tt.out)
		}
	}
}
