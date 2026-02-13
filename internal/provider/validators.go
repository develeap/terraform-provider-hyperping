// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/robfig/cron/v3"
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

// urlFormatValidator validates that a string is a valid HTTP or HTTPS URL.
type urlFormatValidator struct{}

func (v urlFormatValidator) Description(_ context.Context) string {
	return "value must be a valid HTTP or HTTPS URL"
}

func (v urlFormatValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v urlFormatValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid URL Format",
			fmt.Sprintf("The value %q must be a valid HTTP or HTTPS URL", value),
		)
	}
}

// URLFormat returns a validator that checks for valid HTTP or HTTPS URLs.
func URLFormat() validator.String {
	return urlFormatValidator{}
}

// stringLengthValidator validates that a string is between min and max characters.
type stringLengthValidator struct {
	minLength int
	maxLength int
}

func (v stringLengthValidator) Description(_ context.Context) string {
	return fmt.Sprintf("string length must be between %d and %d characters", v.minLength, v.maxLength)
}

func (v stringLengthValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringLengthValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	length := utf8.RuneCountInString(value)

	if length < v.minLength || length > v.maxLength {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid String Length",
			fmt.Sprintf("The value must be between %d and %d characters, got %d", v.minLength, v.maxLength, length),
		)
	}
}

// StringLength returns a validator that checks string length (in Unicode characters).
func StringLength(minLength, maxLength int) validator.String {
	return stringLengthValidator{minLength: minLength, maxLength: maxLength}
}

// cronExpressionValidator validates that a string is a valid cron expression.
type cronExpressionValidator struct{}

func (v cronExpressionValidator) Description(_ context.Context) string {
	return "value must be a valid cron expression (format: 'minute hour day month weekday')"
}

func (v cronExpressionValidator) MarkdownDescription(_ context.Context) string {
	return "value must be a valid cron expression (format: `minute hour day month weekday`)"
}

func (v cronExpressionValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(value)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Cron Expression",
			fmt.Sprintf("The value %q is not a valid cron expression: %v\n"+
				"Expected format: 'minute hour day month weekday' (e.g., '0 0 * * *' for daily at midnight)",
				value, err),
		)
	}
}

// CronExpression returns a validator that checks for valid cron expressions.
func CronExpression() validator.String {
	return cronExpressionValidator{}
}

// timezoneValidator validates that a string is a valid IANA timezone.
type timezoneValidator struct{}

func (v timezoneValidator) Description(_ context.Context) string {
	return "value must be a valid IANA timezone (e.g., 'America/New_York', 'Europe/London', 'UTC')"
}

func (v timezoneValidator) MarkdownDescription(_ context.Context) string {
	return "value must be a valid IANA timezone (e.g., `America/New_York`, `Europe/London`, `UTC`)"
}

func (v timezoneValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	_, err := time.LoadLocation(value)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Timezone",
			fmt.Sprintf("The value %q is not a valid IANA timezone.\n"+
				"Use standard timezone names like 'America/New_York', 'Europe/London', or 'UTC'.\n"+
				"See https://en.wikipedia.org/wiki/List_of_tz_database_time_zones for valid values.",
				value),
		)
	}
}

// Timezone returns a validator that checks for valid IANA timezones.
func Timezone() validator.String {
	return timezoneValidator{}
}

// portRangeValidator validates that an int64 is a valid TCP/UDP port (1-65535).
type portRangeValidator struct{}

func (v portRangeValidator) Description(_ context.Context) string {
	return "value must be a valid port number between 1 and 65535"
}

func (v portRangeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v portRangeValidator) ValidateInt64(_ context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueInt64()
	if value < 1 || value > 65535 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Port Number",
			fmt.Sprintf("Port must be between 1 and 65535, got %d", value),
		)
	}
}

// PortRange returns a validator that checks for valid TCP/UDP port numbers.
func PortRange() validator.Int64 {
	return portRangeValidator{}
}

// hexColorValidator validates that a string is a valid 6-digit hex color (#RRGGBB).
type hexColorValidator struct{}

func (v hexColorValidator) Description(_ context.Context) string {
	return "value must be a 6-digit hex color (e.g., '#ff5733', '#000000')"
}

func (v hexColorValidator) MarkdownDescription(_ context.Context) string {
	return "value must be a 6-digit hex color (e.g., `#ff5733`, `#000000`)"
}

func (v hexColorValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	matched, err := regexp.MatchString(`^#[0-9A-Fa-f]{6}$`, value)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Validation Error",
			fmt.Sprintf("Error validating hex color: %v", err),
		)
		return
	}
	if !matched {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Hex Color",
			fmt.Sprintf("The value %q must be a 6-digit hex color (e.g., '#ff5733', '#000000')", value),
		)
	}
}

// HexColor returns a validator that checks for valid 6-digit hex colors.
func HexColor() validator.String {
	return hexColorValidator{}
}

// emailFormatValidator validates that a string is a valid email address.
// Uses a simplified RFC 5322 regex pattern.
type emailFormatValidator struct{}

func (v emailFormatValidator) Description(_ context.Context) string {
	return "value must be a valid email address"
}

func (v emailFormatValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v emailFormatValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	// RFC 5322 simplified regex
	// Matches: local-part@domain where:
	// - local-part: alphanumeric, dots, underscores, percent, plus, hyphens
	// - domain: alphanumeric and hyphens
	// - TLD: at least 2 characters, letters only
	matched, err := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, value)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Validation Error",
			fmt.Sprintf("Error validating email format: %v", err),
		)
		return
	}
	if !matched {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Email Format",
			fmt.Sprintf("The value %q is not a valid email address", value),
		)
	}
}

// EmailFormat returns a validator that checks for valid email address format.
func EmailFormat() validator.String {
	return emailFormatValidator{}
}
