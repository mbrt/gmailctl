// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"flag"

	"github.com/cpmech/gosl/chk"
)

// ArgToFilename parses an argument as a filename
//  Input:
//   idxArg    -- index of argument; e.g. 0==first, 1==second, etc.
//   fnDefault -- default filename; can be ""
//   ext       -- the file extension to be added; e.g. ".sim"
//   check     -- check for null filename
//  Output:
//   filename -- the filename with extension added
//   fnkey    -- filename key == filename without extension
//  Notes:
//   The first first argument may be a file with extension or not.
//  Examples:
//   If the first argument is "simulation.sim" or "simulation" (with ext=".sim")
//   then the results are: filename="simulation.sim" and fnkey="simulation"
func ArgToFilename(idxArg int, fnDefault, ext string, check bool) (filename, fnkey string) {
	if !flag.Parsed() {
		flag.Parse()
	}
	filename = fnDefault
	if len(flag.Args()) > idxArg {
		filename = flag.Arg(idxArg)
	}
	if FnExt(filename) == "" {
		filename += ext
	}
	fnkey = FnKey(filename)
	if check {
		if filename == "" || fnkey == "" {
			chk.Panic("filename must be given as first argument")
		}
	}
	return
}

// ArgToFloat parses an argument as a float64 value
//  Input:
//   idxArg       -- index of argument; e.g. 0==first, 1==second, etc.
//   defaultValue -- default value
func ArgToFloat(idxArg int, defaultValue float64) float64 {
	if !flag.Parsed() {
		flag.Parse()
	}
	if len(flag.Args()) > idxArg {
		return Atof(flag.Arg(idxArg))
	}
	return defaultValue
}

// ArgToInt parses an argument as an integer value
//  Input:
//   idxArg       -- index of argument; e.g. 0==first, 1==second, etc.
//   defaultValue -- default value
func ArgToInt(idxArg int, defaultValue int) int {
	if !flag.Parsed() {
		flag.Parse()
	}
	if len(flag.Args()) > idxArg {
		return Atoi(flag.Arg(idxArg))
	}
	return defaultValue
}

// ArgToBool parses an argument as a boolean value
//  Input:
//   idxArg       -- index of argument; e.g. 0==first, 1==second, etc.
//   defaultValue -- default value
func ArgToBool(idxArg int, defaultValue bool) bool {
	if !flag.Parsed() {
		flag.Parse()
	}
	if len(flag.Args()) > idxArg {
		return Atob(flag.Arg(idxArg))
	}
	return defaultValue
}

// ArgToString parses an argument as a string
//  Input:
//   idxArg       -- index of argument; e.g. 0==first, 1==second, etc.
//   defaultValue -- default value
func ArgToString(idxArg int, defaultValue string) string {
	if !flag.Parsed() {
		flag.Parse()
	}
	if len(flag.Args()) > idxArg {
		return flag.Arg(idxArg)
	}
	return defaultValue
}

// ArgsTable prints a nice table with input arguments
//  Input:
//   title -- title of table; e.g. INPUT ARGUMENTS
//   data  -- sets of THREE items in the following order:
//                 description, key, value, ...
//                 description, key, value, ...
//                      ...
//                 description, key, value, ...
func ArgsTable(title string, data ...interface{}) (table string) {
	if len(data) < 3 {
		return
	}
	ndat := len(data)
	nlines := ndat / 3
	sizes := []int{0, 0, 0}
	for i := 0; i < nlines; i++ {
		if i*3+2 >= ndat {
			return "ArgsTable: input arguments are not a multiple of 3\n"
		}
		dsc := data[i*3]
		key := data[i*3+1]
		val := data[i*3+2]
		sizes[0] = imax(sizes[0], len(Sf("%v", dsc)))
		sizes[1] = imax(sizes[1], len(Sf("%v", key)))
		sizes[2] = imax(sizes[2], len(Sf("%v", val)))
	}
	sizes[2] = imax(sizes[2], 5) // 5 => 'value'
	strfmt := Sf("%%%dv  %%%dv   %%%dv\n", sizes[0]+1, sizes[1]+1, sizes[2]+1)
	n := sizes[0] + sizes[1] + sizes[2] + 3 + 5
	l := len(title)
	m := (n - l) / 2
	table += StrSpaces(m)
	table += title + "\n"
	table += StrThickLine(n)
	table += Sf(strfmt, "description", "key", "value")
	table += StrThinLine(n)
	for i := 0; i < nlines; i++ {
		dsc := data[i*3]
		key := data[i*3+1]
		val := data[i*3+2]
		table += Sf(strfmt, dsc, key, val)
	}
	table += StrThickLine(n)
	return
}
