// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chk

import (
	"fmt"
	"testing"
)

// Float64 compares two float64 numbers
func Float64(tst *testing.T, msg string, tol, a, b float64) {
	TstDiff(tst, msg, tol, a, b, true)
}

// Complex128 compares two complex128 numbers
func Complex128(tst *testing.T, msg string, tolNorm float64, a, b complex128) {
	TestDiffC(tst, msg, tolNorm, a, b, true)
}

// AnaNum compares analytical versus numerical values
func AnaNum(tst *testing.T, msg string, tol, ana, num float64, verbose bool) {
	e := PrintAnaNum(msg, tol, ana, num, verbose)
	if e != nil {
		TstFail(tst, "%v", e.Error())
	}
}

// AnaNumC compares analytical versus numerical values (complex version)
func AnaNumC(tst *testing.T, msg string, tol float64, ana, num complex128, verbose bool) {
	e := PrintAnaNumC(msg, tol, ana, num, verbose)
	if e != nil {
		TstFail(tst, "%v", e.Error())
	}
}

// String compares two strings
func String(tst *testing.T, a, b string) {
	if a != b {
		TstFail(tst, "%q != %q", a, b)
		return
	}
	PrintOk("%s == %s", a, b)
}

// Int compares two ints
func Int(tst *testing.T, msg string, a, b int) {
	if a != b {
		TstFail(tst, "%s: %d != %d", msg, a, b)
		return
	}
	PrintOk("%s: %d == %d", msg, a, b)
}

// Int32 compares two int32
func Int32(tst *testing.T, msg string, a, b int32) {
	if a != b {
		TstFail(tst, "%s: %d != %d", msg, a, b)
		return
	}
	PrintOk("%s: %d == %d", msg, a, b)
}

// Int64 compares two int64
func Int64(tst *testing.T, msg string, a, b int64) {
	if a != b {
		TstFail(tst, "%s: %d != %d", msg, a, b)
		return
	}
	PrintOk("%s: %d == %d", msg, a, b)
}

// Ints compares two slices of integer. The b slice may be nil indicating that all values are zero
func Ints(tst *testing.T, msg string, a, b []int) {
	if len(a) != len(b) {
		TstFail(tst, "%s len(a)=%d != len(b)=%d", msg, len(a), len(b))
		return
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			TstFail(tst, "%s [%d] %v != %v", msg, i, a[i], b[i])
			return
		}
	}
	PrintOk(msg)
}

// Int32s compares two slices of 32 integer. The b slice may be nil indicating that all values are zero
func Int32s(tst *testing.T, msg string, a, b []int32) {
	if len(a) != len(b) {
		TstFail(tst, "%s len(a)=%d != len(b)=%d", msg, len(a), len(b))
		return
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			TstFail(tst, "%s [%d] %v != %v\n", msg, i, a[i], b[i])
			return
		}
	}
	PrintOk(msg)
}

// Int64s compares two slices of 64 integer. The b slice may be nil indicating that all values are zero
func Int64s(tst *testing.T, msg string, a, b []int64) {
	if len(a) != len(b) {
		TstFail(tst, "%s len(a)=%d != len(b)=%d", msg, len(a), len(b))
		return
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			TstFail(tst, "%s [%d] %v != %v\n", msg, i, a[i], b[i])
			return
		}
	}
	PrintOk(msg)
}

// Bools compare two slices of bool. The b slice may be nil indicating that all values are false
func Bools(tst *testing.T, msg string, a, b []bool) {
	if len(a) != len(b) {
		TstFail(tst, "%s len(a)=%d != len(b)=%d", msg, len(a), len(b))
		return
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			TstFail(tst, "%s [%d] %v != %v\n", msg, i, a[i], b[i])
			return
		}
	}
	PrintOk(msg)
}

// Strings compare two slices of string. The b slice may be nil indicating that all values are "" (empty)
func Strings(tst *testing.T, msg string, a, b []string) {
	if len(a) != len(b) {
		TstFail(tst, "%s len(a)=%d != len(b)=%d", msg, len(a), len(b))
		return
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			TstFail(tst, "%s [%d] %v != %v\n", msg, i, a[i], b[i])
			return
		}
	}
	PrintOk(msg)
}

// Array compares two array. The b slice may be nil indicating that all values are zero
func Array(tst *testing.T, msg string, tol float64, a, b []float64) {
	zero := false
	if len(b) == 0 {
		zero = true
	} else {
		if len(a) != len(b) {
			TstFail(tst, "%s len(a)=%d != len(b)=%d", msg, len(a), len(b))
			return
		}
	}
	for i := 0; i < len(a); i++ {
		var c float64
		if !zero {
			c = b[i]
		}
		if TstDiff(tst, msg+fmt.Sprintf(" [%d] ", i), tol, a[i], c, false) {
			return
		}
	}
	PrintOk(msg)
}

// ArrayC compares two slices of complex nummber. The b slice may be nil indicating that all values are zero
func ArrayC(tst *testing.T, msg string, tol float64, a, b []complex128) {
	zero := false
	if len(b) == 0 {
		zero = true
	} else {
		if len(a) != len(b) {
			TstFail(tst, "%s len(a)=%d != len(b)=%d", msg, len(a), len(b))
			return
		}
	}
	for i := 0; i < len(a); i++ {
		var c complex128
		if !zero {
			c = b[i]
		}
		if TestDiffC(tst, msg, tol, a[i], c, false) {
			return
		}
	}
	PrintOk(msg)
}

// Deep2 compares two nested (depth=2) slice. The b slice may be nil indicating that all values are zero
func Deep2(tst *testing.T, msg string, tol float64, a, b [][]float64) {
	zero := false
	if len(b) == 0 {
		zero = true
	} else {
		if len(a) != len(b) {
			TstFail(tst, "%s len(a)=%d != len(b)=%d", msg, len(a), len(b))
			return
		}
	}
	for i := 0; i < len(a); i++ {
		if !zero {
			if len(a[i]) != len(b[i]) {
				TstFail(tst, "%s len(a[%d])=%d != len(b[%d])=%d", msg, i, len(a[i]), i, len(b[i]))
				return
			}
		}
		for j := 0; j < len(a[i]); j++ {
			var c float64
			if !zero {
				c = b[i][j]
			}
			if TstDiff(tst, msg+fmt.Sprintf(" [%d][%d] ", i, j), tol, a[i][j], c, false) {
				return
			}
		}
	}
	PrintOk(msg)
}

// Deep2c compares two nested (depth=2) slices. The b slice may be nil indicating that all values are zero
func Deep2c(tst *testing.T, msg string, tol float64, a, b [][]complex128) {
	zero := false
	if len(b) == 0 {
		zero = true
	} else {
		if len(a) != len(b) {
			TstFail(tst, "%s len(a)=%d != len(b)=%d", msg, len(a), len(b))
			return
		}
	}
	for i := 0; i < len(a); i++ {
		if !zero {
			if len(a[i]) != len(b[i]) {
				TstFail(tst, "%s len(a[%d])=%d != len(b[%d])=%d", msg, i, len(a[i]), i, len(b[i]))
				return
			}
		}
		for j := 0; j < len(a[i]); j++ {
			var c complex128
			if !zero {
				c = b[i][j]
			}
			if TestDiffC(tst, msg+fmt.Sprintf(" [%d][%d] ", i, j), tol, a[i][j], c, false) {
				return
			}
		}
	}
	PrintOk(msg)
}

// StrDeep2 compares nested slices of strings. The b slice may be nil indicating that all values are zero
func StrDeep2(tst *testing.T, msg string, a, b [][]string) {
	empty := false
	if len(b) == 0 {
		empty = true
	} else {
		if len(a) != len(b) {
			TstFail(tst, "%s len(a)=%d != len(b)=%d", msg, len(a), len(b))
			return
		}
	}
	for i := 0; i < len(a); i++ {
		if !empty {
			if len(a[i]) != len(b[i]) {
				TstFail(tst, "%s len(a[%d])=%d != len(b[%d])=%d", msg, i, len(a[i]), i, len(b[i]))
				return
			}
		}
		for j := 0; j < len(a[i]); j++ {
			var c string
			if !empty {
				c = b[i][j]
			}
			if a[i][j] != c {
				TstFail(tst, "%s [%d][%d] %v != %v", msg, i, j, a[i][j], c)
				return
			}
		}
	}
	PrintOk(msg)
}

// IntDeep2 compares nested slices of ints. The b slice may be nil indicating that all values are zero
func IntDeep2(tst *testing.T, msg string, a, b [][]int) {
	zero := false
	if len(b) == 0 {
		zero = true
	} else {
		if len(a) != len(b) {
			TstFail(tst, "%s len(a)=%d != len(b)=%d", msg, len(a), len(b))
			return
		}
	}
	for i := 0; i < len(a); i++ {
		if !zero {
			if len(a[i]) != len(b[i]) {
				TstFail(tst, "%s len(a[%d])=%d != len(b[%d])=%d", msg, i, len(a[i]), i, len(b[i]))
				return
			}
		}
		for j := 0; j < len(a[i]); j++ {
			var c int
			if !zero {
				c = b[i][j]
			}
			if a[i][j] != c {
				TstFail(tst, "%s [%d][%d] %v != %v", msg, i, j, a[i][j], c)
				return
			}
		}
	}
	PrintOk(msg)
}

// Deep3 compares two deep3 slices. The b slice may be nil indicating that all values are zero
func Deep3(tst *testing.T, msg string, tol float64, a, b [][][]float64) {
	zero := false
	if len(b) == 0 {
		zero = true
	} else {
		if len(a) != len(b) {
			TstFail(tst, "%s len(a)=%d != len(b)=%d", msg, len(a), len(b))
			return
		}
	}
	for i := 0; i < len(a); i++ {
		if !zero {
			if len(a[i]) != len(b[i]) {
				TstFail(tst, "%s len(a[%d])=%d != len(b[%d])=%d", msg, i, len(a[i]), i, len(b[i]))
				return
			}
		}
		for j := 0; j < len(a[i]); j++ {
			if !zero {
				if len(a[i][j]) != len(b[i][j]) {
					TstFail(tst, "%s len(a[%d][%d])=%d != len(b[%d][%d])=%d", msg, i, j, len(a[i][j]), i, j, len(b[i][j]))
					return
				}
			}
			for k := 0; k < len(a[i][j]); k++ {
				var c float64
				if !zero {
					c = b[i][j][k]
				}
				if TstDiff(tst, msg+fmt.Sprintf(" [%d][%d][%d] ", i, j, k), tol, a[i][j][k], c, false) {
					return
				}
			}
		}
	}
	PrintOk(msg)
}

// Deep4 compares two deep4 slices. The b slice may be nil indicating that all values are zero
func Deep4(tst *testing.T, msg string, tol float64, a, b [][][][]float64) {
	zero := false
	if len(b) == 0 {
		zero = true
	} else {
		if len(a) != len(b) {
			TstFail(tst, "%s len(a)=%d != len(b)=%d", msg, len(a), len(b))
			return
		}
	}
	for i := 0; i < len(a); i++ {
		if !zero {
			if len(a[i]) != len(b[i]) {
				TstFail(tst, "%s len(a[%d])=%d != len(b[%d])=%d", msg, i, len(a[i]), i, len(b[i]))
				return
			}
		}
		for j := 0; j < len(a[i]); j++ {
			if !zero {
				if len(a[i][j]) != len(b[i][j]) {
					TstFail(tst, "%s len(a[%d][%d])=%d != len(b[%d][%d])=%d", msg, i, j, len(a[i][j]), i, j, len(b[i][j]))
					return
				}
			}
			for k := 0; k < len(a[i][j]); k++ {
				if !zero {
					if len(a[i][j][k]) != len(b[i][j][k]) {
						TstFail(tst, "%s len(a[%d][%d][%d])=%d != len(b[%d][%d][%d])=%d", msg, i, j, k, len(a[i][j][k]), i, j, k, len(b[i][j][k]))
						return
					}
				}
				for l := 0; l < len(a[i][j][k]); l++ {
					var c float64
					if !zero {
						c = b[i][j][k][l]
					}
					if TstDiff(tst, msg+fmt.Sprintf(" [%d][%d][%d][%d] ", i, j, k, l), tol, a[i][j][k][l], c, false) {
						return
					}
				}
			}
		}
	}
	PrintOk(msg)
}

// Symmetry checks symmetry of SEGMENTS in an even or odd slice of float64
//
//   NOTE: values in X must be sorted ascending
//
func Symmetry(tst *testing.T, msg string, X []float64) {

	// some constants
	npts := len(X)
	l := npts - 1 // last
	even := l%2 == 0
	imax := l/2 + 1
	if !even {
		imax = (l + 1) / 2
	}

	// check symmetry
	for i := 1; i < imax; i++ {
		Δxa := X[i] - X[i-1]
		Δxb := X[l-i+1] - X[l-i]
		AnaNum(tst, msg+": Δxa = Δxb", 0, Δxa, Δxb, Verbose)
		if Δxa != Δxb {
			TstFail(tst, msg+": Δxa must be exactly equal to Δxb")
			return
		}
	}
}
