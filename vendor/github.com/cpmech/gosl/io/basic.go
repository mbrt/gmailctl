// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package io (input/output) implements auxiliary functions for printing,
// parsing, handling files, directories, etc.
package io

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/cpmech/gosl/chk"
)

var (
	// Verbose activates display of messages on console
	Verbose = true

	// ColorsOn activates use of colors on console
	ColorsOn = true
)

// IntSf is the Sprintf for a slice of integers (without brackets)
func IntSf(msg string, slice []int) string {
	return strings.Trim(fmt.Sprintf(msg, slice), "[]")
}

// DblSf is the Sprintf for a slice of float64 (without brackets)
func DblSf(msg string, slice []float64) string {
	return strings.Trim(fmt.Sprintf(msg, slice), "[]")
}

// StrSf is the Sprintf for a slice of string (without brackets)
func StrSf(msg string, slice []string) string {
	return strings.Trim(fmt.Sprintf(msg, slice), "[]")
}

// Sf wraps Sprintf
func Sf(msg string, prm ...interface{}) string {
	return fmt.Sprintf(msg, prm...)
}

// Ff wraps Fprintf
func Ff(b *bytes.Buffer, msg string, prm ...interface{}) {
	fmt.Fprintf(b, msg, prm...)
}

// Atob converts string to bool
func Atob(val string) (bres bool) {
	if strings.ToLower(val) == "true" {
		return true
	}
	if strings.ToLower(val) == "false" {
		return false
	}
	res, err := strconv.Atoi(val)
	if err != nil {
		chk.Panic("cannot parse string representing integer: %s", val)
	}
	if res != 0 {
		bres = true
	}
	return
}

// Atoi converts string to integer
func Atoi(val string) (res int) {
	res, err := strconv.Atoi(val)
	if err != nil {
		chk.Panic("cannot parse string representing integer number: %s", val)
	}
	return
}

// Atof converts string to float64
func Atof(val string) (res float64) {
	res, err := strconv.ParseFloat(val, 64)
	if err != nil {
		chk.Panic("cannot parse string representing float number: %s", val)
	}
	return
}

// Itob converts from integer to bool
//  Note: only zero returns false
//        anything else returns true
func Itob(val int) bool {
	if val == 0 {
		return false
	}
	return true
}

// Btoi converts flag to interger
//  Note: true  => 1
//        false => 0
func Btoi(flag bool) int {
	if flag {
		return 1
	}
	return 0
}

// Btoa converts flag to string
//  Note: true  => "true"
//        false => "false"
func Btoa(flag bool) string {
	if flag {
		return "true"
	}
	return "false"
}

// PrintFormat commands ---------------------------------------------------------

// Pl prints new line
func Pl() {
	if !Verbose {
		return
	}
	fmt.Println()
}

// low intensity

// Pf prints formatted string
func Pf(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	fmt.Printf(msg, prm...)
}

// Pfcyan prints formatted string in cyan
func Pfcyan(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[0;36m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// Pfcyan2 prints formatted string in another shade of cyan
func Pfcyan2(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[38;5;50m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// Pfyel prints formatted string in yellow
func Pfyel(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[0;33m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// Pfdyel prints formatted string in dark yellow
func Pfdyel(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[38;5;58m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// Pfdyel2 prints formatted string in another shade of dark yellow
func Pfdyel2(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[38;5;94m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// Pfred prints formatted string in red
func Pfred(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[0;31m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// Pfgreen prints formatted string in green
func Pfgreen(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[0;32m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// Pfblue prints formatted string in blue
func Pfblue(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[0;34m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// Pfmag prints formatted string in magenta
func Pfmag(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[0;35m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// Pflmag prints formatted string in light magenta
func Pflmag(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[0;95m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// Pfpink prints formatted string in pink
func Pfpink(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[38;5;205m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// Pfdgreen prints formatted string in dark green
func Pfdgreen(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[38;5;22m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// Pfgreen2 prints formatted string in another shade of green
func Pfgreen2(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[38;5;2m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// Pfpurple prints formatted string in purple
func Pfpurple(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[38;5;55m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// Pfgrey prints formatted string in grey
func Pfgrey(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[38;5;59m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// Pfblue2 prints formatted string in another shade of blue
func Pfblue2(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[38;5;69m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// Pfgrey2 prints formatted string in another shade of grey
func Pfgrey2(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[38;5;60m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// Pforan prints formatted string in orange
func Pforan(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[38;5;202m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// high intensity

// PfCyan prints formatted string in high intensity cyan
func PfCyan(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[1;36m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// PfYel prints formatted string in high intensity yello
func PfYel(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[1;33m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// PfRed prints formatted string in high intensity red
func PfRed(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[1;31m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// PfGreen prints formatted string in high intensity green
func PfGreen(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[1;32m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// PfBlue prints formatted string in high intensity blue
func PfBlue(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[1;34m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// PfMag prints formatted string in high intensity magenta
func PfMag(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[1;35m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// PfWhite prints formatted string in high intensity white
func PfWhite(msg string, prm ...interface{}) {
	if !Verbose {
		return
	}
	if ColorsOn {
		fmt.Printf("[1;37m"+msg+"[0m", prm...)
	} else {
		fmt.Printf(msg, prm...)
	}
}

// remove format ----------------------------------------------------------------

// UnColor removes console characters used to show colors
func UnColor(msg string) string {
	if len(msg) < 7 {
		chk.Panic("string must have at least 7 characters (colored)")
	}
	return strings.Trim(msg, "")[6:]
}
