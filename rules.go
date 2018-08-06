package main

const (
	PropertyFrom          = "from"
	PropertyHas           = "hasTheWord"
	PropertyMarkImportant = "shouldAlwaysMarkAsImportant"
	PropertyApplyLabel    = "label"
	PropertyApplyCategory = "smartLabelToApply"
	PropertyDelete        = "shouldTrash"
)

type Entry struct {
	Properties []Property
}

type Property struct {
	Name  string
	Value string
}
