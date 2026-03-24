// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"strings"
)

// formatValidValues formats a slice of string values for use in MarkdownDescription.
// Each value is wrapped in backticks and joined with ", ".
// Example: ["a", "b", "c"] -> "`a`, `b`, `c`"
func formatValidValues(values []string) string {
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = "`" + v + "`"
	}
	return strings.Join(quoted, ", ")
}

// formatValidInts formats a slice of int values for use in MarkdownDescription.
// Each value is wrapped in backticks and joined with ", ".
// Example: [10, 20, 30] -> "`10`, `20`, `30`"
func formatValidInts(values []int) string {
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = fmt.Sprintf("`%d`", v)
	}
	return strings.Join(quoted, ", ")
}

// toInt64Slice converts a []int to []int64 for use with int64validator.OneOf.
func toInt64Slice(values []int) []int64 {
	result := make([]int64, len(values))
	for i, v := range values {
		result[i] = int64(v)
	}
	return result
}
