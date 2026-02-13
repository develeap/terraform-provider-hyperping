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

func TestPortRangeValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     int64
		wantError bool
	}{
		{"valid port 80", 80, false},
		{"valid port 443", 443, false},
		{"valid port 1", 1, false},
		{"valid port 65535", 65535, false},
		{"valid port 3000", 3000, false},
		{"valid port 8080", 8080, false},
		{"valid port 22", 22, false},
		{"valid port 3306", 3306, false},
		{"valid port 5432", 5432, false},
		{"valid port 27017", 27017, false},
		{"invalid port 0", 0, true},
		{"invalid port -1", -1, true},
		{"invalid port -100", -100, true},
		{"invalid port 65536", 65536, true},
		{"invalid port 100000", 100000, true},
		{"invalid port 70000", 70000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := PortRange()
			req := validator.Int64Request{
				Path:        path.Root("port"),
				ConfigValue: types.Int64Value(tt.value),
			}
			resp := &validator.Int64Response{}
			v.ValidateInt64(context.Background(), req, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tt.wantError {
				t.Errorf("PortRange(%d): got error=%v, want error=%v", tt.value, hasError, tt.wantError)
			}
		})
	}
}

func TestPortRange_Null(t *testing.T) {
	v := PortRange()
	req := validator.Int64Request{
		Path:        path.Root("test"),
		ConfigValue: types.Int64Null(),
	}
	resp := &validator.Int64Response{}
	v.ValidateInt64(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("PortRange(null): unexpected error for null value: %v", resp.Diagnostics.Errors())
	}
}

func TestPortRange_Unknown(t *testing.T) {
	v := PortRange()
	req := validator.Int64Request{
		Path:        path.Root("test"),
		ConfigValue: types.Int64Unknown(),
	}
	resp := &validator.Int64Response{}
	v.ValidateInt64(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("PortRange(unknown): unexpected error: %v", resp.Diagnostics.Errors())
	}
}

func TestPortRange_Description(t *testing.T) {
	v := PortRange()
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

func TestHexColorValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"valid lowercase", "#ff5733", false},
		{"valid uppercase", "#FF5733", false},
		{"valid mixed case", "#Ff5733", false},
		{"valid black", "#000000", false},
		{"valid white", "#ffffff", false},
		{"valid white uppercase", "#FFFFFF", false},
		{"valid gray", "#808080", false},
		{"valid default hyperping", "#36b27e", false},
		{"valid all digits", "#123456", false},
		{"valid all letters", "#abcdef", false},
		{"valid mixed", "#a1b2c3", false},
		{"invalid no hash", "ff5733", true},
		{"invalid 3 digit", "#fff", true},
		{"invalid 8 digit", "#ff5733aa", true},
		{"invalid with alpha", "#ff5733ff", true},
		{"invalid chars g", "#gggggg", true},
		{"invalid chars z", "#zzzzzz", true},
		{"invalid empty", "", true},
		{"invalid just hash", "#", true},
		{"invalid 5 digit", "#12345", true},
		{"invalid 7 digit", "#1234567", true},
		{"invalid space", "#ff 5733", true},
		{"invalid special char", "#ff57@3", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := HexColor()
			req := validator.StringRequest{
				Path:        path.Root("accent_color"),
				ConfigValue: types.StringValue(tt.value),
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tt.wantError {
				t.Errorf("HexColor(%q): got error=%v, want error=%v", tt.value, hasError, tt.wantError)
			}
		})
	}
}

func TestHexColor_Null(t *testing.T) {
	v := HexColor()
	req := validator.StringRequest{
		Path:        path.Root("test"),
		ConfigValue: types.StringNull(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("HexColor(null): unexpected error for null value: %v", resp.Diagnostics.Errors())
	}
}

func TestHexColor_Unknown(t *testing.T) {
	v := HexColor()
	req := validator.StringRequest{
		Path:        path.Root("test"),
		ConfigValue: types.StringUnknown(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("HexColor(unknown): unexpected error: %v", resp.Diagnostics.Errors())
	}
}

func TestHexColor_Description(t *testing.T) {
	v := HexColor()
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

func TestEmailFormatValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"valid simple", "user@example.com", false},
		{"valid subdomain", "user@mail.example.com", false},
		{"valid plus", "user+tag@example.com", false},
		{"valid dots", "first.last@example.com", false},
		{"valid numbers", "user123@example456.com", false},
		{"valid hyphen", "user@ex-ample.com", false},
		{"valid underscore", "user_name@example.com", false},
		{"valid percent", "user%test@example.com", false},
		{"valid uppercase", "User@Example.COM", false},
		{"invalid no @", "userexample.com", true},
		{"invalid no domain", "user@", true},
		{"invalid no user", "@example.com", true},
		{"invalid no tld", "user@example", true},
		{"invalid spaces", "user @example.com", true},
		{"invalid double @", "user@@example.com", true},
		{"invalid empty", "", true},
		{"invalid special chars", "user<>@example.com", true},
		{"invalid no domain part", "user@.com", true},
		{"invalid tld too short", "user@example.c", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := EmailFormat()
			req := validator.StringRequest{
				Path:        path.Root("email"),
				ConfigValue: types.StringValue(tt.value),
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tt.wantError {
				t.Errorf("EmailFormat(%q): got error=%v, want error=%v", tt.value, hasError, tt.wantError)
			}
		})
	}
}

func TestEmailFormat_Null(t *testing.T) {
	v := EmailFormat()
	req := validator.StringRequest{
		Path:        path.Root("test"),
		ConfigValue: types.StringNull(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("EmailFormat(null): unexpected error for null value: %v", resp.Diagnostics.Errors())
	}
}

func TestEmailFormat_Unknown(t *testing.T) {
	v := EmailFormat()
	req := validator.StringRequest{
		Path:        path.Root("test"),
		ConfigValue: types.StringUnknown(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("EmailFormat(unknown): unexpected error: %v", resp.Diagnostics.Errors())
	}
}

func TestEmailFormat_Description(t *testing.T) {
	v := EmailFormat()
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
