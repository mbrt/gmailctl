// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utl

import "math"

// BestSquare finds the best square for given size=Nrows * Ncolumns
func BestSquare(size int) (nrow, ncol int) {
	nrow = -1 // not found
	ncol = -1 // not found
	for x := 1; x <= size; x++ {
		if (x * x) >= size {
			if (x * x) == size {
				nrow = x
				ncol = x
				return
			}
			for y := x; y >= 1; y-- {
				if (x * y) == size {
					nrow = x
					ncol = y
					return
				}
			}
		}
	}
	return
}

// BestSquareApprox finds the best square for given size=Nrows * Ncolumns.
// Approximate version; i.e. nrow*ncol may not be equal to size
func BestSquareApprox(size int) (nrow, ncol int) {
	fsize := float64(size)
	ncol = int(math.Floor(math.Sqrt(fsize)))
	nrow = int(math.Ceil(fsize / float64(ncol)))
	return
}

// Iabs performs the absolute operation with ints
func Iabs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// Imin returns the minimum between two integers
func Imin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Imax returns the maximum between two integers
func Imax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Min returns the minimum between two float point numbers
func Min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum between two float point numbers
func Max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// IsPowerOfTwo checks if n is power of 2; i.e. 2⁰, 2¹, 2², 2³, 2⁴, ...
func IsPowerOfTwo(n int) bool {
	if n < 1 {
		return false
	}
	return n&(n-1) == 0
}

// Swap swaps two float64 numbers
func Swap(a, b *float64) {
	*a, *b = *b, *a
}
