// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

// boolPtr returns a pointer to the given bool value.
// Lives in _test.go to exclude from production binary.
func boolPtr(b bool) *bool {
	return &b
}
