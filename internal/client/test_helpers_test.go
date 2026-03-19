// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import "github.com/develeap/terraform-provider-hyperping/internal/provider/testutil"

// Package-level test pointer helpers. These delegate to testutil.Ptr
// to avoid duplicating the implementation across test files.

func boolPtr(b bool) *bool       { return testutil.Ptr(b) }
func strPtr(s string) *string    { return testutil.Ptr(s) }
func stringPtr(s string) *string { return testutil.Ptr(s) }
func intPtr(i int) *int          { return testutil.Ptr(i) }
