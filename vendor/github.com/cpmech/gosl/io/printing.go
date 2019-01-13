// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

// StrThickLine returns a thick line (using '=')
func StrThickLine(n int) (l string) {
	for i := 0; i < n; i++ {
		l += "="
	}
	return l + "\n"
}

// StrThinLine returns a thin line (using '-')
func StrThinLine(n int) (l string) {
	for i := 0; i < n; i++ {
		l += "-"
	}
	return l + "\n"
}

// StrSpaces returns a line with spaces
func StrSpaces(n int) (l string) {
	for i := 0; i < n; i++ {
		l += " "
	}
	return
}
