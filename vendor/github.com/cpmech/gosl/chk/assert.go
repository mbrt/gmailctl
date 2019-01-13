// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chk

// IntAssert asserts that a is equal to b (ints)
func IntAssert(a, b int) {
	if AssertOn {
		if a != b {
			CallerInfo(3)
			CallerInfo(2)
			PanicSimple("Assert failed: ints are different: %d != %d", a, b)
		}
	}
}

// IntAssertLessThan asserts that a < b (ints)
func IntAssertLessThan(a, b int) {
	if AssertOn {
		if a < b {
			return
		}
		CallerInfo(3)
		CallerInfo(2)
		PanicSimple("Assert failed: %d < %d is incorrect", a, b)
	}
}

// IntAssertLessThanOrEqualTo asserts that a ≤ b (ints)
func IntAssertLessThanOrEqualTo(a, b int) {
	if AssertOn {
		if a <= b {
			return
		}
		CallerInfo(3)
		CallerInfo(2)
		PanicSimple("Assert failed: %d ≤ %d is incorrect", a, b)
	}
}

// Float64assert asserts that a is equal to b (floats)
func Float64assert(a, b float64) {
	if AssertOn {
		if a != b {
			CallerInfo(3)
			CallerInfo(2)
			PanicSimple("Assert failed: float64 are different: %g != %g", a, b)
		}
	}
}

// StrAssert asserts that a is equal to b (strings)
func StrAssert(a, b string) {
	if AssertOn {
		if a != b {
			CallerInfo(3)
			CallerInfo(2)
			PanicSimple("Assert failed: strings are different: %s != %s", a, b)
		}
	}
}
