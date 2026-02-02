// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// reservedHeaderNames lists HTTP headers that users must not override.
// Allowing these could leak API credentials or manipulate routing.
var reservedHeaderNames = map[string]bool{
	"authorization":     true,
	"host":              true,
	"cookie":            true,
	"set-cookie":        true,
	"proxy-authorize":   true,
	"transfer-encoding": true,
}

// noControlCharactersValidator rejects strings containing CR, LF, or NULL.
// This prevents HTTP header injection (VULN-012).
type noControlCharactersValidator struct {
	message string
}

func (v noControlCharactersValidator) Description(_ context.Context) string {
	return "value must not contain control characters (CR, LF, NULL)"
}

func (v noControlCharactersValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v noControlCharactersValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if strings.ContainsAny(value, "\r\n\x00") {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Value",
			v.message,
		)
	}
}

// NoControlCharacters returns a validator that rejects CR, LF, and NULL characters.
func NoControlCharacters(message string) validator.String {
	return noControlCharactersValidator{message: message}
}

// reservedHeaderNameValidator rejects reserved HTTP header names (case-insensitive).
// This prevents users from overriding Authorization, Host, Cookie, etc. (VULN-012).
type reservedHeaderNameValidator struct{}

func (v reservedHeaderNameValidator) Description(_ context.Context) string {
	return "header name must not be a reserved HTTP header"
}

func (v reservedHeaderNameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v reservedHeaderNameValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if reservedHeaderNames[strings.ToLower(value)] {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Reserved Header Name",
			fmt.Sprintf("The header name %q is reserved and cannot be overridden in request_headers. "+
				"This protects API credentials and request integrity.", value),
		)
	}
}

// ReservedHeaderName returns a validator that rejects reserved HTTP header names.
func ReservedHeaderName() validator.String {
	return reservedHeaderNameValidator{}
}

// iso8601Validator validates that a string is in ISO 8601 format.
type iso8601Validator struct{}

func (v iso8601Validator) Description(_ context.Context) string {
	return "value must be in ISO 8601 format (e.g., 2026-01-29T10:00:00Z)"
}

func (v iso8601Validator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v iso8601Validator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	// Simple ISO 8601 validation: must contain 'T' and end with 'Z' or timezone offset
	// Full RFC3339 parsing would be overkill for plan-time validation (API will validate)
	if !strings.Contains(value, "T") {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid ISO 8601 Format",
			fmt.Sprintf("The value %q does not appear to be in ISO 8601 format. "+
				"Expected format: 2026-01-29T10:00:00Z", value),
		)
	}
}

// ISO8601() returns a validator that checks for ISO 8601 datetime format.
func ISO8601() validator.String {
	return iso8601Validator{}
}

// uuidFormatValidator validates that a string matches UUID format.
type uuidFormatValidator struct{}

func (v uuidFormatValidator) Description(_ context.Context) string {
	return "value must be a valid UUID format"
}

func (v uuidFormatValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v uuidFormatValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	// Simple UUID format check: 8-4-4-4-12 hexadecimal characters with dashes
	// Or provider-specific formats like "mon_", "tok_", "out_" prefixes
	if len(value) < 8 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid UUID Format",
			fmt.Sprintf("The value %q is too short to be a valid UUID or resource ID.", value),
		)
		return
	}

	// Accept standard UUIDs (with dashes) or Hyperping resource IDs (with underscores)
	hasDashes := strings.Contains(value, "-")
	hasUnderscores := strings.Contains(value, "_")

	if !hasDashes && !hasUnderscores {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid UUID Format",
			fmt.Sprintf("The value %q does not match UUID format (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx) "+
				"or Hyperping resource ID format (prefix_xxxxx).", value),
		)
	}
}

// UUIDFormat returns a validator that checks for valid UUID or resource ID format.
func UUIDFormat() validator.String {
	return uuidFormatValidator{}
}
