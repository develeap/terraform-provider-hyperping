// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	hyperping "github.com/develeap/hyperping-go"
)

// sensitiveLogFieldKeys are structured-log field names whose values are
// always redacted regardless of content.
var sensitiveLogFieldKeys = []string{
	"api_key",
	"apikey",
	"Authorization",
	"authorization",
	"token",
	"access_token",
	"X-Api-Key",
	"x-api-key",
}

// TFLogAdapter adapts the Terraform plugin logging framework to the
// hyperping.Logger interface.
//
// Every Debug call applies tflog masking before emitting:
//   - any field whose key matches a known sensitive header/parameter name is
//     replaced with [REDACTED];
//   - any value matching hyperping.APIKeyPattern (sk_...) is replaced with
//     [REDACTED] regardless of field name, catching cases where a secret is
//     logged under an unexpected key.
//
// This per-call masking is the runtime guarantee: a derived context built in
// provider.Configure does not survive into the per-operation contexts that
// the Terraform framework creates, so masking must be applied here.
type TFLogAdapter struct{}

// NewTFLogAdapter creates a new TFLogAdapter.
func NewTFLogAdapter() *TFLogAdapter {
	return &TFLogAdapter{}
}

// Debug logs a debug-level message using tflog, redacting sensitive fields.
func (l *TFLogAdapter) Debug(ctx context.Context, msg string, fields map[string]interface{}) {
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, sensitiveLogFieldKeys...)
	ctx = tflog.MaskAllFieldValuesRegexes(ctx, hyperping.APIKeyPattern)
	tflog.Debug(ctx, msg, fields)
}

// Ensure TFLogAdapter implements the Logger interface.
var _ hyperping.Logger = (*TFLogAdapter)(nil)
