// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utl

import "github.com/cpmech/gosl/chk"

// Deep3alloc allocates a slice of slice of slice
func Deep3alloc(n1, n2, n3 int) (a [][][]float64) {
	a = make([][][]float64, n1)
	for i := 0; i < n1; i++ {
		a[i] = make([][]float64, n2)
		for j := 0; j < n2; j++ {
			a[i][j] = make([]float64, n3)
		}
	}
	return
}

// Deep4alloc allocates a slice of slice of slice of slice
func Deep4alloc(n1, n2, n3, n4 int) (a [][][][]float64) {
	a = make([][][][]float64, n1)
	for i := 0; i < n1; i++ {
		a[i] = make([][][]float64, n2)
		for j := 0; j < n2; j++ {
			a[i][j] = make([][]float64, n3)
			for k := 0; k < n3; k++ {
				a[i][j][k] = make([]float64, n4)
			}
		}
	}
	return
}

// Deep3set sets deep slice of slice of slice with v values
func Deep3set(a [][][]float64, v float64) {
	for i := 0; i < len(a); i++ {
		for j := 0; j < len(a[i]); j++ {
			for k := 0; k < len(a[i][j]); k++ {
				a[i][j][k] = v
			}
		}
	}
}

// Deep4set sets deep slice of slice of slice of slice with v values
func Deep4set(a [][][][]float64, v float64) {
	for i := 0; i < len(a); i++ {
		for j := 0; j < len(a[i]); j++ {
			for k := 0; k < len(a[i][j]); k++ {
				for l := 0; l < len(a[i][j][k]); l++ {
					a[i][j][k][l] = v
				}
			}
		}
	}
}

// Deep2checkSize checks if dimensions of Deep2 slice are correct
func Deep2checkSize(n1, n2 int, a [][]float64) bool {
	if len(a) != n1 {
		return false
	}
	if n1 == 0 {
		return true
	}
	if len(a[0]) != n2 {
		return false
	}
	return true
}

// Deep3checkSize checks if dimensions of Deep3 slice are correct
func Deep3checkSize(n1, n2, n3 int, a [][][]float64) bool {
	if len(a) != n1 {
		return false
	}
	if n1 == 0 {
		return true
	}
	if len(a[0]) != n2 {
		return false
	}
	if n2 == 0 {
		return true
	}
	if len(a[0][0]) != n3 {
		return false
	}
	return true
}

// Deep2transpose returns the transpose of a deep2 slice
func Deep2transpose(a [][]float64) (aT [][]float64) {
	if len(a) < 1 {
		chk.Panic("input Deep2 slice must be greater than (1,1)\n")
	}
	m, n := len(a), len(a[0])
	aT = Alloc(n, m)
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			aT[j][i] = a[i][j]
		}
	}
	return
}
