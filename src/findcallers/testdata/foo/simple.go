package foo

import (
	f "fmt"
	io "io/ioutil"
)

func a(val int) int {
	f.Println("Checking value")
	if val > 0 {
		return 1
	}
	return 0
}

func Bla(val int) int {
	return -1
}

func B() {
	return 0
}

func main() {
	f.Println(a(1))
	src, err := io.ReadFile("simple.go")
	if err != nil {
		panic(err)
	}
	B()
}
