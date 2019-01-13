// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chk

import (
	"fmt"
	"math"
	"math/cmplx"
	"runtime"
	"testing"
)

// PrintTitle returns the Test Title
func PrintTitle(title string) {
	if Verbose {
		fmt.Printf("\n=== %s =================\n", title)
		return
	}
	fmt.Printf("   . . . testing . . .   %s\n", title)
}

// PrintOk prints "OK" in green (if ColorsOn==true)
func PrintOk(msg string, prm ...interface{}) {
	if Verbose {
		fmt.Printf(msg, prm...)
		if ColorsOn {
			fmt.Println(" [1;32mOK[0m")
			return
		}
		fmt.Println(" OK")
	}
}

// TstFail calls tst.Errorf() with msg and parameters.
// It also prints "FAIL" in red (if ColorsOn==true)
func TstFail(tst *testing.T, msg string, prm ...interface{}) {
	tst.Errorf(msg, prm...)
	if Verbose {
		fmt.Printf(msg, prm...)
		if ColorsOn {
			fmt.Println(" [1;31mFAIL[0m")
			return
		}
		fmt.Println(" FAIL")
	}
}

// TstDiff tests difference between float64
func TstDiff(tst *testing.T, msg string, tol, a, b float64, showOK bool) (failed bool) {
	diff := math.Abs(a - b)
	if math.IsNaN(diff) || math.IsInf(diff, 0) {
		TstFail(tst, "%s NaN or Inf in a=%v b=%v", msg, a, b)
		return true
	}
	if diff > tol {
		TstFail(tst, "%s %v != %v |diff| = %g", msg, a, b, diff)
		return true
	}
	if showOK {
		PrintOk(msg)
	}
	return
}

// TestDiffC tests difference between complex128.
// It also prints "FAIL" or "OK"
func TestDiffC(tst *testing.T, msg string, tol float64, a, b complex128, showOK bool) (failed bool) {
	diff := cmplx.Abs(a - b)
	if math.IsNaN(diff) || math.IsInf(diff, 0) {
		TstFail(tst, "%s NaN or Inf in a=%v b=%v", msg, a, b)
		return true
	}
	if diff > tol {
		TstFail(tst, "%s |diff| = %g", msg, diff)
		return true
	}
	if showOK {
		PrintOk(msg)
	}
	return
}

// PrintAnaNum formats the output of analytical versus numerical comparisons
func PrintAnaNum(msg string, tol, ana, num float64, verbose bool) (e error) {
	diff := math.Abs(ana - num)
	if math.IsNaN(diff) || math.IsInf(diff, 0) {
		if ColorsOn {
			e = Err("[1;31m%s NaN or Inf: %v[0m", msg, diff)
		} else {
			e = Err("%s NaN or Inf: %v", msg, diff)
		}
		return
	}
	if verbose {
		if ColorsOn {
			clr := "[1;32m" // green
			if diff > tol {
				clr = "[1;31m" // red
			}
			fmt.Printf("%s %23v %23v %s%23v[0m\n", msg, ana, num, clr, diff)
		} else {
			fmt.Printf("%s %23v %23v %23v\n", msg, ana, num, diff)
		}
	}
	if diff > tol {
		if ColorsOn {
			e = Err("[1;31m%s |diff| = %g[0m", msg, diff)
		} else {
			e = Err("%s |diff| = %g", msg, diff)
		}
	}
	return
}

// PrintAnaNumC formats the output of analytical versus numerical comparisons (complex version)
func PrintAnaNumC(msg string, tol float64, ana, num complex128, verbose bool) (e error) {
	diffR := math.Abs(real(ana) - real(num))
	diffC := math.Abs(imag(ana) - imag(num))
	if math.IsNaN(diffR) || math.IsInf(diffR, 0) {
		e = Err("[1;31m%s (real part) NaN or Inf: %v[0m", msg, diffR)
		return
	}
	if math.IsNaN(diffC) || math.IsInf(diffC, 0) {
		e = Err("[1;31m%s (imag part) NaN or Inf: %v[0m", msg, diffC)
		return
	}
	if verbose {
		clrR := "[1;32m" // green
		clrC := "[1;32m" // green
		if diffR > tol {
			clrR = "[1;31m" // red
		}
		if diffC > tol {
			clrC = "[1;31m" // red
		}
		f := "%" + fmt.Sprintf("%d", len(msg)) + "s"
		fmt.Printf(f+" %23v  %23v  %s%23v[0m\n", msg, real(ana), real(num), clrR, diffR)
		fmt.Printf(f+" %23vi %23vi %s%23v[0m\n", "", imag(ana), imag(num), clrC, diffC)
	}
	if diffR > tol || diffC > tol {
		e = Err("[1;31m%s |diffR| = %g  |diffC| = %g[0m", msg, diffR, diffC)
	}
	return
}

// CallerInfo returns the file and line positions where an error occurred
//  idx -- use idx=2 to get the caller of Panic
func CallerInfo(idx int) {
	pc, file, line, ok := runtime.Caller(idx)
	if !ok {
		file, line = "?", 0
	}
	var fname string
	f := runtime.FuncForPC(pc)
	if f != nil {
		fname = f.Name()
	}
	if Verbose {
		fmt.Printf("file = %s:%d\n", file, line)
		fmt.Printf("func = %s\n", fname)
	}
}
