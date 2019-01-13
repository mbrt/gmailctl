// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utl

import (
	"math"
	"math/rand"

	"github.com/cpmech/gosl/chk"
)

// ParetoMin compares two vectors using Pareto's optimal criterion
//  Note: minimum dominates (is better)
func ParetoMin(u, v []float64) (uDominates, vDominates bool) {
	chk.IntAssert(len(u), len(v))
	uHasAllLeq := true // all u values are less-than or equal-to v values
	uHasOneLe := false // u has at least one value less-than v
	vHasAllLeq := true // all v values are less-than or equalt-to u values
	vHasOneLe := false // v has at least one value less-than u
	for i := 0; i < len(u); i++ {
		if u[i] > v[i] {
			uHasAllLeq = false
			vHasOneLe = true
		}
		if u[i] < v[i] {
			uHasOneLe = true
			vHasAllLeq = false
		}
	}
	if uHasAllLeq && uHasOneLe {
		uDominates = true
	}
	if vHasAllLeq && vHasOneLe {
		vDominates = true
	}
	return
}

// ParetoMinProb compares two vectors using Pareto's optimal criterion
// φ ∃ [0,1] is a scaling factor that helps v win even if it's not smaller.
// If φ==0, deterministic analysis is carried out. If φ==1, probabilistic analysis is carried out.
// As φ → 1, v "gets more help".
//  Note: (1) minimum dominates (is better)
//        (2) v dominates if !uDominates
func ParetoMinProb(u, v []float64, φ float64) (uDominates bool) {
	chk.IntAssert(len(u), len(v))
	var pu float64
	for i := 0; i < len(u); i++ {
		pu += ProbContestSmall(u[i], v[i], φ)
	}
	pu /= float64(len(u))
	if FlipCoin(pu) {
		uDominates = true
	}
	return
}

// ProbContestSmall computes the probability for a contest between u and v where u wins if it's
// the smaller value. φ ∃ [0,1] is a scaling factor that helps v win even if it's not smaller.
// If φ==0, deterministic analysis is carried out. If φ==1, probabilistic analysis is carried out.
// As φ → 1, v "gets more help".
func ProbContestSmall(u, v, φ float64) float64 {
	u = math.Atan(u)/math.Pi + 1.5
	v = math.Atan(v)/math.Pi + 1.5
	if u < v {
		return v / (v + φ*u)
	}
	if u > v {
		return φ * v / (φ*v + u)
	}
	return 0.5
}

// FlipCoin generates a Bernoulli variable; throw a coin with probability p
func FlipCoin(p float64) bool {
	if p == 1.0 {
		return true
	}
	if p == 0.0 {
		return false
	}
	if rand.Float64() <= p {
		return true
	}
	return false
}

// ParetoFront computes the Pareto optimal front
//  Input:
//   Ovs -- [nsamples][ndim] objective values
//  Output:
//   front -- indices of pareto front
//  Note: this function is slow for large sets
func ParetoFront(Ovs [][]float64) (front []int) {
	dominated := map[int]bool{}
	nsamples := len(Ovs)
	for i := 0; i < nsamples; i++ {
		dominated[i] = false
	}
	for i := 0; i < nsamples; i++ {
		for j := i + 1; j < nsamples; j++ {
			uDominates, vDominates := ParetoMin(Ovs[i], Ovs[j])
			if uDominates {
				dominated[j] = true
			}
			if vDominates {
				dominated[i] = true
			}
		}
	}
	nondom := 0
	for i := 0; i < nsamples; i++ {
		if !dominated[i] {
			nondom++
		}
	}
	front = make([]int, nondom)
	k := 0
	for i := 0; i < nsamples; i++ {
		if !dominated[i] {
			front[k] = i
			k++
		}
	}
	return
}
