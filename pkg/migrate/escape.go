// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package migrate

import "strings"

// EscapeHCL escapes a string for safe inclusion in Terraform HCL.
// It handles backslashes, double quotes, and newlines.
func EscapeHCL(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}

// EscapeShell escapes a string for safe inclusion in a bash script.
// It handles backslashes, double quotes, dollar signs, and backticks.
func EscapeShell(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "$", `\$`)
	s = strings.ReplaceAll(s, "`", "\\`")
	return s
}

// QuoteHCL wraps a string in double quotes after escaping it for HCL.
func QuoteHCL(s string) string {
	return `"` + EscapeHCL(s) + `"`
}
