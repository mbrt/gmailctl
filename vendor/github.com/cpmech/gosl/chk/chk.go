// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package chk contains functions for checking and testing computations
package chk

import "fmt"

var (
	// AssertOn activates or deactivates asserts
	AssertOn = true

	// Verbose turn on verbose mode
	Verbose = false

	// ColorsOn turn on use of colours on console
	ColorsOn = true
)

// PanicSimple panicks without calling CallerInfo
func PanicSimple(msg string, prm ...interface{}) {
	panic(fmt.Sprintf(msg, prm...))
}

// Panic calls CallerInfo and panicks
func Panic(msg string, prm ...interface{}) {
	CallerInfo(4)
	CallerInfo(3)
	CallerInfo(2)
	panic(fmt.Sprintf(msg, prm...))
}

// Err returns a new error
func Err(msg string, prm ...interface{}) error {
	return fmt.Errorf(msg, prm...)
}
