// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utl

import "github.com/cpmech/gosl/io"

// List implements a tabular list with variable number of columns
//  Example:
//    Vals = [][]float64{
//             {0.0},
//             {1.0, 1.1, 1.2, 1.3},
//             {2.0, 2.1},
//             {3.0, 3.1, 3.2},
//           }
type List struct {
	Vals [][]float64 // values
}

// Append appends items to List
func (o *List) Append(rowidx int, value float64) {
	size := len(o.Vals)

	// fill new rows if necessary
	if rowidx >= size {
		for i := size; i <= rowidx; i++ {
			o.Vals = append(o.Vals, []float64{})
		}
	}

	// append value
	o.Vals[rowidx] = append(o.Vals[rowidx], value)
}

// SerialList implements a tabular list with variable number of columns
// using a serial representation
//  Example:
//      0.0
//      1.0  1.1  1.2  1.3
//      2.0  2.1
//      3.0  3.1  3.2
//  becomes:
//             (0)   (1)    2    3    4   (5)    6   (7)    8    9   (10)
//      Vals = 0.0 | 1.0  1.1  1.2  1.3 | 2.0  2.1 | 3.0  3.1  3.2 |
//      Ptrs = 0 1 5 7 10
//  Notes:
//      len(Ptrs) = nrows + 1
//      Ptrs[len(Ptrs)-1] = len(Vals)
type SerialList struct {
	Vals []float64 // values
	Ptrs []int     // pointers
}

// Append appends item to SerialList
func (o *SerialList) Append(startRow bool, value float64) {
	if startRow {
		if len(o.Ptrs) == 0 {
			o.Ptrs = []int{0, 0}
		} else {
			start := o.Ptrs[len(o.Ptrs)-1]
			o.Ptrs = append(o.Ptrs, start)
		}
	}
	o.Vals = append(o.Vals, value)
	o.Ptrs[len(o.Ptrs)-1]++
}

// Print prints the souble-serial-list
func (o SerialList) Print(fmt string) {
	for i := 0; i < len(o.Ptrs)-1; i++ {
		for j := o.Ptrs[i]; j < o.Ptrs[i+1]; j++ {
			io.Pf(fmt, o.Vals[j])
		}
		io.Pf("\n")
	}
}
