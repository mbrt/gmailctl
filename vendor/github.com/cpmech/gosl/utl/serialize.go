// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utl

import "github.com/cpmech/gosl/io"

// SerializeDeep2 converts a matrix into a column-major array
func SerializeDeep2(a [][]float64) (v []float64) {
	m, n, k := len(a), len(a[0]), 0
	v = make([]float64, m*n)
	for j := 0; j < n; j++ {
		for i := 0; i < m; i++ {
			v[k] = a[i][j]
			k++
		}
	}
	return
}

// DeserializeDeep2 converts a column-major array to a matrix
func DeserializeDeep2(v []float64, m, n int) (a [][]float64) {
	a = make([][]float64, m)
	for i := 0; i < m; i++ {
		a[i] = make([]float64, n)
	}
	for k := 0; k < m*n; k++ {
		i, j := k%m, k/m
		a[i][j] = v[k]
	}
	return
}

// SerializeDeep3 serializes an array of array of array in column-compressed format
func SerializeDeep3(A [][][]float64) (I, P []int, S []float64) {
	i, p := 0, 0
	for _, a := range A {
		for _, b := range a {
			i++
			p += len(b)
		}
	}
	I = make([]int, i)
	P = make([]int, i+1)
	S = make([]float64, p)
	i, p, k := 0, 0, 0
	for j, a := range A {
		for _, b := range a {
			I[i] = j
			i++
			p += len(b)
			P[i] = p
			for _, v := range b {
				S[k] = v
				k++
			}
		}
	}
	return
}

// Deep3GetInfo returns information of serialized array of array of array
func Deep3GetInfo(I, P []int, S []float64, verbose bool) (nitems, nrows, ncolsTot int, ncols []int) {
	if verbose {
		io.Pf("I = %v\n", I)
		io.Pf("P = %v\n", P)
		io.Pf("S = %v\n", S)
	}
	nitems = P[len(P)-1]
	nrows = I[len(I)-1] + 1
	ncols = make([]int, nrows)
	for _, j := range I {
		ncolsTot++
		ncols[j]++
	}
	if verbose {
		io.Pf("nitems    = %v\n", nitems)
		io.Pf("nrows     = %v\n", nrows)
		io.Pf("ncols_tot = %v\n", ncolsTot)
		io.Pf("ncols     = %v\n", ncols)
	}
	return
}

// DeserializeDeep3 deserializes an array of array of array in column-compressed format
func DeserializeDeep3(I, P []int, S []float64, debug bool) (A [][][]float64) {
	_, nrows, _, ncols := Deep3GetInfo(I, P, S, false)
	A = make([][][]float64, nrows)
	for i := 0; i < nrows; i++ {
		A[i] = make([][]float64, ncols[i])
	}
	iprev := 0 // previous i
	j := 0     // column index
	for l, i := range I {
		nitems := P[l+1] - P[l]
		if i != iprev { // jumped to new column
			j = 0
		}
		if debug {
			io.Pf("l=%v  i=%v  nitems=%v  j=%v\n", l, i, nitems, j)
		}
		for k, p := 0, P[l]; p < P[l+1]; k, p = k+1, p+1 {
			if debug {
				io.Pf("  k=%v  p=%v  s=%v\n", k, p, S[p])
			}
			if k == 0 {
				A[i][j] = make([]float64, nitems)
			}
			A[i][j][k] = S[p]
		}
		iprev = i
		j++
	}
	return
}
