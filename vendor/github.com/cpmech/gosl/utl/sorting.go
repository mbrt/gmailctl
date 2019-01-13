// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utl

import (
	"sort"

	"github.com/cpmech/gosl/chk"
	"github.com/cpmech/gosl/io"
)

// IntSort3 sorts 3 values in ascending order
func IntSort3(a, b, c *int) {
	if *b < *a {
		*a, *b = *b, *a
	}
	if *c < *b {
		*b, *c = *c, *b
	}
	if *b < *a {
		*a, *b = *b, *a
	}
}

// IntSort4 sort four values in ascending order
func IntSort4(a, b, c, d *int) {
	if *b < *a {
		*b, *a = *a, *b
	}
	if *c < *b {
		*c, *b = *b, *c
	}
	if *d < *c {
		*d, *c = *c, *d
	}
	if *b < *a {
		*b, *a = *a, *b
	}
	if *c < *b {
		*c, *b = *b, *c
	}
	if *b < *a {
		*b, *a = *a, *b
	}
}

// Sort3 sorts 3 values in ascending order
func Sort3(a, b, c *float64) {
	if *b < *a {
		*a, *b = *b, *a
	}
	if *c < *b {
		*b, *c = *c, *b
	}
	if *b < *a {
		*a, *b = *b, *a
	}
}

// Sort3Desc sorts 3 values in descending order
func Sort3Desc(a, b, c *float64) {
	if *b > *a {
		*a, *b = *b, *a
	}
	if *c > *b {
		*b, *c = *c, *b
	}
	if *b > *a {
		*a, *b = *b, *a
	}
}

// Sort4 sort four values in ascending order
func Sort4(a, b, c, d *float64) {
	if *b < *a {
		*b, *a = *a, *b
	}
	if *c < *b {
		*c, *b = *b, *c
	}
	if *d < *c {
		*d, *c = *c, *d
	}
	if *b < *a {
		*b, *a = *a, *b
	}
	if *c < *b {
		*c, *b = *b, *c
	}
	if *b < *a {
		*b, *a = *a, *b
	}
}

// IntGetSorted returns a sorted (increasing) copy of 'A'
func IntGetSorted(A []int) (sortedA []int) {
	sortedA = make([]int, len(A))
	copy(sortedA, A)
	sort.Ints(sortedA)
	return
}

// GetSorted returns a sorted (increasing) copy of 'A'
func GetSorted(A []float64) (sortedA []float64) {
	sortedA = make([]float64, len(A))
	copy(sortedA, A)
	sort.Float64s(sortedA)
	return
}

// Quadruple //////////////////////////////////////////////////////////////////////////////////////

// Quadruple helps to sort a quadruple of 1 int and 3 float64s
type Quadruple struct {
	I int
	X float64
	Y float64
	Z float64
}

// Quadruples helps to sort quadruples
type Quadruples []*Quadruple

// BuildQuadruples initialise Quadruples with i, x, y, and z
//  Note: i, x, y, or z can be nil; but at least one of them must be non nil
func BuildQuadruples(i []int, x, y, z []float64) (q Quadruples) {
	ni := len(i)
	nx := len(x)
	ny := len(y)
	nz := len(z)
	_, n := IntMinMax([]int{ni, nx, ny, nz})
	q = make([]*Quadruple, n)
	var ival int
	var xval, yval, zval float64
	for k := 0; k < n; k++ {
		if ni > k {
			ival = i[k]
		}
		if nx > k {
			xval = x[k]
		}
		if ny > k {
			yval = y[k]
		}
		if nz > k {
			zval = z[k]
		}
		q[k] = &Quadruple{ival, xval, yval, zval}
	}
	return
}

// I returns the 'i' in quadruples
func (o Quadruples) I() (i []int) {
	i = make([]int, len(o))
	for k := 0; k < len(o); k++ {
		i[k] = o[k].I
	}
	return
}

// X returns the 'x' in quadruples
func (o Quadruples) X() (x []float64) {
	x = make([]float64, len(o))
	for k := 0; k < len(o); k++ {
		x[k] = o[k].X
	}
	return
}

// Y returns the 'y' in quadruples
func (o Quadruples) Y() (y []float64) {
	y = make([]float64, len(o))
	for k := 0; k < len(o); k++ {
		y[k] = o[k].Y
	}
	return
}

// Z returns the 'z' in quadruples
func (o Quadruples) Z() (z []float64) {
	z = make([]float64, len(o))
	for k := 0; k < len(o); k++ {
		z[k] = o[k].Z
	}
	return
}

// Len returns the length of Quadruples
func (o Quadruples) Len() int { return len(o) }

// Swap swaps two quadruples
func (o Quadruples) Swap(i, j int) { o[i], o[j] = o[j], o[i] }

// String returns the string representation of Quadruples
func (o Quadruples) String() string {
	res := ""
	for _, p := range o {
		res += io.Sf("%6d%20g%20g%20g\n", p.I, p.X, p.Y, p.Z)
	}
	return res
}

// QuadruplesByI defines struct to sort Quadruples by I
type QuadruplesByI struct{ Quadruples }

// Less compares two QuadruplesI
func (o QuadruplesByI) Less(i, j int) bool { return o.Quadruples[i].I < o.Quadruples[j].I }

// QuadruplesByX defines struct to sort Quadruples by X
type QuadruplesByX struct{ Quadruples }

// Less compares two QuadruplesX
func (o QuadruplesByX) Less(i, j int) bool { return o.Quadruples[i].X < o.Quadruples[j].X }

// QuadruplesByY Sort Quadruples by Y
type QuadruplesByY struct{ Quadruples }

// Less compares two QuadruplesY
func (o QuadruplesByY) Less(i, j int) bool { return o.Quadruples[i].Y < o.Quadruples[j].Y }

// QuadruplesByZ defines struct to Sort Quadruples by Z
type QuadruplesByZ struct{ Quadruples }

// Less compares two QuadruplesZ
func (o QuadruplesByZ) Less(i, j int) bool { return o.Quadruples[i].Z < o.Quadruples[j].Z }

// SortQuadruples sorts i, x, y, and z by "i", "x", "y", or "z"
//  Note: either i, x, y, or z can be nil; i.e. at least one of them must be non nil
func SortQuadruples(i []int, x, y, z []float64, by string) (I []int, X, Y, Z []float64) {
	q := BuildQuadruples(i, x, y, z)
	switch by {
	case "i":
		sort.Sort(QuadruplesByI{q})
	case "x":
		sort.Sort(QuadruplesByX{q})
	case "y":
		sort.Sort(QuadruplesByY{q})
	case "z":
		sort.Sort(QuadruplesByZ{q})
	default:
		chk.Panic("sort quadruples command must be 'i', 'x', 'y', or 'z'. by == '%s' is invalid\n", by)
	}
	if i != nil {
		I = q.I()
	}
	if x != nil {
		X = q.X()
	}
	if y != nil {
		Y = q.Y()
	}
	if z != nil {
		Z = q.Z()
	}
	return
}

// Str => ??? maps /////////////////////////////////////////////////////////////////////////////////

// StrIntMapSort returns sorted keys of map[string]int
func StrIntMapSort(m map[string]int) (sortedKeys []string) {
	sortedKeys = make([]string, len(m))
	i := 0
	for key := range m {
		sortedKeys[i] = key
		i++
	}
	sort.Strings(sortedKeys)
	return
}

// StrFltMapSort returns sorted keys of map[string]float64
func StrFltMapSort(m map[string]float64) (sortedKeys []string) {
	sortedKeys = make([]string, len(m))
	i := 0
	for key := range m {
		sortedKeys[i] = key
		i++
	}
	sort.Strings(sortedKeys)
	return
}

// StrBoolMapSort returns sorted keys of map[string]bool
func StrBoolMapSort(m map[string]bool) (sortedKeys []string) {
	sortedKeys = make([]string, len(m))
	i := 0
	for key := range m {
		sortedKeys[i] = key
		i++
	}
	sort.Strings(sortedKeys)
	return
}

// StrIntMapSortSplit returns sorted keys of map[string]int and sorted values
func StrIntMapSortSplit(m map[string]int) (sortedKeys []string, sortedVals []int) {
	sortedKeys = make([]string, len(m))
	sortedVals = make([]int, len(m))
	i := 0
	for key := range m {
		sortedKeys[i] = key
		i++
	}
	sort.Strings(sortedKeys)
	for j, key := range sortedKeys {
		sortedVals[j] = m[key]
	}
	return
}

// StrFltMapSortSplit returns sorted keys of map[string]float64 and sorted values
func StrFltMapSortSplit(m map[string]float64) (sortedKeys []string, sortedVals []float64) {
	sortedKeys = make([]string, len(m))
	sortedVals = make([]float64, len(m))
	i := 0
	for key := range m {
		sortedKeys[i] = key
		i++
	}
	sort.Strings(sortedKeys)
	for j, key := range sortedKeys {
		sortedVals[j] = m[key]
	}
	return
}

// StrBoolMapSortSplit returns sorted keys of map[string]bool and sorted values
func StrBoolMapSortSplit(m map[string]bool) (sortedKeys []string, sortedVals []bool) {
	sortedKeys = make([]string, len(m))
	sortedVals = make([]bool, len(m))
	i := 0
	for key := range m {
		sortedKeys[i] = key
		i++
	}
	sort.Strings(sortedKeys)
	for j, key := range sortedKeys {
		sortedVals[j] = m[key]
	}
	return
}

// Int => ??? maps /////////////////////////////////////////////////////////////////////////////////

// IntBoolMapSort returns sorted keys of map[int]bool
func IntBoolMapSort(m map[int]bool) (sortedKeys []int) {
	sortedKeys = make([]int, len(m))
	i := 0
	for key := range m {
		sortedKeys[i] = key
		i++
	}
	sort.Ints(sortedKeys)
	return
}

// low level implementations ///////////////////////////////////////////////////////////////////////

// swap swaps two float64 numbers
func swap(a, b *float64) { *a, *b = *b, *a }

// iswap swaps two int numbers
func iswap(a, b *int) { *a, *b = *b, *a }

// Qsort sort an array arr[0..n-1] into ascending numerical order using the Quicksort algorithm.
// arr is replaced on output by its sorted rearrangement. Normally, the optional argument m should
// be omitted, but if it is set to a positive value, then only the first m elements of arr are
// sorted. Implementation from [1]
//   Reference:
//   [1] Press WH, Teukolsky SA, Vetterling WT, Fnannery BP (2007) Numerical Recipes: The Art of
//       Scientific Computing. Third Edition. Cambridge University Press. 1235p.
func Qsort(arr []float64) {
	M := 7       // size of subarrays sorted by straight insertion
	NSTACK := 64 // required auxiliary storage.
	istack := make([]int, NSTACK)
	var i, j, k int
	jstack := -1
	l := 0
	n := len(arr)
	var a float64
	ir := n - 1
	for { // Insertion sort when subarray small enough.
		if ir-l < M {
			for j = l + 1; j <= ir; j++ {
				a = arr[j]
				for i = j - 1; i >= l; i-- {
					if arr[i] <= a {
						break
					}
					arr[i+1] = arr[i]
				}
				arr[i+1] = a
			}
			if jstack < 0 {
				break
			}
			ir = istack[jstack] // Pop stack and begin a new round of partitioning.
			jstack--
			l = istack[jstack]
			jstack--
		} else {
			k = (l + ir) >> 1 // Choose median of left, center, and right elements as partitioning element a. Also rearrange so that a[l] ≤ a[l+1] ≤ a[ir].
			swap(&arr[k], &arr[l+1])
			if arr[l] > arr[ir] {
				swap(&arr[l], &arr[ir])
			}
			if arr[l+1] > arr[ir] {
				swap(&arr[l+1], &arr[ir])
			}
			if arr[l] > arr[l+1] {
				swap(&arr[l], &arr[l+1])
			}
			i = l + 1 // Initialize pointers for partitioning.
			j = ir
			a = arr[l+1] // Partitioning element.
			for {        // Beginning of innermost loop.
				// Scan up to find element > a.
				for { // do i++; while (arr[i] < a);
					i++
					if arr[i] >= a {
						break
					}
				}
				// Scan down to find element < a.
				for { // do j--; while (arr[j] > a);
					j--
					if arr[j] <= a {
						break
					}
				}
				if j < i {
					break
				}
				// Pointers crossed. Partitioning complete.
				swap(&arr[i], &arr[j]) // Exchange elements.
			} // End of innermost loop.
			arr[l+1] = arr[j] // Insert partitioning element.
			arr[j] = a
			jstack += 2
			// Push pointers to larger subarray on stack; process smaller subarray immediately.

			if jstack >= NSTACK {
				chk.Panic("NSTACK=%d too small in sort.", NSTACK)
			}
			if ir-i+1 >= j-l {
				istack[jstack] = ir
				istack[jstack-1] = i
				ir = j - 1
			} else {
				istack[jstack] = j - 1
				istack[jstack-1] = l
				l = i
			}
		}
	}
}

// Qsort2 sorts an array arr[0..n-1] into ascending order using Quicksort, while making the
// corresponding rearrangment of the array brr[0..n-1]. Implementation from [1]
//   Reference:
//   [1] Press WH, Teukolsky SA, Vetterling WT, Fnannery BP (2007) Numerical Recipes: The Art of
//       Scientific Computing. Third Edition. Cambridge University Press. 1235p.
func Qsort2(arr, brr []float64) {
	M := 7       // size of subarrays sorted by straight insertion
	NSTACK := 64 // required auxiliary storage.
	istack := make([]int, NSTACK)
	var i, ir, j, k int
	jstack := -1
	l := 0
	n := len(arr)
	var a, b float64
	ir = n - 1
	for { // Insertion sort when subarray small enough.
		if ir-l < M {
			for j = l + 1; j <= ir; j++ {
				a = arr[j]
				b = brr[j]
				for i = j - 1; i >= l; i-- {
					if arr[i] <= a {
						break
					}
					arr[i+1] = arr[i]
					brr[i+1] = brr[i]
				}
				arr[i+1] = a
				brr[i+1] = b
			}
			if jstack < 0 {
				break
			}
			ir = istack[jstack] // Pop stack and begin a new round of partitioning.
			jstack--
			l = istack[jstack]
			jstack--
		} else {
			k = (l + ir) >> 1 // Choose median of left, center, and right elements as partitioning element a. Also rearrange so that a[l] ≤ a[l+1] ≤ a[ir].
			swap(&arr[k], &arr[l+1])
			swap(&brr[k], &brr[l+1])
			if arr[l] > arr[ir] {
				swap(&arr[l], &arr[ir])
				swap(&brr[l], &brr[ir])
			}
			if arr[l+1] > arr[ir] {
				swap(&arr[l+1], &arr[ir])
				swap(&brr[l+1], &brr[ir])
			}
			if arr[l] > arr[l+1] {
				swap(&arr[l], &arr[l+1])
				swap(&brr[l], &brr[l+1])
			}
			i = l + 1 // Initialize pointers for partitioning.
			j = ir
			a = arr[l+1] // Partitioning element.
			b = brr[l+1]
			for { // Beginning of innermost loop.
				// Scan up to find element > a.
				for { // do i++; while (arr[i] < a);
					i++
					if arr[i] >= a {
						break
					}
				}
				// Scan down to find element < a.
				for { // do j--; while (arr[j] > a);
					j--
					if arr[j] <= a {
						break
					}
				}
				if j < i { // Pointers crossed. Partitioning complete.
					break
				}
				swap(&arr[i], &arr[j]) // Exchange elements of both arrays.
				swap(&brr[i], &brr[j])
			}
			arr[l+1] = arr[j] // Insert partitioning element in both arrays.
			arr[j] = a
			brr[l+1] = brr[j]
			brr[j] = b
			jstack += 2
			// Push pointers to larger subarray on stack; process smaller subarray immediately.
			if jstack >= NSTACK {
				chk.Panic("NSTACK=%d too small in qsort2.", NSTACK)
			}
			if ir-i+1 >= j-l {
				istack[jstack] = ir
				istack[jstack-1] = i
				ir = j - 1
			} else {
				istack[jstack] = j - 1
				istack[jstack-1] = l
				l = i
			}
		}
	}
}

// Sorter /////////////////////////////////////////////////////////////////////////////////////////

// Sorter builds an index list to sort arrays of any type.
//   Reference:
//   [1] Press WH, Teukolsky SA, Vetterling WT, Fnannery BP (2007) Numerical Recipes: The Art of
//       Scientific Computing. Third Edition. Cambridge University Press. 1235p.
type Sorter struct {
	Index []int
}

// Init builds an index indx[0..n-1] to sort an array a[0..n-1] such that a[indx[j]] is in
// ascending order for j=0,1,...,n-1.
//   Input:
//     n    -- number of items in the array to be sorted. n must be ≤ len(a)
//     less -- a function that returns true if a[i] < a[j]
//   Reference:
//   [1] Press WH, Teukolsky SA, Vetterling WT, Fnannery BP (2007) Numerical Recipes: The Art of
//       Scientific Computing. Third Edition. Cambridge University Press. 1235p.
func (o *Sorter) Init(n int, less func(i, j int) bool) {
	M := 7       // size of subarrays sorted by straight insertion
	NSTACK := 64 // required auxiliary storage.
	istack := make([]int, NSTACK)
	o.Index = make([]int, n)
	jstack := -1
	l := 0
	ir := n - 1
	var i, j, k, indxt int
	for j = 0; j < n; j++ {
		o.Index[j] = j
	}
	for { // Insertion sort when subarray small enough.
		if ir-l < M {
			for j = l + 1; j <= ir; j++ {
				indxt = o.Index[j]
				for i = j - 1; i >= l; i-- {
					if less(o.Index[i], indxt) {
						break
					}
					o.Index[i+1] = o.Index[i]
				}
				o.Index[i+1] = indxt
			}
			if jstack < 0 {
				break
			}
			ir = istack[jstack] // Pop stack and begin a new round of partitioning.
			jstack--
			l = istack[jstack]
			jstack--
		} else {
			k = (l + ir) >> 1 // Choose median of left, center, and right elements as partitioning element a. Also rearrange so that a[l] ≤ a[l+1] ≤ a[ir].
			iswap(&o.Index[k], &o.Index[l+1])
			if !less(o.Index[l], o.Index[ir]) {
				iswap(&o.Index[l], &o.Index[ir])
			}
			if !less(o.Index[l+1], o.Index[ir]) {
				iswap(&o.Index[l+1], &o.Index[ir])
			}
			if !less(o.Index[l], o.Index[l+1]) {
				iswap(&o.Index[l], &o.Index[l+1])
			}
			i = l + 1 // Initialize pointers for partitioning.
			j = ir
			indxt = o.Index[l+1]
			for {
				// Scan up to find element > a.
				for { // do i++; while (a[indx[i]] < a[indxt]);
					i++
					if !less(o.Index[i], indxt) {
						break
					}
				}
				// Scan down to find element < a.
				for { // do j--; while (a[o.indx[j]] > a[indxt]);
					j--
					if less(o.Index[j], indxt) {
						break
					}
				}
				if j < i {
					break
				}
				// Pointers crossed. Partitioning complete.
				iswap(&o.Index[i], &o.Index[j])
			}
			o.Index[l+1] = o.Index[j] // Insert partitioning element.
			o.Index[j] = indxt
			jstack += 2
			// Push pointers to larger subarray on stack; process smaller subarray immediately.
			if jstack >= NSTACK {
				chk.Panic("NSTACK=%d too small in IndexSort.Build.", NSTACK)
			}
			if ir-i+1 >= j-l {
				istack[jstack] = ir
				istack[jstack-1] = i
				ir = j - 1
			} else {
				istack[jstack] = j - 1
				istack[jstack-1] = l
				l = i
			}
		}
	}
}

// GetSorted returns a copy of array 'a' sorted according to the previously built index.
// NOTE: the copy may be smaller if the index was built with a smaller set
func (o *Sorter) GetSorted(a []float64) (b []float64) {
	b = make([]float64, len(o.Index))
	for i := 0; i < len(o.Index); i++ {
		b[i] = a[o.Index[i]]
	}
	return
}

// GetSortedI returns a copy of array 'a' sorted according to the previously built index.
// NOTE: the copy may be smaller if the index was built with a smaller set
func (o *Sorter) GetSortedI(a []int) (b []int) {
	b = make([]int, len(o.Index))
	for i := 0; i < len(o.Index); i++ {
		b[i] = a[o.Index[i]]
	}
	return
}
