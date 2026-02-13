// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNoControlCharacters(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"clean header name", "X-Custom-Header", false},
		{"clean header value", "Bearer token123", false},
		{"LF injection", "X-Injected\nEvil: payload", true},
		{"CR injection", "X-Injected\rEvil: payload", true},
		{"CRLF injection", "X-Injected\r\nEvil: payload", true},
		{"NULL byte", "X-Null\x00Byte", true},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NoControlCharacters("test message")
			req := validator.StringRequest{
				Path:        path.Root("test"),
				ConfigValue: types.StringValue(tt.input),
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tt.wantError {
				t.Errorf("NoControlCharacters(%q): got error=%v, want error=%v", tt.input, hasError, tt.wantError)
			}
		})
	}
}

func TestReservedHeaderName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"Authorization blocked", "Authorization", true},
		{"authorization lowercase blocked", "authorization", true},
		{"AUTHORIZATION uppercase blocked", "AUTHORIZATION", true},
		{"Host blocked", "Host", true},
		{"Cookie blocked", "Cookie", true},
		{"Set-Cookie blocked", "Set-Cookie", true},
		{"Proxy-Authorize blocked", "Proxy-Authorize", true},
		{"Transfer-Encoding blocked", "Transfer-Encoding", true},
		{"custom header allowed", "X-Custom", false},
		{"Accept allowed", "Accept", false},
		{"Content-Type allowed", "Content-Type", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := ReservedHeaderName()
			req := validator.StringRequest{
				Path:        path.Root("name"),
				ConfigValue: types.StringValue(tt.input),
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tt.wantError {
				t.Errorf("ReservedHeaderName(%q): got error=%v, want error=%v", tt.input, hasError, tt.wantError)
			}
		})
	}
}

func TestISO8601(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid with Z", "2026-01-29T10:00:00Z", false},
		{"valid with timezone", "2026-01-29T10:00:00+05:00", false},
		{"valid end of year", "2026-12-31T23:59:59Z", false},
		{"invalid - missing time", "2026-01-29", true},
		{"invalid - space separator", "2026-01-29 10:00:00", true},
		{"invalid - not a date", "invalid", true},
		{"invalid - empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := ISO8601()
			req := validator.StringRequest{
				Path:        path.Root("test"),
				ConfigValue: types.StringValue(tt.input),
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tt.wantError {
				t.Errorf("ISO8601(%q): got error=%v, want error=%v", tt.input, hasError, tt.wantError)
			}
		})
	}
}

func TestISO8601_Null(t *testing.T) {
	v := ISO8601()
	req := validator.StringRequest{
		Path:        path.Root("test"),
		ConfigValue: types.StringNull(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("ISO8601(null): unexpected error for null value: %v", resp.Diagnostics.Errors())
	}
}

func TestUUIDFormat(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"standard UUID", "550e8400-e29b-41d4-a716-446655440000", false},
		{"monitor ID", "mon_abc123def456", false},
		{"token ID", "tok_xyz789abc", false},
		{"outage ID", "out_incident123", false},
		{"incident ID", "inc_test123", false},
		{"invalid - too short", "short", true},
		{"invalid - no separators", "nodashesorunderscores", true},
		{"invalid - empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := UUIDFormat()
			req := validator.StringRequest{
				Path:        path.Root("test"),
				ConfigValue: types.StringValue(tt.input),
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tt.wantError {
				t.Errorf("UUIDFormat(%q): got error=%v, want error=%v", tt.input, hasError, tt.wantError)
			}
		})
	}
}

func TestUUIDFormat_Null(t *testing.T) {
	v := UUIDFormat()
	req := validator.StringRequest{
		Path:        path.Root("test"),
		ConfigValue: types.StringNull(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("UUIDFormat(null): unexpected error for null value: %v", resp.Diagnostics.Errors())
	}
}

// Description method tests for coverage
func TestNoControlCharacters_Description(t *testing.T) {
	v := NoControlCharacters("test message")
	ctx := context.Background()

	desc := v.Description(ctx)
	if desc == "" {
		t.Error("expected non-empty description")
	}

	mdDesc := v.MarkdownDescription(ctx)
	if mdDesc == "" {
		t.Error("expected non-empty markdown description")
	}

	if desc != mdDesc {
		t.Error("description and markdown description should match")
	}
}

func TestReservedHeaderName_Description(t *testing.T) {
	v := ReservedHeaderName()
	ctx := context.Background()

	desc := v.Description(ctx)
	if desc == "" {
		t.Error("expected non-empty description")
	}

	mdDesc := v.MarkdownDescription(ctx)
	if mdDesc == "" {
		t.Error("expected non-empty markdown description")
	}

	if desc != mdDesc {
		t.Error("description and markdown description should match")
	}
}

func TestISO8601_Description(t *testing.T) {
	v := ISO8601()
	ctx := context.Background()

	desc := v.Description(ctx)
	if desc == "" {
		t.Error("expected non-empty description")
	}

	mdDesc := v.MarkdownDescription(ctx)
	if mdDesc == "" {
		t.Error("expected non-empty markdown description")
	}

	if desc != mdDesc {
		t.Error("description and markdown description should match")
	}
}

func TestUUIDFormat_Description(t *testing.T) {
	v := UUIDFormat()
	ctx := context.Background()

	desc := v.Description(ctx)
	if desc == "" {
		t.Error("expected non-empty description")
	}

	mdDesc := v.MarkdownDescription(ctx)
	if mdDesc == "" {
		t.Error("expected non-empty markdown description")
	}

	if desc != mdDesc {
		t.Error("description and markdown description should match")
	}
}

// Unknown value tests
func TestNoControlCharacters_Unknown(t *testing.T) {
	v := NoControlCharacters("test message")
	req := validator.StringRequest{
		Path:        path.Root("test"),
		ConfigValue: types.StringUnknown(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("NoControlCharacters(unknown): unexpected error: %v", resp.Diagnostics.Errors())
	}
}

func TestReservedHeaderName_NullUnknown(t *testing.T) {
	tests := []struct {
		name  string
		value types.String
	}{
		{"null", types.StringNull()},
		{"unknown", types.StringUnknown()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := ReservedHeaderName()
			req := validator.StringRequest{
				Path:        path.Root("test"),
				ConfigValue: tt.value,
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			if resp.Diagnostics.HasError() {
				t.Errorf("ReservedHeaderName(%s): unexpected error: %v", tt.name, resp.Diagnostics.Errors())
			}
		})
	}
}

func TestISO8601_Unknown(t *testing.T) {
	v := ISO8601()
	req := validator.StringRequest{
		Path:        path.Root("test"),
		ConfigValue: types.StringUnknown(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("ISO8601(unknown): unexpected error: %v", resp.Diagnostics.Errors())
	}
}

func TestUUIDFormat_Unknown(t *testing.T) {
	v := UUIDFormat()
	req := validator.StringRequest{
		Path:        path.Root("test"),
		ConfigValue: types.StringUnknown(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("UUIDFormat(unknown): unexpected error: %v", resp.Diagnostics.Errors())
	}
}

func TestURLFormatValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"valid https", "https://example.com", false},
		{"valid http", "http://api.example.com/health", false},
		{"valid with path", "https://example.com/api/v1/health", false},
		{"valid with port", "https://example.com:8080", false},
		{"valid with query", "https://example.com/api?key=value", false},
		{"valid with fragment", "https://example.com/page#section", false},
		{"invalid ftp", "ftp://example.com", true},
		{"invalid no scheme", "example.com", true},
		{"invalid typo", "htp://example.com", true},
		{"invalid empty", "", true},
		{"invalid just scheme", "https://", true},
		{"invalid file scheme", "file:///path/to/file", true},
		{"invalid javascript", "javascript:alert(1)", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := URLFormat()
			req := validator.StringRequest{
				Path:        path.Root("url"),
				ConfigValue: types.StringValue(tt.value),
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tt.wantError {
				t.Errorf("URLFormat(%q): got error=%v, want error=%v", tt.value, hasError, tt.wantError)
			}
		})
	}
}

func TestURLFormat_Null(t *testing.T) {
	v := URLFormat()
	req := validator.StringRequest{
		Path:        path.Root("test"),
		ConfigValue: types.StringNull(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("URLFormat(null): unexpected error for null value: %v", resp.Diagnostics.Errors())
	}
}

func TestURLFormat_Unknown(t *testing.T) {
	v := URLFormat()
	req := validator.StringRequest{
		Path:        path.Root("test"),
		ConfigValue: types.StringUnknown(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("URLFormat(unknown): unexpected error: %v", resp.Diagnostics.Errors())
	}
}

func TestURLFormat_Description(t *testing.T) {
	v := URLFormat()
	ctx := context.Background()

	desc := v.Description(ctx)
	if desc == "" {
		t.Error("expected non-empty description")
	}

	mdDesc := v.MarkdownDescription(ctx)
	if mdDesc == "" {
		t.Error("expected non-empty markdown description")
	}

	if desc != mdDesc {
		t.Error("description and markdown description should match")
	}
}

func TestStringLengthValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		min       int
		max       int
		value     string
		wantError bool
	}{
		{"valid in range", 1, 255, "Valid Name", false},
		{"valid at min", 1, 255, "X", false},
		{"valid at max", 1, 255, string(make([]byte, 255)), false},
		{"invalid too short", 1, 255, "", true},
		{"invalid too long", 1, 255, string(make([]byte, 256)), true},
		{"valid exact length", 5, 5, "hello", false},
		{"invalid one under min", 5, 10, "test", true},
		{"invalid one over max", 5, 10, "hello world", true},
		{"valid unicode single char", 1, 10, "😀", false},
		{"valid unicode multi char", 1, 10, "Hello 世界", false},
		{"invalid unicode too long", 1, 5, "Hello 世界", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := StringLength(tt.min, tt.max)
			req := validator.StringRequest{
				Path:        path.Root("name"),
				ConfigValue: types.StringValue(tt.value),
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tt.wantError {
				t.Errorf("StringLength(%d,%d)(%q): got error=%v, want error=%v",
					tt.min, tt.max, tt.value, hasError, tt.wantError)
			}
		})
	}
}

func TestStringLength_Null(t *testing.T) {
	v := StringLength(1, 255)
	req := validator.StringRequest{
		Path:        path.Root("test"),
		ConfigValue: types.StringNull(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("StringLength(null): unexpected error for null value: %v", resp.Diagnostics.Errors())
	}
}

func TestStringLength_Unknown(t *testing.T) {
	v := StringLength(1, 255)
	req := validator.StringRequest{
		Path:        path.Root("test"),
		ConfigValue: types.StringUnknown(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("StringLength(unknown): unexpected error: %v", resp.Diagnostics.Errors())
	}
}

func TestStringLength_Description(t *testing.T) {
	v := StringLength(1, 255)
	ctx := context.Background()

	desc := v.Description(ctx)
	if desc == "" {
		t.Error("expected non-empty description")
	}

	mdDesc := v.MarkdownDescription(ctx)
	if mdDesc == "" {
		t.Error("expected non-empty markdown description")
	}

	if desc != mdDesc {
		t.Error("description and markdown description should match")
	}
}

func TestCronExpressionValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"valid daily midnight", "0 0 * * *", false},
		{"valid every 15 min", "*/15 * * * *", false},
		{"valid business hours", "0 9-17 * * 1-5", false},
		{"valid specific time", "30 14 * * *", false},
		{"valid every hour", "0 * * * *", false},
		{"valid complex", "0,30 9-17 * * 1-5", false},
		{"invalid text", "invalid", true},
		{"invalid minute 60", "60 * * * *", true},
		{"invalid day 32", "* * 32 * *", true},
		{"invalid empty", "", true},
		{"invalid too few fields", "0 0 *", true},
		{"invalid too many fields", "0 0 * * * * *", true},
		{"invalid month 13", "0 0 * 13 *", true},
		{"invalid weekday 8", "0 0 * * 8", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := CronExpression()
			req := validator.StringRequest{
				Path:        path.Root("cron"),
				ConfigValue: types.StringValue(tt.value),
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tt.wantError {
				t.Errorf("CronExpression(%q): got error=%v, want error=%v", tt.value, hasError, tt.wantError)
			}
		})
	}
}

func TestCronExpression_Null(t *testing.T) {
	v := CronExpression()
	req := validator.StringRequest{
		Path:        path.Root("test"),
		ConfigValue: types.StringNull(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("CronExpression(null): unexpected error for null value: %v", resp.Diagnostics.Errors())
	}
}

func TestCronExpression_Unknown(t *testing.T) {
	v := CronExpression()
	req := validator.StringRequest{
		Path:        path.Root("test"),
		ConfigValue: types.StringUnknown(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("CronExpression(unknown): unexpected error: %v", resp.Diagnostics.Errors())
	}
}

func TestCronExpression_Description(t *testing.T) {
	v := CronExpression()
	ctx := context.Background()

	desc := v.Description(ctx)
	if desc == "" {
		t.Error("expected non-empty description")
	}

	mdDesc := v.MarkdownDescription(ctx)
	if mdDesc == "" {
		t.Error("expected non-empty markdown description")
	}
}

func TestTimezoneValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"valid UTC", "UTC", false},
		{"valid America/New_York", "America/New_York", false},
		{"valid Europe/London", "Europe/London", false},
		{"valid Asia/Tokyo", "Asia/Tokyo", false},
		{"valid America/Los_Angeles", "America/Los_Angeles", false},
		{"valid Europe/Paris", "Europe/Paris", false},
		{"valid Australia/Sydney", "Australia/Sydney", false},
		{"valid Africa/Cairo", "Africa/Cairo", false},
		{"valid EST", "EST", false},     // EST is a valid timezone abbreviation
		{"valid Local", "Local", false}, // Local is a valid timezone
		{"invalid New York", "New York", true},
		{"invalid random", "RandomTimezone", true},
		{"invalid number", "12345", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Timezone()
			req := validator.StringRequest{
				Path:        path.Root("timezone"),
				ConfigValue: types.StringValue(tt.value),
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tt.wantError {
				t.Errorf("Timezone(%q): got error=%v, want error=%v", tt.value, hasError, tt.wantError)
			}
		})
	}
}

func TestTimezone_Null(t *testing.T) {
	v := Timezone()
	req := validator.StringRequest{
		Path:        path.Root("test"),
		ConfigValue: types.StringNull(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Timezone(null): unexpected error for null value: %v", resp.Diagnostics.Errors())
	}
}

func TestTimezone_Unknown(t *testing.T) {
	v := Timezone()
	req := validator.StringRequest{
		Path:        path.Root("test"),
		ConfigValue: types.StringUnknown(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Timezone(unknown): unexpected error: %v", resp.Diagnostics.Errors())
	}
}

func TestTimezone_Description(t *testing.T) {
	v := Timezone()
	ctx := context.Background()

	desc := v.Description(ctx)
	if desc == "" {
		t.Error("expected non-empty description")
	}

	mdDesc := v.MarkdownDescription(ctx)
	if mdDesc == "" {
		t.Error("expected non-empty markdown description")
	}
}
