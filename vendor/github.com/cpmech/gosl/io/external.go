// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"bytes"
	"os/exec"

	"github.com/cpmech/gosl/chk"
)

// RunCmd runs external command
func RunCmd(verbose bool, cmd string, args ...string) string {
	exe := exec.Command(cmd, args...)
	var out, serr bytes.Buffer
	exe.Stdout = &out
	exe.Stderr = &serr
	err := exe.Run()
	if err != nil {
		chk.Panic("%v\n", err)
	}
	if verbose {
		Pf("\n")
		Pf("%s", out.String())
	}
	return out.String()
}

// CopyFileOver copies file (Linux only with cp), overwriting if it exists already
func CopyFileOver(destination, source string) {
	RunCmd(false, "cp", "-rf", source, destination)
}

// Pipeline strings together the given exec.Cmd commands in a similar fashion
// to the Unix pipeline. Each command's standard output is connected to the
// standard input of the next command, and the output of the final command in
// the pipeline is returned, along with the collected standard error of all
// commands and the first error found (if any).
//
// by Kyle Lemons
//
// To provide input to the pipeline, assign an io.Reader to the first's Stdin.
func Pipeline(cmds ...*exec.Cmd) (pipeLineOutput, collectedStandardError []byte) {

	// Require at least one command
	if len(cmds) < 1 {
		return nil, nil
	}

	// Collect the output from the command(s)
	var output bytes.Buffer
	var stderr bytes.Buffer

	last := len(cmds) - 1
	for i, cmd := range cmds[:last] {
		// Connect each command's stdin to the previous command's stdout
		var err error
		if cmds[i+1].Stdin, err = cmd.StdoutPipe(); err != nil {
			chk.Panic("%v\n", err)
		}
		// Connect each command's stderr to a buffer
		cmd.Stderr = &stderr
	}

	// Connect the output and error for the last command
	cmds[last].Stdout, cmds[last].Stderr = &output, &stderr

	// Start each command
	for _, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			chk.Panic("%v\n", err)
		}
	}

	// Wait for each command to complete
	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			chk.Panic("%v\n", err)
		}
	}

	// Return the pipeline output and the collected standard error
	return output.Bytes(), stderr.Bytes()
}
