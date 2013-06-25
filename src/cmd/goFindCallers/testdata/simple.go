package main

import (
	f "fmt"
)

func a(val int) int {
	f.Println("Checking value")
	if val > 0 {
		return 1
	}
	return 0
}

func bla(val int) int {
	return -1
}

func main() {
	f.Println(a(1))
	panic("bla")
}
