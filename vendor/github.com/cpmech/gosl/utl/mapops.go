// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utl

// NOTE: Like slices, maps hold references to an underlying data structure.
// If you pass a map to a function that changes the contents of the map,
// the changes will be visible in the caller [Effective Go].

// IntIntsMapAppend appends a new item to a map of slice.
//  Note: this function creates a new slice in the map if key is not found.
func IntIntsMapAppend(m map[int][]int, key int, item int) {
	if slice, ok := m[key]; ok {
		slice = append(slice, item)
		m[key] = slice
	} else {
		m[key] = []int{item}
	}
}

// StrIntsMapAppend appends a new item to a map of slice.
//  Note: this function creates a new slice in the map if key is not found.
func StrIntsMapAppend(m map[string][]int, key string, item int) {
	if slice, ok := m[key]; ok {
		slice = append(slice, item)
		m[key] = slice
	} else {
		m[key] = []int{item}
	}
}

// StrFltsMapAppend appends a new item to a map of slice.
//  Note: this function creates a new slice in the map if key is not found.
func StrFltsMapAppend(m map[string][]float64, key string, item float64) {
	if slice, ok := m[key]; ok {
		slice = append(slice, item)
		m[key] = slice
	} else {
		m[key] = []float64{item}
	}
}
