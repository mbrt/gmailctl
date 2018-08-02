package main

const (
	From      = "from"
	Important = "shouldAlwaysMarkAsImportant"
	Label     = "label"
	Category  = "smartLabelToApply"
)

type Entry struct {
	Properties []Property
}

type Property struct {
	Name  string
	Value string
}
