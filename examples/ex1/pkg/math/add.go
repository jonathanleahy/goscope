package math

import "fmt"

// Add returns the sum of two integers.
// It validates inputs before adding them.
func Add(a, b int) int {
	if !validateInputs(a, b) {
		fmt.Println("invalid inputs")
		return 0
	}
	return a + b
}

// Sub returns the difference between two integers.
func Sub(a, b int) int {
	return a - b
}
