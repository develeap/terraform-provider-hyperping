// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	hyperping "github.com/develeap/hyperping-go"
)

// TFLogAdapter adapts the Terraform plugin logging framework to the hyperping.Logger interface.
type TFLogAdapter struct{}

// NewTFLogAdapter creates a new TFLogAdapter.
func NewTFLogAdapter() *TFLogAdapter {
	return &TFLogAdapter{}
}

// Debug logs a debug-level message using tflog.
func (l *TFLogAdapter) Debug(ctx context.Context, msg string, fields map[string]interface{}) {
	tflog.Debug(ctx, msg, fields)
}

// Ensure TFLogAdapter implements the Logger interface.
var _ hyperping.Logger = (*TFLogAdapter)(nil)
