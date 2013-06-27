package main

import (
	"findcallers/testdata/foo"
	"fmt"
	"io/ioutil"
)

func a() int {
	return 0
}

func main() {
	a()
	fmt.Println("Hallo World!")
	src, err := ioutil.ReadFile("hello.go")
	if err != nil {
		panic(err)
	}
	foo.B()
}
