// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utl

import (
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"

	"github.com/cpmech/gosl/chk"
	"github.com/cpmech/gosl/io"
)

// Prof runs either CPU profiling or MEM profiling
//
//  INPUT:
//    mem    -- memory profiling instead of CPU profiling
//    silent -- hide messages
//
//  OUTPUT:
//    returns a "stop()" function to be called before exiting
//    output files are saved to "/tmp/gosl/cpu.pprof"; or
//                              "/tmp/gosl/mem.pprof"
//  Run analysis with:
//      go tool pprof <binary-goes-here> /tmp/gosl/cpu.pprof
//  or
//      go tool pprof <binary-goes-here> /tmp/gosl/mem.pprof
//
//  Example of use (notice the last parentheses):
//
//       func main() {
//          defer utl.Prof(false, false)()
//          ...
//       }
//
func Prof(mem, silent bool) func() {
	if mem {
		return ProfMEM("/tmp/gosl", "mem.pprof", silent)
	}
	return ProfCPU("/tmp/gosl", "cpu.pprof", silent)
}

// ProfCPU activates CPU profiling
//
//  OUTPUT: returns a "stop()" function to be called before exiting
//
//  Example of use (notice the last parentheses):
//
//       func main() {
//          defer ProfCPU("/tmp", "cpu.pprof", true)()
//          ...
//       }
//
//  Run analysis with:
//      go tool pprof <binary-goes-here> /tmp/cpu.pprof
//
func ProfCPU(dirout, filename string, silent bool) func() {
	os.MkdirAll(dirout, 0777)
	fn := filepath.Join(dirout, filename)
	f, err := os.Create(fn)
	if err != nil {
		chk.Panic("cannot create file:\n%v\n", err)
	}
	if !silent {
		io.Pfcyan("CPU profiling => %s\n", fn)
	}
	pprof.StartCPUProfile(f)
	return func() {
		pprof.StopCPUProfile()
		f.Close()
		if !silent {
			io.Pfcyan("CPU profiling finished\n")
			io.Pfcyan("go tool pprof <binary-goes-here> %s/%s\n", dirout, filename)
		}
	}
}

// ProfMEM activates memory profiling
//
//  OUTPUT: returns a "stop()" function to be called before exiting
//
//  Example of use (notice the last parentheses):
//
//       func main() {
//          defer ProfMEM("/tmp", "mem.pprof", true)()
//          ...
//       }
//
//  Run analysis with:
//      go tool pprof <binary-goes-here> /tmp/mem.pprof
//
func ProfMEM(dirout, filename string, silent bool) func() {
	os.MkdirAll(dirout, 0777)
	fn := filepath.Join(dirout, filename)
	f, err := os.Create(fn)
	if err != nil {
		chk.Panic("cannot create file:\n%v\n", err)
	}
	if !silent {
		io.Pfcyan("MEM profiling => %s\n", fn)
	}
	return func() {
		pprof.WriteHeapProfile(f)
		f.Close()
		if !silent {
			io.Pfcyan("MEM profiling finished\n")
			io.Pfcyan("go tool pprof <binary-goes-here> %s/%s\n", dirout, filename)
		}
	}
}

// PrintMemStat prints memory statistics
func PrintMemStat(msg string) {
	var kbSize uint64 = 1024
	var mbSize uint64 = 1048576
	var gbSize uint64 = 1073741824
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	io.PfYel("%s\n", msg)
	io.Pfyel("Alloc      = %v [KB]  %v [MB]  %v [GB]\n", mem.Alloc/kbSize, mem.Alloc/mbSize, mem.Alloc/gbSize)
	io.Pfyel("HeapAlloc  = %v [KB]  %v [MB]  %v [GB]\n", mem.HeapAlloc/kbSize, mem.HeapAlloc/mbSize, mem.HeapAlloc/gbSize)
	io.Pfyel("Sys        = %v [KB]  %v [MB]  %v [GB]\n", mem.Sys/kbSize, mem.Sys/mbSize, mem.Sys/gbSize)
	io.Pfyel("HeapSys    = %v [KB]  %v [MB]  %v [GB]\n", mem.HeapSys/kbSize, mem.HeapSys/mbSize, mem.HeapSys/gbSize)
	io.Pfyel("TotalAlloc = %v [KB]  %v [MB]  %v [GB]\n", mem.TotalAlloc/kbSize, mem.TotalAlloc/mbSize, mem.TotalAlloc/gbSize)
	io.Pfyel("Mallocs    = %v\n", mem.Mallocs)
	io.Pfyel("Frees      = %v\n", mem.Frees)
}
