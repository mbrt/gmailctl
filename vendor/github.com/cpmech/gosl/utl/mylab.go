// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package utl implements functions for simplifying calculations and allocation of structures
// such as slices and slices of slices. It also contains functions for sorting quantities.
package utl

import (
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/cpmech/gosl/chk"
	"github.com/cpmech/gosl/io"
)

// functions /////////////////////////////////////////////////////////////////////////////////////

// Digits returns the nubmer of digits
func Digits(maxint int) (ndigits int, format string) {
	ndigits = int(math.Log10(float64(maxint))) + 1
	format = io.Sf("%%%dd", ndigits)
	return
}

// Expon returns the exponent
func Expon(val float64) (ndigits int) {
	if val == 0.0 {
		return
	}
	ndigits = int(math.Log10(math.Abs(val)))
	return
}

// slices of string //////////////////////////////////////////////////////////////////////////////

// StrVals allocates a slice of strings with size==n, filled with val
func StrVals(n int, val string) (s []string) {
	s = make([]string, n)
	for i := 0; i < n; i++ {
		s[i] = val
	}
	return
}

// StrAlloc allocates a matrix of strings
func StrAlloc(m, n int) (mat [][]string) {
	mat = make([][]string, m)
	for i := 0; i < m; i++ {
		mat[i] = make([]string, n)
	}
	return
}

// slices of int /////////////////////////////////////////////////////////////////////////////////

// IntFill fills a slice of integers
func IntFill(s []int, val int) {
	for i := 0; i < len(s); i++ {
		s[i] = val
	}
}

// IntVals allocates a slice of integers with size==n, filled with val
func IntVals(n int, val int) (s []int) {
	s = make([]int, n)
	for i := 0; i < n; i++ {
		s[i] = val
	}
	return
}

// IntAlloc allocates a matrix of integers
func IntAlloc(m, n int) (mat [][]int) {
	mat = make([][]int, m)
	for i := 0; i < m; i++ {
		mat[i] = make([]int, n)
	}
	return
}

// IntCopy returns a copy of slice of ints
func IntCopy(in []int) (out []int) {
	out = make([]int, len(in))
	copy(out, in)
	return
}

// IntClone allocates and clones a matrix of integers
func IntClone(a [][]int) (b [][]int) {
	b = make([][]int, len(a))
	for i := 0; i < len(a); i++ {
		b[i] = make([]int, len(a[i]))
		for j := 0; j < len(a[i]); j++ {
			b[i][j] = a[i][j]
		}
	}
	return
}

// IntRange generates a slice of integers from 0 to n-1
func IntRange(n int) (res []int) {
	if n <= 0 {
		return
	}
	res = make([]int, n)
	for i := 0; i < n; i++ {
		res[i] = i
	}
	return
}

// IntRange2 generates slice of integers from start to stop (but not stop)
func IntRange2(start, stop int) []int {
	return IntRange3(start, stop, 1)
}

// IntRange3 generates a slice of integers from start to stop (but not stop), afer each 'step'
func IntRange3(start, stop, step int) (res []int) {
	switch {
	case stop == start:
		return
	case stop > start:
		n := (stop - start) / step
		stop = start + step*n
		res = make([]int, n)
		for i, v := 0, start; v < stop; i, v = i+1, v+step {
			res[i] = v
		}
	case stop < start:
		if step > 0 {
			return
		}
		n := (stop - start) / step
		stop = start + step*n
		res = make([]int, n)
		for i, v := 0, start; v > stop; i, v = i+1, v+step {
			res[i] = v
		}
	}
	return
}

// IntAddScalar adds a scalar to all values in a slice of integers
func IntAddScalar(a []int, s int) (res []int) {
	res = make([]int, len(a))
	for i := 0; i < len(a); i++ {
		res[i] = a[i] + s
	}
	return
}

// IntUnique returns a unique and sorted slice of integers
func IntUnique(slices ...[]int) (res []int) {
	if len(slices) == 0 {
		return
	}
	nn := 0
	for i := 0; i < len(slices); i++ {
		nn += len(slices[i])
	}
	res = make([]int, 0, nn)
	for i := 0; i < len(slices); i++ {
		a := make([]int, len(slices[i]))
		copy(a, slices[i])
		sort.Ints(a)
		for j := 0; j < len(a); j++ {
			idx := sort.SearchInts(res, a[j])
			if idx < len(res) && res[idx] == a[j] {
				continue // found
			} else {
				if idx == len(res) { // append
					res = append(res, a[j])
				} else { // insert
					res = append(res[:idx], append([]int{a[j]}, res[idx:]...)...)
				}
			}
		}
	}
	return
}

// IntPy returns a Python string representing a slice of integers
func IntPy(a []int) (res string) {
	res = "["
	for i := 0; i < len(a); i++ {
		res += strconv.Itoa(a[i])
		if i < len(a)-1 {
			res += ", "
		}
	}
	res += "]"
	return
}

// slices of float64 /////////////////////////////////////////////////////////////////////////////

// GetMapped returns a new slice such that: Y[i] = filter(X[i])
func GetMapped(X []float64, filter func(x float64) float64) (Y []float64) {
	Y = make([]float64, len(X))
	for i := 0; i < len(X); i++ {
		Y[i] = filter(X[i])
	}
	return
}

// GetMapped2 returns a new slice of slice such that: Y[i][j] = filter(X[i][j])
//  NOTE: each row in X may have different number of columns; i.e. len(X[0]) may be != len(X[1])
func GetMapped2(X [][]float64, filter func(x float64) float64) (Y [][]float64) {
	Y = make([][]float64, len(X))
	for i := 0; i < len(X); i++ {
		Y[i] = make([]float64, len(X[i]))
		for j := 0; j < len(X[i]); j++ {
			Y[i][j] = filter(X[i][j])
		}
	}
	return
}

// Alloc allocates a slice of slices of float64
func Alloc(m, n int) (mat [][]float64) {
	mat = make([][]float64, m)
	for i := 0; i < m; i++ {
		mat[i] = make([]float64, n)
	}
	return
}

// Fill fills a slice of float64
func Fill(s []float64, val float64) {
	for i := 0; i < len(s); i++ {
		s[i] = val
	}
}

// Ones generates a slice of float64 with ones
func Ones(n int) (res []float64) {
	res = make([]float64, n)
	for i := 0; i < n; i++ {
		res[i] = 1.0
	}
	return
}

// Vals generates a slice of float64 filled with v
func Vals(n int, v float64) (res []float64) {
	res = make([]float64, n)
	for i := 0; i < n; i++ {
		res[i] = v
	}
	return
}

// GetCopy gets a copy of slice of float64
func GetCopy(in []float64) (out []float64) {
	out = make([]float64, len(in))
	copy(out, in)
	return
}

// GetReversed return a copy with reversed items
func GetReversed(in []float64) (out []float64) {
	n := len(in)
	out = make([]float64, n)
	for i := 0; i < n; i++ {
		out[n-1-i] = in[i]
	}
	return
}

// Clone allocates and clones a matrix of float64
func Clone(a [][]float64) (b [][]float64) {
	b = make([][]float64, len(a))
	for i := 0; i < len(a); i++ {
		b[i] = make([]float64, len(a[i]))
		for j := 0; j < len(a[i]); j++ {
			b[i][j] = a[i][j]
		}
	}
	return
}

// LinSpace returns evenly spaced numbers over a specified closed interval.
func LinSpace(start, stop float64, num int) (res []float64) {
	if num <= 0 {
		return []float64{}
	}
	if num == 1 {
		return []float64{start}
	}
	step := (stop - start) / float64(num-1)
	res = make([]float64, num)
	res[0] = start
	for i := 1; i < num; i++ {
		res[i] = start + float64(i)*step
	}
	res[num-1] = stop
	return
}

// LinSpaceOpen returns evenly spaced numbers over a specified open interval.
func LinSpaceOpen(start, stop float64, num int) (res []float64) {
	if num <= 0 {
		return []float64{}
	}
	step := (stop - start) / float64(num)
	res = make([]float64, num)
	res[0] = start
	for i := 1; i < num; i++ {
		res[i] = start + float64(i)*step
	}
	return
}

// NonlinSpace generates N points such that the ratio between the last segment (or the middle
// segment) to the first one is equal to a given constant R.
//
//   The ratio between the last (or middle largest) segment to the first one is:
//
//       ΔxL   = Δx0 ⋅ R
//
//   The ratio between successive segments is
//
//                k-1
//       Δx[k] = α    ⋅ Δx[0]
//
//   Unsymmetricc case:
//
//     |--|-------|------------|-----------------|
//      Δx0                            ΔxL
//
//   Symmetric case with odd number of spacings:
//
//     |---|--------|---------------|--------|---|
//      Δx0                ΔxL                Δx0
//
//   Symmetric case with even number of spacings:
//
//     |---|----------------|----------------|---|
//      Δx0       ΔxL               ΔxL       Δx0
//
func NonlinSpace(xa, xb float64, N int, R float64, symmetric bool) (x []float64) {

	// initialise data
	if N < 2 {
		N = 2
	}
	x = make([]float64, N)
	l := N - 1
	x[0] = xa
	x[l] = xb
	if N == 2 {
		return
	}

	// uniform grid
	if R == 1.0 {
		return LinSpace(xa, xb, N)
	}

	// symmetric grid
	if symmetric {

		// even number of segments
		if l%2 == 0 {
			if N == 3 {
				x[1] = (xa + xb) / 2.0
				x[2] = xb
				return
			}
			nh := float64((N - 1) / 2)
			xh := 0.5 * (xa + xb)
			α := math.Pow(R, 1.0/(nh-1.0))
			m := (1.0 - α) / (1.0 - math.Pow(α, nh))
			δx := (xh - xa) * m
			for i := 1; i < l/2+1; i++ {
				x[i] = x[i-1] + δx
				if i < l/2 { // adding backwards
					x[l-i] = x[N-i] - δx
				}
				δx *= α
			}

			// odd number of segments
		} else {
			n := float64(N - 1)
			α := math.Pow(R, 2.0/(n-1.0))
			m := (1.0 - α) / (2.0 - math.Pow(α, (n+1.0)/2.0) - math.Pow(α, (n-1.0)/2.0))
			δx := (xb - xa) * m
			for i := 1; i < (N+2)/2; i++ {
				x[i] = x[i-1] + δx
				if i < (N+2)/2-1 { // adding backwards
					x[l-i] = x[N-i] - δx
				}
				δx *= α
			}
		}
		return
	}

	// unsymmetric grid
	n := float64(N - 1)
	α := math.Pow(R, 1.0/(n-1.0))
	m := (1.0 - α) / (1.0 - math.Pow(α, n))
	δx := (xb - xa) * m
	for i := 1; i < l; i++ {
		x[i] = x[i-1] + δx
		δx *= α
	}
	return
}

// ToStrings converts a slice of float64 to a slice of strings
func ToStrings(v []float64, format string) (s []string) {
	s = make([]string, len(v))
	for i := 0; i < len(v); i++ {
		s[i] = io.Sf(format, v[i])
	}
	return
}

// FromStrings converts a slice of strings to a slice of float64
func FromStrings(s []string) (v []float64) {
	v = make([]float64, len(s))
	for i := 0; i < len(s); i++ {
		v[i] = io.Atof(s[i])
	}
	return
}

// FromString splits a string with numbers separeted by spaces into float64
func FromString(s string) (r []float64) {
	ss := strings.Fields(s)
	r = make([]float64, len(ss))
	for i, v := range ss {
		r[i] = io.Atof(v)
	}
	return
}

// meshgrid //////////////////////////////////////////////////////////////////////////////////////

// MeshGrid2d creates a grid with x-y coordinates
//  X, Y -- [ny][nx]
func MeshGrid2d(xmin, xmax, ymin, ymax float64, nx, ny int) (X, Y [][]float64) {
	if nx < 2 || ny < 2 {
		return
	}
	dx := (xmax - xmin) / float64(nx-1)
	dy := (ymax - ymin) / float64(ny-1)
	X = make([][]float64, ny)
	Y = make([][]float64, ny)
	for i := 0; i < ny; i++ {
		X[i] = make([]float64, nx)
		Y[i] = make([]float64, nx)
		for j := 0; j < nx; j++ {
			X[i][j] = xmin + float64(j)*dx
			Y[i][j] = ymin + float64(i)*dy
		}
	}
	return
}

// MeshGrid2dF creates a grid with x-y coordinates and evaluates z=f(x,y)
//  X, Y, Z -- [ny][nx]
func MeshGrid2dF(xmin, xmax, ymin, ymax float64, nx, ny int, f func(x, y float64) float64) (X, Y, Z [][]float64) {
	if nx < 2 || ny < 2 {
		return
	}
	dx := (xmax - xmin) / float64(nx-1)
	dy := (ymax - ymin) / float64(ny-1)
	X = make([][]float64, ny)
	Y = make([][]float64, ny)
	Z = make([][]float64, ny)
	for i := 0; i < ny; i++ {
		X[i] = make([]float64, nx)
		Y[i] = make([]float64, nx)
		Z[i] = make([]float64, nx)
		for j := 0; j < nx; j++ {
			X[i][j] = xmin + float64(j)*dx
			Y[i][j] = ymin + float64(i)*dy
			Z[i][j] = f(X[i][j], Y[i][j])
		}
	}
	return
}

// MeshGrid2dFG creates a grid with x-y coordinates and evaluates (z,u,v)=fg(x,y)
//  X, Y, Z, U, V -- [ny][nx]
func MeshGrid2dFG(xmin, xmax, ymin, ymax float64, nx, ny int, fg func(x, y float64) (z, u, v float64)) (X, Y, Z, U, V [][]float64) {
	if nx < 2 || ny < 2 {
		return
	}
	dx := (xmax - xmin) / float64(nx-1)
	dy := (ymax - ymin) / float64(ny-1)
	X = make([][]float64, ny)
	Y = make([][]float64, ny)
	Z = make([][]float64, ny)
	U = make([][]float64, ny)
	V = make([][]float64, ny)
	for i := 0; i < ny; i++ {
		X[i] = make([]float64, nx)
		Y[i] = make([]float64, nx)
		Z[i] = make([]float64, nx)
		U[i] = make([]float64, nx)
		V[i] = make([]float64, nx)
		for j := 0; j < nx; j++ {
			X[i][j] = xmin + float64(j)*dx
			Y[i][j] = ymin + float64(i)*dy
			Z[i][j], U[i][j], V[i][j] = fg(X[i][j], Y[i][j])
		}
	}
	return
}

// MeshGrid2dV creates a grid with x-y coordinates given x and y values
//  X, Y -- [len(yVals)][len(xVals)]
func MeshGrid2dV(xVals, yVals []float64) (X, Y [][]float64) {
	nx, ny := len(xVals), len(yVals)
	if nx < 2 || ny < 2 {
		return
	}
	X = make([][]float64, ny)
	Y = make([][]float64, ny)
	for i := 0; i < ny; i++ {
		X[i] = make([]float64, nx)
		Y[i] = make([]float64, nx)
		for j := 0; j < nx; j++ {
			X[i][j] = xVals[j]
			Y[i][j] = yVals[i]
		}
	}
	return
}

// more functions ////////////////////////////////////////////////////////////////////////////////

// Scaling computes a scaled version of the input slice with results in [0.0, 1.0]
//  Input:
//   x       -- values
//   ds      -- δs value to be added to all 's' values
//   tol     -- tolerance for capturing xmax ≅ xmin
//   reverse -- compute reverse series;
//              i.e. 's' decreases from 1 to 0 while x goes from xmin to xmax
//   useinds -- if (xmax-xmin)<tol, use indices to generate the 's' slice;
//              otherwise, 's' will be filled with δs + zeros
//  Ouptut:
//   s          -- scaled series; pre--allocated with len(s) == len(x)
//   xmin, xmax -- min(x) and max(x)
func Scaling(s, x []float64, ds, tol float64, reverse, useinds bool) (xmin, xmax float64) {
	if len(x) < 2 {
		return
	}
	n := len(x)
	chk.IntAssert(len(s), n)
	xmin, xmax = x[0], x[0]
	for i := 1; i < n; i++ {
		if x[i] < xmin {
			xmin = x[i]
		}
		if x[i] > xmax {
			xmax = x[i]
		}
	}
	dx := xmax - xmin
	if dx < tol {
		if useinds {
			N := float64(n - 1)
			if reverse {
				for i := 0; i < n; i++ {
					s[i] = ds + float64(n-1-i)/N
				}
				return
			}
			for i := 0; i < n; i++ {
				s[i] = ds + float64(i)/N
			}
			return
		}
		for i := 0; i < n; i++ {
			s[i] = ds
		}
		return
	}
	if reverse {
		for i := 0; i < n; i++ {
			s[i] = ds + (xmax-x[i])/dx
		}
		return
	}
	for i := 0; i < n; i++ {
		s[i] = ds + (x[i]-xmin)/dx
	}
	return
}

// CumSum returns the cumulative sum of the elements in p
//  Input:
//   p -- values
//  Output:
//   cs -- cumulated sum; pre-allocated with len(cs) == len(p)
func CumSum(cs, p []float64) {
	if len(p) < 1 {
		return
	}
	chk.IntAssert(len(cs), len(p))
	cs[0] = p[0]
	for i := 1; i < len(p); i++ {
		cs[i] = cs[i-1] + p[i]
	}
}

// GtPenalty implements a 'greater than' penalty function where
// x must be greater than b; otherwise the error is magnified
func GtPenalty(x, b, penaltyM float64) float64 {
	if x > b {
		return 0.0
	}
	return penaltyM*(b-x) + 1e-16 // must add small number because x must be greater than b
}

// GtePenalty implements a 'greater than or equal' penalty function where
// x must be greater than b or equal to be; otherwise the error is magnified
func GtePenalty(x, b, penaltyM float64) float64 {
	if x >= b {
		return 0.0
	}
	return penaltyM * (b - x)
}

// GetColumn returns the column of a matrix of float64
func GetColumn(j int, v [][]float64) (x []float64) {
	x = make([]float64, len(v))
	for i := 0; i < len(v); i++ {
		x[i] = v[i][j]
	}
	return
}

// L2norm returns the Euclidean distance between p and q
func L2norm(p, q []float64) (dist float64) {
	for i := 0; i < len(p); i++ {
		dist += math.Pow(p[i]-q[i], 2.0)
	}
	return math.Sqrt(dist)
}

// Dot3d returns the dot product between two 3D vectors
func Dot3d(u, v []float64) (s float64) {
	return u[0]*v[0] + u[1]*v[1] + u[2]*v[2]
}

// Cross3d computes the cross product of two 3D vectors u and w
//  w = u cross v
//  Note: w must be pre-allocated
func Cross3d(w, u, v []float64) {
	w[0] = u[1]*v[2] - u[2]*v[1]
	w[1] = u[2]*v[0] - u[0]*v[2]
	w[2] = u[0]*v[1] - u[1]*v[0]
}

// ArgMinMax finds the indices of min and max arguments
func ArgMinMax(v []float64) (imin, imax int) {
	if len(v) < 1 {
		return
	}
	vmin, vmax := v[0], v[0]
	imin, imax = 0, 0
	for i := 1; i < len(v); i++ {
		if v[i] < vmin {
			imin = i
			vmin = v[i]
		}
		if v[i] > vmax {
			imax = i
			vmax = v[i]
		}
	}
	return
}

// FromInts returns a new slice of float64 from a slice of ints
func FromInts(a []int) (b []float64) {
	b = make([]float64, len(a))
	for i, x := range a {
		b[i] = float64(x)
	}
	return
}

// FromFloat64s returns a new slice of int from a slice of float64
func FromFloat64s(a []float64) (b []int) {
	b = make([]int, len(a))
	for i, x := range a {
		b[i] = int(x)
	}
	return
}

// bool //////////////////////////////////////////////////////////////////////////////////////////

// AllTrue returns true if all values are true
func AllTrue(values []bool) bool {
	for _, v := range values {
		if !v {
			return false
		}
	}
	return true
}

// AllFalse returns true if all values are false
func AllFalse(values []bool) bool {
	for _, v := range values {
		if v {
			return false
		}
	}
	return true
}
