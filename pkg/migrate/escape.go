// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package migrate

import "strings"

// EscapeHCL escapes a string for safe inclusion in Terraform HCL.
//
// In addition to backslashes, quotes, and control characters, this function
// neutralizes HCL template-interpolation sequences (${...} and %{...}) by
// doubling their leading sigil. This is critical when emitting untrusted
// strings (monitor names, URLs, header values, etc.) into a generated .tf
// file: without it, a value such as `${file("/etc/passwd")}` would be
// evaluated by Terraform at plan time, enabling local-file exfiltration or
// arbitrary HCL function execution against the operator's machine.
//
// The template-escape replacements MUST run after the backslash/quote
// escaping (so we never accidentally double a sigil we just produced) and
// before the result is wrapped in quotes by QuoteHCL.
func EscapeHCL(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	// Neutralize HCL template-interpolation sequences.
	s = strings.ReplaceAll(s, "${", "$${")
	s = strings.ReplaceAll(s, "%{", "%%{")
	return s
}

// EscapeShell escapes a string for safe inclusion in a bash script.
// It handles backslashes, double quotes, dollar signs, backticks, and
// embedded newlines/carriage returns (which would otherwise break out of
// a single shell argument).
func EscapeShell(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "$", `\$`)
	s = strings.ReplaceAll(s, "`", "\\`")
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	return s
}

// QuoteHCL wraps a string in double quotes after escaping it for HCL.
// Prefer this over fmt's %q verb when emitting untrusted data into
// generated Terraform configuration: %q does not escape HCL template
// sequences (${...} and %{...}).
func QuoteHCL(s string) string {
	return `"` + EscapeHCL(s) + `"`
}
