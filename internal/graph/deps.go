package graph

import "fmt"

// Re-implement some of the dependencies used in munkres.go

// IntAssertLessThan panics if a >= b.
func IntAssertLessThan(a, b int) {
	if a >= b {
		panic(fmt.Sprintf("assert %d < %d failed", a, b))
	}
}

// Panic just panics.
func Panic(a string) {
	panic(a)
}

// Imax returns the maximum between two integers
func Imax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Min returns the minimum between two float point numbers.
func Min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Alloc allocates a slice of slices of float64.
func Alloc(m, n int) (mat [][]float64) {
	mat = make([][]float64, m)
	for i := 0; i < m; i++ {
		mat[i] = make([]float64, n)
	}
	return
}

// IntAlloc allocates a matrix of integers.
func IntAlloc(m, n int) (mat [][]int) {
	mat = make([][]int, m)
	for i := 0; i < m; i++ {
		mat[i] = make([]int, n)
	}
	return
}

// Sf wraps Sprintf
func Sf(msg string, prm ...interface{}) string {
	return fmt.Sprintf(msg, prm...)
}
