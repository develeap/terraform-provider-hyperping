// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package migrate

import (
	"fmt"
	"strings"
)

// SanitizeOpts configures resource name sanitization behavior.
type SanitizeOpts struct {
	// DigitPrefix is prepended (with underscore) when the name starts with a digit.
	// Default: "monitor".
	DigitPrefix string
	// EmptyFallback is used when the sanitized name is empty.
	// Default: "unnamed_monitor".
	EmptyFallback string
}

// SanitizeResourceName converts a human-readable name to a valid Terraform resource name.
// It lowercases, replaces non-alphanumeric characters with underscores, collapses
// consecutive underscores, trims leading/trailing underscores, and prepends a prefix
// if the result starts with a digit.
func SanitizeResourceName(name string) string {
	return SanitizeResourceNameWith(name, SanitizeOpts{
		DigitPrefix:   "monitor",
		EmptyFallback: "unnamed_monitor",
	})
}

// SanitizeResourceNameWith converts a name to a valid Terraform resource name
// using the provided options for digit-leading prefix and empty-name fallback.
func SanitizeResourceNameWith(name string, opts SanitizeOpts) string {
	safe := strings.ToLower(name)

	var result strings.Builder
	for _, ch := range safe {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			result.WriteRune(ch)
		} else {
			result.WriteRune('_')
		}
	}

	s := result.String()

	// Collapse consecutive underscores
	for strings.Contains(s, "__") {
		s = strings.ReplaceAll(s, "__", "_")
	}

	// Trim leading/trailing underscores
	s = strings.Trim(s, "_")

	// Ensure it starts with a letter
	if s != "" && s[0] >= '0' && s[0] <= '9' {
		prefix := opts.DigitPrefix
		if prefix == "" {
			prefix = "monitor"
		}
		s = prefix + "_" + s
	}

	if s == "" {
		s = opts.EmptyFallback
		if s == "" {
			s = "unnamed_monitor"
		}
	}

	return s
}

// DeduplicateResourceName appends a numeric suffix when a name has already been used.
// The seen map tracks how many times each name has appeared.
// First occurrence: returns name as-is. Second: returns "name_2", third: "name_3", etc.
func DeduplicateResourceName(name string, seen map[string]int) string {
	seen[name]++
	if seen[name] == 1 {
		return name
	}
	return fmt.Sprintf("%s_%d", name, seen[name])
}
