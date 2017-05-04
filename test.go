package main

import (
	"fmt"
	"math"
)

func main() {
	/// UNDERSTANDING BINARY SHIFT /////
	var shift uint = 10

	fmt.Println("8 * 1 <<", shift, " : ", 8<<shift)
	fmt.Println("8 * 2 ^", shift, " : ", (8 * math.Pow(2, float64(shift))))
}
