package main

import (
	"fmt"
	"io/ioutil"
)

func a() int {
	return 0
}

func main() {
	fmt.Print("Hallo World!")
	src, err := ioutil.ReadFile("hello.go")
	if err != nil {
		panic(err)
	}
}
