// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utl

import "math"

// OutFcnType is a function that sets the "u" array with an output @ time "t"
type OutFcnType func(u []float64, t float64)

// Outputter helps with the output of (numerical) results
//
//   The time loop can be something like:
//     t := 0.0
//     for tidx := 0; tidx < outper.Nsteps; tidx++ {
//         t += outper.Dt
//         ... // do something
//         outper.MaybeNow(tidx, t)
//     }
//
type Outputter struct {
	Dt     float64     // time step
	DtOut  float64     // time step for output
	Tmax   float64     // final time (eventually increased to accommodate all time steps)
	Nsteps int         // number of time steps
	Every  int         // increment to output after every time step
	Tidx   int         // the index of the time step for the next output
	Nmax   int         // max number of outputs
	Idx    int         // index in the output arrays for the next output == num.outputs in the end
	T      []float64   // saved t values len = Nmax
	U      [][]float64 // saved u values @ t. [Nmax][nu]
	Fcn    OutFcnType  // function to process output. may be nil
}

// NewOutputter creates a new Outputter
//  dt     -- time step
//  dtOut  -- time step for output
//  tmax   -- final time (of simulation)
//  outFcn -- callback function to perform output (may be nil)
//  NOTE: a first output will be processed if outFcn != nil
func NewOutputter(dt, dtOut, tmax float64, nu int, outFcn OutFcnType) (o *Outputter) {
	if dtOut < dt {
		dtOut = dt
	}
	o = new(Outputter)
	o.Dt = dt
	o.DtOut = dtOut
	o.Nsteps = int(math.Ceil(tmax / o.Dt))
	o.Tmax = float64(o.Nsteps) * o.Dt // fix tmax
	o.Every = int(o.DtOut / o.Dt)
	o.Nmax = int(math.Ceil(float64(o.Nsteps)/float64(o.Every))) + 1
	o.T = make([]float64, o.Nmax)
	o.U = Alloc(o.Nmax, nu)
	if outFcn != nil {
		o.Fcn = outFcn
		o.Fcn(o.U[o.Idx], 0)
		if o.Every > 1 {
			o.Tidx = o.Every - 1 // use -1 here only for the first output
		}
		o.Idx++
	}
	return
}

// MaybeNow process the output if tidx == NextIdxT
func (o *Outputter) MaybeNow(tidx int, t float64) {
	if o.Fcn == nil {
		return
	}
	if tidx == o.Tidx || tidx == o.Nsteps-1 { // always process the last one
		o.T[o.Idx] = t
		o.Fcn(o.U[o.Idx], t)
		o.Tidx += o.Every
		o.Idx++
	}
}

// GetITout returns indices and output times
//  Input:
//    all_output_times  -- array with all output times. ex: [0,0.1,0.2,0.22,0.3,0.4]
//    time_stations_out -- time stations for output: ex: [0,0.2,0.4]
//    tol               -- tolerance to compare times
//  Output:
//    iout -- indices of times in all_output_times
//    tout -- times corresponding to iout
//  Notes:
//    use -1 in all_output_times to enforce output of last timestep
func GetITout(allOutputTimes, timeStationsOut []float64, tol float64) (I []int, T []float64) {
	lowerIndex := 0                     // lower index in all_output_times
	lenAoTimes := len(allOutputTimes)   // length of a_o_times
	for _, t := range timeStationsOut { // for all requested output times
		if t < 0 { // final time
			k := len(allOutputTimes) - 1     // last output time index
			I = append(I, k)                 // append index to iout
			T = append(T, allOutputTimes[k]) // append last output time
			continue                         // skip search
		}
		for k := lowerIndex; k < lenAoTimes; k++ { // search within a_o_times
			if math.Abs(t-allOutputTimes[k]) < tol { // found near match
				lowerIndex++                     // update index
				I = append(I, k)                 // add index to iout
				T = append(T, allOutputTimes[k]) // add time to tout
				break                            // stop searching for this 't'
			}
			if allOutputTimes[k] > t { // failed to search for 't'
				lowerIndex = k // update idx to start from here on
				break          // skip this 't' and try the next one
			}
		}
	}
	return
}

// GetStrides returns nReq indices from 0 (inclusive) to nTotal (inclusive)
//   Input:
//     nTotal -- total number of intices
//     nReq -- required indices
//   Example:
//     GetStrides(2001, 5) => [0 400 800 1200 1600 2000 2001]
//
//   NOTE: GetStrides will always include nTotal as the last item in I
//
func GetStrides(nTotal, nReq int) (I []int) {
	if nReq > nTotal {
		nReq = nTotal
	}
	lt := nTotal / nReq
	if lt < 1 {
		lt = 1
	}
	I = IntRange3(0, nTotal, lt)
	if I[len(I)-1] != nTotal {
		I = append(I, nTotal)
	}
	return
}
