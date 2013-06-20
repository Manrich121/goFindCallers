package main

import (
	"fmt"
)

func a(val int) int {
	fmt.Println("Checking value")
	if val > 0 {
		return 1
	}
	return 0
}

func b(val int) int {
	return -1
}

func main() {
	fmt.Println(a(1))
}
