// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chk

import (
	"fmt"
	"math"
	"testing"
)

// DerivVecSca checks the derivative of vector w.r.t scalar by comparing with numerical solution
// obtained with central differences (5-point rule)
//   Check:
//              d{f} │             {f}:vector   x:scalar
//        {g} = ———— │      with   {g}:vector
//               dx  │xAt          len(g) == len(f)
//   Input:
//     tst  -- testing.T structure
//     msg  -- message about this test
//     tol  -- tolerance to compare gAna with gNum
//     gAna -- [vector] analytical (or other kind) derivative dfdx. size=len(f)
//     xAt  -- [scalar] position to compute dfdx
//     h    -- initial stepsize; e.g. 1e-1
//     verb -- verbose: show messages
//     fcn  -- [vector] function f(x). x is scalar
func DerivVecSca(tst *testing.T, msg string, tol float64, gAna []float64, xAt, h float64,
	verb bool, fcn func(f []float64, x float64)) {

	// run test
	f := make([]float64, len(gAna))
	for i, ana := range gAna {
		fi := func(x float64) float64 {
			fcn(f, x)
			return f[i]
		}
		DerivScaSca(tst, fmt.Sprintf("%s%d", msg, i), tol, ana, xAt, h, verb, fi)
	}
}

// DerivScaVec checks the derivative of scalar w.r.t vector by comparing with numerical solution
// obtained with central differences (5-point rule)
//   Check:
//               df  │               f:scalar   {x}:vector
//        {g} = ———— │        with   {g}:vector
//              d{x} │{xAt}          len(g) == len(x) == len(xAt)
//   Input:
//     tst  -- testing.T structure
//     msg  -- message about this test
//     tol  -- tolerance to compare gAna with gNum
//     gAna -- [vector] analytical (or other kind) derivative dfdx. size=len(x)=len(xAt)
//     xAt  -- [vector] position to compute dfdx
//     h    -- initial stepsize; e.g. 1e-1
//     verb -- verbose: show messages
//     fcn  -- [scalar] function f(x). x is vector
func DerivScaVec(tst *testing.T, msg string, tol float64, gAna, xAt []float64, h float64,
	verb bool, fcn func(x []float64) float64) {

	// check input
	ndim := len(xAt)
	if len(gAna) != ndim {
		tst.Errorf("length of gAna vector must be equal to the length of xAt vector. %d != %d", len(gAna), ndim)
		return
	}

	// run test
	xTmp := make([]float64, ndim)
	for i := 0; i < ndim; i++ {
		copy(xTmp, xAt)
		fi := func(x float64) float64 {
			xTmp[i] = x
			return fcn(xTmp)
		}
		DerivScaSca(tst, fmt.Sprintf("%s%d", msg, i), tol, gAna[i], xAt[i], h, verb, fi)
	}
}

// DerivVecVec checks the derivative of vector w.r.t vector by comparing with numerical solution
// obtained with central differences (5-point rule)
//   Checks:
//              d{f} │               {f}:vector   {x}:vector
//        [g] = ———— │        with   [g]:matrix
//              d{x} │{xAt}          rows(g)==len(f)  cols(g)==len(x)==len(xAt)
//   Input:
//     tst  -- testing.T structure
//     msg  -- message about this test
//     tol  -- tolerance to compare gAna with gNum
//     gAna -- [matrix] analytical (or other kind) derivative dfdx. size=(len(f),len(x))
//     xAt  -- [vector] position to compute dfdx
//     h    -- initial stepsize; e.g. 1e-1
//     verb -- verbose: show messages
//     fcn  -- [vector] function f(x). x is vector
func DerivVecVec(tst *testing.T, msg string, tol float64, gAna [][]float64, xAt []float64, h float64,
	verb bool, fcn func(f, x []float64)) {

	// check input
	nrow := len(gAna)
	if nrow < 1 {
		tst.Errorf("number of rows of gAna matrix must be greater than or equal to 1\n")
		return
	}
	ncol := len(gAna[0])
	ndim := len(xAt)
	if ncol != ndim {
		tst.Errorf("number of columns in gAna matrix must be equal to len(xAt). %d != %d\n", ncol, ndim)
		return
	}

	// run test
	f := make([]float64, nrow)
	xTmp := make([]float64, ndim)
	for i := 0; i < nrow; i++ {
		ncol = len(gAna[i])
		if ncol != ndim {
			tst.Errorf("number of columns in gAna matrix must be equal to len(xAt). %d != %d\n", ncol, ndim)
			return
		}
		for j := 0; j < ndim; j++ {
			copy(xTmp, xAt)
			fij := func(x float64) float64 {
				xTmp[j] = x
				fcn(f, xTmp)
				return f[i]
			}
			DerivScaSca(tst, fmt.Sprintf("%s%d%d", msg, i, j), tol, gAna[i][j], xAt[j], h, verb, fij)
		}
	}
}

// DerivScaSca checks the derivative of scalar w.r.t scalar by comparing with numerical solution
// obtained with central differences (5-point rule)
//   Checks:
//             df │
//         g = —— │      with   f:scalar,  x:scalar
//             dx │xAt          g:scalar
//   Input:
//     tst  -- testing.T structure
//     msg  -- message about this test
//     tol  -- tolerance to compare gAna with gNum
//     gAna -- [scalar] analytical (or other kind) derivative dfdx
//     xAt  -- [scalar] position to compute dfdx
//     h    -- initial stepsize; e.g. 1e-1
//     verb -- verbose: show messages
//     fcn  -- [scalar] function f(x). x is scalar
func DerivScaSca(tst *testing.T, msg string, tol, gAna, xAt, h float64, verb bool, fcn func(x float64) float64) {

	// call centralDeriv first
	res, round, trunc := centralDeriv(fcn, xAt, h)
	numerr := round + trunc

	// check rounding error
	if round < trunc && (round > 0 && trunc > 0) {

		// compute an optimised stepsize to minimize the total error, using the scaling of the
		// truncation error (O(h^2)) and rounding error (O(1/h)).
		hOpt := h * math.Pow(round/(2.0*trunc), 1.0/3.0)
		rOpt, roundOpt, truncOpt := centralDeriv(fcn, xAt, hOpt)
		errorOpt := roundOpt + truncOpt

		// check that the new error is smaller, and that the new derivative is consistent with the
		// error bounds of the original estimate.
		if errorOpt < numerr && math.Abs(rOpt-res) < 4.0*numerr {
			res = rOpt
		}
	}

	// compare
	AnaNum(tst, msg, tol, gAna, res, verb)
}

// centralDeriv Computes the derivative using the 5-point rule (x-h, x-h/2, x, x+h/2, x+h).
func centralDeriv(f func(x float64) float64, x float64, h float64) (res, absErrRound, absErrTrunc float64) {

	// Compute the derivative using the 5-point rule (x-h, x-h/2, x, x+h/2, x+h).
	// Note that the central point is not used.
	// Compute the error using the difference between the 5-point and the 3-point rule (x-h,x,x+h).
	// Again the central point is not used.
	fm1 := f(x - h)
	fp1 := f(x + h)
	fmh := f(x - h/2.0)
	fph := f(x + h/2.0)
	EPS := 1.0e-15 // smallest number satisfying 1.0 + EPS > 1.0
	r3 := 0.5 * (fp1 - fm1)
	r5 := (4.0/3.0)*(fph-fmh) - (1.0/3.0)*r3
	e3 := (math.Abs(fp1) + math.Abs(fm1)) * EPS
	e5 := 2.0*(math.Abs(fph)+math.Abs(fmh))*EPS + e3

	// The next term is due to finite precision in x+h = O (eps * x)
	dy := max(math.Abs(r3/h), math.Abs(r5/h)) * (math.Abs(x) / h) * EPS

	// The truncation error in the r5 approximation itself is O(h^4).
	// However, for safety, we estimate the error from r5-r3, which is O(h^2).
	// By scaling h we will minimise this estimated error, not the actual truncation error in r5.
	res = r5 / h
	absErrTrunc = math.Abs((r5 - r3) / h) // Estimated truncation error O(h^2)
	absErrRound = math.Abs(e5/h) + dy     // Rounding error (cancellations)
	return
}
