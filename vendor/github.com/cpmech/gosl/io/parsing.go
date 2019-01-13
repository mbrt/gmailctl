// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"strings"
	"unicode"

	"github.com/cpmech/gosl/chk"
)

// ExtractStrPair extracts the pair from, e.g., "key:val"
//  Note: it returs empty strings if any is not found
func ExtractStrPair(pair, sep string) (key, val string) {
	res := strings.Split(pair, sep)
	if len(res) > 0 {
		key = strings.TrimSpace(res[0])
	}
	if len(res) > 1 {
		val = strings.TrimSpace(res[1])
	}
	return
}

// Keycode extracts a keycode from a string such as "!typeA:keycodeA !typeB:keycodeB!typeC:keycodeC"
//  Note: String == "!keyA !typeB:valB" is also valid
func Keycode(String string, Type string) (keycode string, found bool) {
	if len(String) < 1 || Type == "" {
		return "", false
	}
	if String[0] != '!' {
		chk.Panic("first character in keycode string must be an exclamation mark !\nstring = \"%v\"", String)
	}
	for _, s := range strings.Split(String[1:], "!") { // [1:] => skip first "!"
		ss := strings.TrimSpace(s)
		sss := strings.Split(ss, ":")
		for i, val := range sss {
			sss[i] = strings.TrimSpace(val)
		}
		if sss[0] == Type {
			if len(sss) > 1 {
				return sss[1], true
			}
			return "", true
		}
	}
	return
}

// Keycodes extracts keys from a keycode (extra) string
//  Example: "!keyA !typeB:valB" => keycodes = [keyA, typeB]
func Keycodes(String string) (keycodes []string) {
	if len(String) < 1 {
		return
	}
	if String[0] != '!' {
		chk.Panic("first character in keycode string must be an exclamation mark !\nstring = \"%v\"", String)
	}
	for _, s := range strings.Split(String[1:], "!") { // [1:] => skip first "!"
		ss := strings.TrimSpace(s)
		sss := strings.Split(ss, ":")
		for i, val := range sss {
			sss[i] = strings.TrimSpace(val)
		}
		//if len(sss) < 1 { // TODO: remove this since this will not happen
		//chk.Panic(_parsing_err4, "Keycodes", String)
		//}
		keycodes = append(keycodes, sss[0])
	}
	return
}

// JoinKeys3 joins keys from 3 slices into a string with sets separated by sep
// and keys separeted by spaces
func JoinKeys3(k0, k1, k2 []string, sep string) (res string) {
	res = strings.Join(k0, " ") + sep
	if len(k1) > 0 {
		res += " " + strings.Join(k1, " ")
	}
	res += sep
	if len(k2) > 0 {
		res += " " + strings.Join(k2, " ")
	}
	return
}

// SplitKeys3 splits a string with three sets of keys separated by comma
func SplitKeys3(res string) (k0, k1, k2 []string) {
	sets := strings.Split(res, ",")
	if len(sets) != 3 {
		chk.Panic("string '%s' does not contain 3 subsets separated by 2 commas. sets = %v", res, sets)
	}
	s0 := strings.TrimSpace(sets[0])
	s1 := strings.TrimSpace(sets[1])
	s2 := strings.TrimSpace(sets[2])
	if len(s0) > 0 {
		k0 = strings.Split(s0, " ")
	}
	if len(s1) > 0 {
		k1 = strings.Split(s1, " ")
	}
	if len(s2) > 0 {
		k2 = strings.Split(s2, " ")
	}
	return
}

// JoinKeys4 joins keys from 4 slices into a string with sets separated by sep
// and keys separeted by spaces
func JoinKeys4(k0, k1, k2, k3 []string, sep string) (res string) {
	res = strings.Join(k0, " ") + sep
	if len(k1) > 0 {
		res += " " + strings.Join(k1, " ")
	}
	res += sep
	if len(k2) > 0 {
		res += " " + strings.Join(k2, " ")
	}
	res += sep
	if len(k3) > 0 {
		res += " " + strings.Join(k3, " ")
	}
	return
}

// SplitKeys4 splits a string with four sets of keys separated by comma
func SplitKeys4(res string) (k0, k1, k2, k3 []string) {
	sets := strings.Split(res, ",")
	if len(sets) != 4 {
		chk.Panic("string '%s' does not contain 4 subsets separated by 3 commas. sets = %v", res, sets)
	}
	s0 := strings.TrimSpace(sets[0])
	s1 := strings.TrimSpace(sets[1])
	s2 := strings.TrimSpace(sets[2])
	s3 := strings.TrimSpace(sets[3])
	if len(s0) > 0 {
		k0 = strings.Split(s0, " ")
	}
	if len(s1) > 0 {
		k1 = strings.Split(s1, " ")
	}
	if len(s2) > 0 {
		k2 = strings.Split(s2, " ")
	}
	if len(s3) > 0 {
		k3 = strings.Split(s3, " ")
	}
	return
}

// JoinKeys join keys separeted by spaces
func JoinKeys(keys []string) string {
	return strings.Join(keys, " ")
}

// JoinKeysPre join keys separeted by spaces with a prefix
func JoinKeysPre(prefix string, keys []string) (res string) {
	for i, val := range keys {
		res += prefix + val
		if i < len(keys)-1 {
			res += " "
		}
	}
	return
}

// SplitKeys splits keys separeted by spaces
func SplitKeys(keys string) []string {
	return strings.Split(keys, " ")
}

// SplitSpacesQuoted splits string with quoted substrings. e.g. "  a,b, 'c', \"d\"  "
func SplitSpacesQuoted(str string) (res []string) {
	lastQuote := rune(0)
	f := func(c rune) bool {
		switch {
		case c == lastQuote:
			lastQuote = rune(0)
			return false
		case lastQuote != rune(0):
			return false
		case unicode.In(c, unicode.Quotation_Mark):
			lastQuote = c
			return false
		default:
			return unicode.IsSpace(c)

		}
	}
	return strings.FieldsFunc(str, f)
}

// SplitWithinParentheses extracts arguments (substrings) within brackets
// e.g.: "(arg1, (arg2.1, arg2.2),  arg3, arg4, (arg5.1,arg5.2,  arg5.3 ) )"
func SplitWithinParentheses(s string) (res []string) {
	trim := func(l, pfix, sfix string) string {
		l = strings.TrimSpace(l)
		l = strings.TrimPrefix(l, pfix)
		l = strings.TrimSuffix(l, sfix)
		return l
	}
	s = trim(s, "(", ")")
	s = strings.Replace(s, "(", "'", -1)
	s = strings.Replace(s, ")", "'", -1)
	s = strings.Replace(s, ",", " ", -1)
	res = SplitSpacesQuoted(s)
	for i := 0; i < len(res); i++ {
		res[i] = trim(res[i], "'", "'")
	}
	return
}

// SplitInts splits space-separated integers
func SplitInts(str string) (res []int) {
	vals := strings.Fields(str)
	res = make([]int, len(vals))
	for i, v := range vals {
		res[i] = Atoi(v)
	}
	return
}

// SplitFloats splits space-separated float numbers
func SplitFloats(str string) (res []float64) {
	vals := strings.Fields(str)
	res = make([]float64, len(vals))
	for i, v := range vals {
		res[i] = Atof(v)
	}
	return
}
