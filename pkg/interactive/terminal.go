// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package interactive

import (
	"io"
	"os"

	"github.com/mattn/go-isatty"
)

// IsInteractive checks if the terminal is interactive (TTY).
func IsInteractive() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}

// IsTTY checks if the writer is a TTY.
func IsTTY(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
	}
	return false
}

// SupportsANSI checks if the terminal supports ANSI codes.
func SupportsANSI() bool {
	// Check NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check if stdout is a terminal
	return IsTTY(os.Stdout)
}
