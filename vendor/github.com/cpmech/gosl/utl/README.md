# Gosl. utl. Utilities. Lists. Dictionaries. Simple Numerics

[![GoDoc](https://godoc.org/github.com/cpmech/gosl/utl?status.svg)](https://godoc.org/github.com/cpmech/gosl/utl) 

More information is available in **[the documentation of this package](https://godoc.org/github.com/cpmech/gosl/utl).**

This package implements functions for simplifying numeric calculations such as finding the maximum
and minimum of lists (i.e. slices), allocation of _deep_ structures such as slices of slices, and
generation of _arrays_. It also contains functions for sorting quantities and updating dictionaries
(i.e. maps).

This package does not aim for high performance linear algebra computations. For that purpose, we
have the `la` package. Nonetheless, `utl` package is OK for _small computations_ such as for vectors
in the 3D space. It also tries to use the best algorithms for sorting that are implemented in the
standard Go library and others.

Example of what the functions here can do:
* Generate lists of integers
* Generate lists of float64s
* Cumulative sums
* Handling tables of float64s
* Pareto fronts
* Slices and deep (nested) slices up to the 4th depth
* Allocate deep slices
* Serialization of deep slices
* Sorting
* Maps and _dictionaries_
* Append to maps of slices of float64
* Find the _best square_ for given `size = numberOfRows * numberOfColumns`
* ...
