// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

import (
	"testing"

	"github.com/cpmech/gosl/chk"
	"github.com/cpmech/gosl/io"
)

func Test_munkres01(tst *testing.T) {

	//verbose()
	chk.PrintTitle("munkres01")

	C := [][]float64{
		{1, 2, 3},
		{2, 4, 6},
		{3, 6, 9},
	}
	Ccor := [][]float64{
		{0, 1, 2},
		{0, 2, 4},
		{0, 3, 6},
	}
	Mcor := [][]MaskType{
		{StarType, NoneType, NoneType},
		{NoneType, NoneType, NoneType},
		{NoneType, NoneType, NoneType},
	}

	var mnk Munkres
	mnk.Init(len(C), len(C[0]))
	mnk.SetCostMatrix(C)

	// 1:
	io.PfYel("1: after step 0:\n")
	io.Pf("%v", mnk.StrCostMatrix())

	// 2: step 1
	nextStep := mnk.step1()
	io.PfYel("\n2: after step 1:\n")
	io.Pf("%v", mnk.StrCostMatrix())
	chk.IntAssert(nextStep, 2)
	chk.Deep2(tst, "C", 1e-17, mnk.C, Ccor)
	chk.Bools(tst, "row_covered", mnk.rowCovered, []bool{false, false, false})
	chk.Bools(tst, "col_covered", mnk.colCovered, []bool{false, false, false})

	// 3: step 2
	nextStep = mnk.step2()
	io.PfYel("\n3: after step 2:\n")
	io.Pf("%v", mnk.StrCostMatrix())
	chk.IntAssert(nextStep, 3)
	chk.Deep2(tst, "C", 1e-17, mnk.C, Ccor)
	checkMaskMatrix(tst, "M", mnk.M, Mcor)
	chk.Bools(tst, "row_covered", mnk.rowCovered, []bool{false, false, false})
	chk.Bools(tst, "col_covered", mnk.colCovered, []bool{false, false, false})

	// 4: step 3
	nextStep = mnk.step3()
	io.PfYel("\n4: after step 3:\n")
	io.Pf("%v", mnk.StrCostMatrix())
	chk.IntAssert(nextStep, 4)
	chk.Deep2(tst, "C", 1e-17, mnk.C, Ccor)
	checkMaskMatrix(tst, "M", mnk.M, Mcor)
	chk.Bools(tst, "row_covered", mnk.rowCovered, []bool{false, false, false})
	chk.Bools(tst, "col_covered", mnk.colCovered, []bool{true, false, false})

	// 5: step 4
	nextStep = mnk.step4()
	io.PfYel("\n5: after step 4:\n")
	io.Pf("%v", mnk.StrCostMatrix())
	chk.IntAssert(nextStep, 6)
	chk.Deep2(tst, "C", 1e-17, mnk.C, Ccor)
	checkMaskMatrix(tst, "M", mnk.M, Mcor)
	chk.Bools(tst, "row_covered", mnk.rowCovered, []bool{false, false, false})
	chk.Bools(tst, "col_covered", mnk.colCovered, []bool{true, false, false})

	// 6: step 6
	Ccor = [][]float64{
		{0, 0, 1},
		{0, 1, 3},
		{0, 2, 5},
	}
	nextStep = mnk.step6()
	io.PfYel("\n6: after step 6:\n")
	io.Pf("%v", mnk.StrCostMatrix())
	chk.IntAssert(nextStep, 4)
	chk.Deep2(tst, "C", 1e-17, mnk.C, Ccor)
	checkMaskMatrix(tst, "M", mnk.M, Mcor)
	chk.Bools(tst, "row_covered", mnk.rowCovered, []bool{false, false, false})
	chk.Bools(tst, "col_covered", mnk.colCovered, []bool{true, false, false})

	// 7: step 4 again (1)
	Mcor = [][]MaskType{
		{StarType, PrimeType, NoneType},
		{PrimeType, NoneType, NoneType},
		{NoneType, NoneType, NoneType},
	}
	nextStep = mnk.step4()
	io.PfYel("\n7: after step 4 again (1):\n")
	io.Pf("%v", mnk.StrCostMatrix())
	chk.IntAssert(nextStep, 5)
	chk.Deep2(tst, "C", 1e-17, mnk.C, Ccor)
	checkMaskMatrix(tst, "M", mnk.M, Mcor)
	chk.Bools(tst, "row_covered", mnk.rowCovered, []bool{true, false, false})
	chk.Bools(tst, "col_covered", mnk.colCovered, []bool{false, false, false})

	// 8: step 5
	Mcor = [][]MaskType{
		{NoneType, StarType, NoneType},
		{StarType, NoneType, NoneType},
		{NoneType, NoneType, NoneType},
	}
	nextStep = mnk.step5()
	io.PfYel("\n8: after step 5:\n")
	io.Pf("%v", mnk.StrCostMatrix())
	chk.IntAssert(nextStep, 3)
	chk.Deep2(tst, "C", 1e-17, mnk.C, Ccor)
	checkMaskMatrix(tst, "M", mnk.M, Mcor)
	chk.Bools(tst, "row_covered", mnk.rowCovered, []bool{false, false, false})
	chk.Bools(tst, "col_covered", mnk.colCovered, []bool{false, false, false})

	// 9: step 3 again (1)
	nextStep = mnk.step3()
	io.PfYel("\n9: after step 3 again (1):\n")
	io.Pf("%v", mnk.StrCostMatrix())
	chk.IntAssert(nextStep, 4)
	chk.Deep2(tst, "C", 1e-17, mnk.C, Ccor)
	checkMaskMatrix(tst, "M", mnk.M, Mcor)
	chk.Bools(tst, "row_covered", mnk.rowCovered, []bool{false, false, false})
	chk.Bools(tst, "col_covered", mnk.colCovered, []bool{true, true, false})

	// 10: step 4 again (2)
	nextStep = mnk.step4()
	io.PfYel("\n10: after step 4 again (2):\n")
	io.Pf("%v", mnk.StrCostMatrix())
	chk.IntAssert(nextStep, 6)
	chk.Deep2(tst, "C", 1e-17, mnk.C, Ccor)
	checkMaskMatrix(tst, "M", mnk.M, Mcor)
	chk.Bools(tst, "row_covered", mnk.rowCovered, []bool{false, false, false})
	chk.Bools(tst, "col_covered", mnk.colCovered, []bool{true, true, false})

	// 11: step 6 again (1)
	Ccor = [][]float64{
		{0, 0, 0},
		{0, 1, 2},
		{0, 2, 4},
	}
	nextStep = mnk.step6()
	io.PfYel("\n11: after step 6 again (1):\n")
	io.Pf("%v", mnk.StrCostMatrix())
	chk.IntAssert(nextStep, 4)
	chk.Deep2(tst, "C", 1e-17, mnk.C, Ccor)
	checkMaskMatrix(tst, "M", mnk.M, Mcor)
	chk.Bools(tst, "row_covered", mnk.rowCovered, []bool{false, false, false})
	chk.Bools(tst, "col_covered", mnk.colCovered, []bool{true, true, false})

	// 12: step 4 again (3)
	Mcor = [][]MaskType{
		{NoneType, StarType, PrimeType},
		{StarType, NoneType, NoneType},
		{NoneType, NoneType, NoneType},
	}
	nextStep = mnk.step4()
	io.PfYel("\n12: after step 4 again (3):\n")
	io.Pf("%v", mnk.StrCostMatrix())
	chk.IntAssert(nextStep, 6)
	chk.Deep2(tst, "C", 1e-17, mnk.C, Ccor)
	checkMaskMatrix(tst, "M", mnk.M, Mcor)
	chk.Bools(tst, "row_covered", mnk.rowCovered, []bool{true, false, false})
	chk.Bools(tst, "col_covered", mnk.colCovered, []bool{true, false, false})

	// 13: step 6 again (2)
	Ccor = [][]float64{
		{1, 0, 0},
		{0, 0, 1},
		{0, 1, 3},
	}
	nextStep = mnk.step6()
	io.PfYel("\n13: after step 6 again (2):\n")
	io.Pf("%v", mnk.StrCostMatrix())
	chk.IntAssert(nextStep, 4)
	chk.Deep2(tst, "C", 1e-17, mnk.C, Ccor)
	checkMaskMatrix(tst, "M", mnk.M, Mcor)
	chk.Bools(tst, "row_covered", mnk.rowCovered, []bool{true, false, false})
	chk.Bools(tst, "col_covered", mnk.colCovered, []bool{true, false, false})

	// 14: step 4 again (4)
	Mcor = [][]MaskType{
		{NoneType, StarType, PrimeType},
		{StarType, PrimeType, NoneType},
		{PrimeType, NoneType, NoneType},
	}
	nextStep = mnk.step4()
	io.PfYel("\n14: after step 4 again (4):\n")
	io.Pf("%v", mnk.StrCostMatrix())
	chk.IntAssert(nextStep, 5)
	chk.Deep2(tst, "C", 1e-17, mnk.C, Ccor)
	checkMaskMatrix(tst, "M", mnk.M, Mcor)
	chk.Bools(tst, "row_covered", mnk.rowCovered, []bool{true, true, false})
	chk.Bools(tst, "col_covered", mnk.colCovered, []bool{false, false, false})

	// 15: step 5 again (1)
	Mcor = [][]MaskType{
		{NoneType, NoneType, StarType},
		{NoneType, StarType, NoneType},
		{StarType, NoneType, NoneType},
	}
	nextStep = mnk.step5()
	io.PfYel("\n15: after step 5 again (1):\n")
	io.Pf("%v", mnk.StrCostMatrix())
	chk.IntAssert(nextStep, 3)
	chk.Deep2(tst, "C", 1e-17, mnk.C, Ccor)
	checkMaskMatrix(tst, "M", mnk.M, Mcor)
	chk.Bools(tst, "row_covered", mnk.rowCovered, []bool{false, false, false})
	chk.Bools(tst, "col_covered", mnk.colCovered, []bool{false, false, false})

	// 15: step 3 again (2)
	nextStep = mnk.step3()
	io.PfYel("\n15: after step 3 again (2):\n")
	io.Pf("%v", mnk.StrCostMatrix())
	chk.IntAssert(nextStep, 7)
	chk.Deep2(tst, "C", 1e-17, mnk.C, Ccor)
	checkMaskMatrix(tst, "M", mnk.M, Mcor)
	chk.Bools(tst, "row_covered", mnk.rowCovered, []bool{false, false, false})
	chk.Bools(tst, "col_covered", mnk.colCovered, []bool{true, true, true})
}

func Test_munkres02(tst *testing.T) {

	//verbose()
	chk.PrintTitle("munkres02")

	C := [][]float64{
		{2, 1},
		{1, 1},
	}
	var mnk Munkres
	mnk.Init(len(C), len(C[0]))
	mnk.SetCostMatrix(C)
	mnk.Run()
	chk.Ints(tst, "links", mnk.Links, []int{1, 0}) // 0 goes with 1 and 1 goes with 0
	chk.Float64(tst, "cost", 1e-17, mnk.Cost, 2)

	C = [][]float64{
		{2, 2},
		{4, 3},
	}
	mnk.Init(len(C), len(C[0]))
	mnk.SetCostMatrix(C)
	mnk.Run()
	chk.Ints(tst, "links", mnk.Links, []int{0, 1}) // 0 goes 0 and 1 goes with 1
	chk.Float64(tst, "cost", 1e-17, mnk.Cost, 5)

	C = [][]float64{
		{2, 2},
		{1, 3},
	}
	mnk.Init(len(C), len(C[0]))
	mnk.SetCostMatrix(C)
	mnk.Run()
	chk.Ints(tst, "links", mnk.Links, []int{1, 0}) // 0 goes 1 and 1 goes 0
	chk.Float64(tst, "cost", 1e-17, mnk.Cost, 3)

	C = [][]float64{
		{2, 1},
		{2, 1},
		{1, 1},
		{1, 1},
	}
	mnk.Init(len(C), len(C[0]))
	mnk.SetCostMatrix(C)
	mnk.Run()
	chk.Ints(tst, "links", mnk.Links, []int{-1, -1, 1, 0})
	chk.Float64(tst, "cost", 1e-17, mnk.Cost, 2)

	C = [][]float64{
		{1, 2, 3},
		{6, 5, 4},
	}
	mnk.Init(len(C), len(C[0]))
	mnk.SetCostMatrix(C)
	mnk.Run()
	chk.Ints(tst, "links", mnk.Links, []int{0, 2}) // 0 goes with 0 and 1 goes with 2
	chk.Float64(tst, "cost", 1e-17, mnk.Cost, 5)

	C = [][]float64{
		{1, 2, 3},
		{6, 5, 4},
		{1, 1, 1},
	}
	mnk.Init(len(C), len(C[0]))
	mnk.SetCostMatrix(C)
	mnk.Run()
	chk.Ints(tst, "links", mnk.Links, []int{0, 2, 1}) // 0 goes with 0, 1 goes with 2 and 2 goes with 1
	chk.Float64(tst, "cost", 1e-17, mnk.Cost, 6)

	C = [][]float64{
		{2, 4, 7, 9},
		{3, 9, 5, 1},
		{8, 2, 9, 7},
	}
	mnk.Init(len(C), len(C[0]))
	mnk.SetCostMatrix(C)
	mnk.Run()
	chk.Ints(tst, "links", mnk.Links, []int{0, 3, 1})
	chk.Float64(tst, "cost", 1e-17, mnk.Cost, 5)

	C = [][]float64{
		{1, 2, 3},
		{2, 4, 6},
		{3, 6, 9},
	}
	Ccor := [][]float64{
		{1, 0, 0},
		{0, 0, 1},
		{0, 1, 3},
	}
	mnk.Init(len(C), len(C[0]))
	mnk.SetCostMatrix(C)
	mnk.Run()
	chk.Deep2(tst, "C", 1e-17, mnk.C, Ccor)
	chk.Ints(tst, "links", mnk.Links, []int{2, 1, 0}) // 0 goes with 2, 1 goes with 1 and 2 goes with 0
	chk.Float64(tst, "cost", 1e-17, mnk.Cost, 10)

	// from https://projecteuler.net/index.php?section=problems&id=345
	C = [][]float64{
		{7, 53, 183, 439, 863},
		{497, 383, 563, 79, 973},
		{287, 63, 343, 169, 583},
		{627, 343, 773, 959, 943},
		{767, 473, 103, 699, 303},
	}
	for i := 0; i < len(C); i++ {
		for j := 0; j < len(C[i]); j++ {
			C[i][j] *= -1
		}
	}
	mnk.Init(len(C), len(C[0]))
	mnk.SetCostMatrix(C)
	mnk.Run()
	chk.Ints(tst, "links", mnk.Links, []int{4, 1, 2, 3, 0})
}

func Test_munkres03(tst *testing.T) {

	//verbose()
	chk.PrintTitle("munkres03. Euler problem 345")

	C := [][]float64{
		{7, 53, 183, 439, 863, 497, 383, 563, 79, 973, 287, 63, 343, 169, 583},
		{627, 343, 773, 959, 943, 767, 473, 103, 699, 303, 957, 703, 583, 639, 913},
		{447, 283, 463, 29, 23, 487, 463, 993, 119, 883, 327, 493, 423, 159, 743},
		{217, 623, 3, 399, 853, 407, 103, 983, 89, 463, 290, 516, 212, 462, 350},
		{960, 376, 682, 962, 300, 780, 486, 502, 912, 800, 250, 346, 172, 812, 350},
		{870, 456, 192, 162, 593, 473, 915, 45, 989, 873, 823, 965, 425, 329, 803},
		{973, 965, 905, 919, 133, 673, 665, 235, 509, 613, 673, 815, 165, 992, 326},
		{322, 148, 972, 962, 286, 255, 941, 541, 265, 323, 925, 281, 601, 95, 973},
		{445, 721, 11, 525, 473, 65, 511, 164, 138, 672, 18, 428, 154, 448, 848},
		{414, 456, 310, 312, 798, 104, 566, 520, 302, 248, 694, 976, 430, 392, 198},
		{184, 829, 373, 181, 631, 101, 969, 613, 840, 740, 778, 458, 284, 760, 390},
		{821, 461, 843, 513, 17, 901, 711, 993, 293, 157, 274, 94, 192, 156, 574},
		{34, 124, 4, 878, 450, 476, 712, 914, 838, 669, 875, 299, 823, 329, 699},
		{815, 559, 813, 459, 522, 788, 168, 586, 966, 232, 308, 833, 251, 631, 107},
		{813, 883, 451, 509, 615, 77, 281, 613, 459, 205, 380, 274, 302, 35, 805},
	}
	for i := 0; i < len(C); i++ {
		for j := 0; j < len(C[i]); j++ {
			C[i][j] *= -1
		}
	}

	var mnk Munkres
	mnk.Init(len(C), len(C[0]))
	mnk.SetCostMatrix(C)
	mnk.Run()
	io.Pforan("links = %v\n", mnk.Links)
	io.Pforan("cost = %v  (13938)\n", -mnk.Cost)
	chk.Ints(tst, "links", mnk.Links, []int{9, 10, 7, 4, 3, 0, 13, 2, 14, 11, 6, 5, 12, 8, 1})
	chk.Float64(tst, "cost", 1e-17, mnk.Cost, -13938)
}

func checkMaskMatrix(tst *testing.T, msg string, res, correct [][]MaskType) {
	if len(res) != len(correct) {
		io.Pf("%s. len(res)=%d != len(correct)=%d\n", msg, len(res), len(correct))
		tst.Errorf("%s failed: res and correct matrices have different lengths. %d != %d", msg, len(res), len(correct))
		return
	}
	for i := 0; i < len(res); i++ {
		if len(res[i]) != len(correct[i]) {
			io.Pf("%s. len(res[%d])=%d != len(correct[%d])=%d\n", msg, i, len(res[i]), i, len(correct[i]))
			tst.Errorf("%s failed: matrices have different number of columns", msg)
			return
		}
		for j := 0; j < len(res[i]); j++ {
			if res[i][j] != correct[i][j] {
				io.Pf("[%d,%d] %v != %v\n", i, j, res[i][j], correct[i][j])
				tst.Errorf("%s failed: different int matrices:\n [%d,%d] item is wrong: %v != %v", msg, i, j, res[i][j], correct[i][j])
				return
			}
		}
	}
	chk.PrintOk(msg)
}

func Test_munkres04(tst *testing.T) {

	//verbose()
	chk.PrintTitle("munkres04. row and column matrices")

	C := [][]float64{
		{1.0},
		{2.0},
		{0.5},
		{3.0},
		{4.0},
	}

	var mnk Munkres
	mnk.Init(len(C), len(C[0]))
	mnk.SetCostMatrix(C)
	mnk.Run()
	io.Pforan("links = %v\n", mnk.Links)
	io.Pforan("cost = %v  (13938)\n", mnk.Cost)
	chk.Ints(tst, "links", mnk.Links, []int{-1, -1, 0, -1, -1})
	chk.Float64(tst, "cost", 1e-17, mnk.Cost, 0.5)

	C = [][]float64{
		{1.0, 2.0, 0.5, 3.0, 4.0},
	}
	mnk.Init(len(C), len(C[0]))
	mnk.SetCostMatrix(C)
	mnk.Run()
	io.Pforan("links = %v\n", mnk.Links)
	io.Pforan("cost = %v  (13938)\n", mnk.Cost)
	chk.Ints(tst, "links", mnk.Links, []int{2})
	chk.Float64(tst, "cost", 1e-17, mnk.Cost, 0.5)
}

func Test_munkres05(tst *testing.T) {

	//verbose()
	chk.PrintTitle("munkres05. issue (22/Nov/2016)") // fixed: use square matrix internally

	C := [][]float64{
		{11757.0, 6957.0},
		{28985.0, 24171.0},
		{33857.0, 29057.0},
	}

	D := [][]float64{
		{11757.0, 6957.0},
		{33857.0, 29057.0},
		{28985.0, 24171.0},
	}

	var mnkC Munkres
	mnkC.Init(len(C), len(C[0]))
	mnkC.SetCostMatrix(C)
	mnkC.Run()
	io.Pforan("C: links = %v\n", mnkC.Links)
	io.Pforan("C: cost = %v\n", mnkC.Cost)
	chk.Ints(tst, "C: links", mnkC.Links, []int{0, 1, -1})
	chk.Float64(tst, "C: cost", 1e-17, mnkC.Cost, 35928)

	io.Pf("\n")

	var mnkD Munkres
	mnkD.Init(len(D), len(D[0]))
	mnkD.SetCostMatrix(D)
	mnkD.Run()
	io.Pforan("D: links = %v\n", mnkD.Links)
	io.Pforan("D: cost = %v\n", mnkD.Cost)
	chk.Ints(tst, "D: links", mnkD.Links, []int{0, -1, 1})
	chk.Float64(tst, "D: cost", 1e-17, mnkD.Cost, 35928)
}
