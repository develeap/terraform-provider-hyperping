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
		{"invalid - no separators", "nohyphensorunderscores", true},
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
