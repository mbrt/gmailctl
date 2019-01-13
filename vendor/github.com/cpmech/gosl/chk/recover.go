// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chk

import (
	"fmt"
	"os"
	"testing"
)

// Recover catches panics and call os.Exit(1) on 'panic'
func Recover() {
	if err := recover(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
}

// RecoverTst catches panics in tests. Test will fail on 'panic'
func RecoverTst(tst *testing.T) {
	if err := recover(); err != nil {
		tst.Errorf("%v\n", err)
		tst.FailNow()
	}
}

// RecoverTstPanicIsOK catches panics in tests. Test must 'panic' to be OK
func RecoverTstPanicIsOK(tst *testing.T) {
	if err := recover(); err == nil {
		tst.Errorf("Test should have panicked\n")
		tst.FailNow()
	}
}
