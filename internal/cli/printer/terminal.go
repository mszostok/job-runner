/*
Copyright 2019 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package printer is copied and simplified code from: https://github.com/kubernetes-sigs/kind/blob/b6bc112522651d98c81823df56b7afa511459a3b/pkg/internal/env/term.go#L51
// It is shared with the same license.
package printer

import (
	"io"
	"os"
	"runtime"

	"github.com/mattn/go-isatty"
)

// IsSmartTerminal returns true if the writer w is a terminal AND
// we think that the terminal is smart enough to use VT escape codes etc.
func IsSmartTerminal(w io.Writer) bool {
	return isSmartTerminal(w, runtime.GOOS, os.LookupEnv)
}

func isSmartTerminal(w io.Writer, GOOS string, lookupEnv func(string) (string, bool)) bool {
	// Not smart if it's not a tty
	f, ok := (w).(*os.File)
	if !ok {
		return false
	}
	if !isatty.IsTerminal(f.Fd()) {
		return false
	}

	// getenv helper for when we only care about the value
	getenv := func(e string) string {
		v, _ := lookupEnv(e)
		return v
	}

	// Explicitly dumb terminals are not smart
	// https://en.wikipedia.org/wiki/Computer_terminal#Dumb_terminals
	term := getenv("TERM")
	if term == "dumb" {
		return false
	}
	// st has some bug
	// https://github.com/kubernetes-sigs/kind/issues/1892
	if term == "st-256color" {
		return false
	}

	// On Windows WT_SESSION is set by the modern terminal component.
	// Older terminals have poor support for UTF-8, VT escape codes, etc.
	if GOOS == "windows" && getenv("WT_SESSION") == "" {
		return false
	}

	/* CI Systems with bad Fake TTYs */
	// Travis CI
	// https://github.com/kubernetes-sigs/kind/issues/1478
	// We can detect it with documented magical environment variables
	// https://docs.travis-ci.com/user/environment-variables/#default-environment-variables
	if getenv("HAS_JOSH_K_SEAL_OF_APPROVAL") == "true" && getenv("TRAVIS") == "true" {
		return false
	}

	// OK, we'll assume it's smart now, given no evidence otherwise.
	return true
}
