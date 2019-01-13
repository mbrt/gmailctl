// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"bytes"
	"strings"
	"time"
)

// FcnConvertNum is a function to convert number to string
type FcnConvertNum func(row int, x float64) string

// FcnRow is a function that returns the row value as string
type FcnRow func(row int) string

// Report holds data to generate LaTeX and PDF files
type Report struct {

	// configuration
	Title     string // title of pdf
	Author    string // author of pdf
	Landscape bool   // to format paper

	// default options
	TablePos    string  // default table positioning key; e.g. !t (to be written as [!t])
	NumFmt      string  // default number formatting string; e.g. "%v", "%g" or "%.3f"
	TableFontSz string  // default table fontsize string; e.g. \scriptsize
	TableColSep float64 // default table column separation in 'em'; e.g. 0.5 => \setlength{\tabcolsep}{0.5em}
	NotesFmt    string  // default table notes format; e.g. 'c' or 'p{7cm}'
	NotesFontSz string  // default table notes font size; e.g. \scriptsize
	RowGapStep  int     // step when looping over rows to add an extra gap; e.g. 2 => after every two rows
	RowGapPt    int     // row gap in points; e.g. 12 => 12 pt. 0 means no gap

	// options
	DoNotAlignTable   bool // align coluns in TeX table (has to loop over rows first...)
	DoNotUseGeomPkg   bool // do not use package geometry for margins
	DoNotGeneratePDF  bool // do not generate pdf when writing tex files
	DoNotShowMessages bool // do not show messages

	// internal
	buffers []*bytes.Buffer // buffers created by add commands
	tables  map[string]int  // maps table label to index in buffers
}

// Reset clears report
func (o *Report) Reset() {
	for _, buffer := range o.buffers {
		if buffer != nil {
			buffer.Reset()
		}
	}
}

// AddSection adds section and subsections to report
func (o *Report) AddSection(name string, level int) {
	sec := "section"
	for i := 0; i < level; i++ {
		if i < 2 {
			sec = "sub" + sec
		}
	}
	buffer := new(bytes.Buffer)
	Ff(buffer, "\n")
	Ff(buffer, "\\%s{%s}\n", sec, name)
	o.buffers = append(o.buffers, buffer)
}

// AddTex adds TeX commands
func (o *Report) AddTex(commands string) {
	buffer := new(bytes.Buffer)
	Ff(buffer, "\n%s\n", commands)
	o.buffers = append(o.buffers, buffer)
}

// AddTable adds tex table to report
//   caption -- caption of table
//   label -- label of table
//   keys -- column keys
//   T -- table of values: key => row value
//   key2tex -- maps key to tex formatted text of this key (i.e. equation). may be nil
//   key2convert -- maps key to function to convert numbers to string in that column. may be nil
func (o *Report) AddTable(caption, label, notes string, keys []string, T map[string][]float64, key2tex map[string]string, key2numfmt map[string]FcnConvertNum) {

	// new buffer
	buffer := new(bytes.Buffer)

	// fix default parameters
	o.fixDefaults()

	// find column widths and set formatting string
	strfmt := make([]string, len(keys)) // for each column
	if !o.DoNotAlignTable {
		widths := make([]int, len(keys)) // column widths
		for j, key := range keys {
			if key2tex == nil {
				widths[j] = imax(widths[j], len(key))
			} else {
				widths[j] = imax(widths[j], len(key2tex[key]))
			}
			for i, v := range T[key] {
				if key2numfmt == nil {
					widths[j] = imax(widths[j], len(Sf(o.NumFmt, v)))
				} else {
					widths[j] = imax(widths[j], len(key2numfmt[key](i, v)))
				}
			}
		}
		for j, width := range widths {
			strfmt[j] = "%" + Sf("%d", width) + "s"
		}
	} else {
		for j := 0; j < len(keys); j++ {
			strfmt[j] = "%s"
		}
	}

	// start table and tabular
	ncols := len(keys)
	o.startTableAndTabular(buffer, ncols, caption)

	// header
	for j, key := range keys {
		if j > 0 {
			Ff(buffer, " & ")
		}
		if key2tex == nil {
			Ff(buffer, strfmt[j], key)
		} else {
			Ff(buffer, strfmt[j], key2tex[key])
		}
	}
	Ff(buffer, " \\\\ \\midrule\n")

	// rows
	nrows := len(T[keys[0]])
	for i := 0; i < nrows; i++ {
		if i > 0 {
			Ff(buffer, "\n")
		}
		for j, key := range keys {
			if j > 0 {
				Ff(buffer, " & ")
			}
			if key2numfmt == nil {
				Ff(buffer, strfmt[j], Sf(o.NumFmt, T[key][i]))
			} else {
				Ff(buffer, strfmt[j], key2numfmt[key](i, T[key][i]))
			}
		}
		Ff(buffer, " \\\\")
		if o.RowGapPt > 0 && o.RowGapStep > 0 {
			if (i+1)%o.RowGapStep == 0 && i < nrows-1 {
				Ff(buffer, " \\rule{0pt}{%dpt}", o.RowGapPt)
			}
		}
	}

	// end tabular and table
	o.endTableAndTabular(buffer, ncols, label, notes)

	// map with labels and buffers
	if o.tables == nil {
		o.tables = make(map[string]int)
	}
	o.tables[label] = len(o.buffers)
	o.buffers = append(o.buffers, buffer)
}

// AddTableF adds tex table to report by using a map of functions to extract row values
//   caption -- caption of table
//   label -- label of table
//   keys -- column keys
//   nrows -- number of rows
//   F -- map of functions: key => function returning formatted row value
//   key2tex -- maps key to tex formatted text of this key (i.e. equation). may be nil
func (o *Report) AddTableF(caption, label, notes string, keys []string, nrows int, F map[string]FcnRow, key2tex map[string]string) {

	// new buffer
	buffer := new(bytes.Buffer)

	// fix default parameters
	o.fixDefaults()

	// find column widths and set formatting string
	strfmt := make([]string, len(keys)) // for each column
	if !o.DoNotAlignTable {
		widths := make([]int, len(keys)) // column widths
		for j, key := range keys {
			if key2tex == nil {
				widths[j] = imax(widths[j], len(key))
			} else {
				widths[j] = imax(widths[j], len(key2tex[key]))
			}
			for i := 0; i < nrows; i++ {
				widths[j] = imax(widths[j], len(F[key](i)))
			}
		}
		for j, width := range widths {
			strfmt[j] = "%" + Sf("%d", width) + "s"
		}
	} else {
		for j := 0; j < len(keys); j++ {
			strfmt[j] = "%s"
		}
	}

	// start table and tabular
	ncols := len(keys)
	o.startTableAndTabular(buffer, ncols, caption)

	// header
	for j, key := range keys {
		if j > 0 {
			Ff(buffer, " & ")
		}
		if key2tex == nil {
			Ff(buffer, strfmt[j], key)
		} else {
			Ff(buffer, strfmt[j], key2tex[key])
		}
	}
	Ff(buffer, " \\\\ \\midrule\n")

	// rows
	for i := 0; i < nrows; i++ {
		if i > 0 {
			Ff(buffer, "\n")
		}
		for j, key := range keys {
			if j > 0 {
				Ff(buffer, " & ")
			}
			Ff(buffer, strfmt[j], F[key](i))
		}
		Ff(buffer, " \\\\")
		if o.RowGapPt > 0 && o.RowGapStep > 0 {
			if (i+1)%o.RowGapStep == 0 && i < nrows-1 {
				Ff(buffer, " \\rule{0pt}{%dpt}", o.RowGapPt)
			}
		}
	}

	// end tabular and table
	o.endTableAndTabular(buffer, ncols, label, notes)

	// map with labels and buffers
	if o.tables == nil {
		o.tables = make(map[string]int)
	}
	o.tables[label] = len(o.buffers)
	o.buffers = append(o.buffers, buffer)
}

// WriteTexPdf writes tex file and generates pdf file
//  extra -- extra LaTeX commands; may be nil
func (o *Report) WriteTexPdf(dirout, fnkey string, extra *bytes.Buffer) {

	// header
	pdf := new(bytes.Buffer)
	if o.Landscape {
		Ff(pdf, "\\documentclass[a4paper,landscape]{article}\n")
	} else {
		Ff(pdf, "\\documentclass[a4paper]{article}\n")
	}
	Ff(pdf, "\\usepackage{amsmath}\n")
	Ff(pdf, "\\usepackage{amssymb}\n")
	Ff(pdf, "\\usepackage{booktabs}\n")
	if !o.DoNotUseGeomPkg {
		Ff(pdf, "\\usepackage[margin=1.5cm,footskip=0.5cm]{geometry}\n")
	}

	// title and author
	hasTitleOrAuthor := false
	if o.Title != "" {
		Ff(pdf, "\n")
		Ff(pdf, "\\title{%s}\n", o.Title)
		hasTitleOrAuthor = true
	}
	if o.Author != "" {
		Ff(pdf, "\\author{%s}\n", o.Author)
		hasTitleOrAuthor = true
	}

	// begin document
	Ff(pdf, "\n")
	Ff(pdf, "\\begin{document}\n")
	if hasTitleOrAuthor {
		Ff(pdf, "\\maketitle\n")
	}

	// write buffers
	for _, buffer := range o.buffers {
		if buffer != nil {
			Ff(pdf, "%v\n", buffer)
		}
	}

	// write extra LaTeX commands
	if extra != nil {
		Ff(pdf, "\n%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%% extra commands %%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%\n\n")
		Ff(pdf, "%v\n", extra)
	}

	// end document
	Ff(pdf, "\n")
	Ff(pdf, "\\end{document}\n")

	// write TeX file
	fn := fnkey + ".tex"
	WriteFileD(dirout, fn, pdf)

	// run pdflatex
	if !o.DoNotGeneratePDF {
		RunCmd(false, "pdflatex", "-interaction=batchmode", "-halt-on-error", "-output-directory="+dirout, fn)
		if !o.DoNotShowMessages {
			PfBlue("file <%s/%s.pdf> generated\n", dirout, fnkey)
		}
	}
	return
}

// WriteTexTables writes tables to tex file
func (o *Report) WriteTexTables(dirout string, label2fnkey map[string]string) {
	for label, fnkey := range label2fnkey {
		fn := fnkey + ".tex"
		if idx, ok := o.tables[label]; ok {
			WriteFileD(dirout, fn, o.buffers[idx])
			if !o.DoNotShowMessages {
				PfBlue("file <%s/%s> generated\n", dirout, fn)
			}
		}
	}
	return
}

// TexNum returns a string representation in TeX format of a real number.
// scientificNotation:
//   peforms the conversion of numbers into scientific notation where
//   the exponent notation with e{+-}{##} is converted into \cdot 10^{{+-}##}
func TexNum(fmt string, num float64, scientificNotation bool) (l string) {
	if fmt == "" {
		fmt = "%g"
	}
	l = Sf(fmt, num)
	if scientificNotation {
		s := strings.Split(l, "e-")
		if len(s) == 2 {
			e := s[1]
			if e == "00" {
				l = s[0]
				return
			}
			if e[0] == '0' {
				e = string(e[1])
			}
			l = s[0] + "\\cdot 10^{-" + e + "}"
		}
		s = strings.Split(l, "e+")
		if len(s) == 2 {
			e := s[1]
			if e == "00" {
				l = s[0]
				return
			}
			if e[0] == '0' {
				e = string(e[1])
			}
			l = s[0] + "\\cdot 10^{" + e + "}"
		}
	}
	return
}

// RoundDuration rounds duration
//  Input:
//    d -- duration
//    r -- rounding amount in milliseconds; e.g. 1e6 [ms]
//         NOTE: use -1 for default value [1e6]
func RoundDuration(d, r time.Duration) time.Duration {
	if r <= 0 {
		r = 1e6
	}
	neg := d < 0
	if neg {
		d = -d
	}
	if m := d % r; m+m < r {
		d = d - m
	} else {
		d = d + r - m
	}
	if neg {
		return -d
	}
	return d
}

// auxiliary //////////////////////////////////////////////////////////////////////////////////////

// startTableAndTabular starts table and tabular
func (o *Report) startTableAndTabular(buffer *bytes.Buffer, ncols int, caption string) {

	// start table
	Ff(buffer, "\n")
	Ff(buffer, "\\begin{table*} [%s] \\centering\n", o.TablePos)
	Ff(buffer, "\\caption{%s}\n", caption)

	// set fontsize and column separation
	Ff(buffer, o.TableFontSz)
	Ff(buffer, "\\setlength{\\tabcolsep}{%gem}\n", o.TableColSep)

	// start tabular
	cc := ""
	for i := 0; i < ncols; i++ {
		cc += "c"
	}
	Ff(buffer, "\\begin{tabular}[c]{%s} \\toprule\n", cc)
}

// endTableAndTabular ends table and tabular
func (o *Report) endTableAndTabular(buffer *bytes.Buffer, ncols int, label, notes string) {
	Ff(buffer, "\n")
	if notes != "" {
		Ff(buffer, "\\midrule\n")
		Ff(buffer, "\\multicolumn{%d}{%s}{\n", ncols, o.NotesFmt)
		Ff(buffer, "%s\n", o.NotesFontSz)
		Ff(buffer, "%s\n", notes)
		Ff(buffer, "} \\\\\n")
	}
	Ff(buffer, "\\bottomrule\n")
	Ff(buffer, "\\end{tabular}\n")
	Ff(buffer, "\\label{tab:%s}\n", label)
	Ff(buffer, "\\end{table*}")
}

// fixDefaults fix default values
func (o *Report) fixDefaults() {

	// default table positioning key
	if o.TablePos == "" {
		o.TablePos = "h"
	}

	// default number formatting string
	if o.NumFmt == "" {
		o.NumFmt = "%g"
	}

	// default table fontsize string
	if o.TableFontSz == "" {
		//o.TableFontSz = "\\scriptsize"
	}

	// default table column separation in 'em'; e.g. 0.5 =>
	if o.TableColSep <= 0 {
		o.TableColSep = 0.5
	}

	// default table notes format
	if o.NotesFmt == "" {
		o.NotesFmt = "l"
	}

	// default table notes font size
	if o.NotesFontSz == "" {
		o.NotesFontSz = "\\scriptsize"
	}
}
