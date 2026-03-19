// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package testutil

// Ptr returns a pointer to the given value. Used in tests to create
// inline pointers for optional fields.
func Ptr[T any](v T) *T { return &v }
