// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

import (
	"math"
)

// MaskType defines the type of mask
type MaskType int

const (

	// NoneType defines the NONE mask type
	NoneType MaskType = iota

	// StarType defines the STAR mask type
	StarType

	// PrimeType defines the PRIME mask type
	PrimeType
)

// Munkres (Hungarian algorithm) method to solve the assignment problem
//
//	based on code by Bob Pilgrim from http://csclab.murraystate.edu/bob.pilgrim/445/munkres.html
//	Note: this method runs in O(n³), in the worst case; therefore is not efficient for large matrix
//	 Example:
//	         $ | Clean  Sweep   Wash
//	    -------|--------------------
//	    Fry    |   [2]      3      3
//	    Leela  |     3    [2]      3
//	    Bender |     3      3    [2]
//	    minimum cost = 6
//
//	Note: cost will be minimised
type Munkres struct {

	// main
	C     [][]float64 // [nrow][ncol] cost matrix
	Cori  [][]float64 // [nrow][ncol] original cost matrix
	Links []int       // [nrow] will contain links/assignments after Run(), where j := o.Links[i] means that i is assigned to j. -1 means no assignment/link
	Cost  float64     // total cost after Run() and links are established

	// auxiliary
	M          [][]MaskType // [nrow][ncol] mask matrix. If Mij==1, then Cij is a starred zero. If Mij==2, then Cij is a primed zero
	path       [][]int      // path
	rowCovered []bool       // indicates whether a row is covered or not
	colCovered []bool       // indicates whether a column is covered or not
	nrowOri    int          // (original) number of rows in cost matrix
	ncolOri    int          // (original) number of column in cost matrix
	nrow       int          // number of rows in cost/mask matrix
	ncol       int          // number of column in cost/mask matrix
	pathRow0   int          // first row in path
	pathCol0   int          // first col in path
}

// Init initialises Munkres' structure
func (o *Munkres) Init(nrow, ncol int) {
	IntAssertLessThan(0, nrow) // nrow > 1
	IntAssertLessThan(0, ncol) // ncol > 1
	o.nrow, o.ncol = nrow, ncol
	o.nrowOri, o.ncolOri = nrow, ncol
	if o.nrow != o.ncol { // make it square. padded entries will have zero cost
		o.nrow = Imax(o.nrow, o.ncol)
		o.ncol = o.nrow
	}
	o.C = Alloc(o.nrow, o.ncol)
	o.M = make([][]MaskType, o.nrow)
	for i := 0; i < o.nrow; i++ {
		o.M[i] = make([]MaskType, o.ncol)
	}
	o.Links = make([]int, o.nrowOri)
	npath := 2*o.nrow + 1 // TODO: check this
	o.path = IntAlloc(npath, 2)
	o.rowCovered = make([]bool, o.nrow)
	o.colCovered = make([]bool, o.ncol)
}

// SetCostMatrix sets cost matrix by copying from C to internal o.C
//
//	Note: costs must be positive
func (o *Munkres) SetCostMatrix(C [][]float64) {
	o.Cori = C
	for i := 0; i < o.nrowOri; i++ {
		for j := 0; j < o.ncolOri; j++ {
			o.C[i][j] = C[i][j]
			o.M[i][j] = NoneType
			if math.IsNaN(o.C[i][j]) {
				Panic("cannot set cost matrix because of NaN value")
			}
		}
		o.rowCovered[i] = false
	}
	for j := 0; j < o.ncol; j++ {
		o.colCovered[j] = false
	}
	o.pathRow0 = 0
	o.pathCol0 = 0
}

// Run runs the iterative algorithm
//
//	Output:
//	 o.Links -- will contain assignments, where len(assignments) == nrow and
//	            j := o.Links[i] means that i is assigned to j
//	            -1 means no assignment/link
//	 o.Cost -- will have the total cost by following links
func (o *Munkres) Run() {

	// column matrix
	if o.ncol == 1 {
		o.Cost = o.C[0][0]
		isel := 0
		o.Links[0] = -1
		for i := 1; i < o.nrow; i++ {
			o.Links[i] = -1
			if o.C[i][0] < o.Cost {
				o.Cost = o.C[i][0]
				isel = i
			}
		}
		o.Links[isel] = 0
		return
	}

	// row matrix
	if o.nrow == 1 {
		o.Cost = o.C[0][0]
		jsel := 0
		for j := 1; j < o.ncol; j++ {
			if o.C[0][j] < o.Cost {
				o.Cost = o.C[0][j]
				jsel = j
			}
		}
		o.Links[0] = jsel
		return
	}

	// run Munkres algorithm
	step := 1
	done := false
	for !done {
		switch step {
		case 1:
			step = o.step1() // returns 2
		case 2:
			step = o.step2() // returns 3
		case 3:
			step = o.step3() // returns 4 or 7
		case 4:
			step = o.step4() // returns 5 or 6
		case 5:
			step = o.step5() // returns 3
		case 6:
			step = o.step6() // returns 4
		case 7:
			done = true
		}
	}

	// compute cost and set links
	o.Cost = 0
	for i := 0; i < o.nrowOri; i++ {
		o.Links[i] = -1
		for j := 0; j < o.ncolOri; j++ {
			if o.M[i][j] == StarType {
				o.Links[i] = j
				o.Cost += o.Cori[i][j]
				break
			}
		}
	}
}

// steps //////////////////////////////////////////////////////////////////////////////////////////

// step1: for each row of the cost matrix, find the smallest element and subtract it from every
// element in its row. nextStep = 2
func (o *Munkres) step1() (nextStep int) {
	var xmin float64
	for i := 0; i < o.nrow; i++ {
		xmin = o.C[i][0]
		for j := 1; j < o.ncol; j++ {
			xmin = Min(xmin, o.C[i][j])
		}
		for j := 0; j < o.ncol; j++ {
			o.C[i][j] -= xmin
		}
	}
	return 2
}

// step2: find a zero (Z) in the resulting matrix. If there is no starred zero in its row or column,
// star Z. Repeat for each element in the matrix. Check to see if Cij is a zero value and if its
// column or row is not already covered. If not, then star this zero and cover its row and column.
// Uncover all rows and columns before leaving. nextStep = 3
func (o *Munkres) step2() (nextStep int) {
	for i := 0; i < o.nrow; i++ {
		for j := 0; j < o.ncol; j++ {
			if !o.rowCovered[i] && !o.colCovered[j] {
				if o.C[i][j] == 0 {
					o.M[i][j] = StarType
					o.rowCovered[i] = true
					o.colCovered[j] = true
				}
			}
		}
	}
	for i := 0; i < o.nrow; i++ {
		o.rowCovered[i] = false
	}
	for j := 0; j < o.ncol; j++ {
		o.colCovered[j] = false
	}
	return 3
}

// step3: cover each column containing a starred zero. If min(n,m) columns are covered, the starred
// zeros describe a complete set of unique assignments and the process is completed (nextStep=7);
// otherwise nextStep=4.
func (o *Munkres) step3() (nextStep int) {
	for i := 0; i < o.nrow; i++ {
		for j := 0; j < o.ncol; j++ {
			if o.M[i][j] == StarType {
				o.colCovered[j] = true
			}
		}
	}
	count := 0
	for j := 0; j < o.ncol; j++ {
		if o.colCovered[j] {
			count++
		}
	}
	nextStep = 4 // not all columns are covered
	if count >= o.ncol || count >= o.nrow {
		nextStep = 7 // all covered
	}
	return
}

// step4: find a noncovered zero and prime it. Change cost matrix by adding min value to every
// element of covered rows and subtracting it from elements of uncovered columns until a noncovered
// zero is found.
func (o *Munkres) step4() (nextStep int) {
	row, col, done := -1, -1, false
	for !done {
		row, col = o.findNoncovZero()
		if row == -1 {
			done = true
			nextStep = 6
		} else {
			o.M[row][col] = PrimeType
			colStar := o.findStarInRow(row)
			if colStar >= 0 {
				col = colStar
				o.rowCovered[row] = true
				o.colCovered[col] = false
			} else {
				done = true
				nextStep = 5
				o.pathRow0 = row
				o.pathCol0 = col
			}
		}
	}
	return
}

// step5: construct a series of alternating primed and starred zeros as follows. Let Z0 represent
// the uncovered primed zero found in Step 4. Let Z1 denote the starred zero in the column of Z0 (if
// any). Let Z2 denote the primed zero in the row of Z1 (there will always be one). Continue until
// the series terminates at a primed zero that has no starred zero in its column. Unstar each
// starred zero of the series, star each primed zero of the series, erase all primes and uncover
// every line in the matrix. nextStep = 3
func (o *Munkres) step5() (nextStep int) {

	// construct series
	r, c, done := -1, -1, false
	count := 1
	o.path[count-1][0] = o.pathRow0
	o.path[count-1][1] = o.pathCol0
	for !done {
		r = o.findStarInCol(o.path[count-1][1])
		if r > -1 {
			count++
			o.path[count-1][0] = r
			o.path[count-1][1] = o.path[count-2][1]
		} else {
			done = true
		}
		if !done {
			c = o.findPrimeInRow(o.path[count-1][0])
			count++
			o.path[count-1][0] = o.path[count-2][0]
			o.path[count-1][1] = c
		}
	}

	// augment path
	for p := 0; p < count; p++ {
		if o.M[o.path[p][0]][o.path[p][1]] == StarType {
			o.M[o.path[p][0]][o.path[p][1]] = NoneType
		} else {
			o.M[o.path[p][0]][o.path[p][1]] = StarType
		}
	}

	// clear covers
	for i := 0; i < o.nrow; i++ {
		o.rowCovered[i] = false
	}
	for j := 0; j < o.ncol; j++ {
		o.colCovered[j] = false
	}

	// erase primes
	for i := 0; i < o.nrow; i++ {
		for j := 0; j < o.ncol; j++ {
			if o.M[i][j] == PrimeType {
				o.M[i][j] = NoneType
			}
		}
	}
	return 3
}

// step6: add min value to every element of each covered row, and subtract it from every element of
// each uncovered column. nextStep = 4
func (o *Munkres) step6() (nextStep int) {

	// find min value
	xmin := math.MaxFloat64
	for i := 0; i < o.nrow; i++ {
		for j := 0; j < o.ncol; j++ {
			if !o.rowCovered[i] && !o.colCovered[j] {
				xmin = Min(xmin, o.C[i][j])
			}
		}
	}

	// add/subtract min value
	for i := 0; i < o.nrow; i++ {
		for j := 0; j < o.ncol; j++ {
			if o.rowCovered[i] {
				o.C[i][j] += xmin
			}
			if !o.colCovered[j] {
				o.C[i][j] -= xmin
			}
		}
	}
	return 4
}

// auxiliary ///////////////////////////////////////////////////////////////////////////////////////

// StrCostMatrix returns a representation of cost matrix with masks and covers
func (o *Munkres) StrCostMatrix() (l string) {
	numfmt := "%v"
	l += Sf("%4v", " ")
	for j := 0; j < o.ncol; j++ {
		if o.colCovered[j] {
			l += Sf("%8v", "T ")
		} else {
			l += Sf("%8v", "F ")
		}
	}
	l += Sf("\n")
	for i := 0; i < o.nrow; i++ {
		if o.rowCovered[i] {
			l += Sf("%4v", "T")
		} else {
			l += Sf("%4v", "F")
		}
		for j := 0; j < o.ncol; j++ {
			s := Sf(numfmt, o.C[i][j])
			switch o.M[i][j] {
			case NoneType:
				s += " "
			case StarType:
				s += "*"
			case PrimeType:
				s += "'"
			}
			l += Sf("%8v", s)
		}
		l += Sf("\n")
	}
	return
}

// findNoncovZero finds the row and column of a non-covered zero entry. -1 means not found
func (o *Munkres) findNoncovZero() (row, col int) {
	i, j, done := 0, 0, false
	row, col = -1, -1
	for !done {
		j = 0
		for true {
			if !o.rowCovered[i] && !o.colCovered[j] {
				if o.C[i][j] == 0 {
					row, col, done = i, j, true
				}
			}
			j++
			if j >= o.ncol || done {
				break
			}
		}
		i++
		if i >= o.nrow {
			done = true
		}
	}
	return
}

// findStarInRow is a method to support step 4
func (o *Munkres) findStarInRow(row int) (col int) {
	col = -1
	for j := 0; j < o.ncol; j++ {
		if o.M[row][j] == StarType {
			col = j
		}
	}
	return
}

// findStarInCol is a method to support step 5
func (o *Munkres) findStarInCol(c int) (r int) {
	r = -1
	for i := 0; i < o.nrow; i++ {
		if o.M[i][c] == StarType {
			r = i
		}
	}
	return
}

// findStarInRow is a method to support step 5
func (o *Munkres) findPrimeInRow(r int) (c int) {
	for j := 0; j < o.ncol; j++ {
		if o.M[r][j] == PrimeType {
			c = j
		}
	}
	return
}

/* Notes:
 1. The algorithm will work even when the minimum values in two or more rows are in the same column.
 2. The algorithm will work even when two or more of the rows contain the same values in the the same order.
 3. The algorithm will work even when all the values are the same (although the result is not very interesting).
 4. Munkres Assignment Algorithm is not exponential run time or intractable; it is of a low order polynomial run time, worst-case O(n3).
 5. Optimality is guaranteed in Munkres Assignment Algorithm.
 6. Setting the cost matrix to C(i,j) = i*j  makes a good testing matrix for this problem.
 7. In this algorithm the range of indices is[0..n-1] rather than [1..n].
 8. Step 3 is an example of the greedy method.  If the minimum values are all in different rows then their positions represent the minimal pairwise assignments.
 9. Step 5 is an example of the Augmenting Path Algorithm (Stable Marriage Problem).
10. Step 6 is an example of constraint relaxation.  It is "giving up" on a particular cost and raising the constraint by the least amount possible.
11. If your implementation is jumping between Step 4 and Step 6 without entering Step 5, you probably have not properly dealt with recognizing that there are no uncovered zeros in Step 4.
12. In the matrix M 1=starred zero and 2=primed zero.  So, if C[i,j] is a starred zero we would set M[i,j]=1.  All other elements in M are set to zero
13. The Munkres assignment algorithm can be implemented as a sparse matrix, but you will need to ensure that the correct (optimal) assignment pairs are active in the sparse cost matrix C
14. Munkres Assignment can be applied to TSP, pattern matching, track initiation, data correlation, and (of course) any pairwise assignment application.
15. Munkres can be extended to rectangular arrays (i.e. more jobs than workers, or more workers than jobs) .
16. The best way to find a maximal assignment is to replace the values ci,j in the cost matrix with C[i,j] = bigval - ci,j.
17. Original Reference: Algorithms for Assignment and Transportation Problems, James Munkres, Journal of the Society for Industrial and Applied Mathematics Volume 5, Number 1, March, 1957
18. Extension to Rectangular Arrays Ref:  F. Burgeois and J.-C. Lasalle. An extension of the Munkres algorithm for the assignment problem to rectangular matrices. Communications of the ACM, 142302-806, 1971.
*/
