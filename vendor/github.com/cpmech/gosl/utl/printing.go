// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utl

import (
	"fmt"
	"math"

	"github.com/cpmech/gosl/io"
)

// PrintDeep3 prints an array of array of array
func PrintDeep3(name string, A [][][]float64) {
	io.Pf("%s = [\n", name)
	for _, a := range A {
		io.Pf("  %v\n", a)
	}
	io.Pf("]\n")
}

// PrintDeep4 prints an array of array of array
func PrintDeep4(name string, A [][][][]float64, format string) {
	res := name + " = \n"
	for _, a := range A {
		for _, b := range a {
			for _, c := range b {
				for _, d := range c {
					if math.Abs(d) <= 1e-3 {
						d = 0
					}
					res += io.Sf(format, d)
				}
				res += "\n"
			}
			res += "\n"
		}
	}
	fmt.Println(res)
}
