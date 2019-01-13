# Gosl. io. Input/output, read/write files, and print commands

[![GoDoc](https://godoc.org/github.com/cpmech/gosl/io?status.svg)](https://godoc.org/github.com/cpmech/gosl/io) 

More information is available in **[the documentation of this package](https://godoc.org/github.com/cpmech/gosl/io).**

This subpackage helps with reading and writing files, printing nice formatted messages (with
colours), and parsing strings.

It has also functions to generate TeX reports.


## Examples

### Read and write files

To write and read a file:

```go
theline := "Hello World !!!"
io.WriteFileSD("/tmp/gosl", "filestring.txt", theline)

f, err := io.OpenFileR("/tmp/gosl/filestring.txt")
if err != nil {
    chk.Panic("%v", err)
}

io.ReadLinesFile(f, func(idx int, line string) (stop bool) {
    Pforan("line = %v\n", line)
    chk.String(tst, line, theline)
    return
})
```

### Read table and generate LaTeX report

To read a table with results separeted by spaces and then generate a LaTeX report:

```go
keys, values := io.ReadTableOrPanic("data/table01.dat")

rpt := io.Report{
    Title:  "Gosl test",
    Author: "Gosl authors",
}

rpt.AddSection("Introduction", 0) // 0 is the section level
rpt.AddTex("In this test, we add one table and one equation to the LaTeX document.")
rpt.AddTex("Then we generate a PDF files in the temporary directory.")
rpt.AddTex("The numbers in the rows of the table have a fancy format.")

rpt.AddSection("MyTable", 1) // 1 is the section level (i.e. subsection)
rpt.AddTable(keys, values, "Results from simulation.", "results", nil, nil)

err := rpt.WriteTexPdf("/tmp/gosl", "test_texpdf02", nil)
if err != nil {
    chk.Panic("%v", err)
}
```
