// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFlexibleString_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected FlexibleString
		wantErr  bool
	}{
		{
			name:     "string value",
			input:    `"200"`,
			expected: FlexibleString("200"),
			wantErr:  false,
		},
		{
			name:     "integer number",
			input:    `200`,
			expected: FlexibleString("200"),
			wantErr:  false,
		},
		{
			name:     "float number",
			input:    `200.5`,
			expected: FlexibleString("200.5"),
			wantErr:  false,
		},
		{
			name:     "empty string",
			input:    `""`,
			expected: FlexibleString(""),
			wantErr:  false,
		},
		{
			name:     "zero number",
			input:    `0`,
			expected: FlexibleString("0"),
			wantErr:  false,
		},
		{
			name:     "negative number",
			input:    `-1`,
			expected: FlexibleString("-1"),
			wantErr:  false,
		},
		{
			name:     "boolean - should error",
			input:    `true`,
			expected: FlexibleString(""),
			wantErr:  true,
		},
		{
			name:     "null becomes empty string",
			input:    `null`,
			expected: FlexibleString(""),
			wantErr:  false,
		},
		{
			name:     "array - should error",
			input:    `[1,2,3]`,
			expected: FlexibleString(""),
			wantErr:  true,
		},
		{
			name:     "object - should error",
			input:    `{"key":"value"}`,
			expected: FlexibleString(""),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result FlexibleString
			err := json.Unmarshal([]byte(tt.input), &result)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for input %s, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %s: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("UnmarshalJSON() = %q, want %q", result, tt.expected)
				}
			}
		})
	}
}

func TestFlexibleString_String(t *testing.T) {
	tests := []struct {
		name     string
		input    FlexibleString
		expected string
	}{
		{
			name:     "normal string",
			input:    FlexibleString("200"),
			expected: "200",
		},
		{
			name:     "empty string",
			input:    FlexibleString(""),
			expected: "",
		},
		{
			name:     "numeric string",
			input:    FlexibleString("404"),
			expected: "404",
		},
		{
			name:     "string with special characters",
			input:    FlexibleString("status-code: 200"),
			expected: "status-code: 200",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.String()
			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFlexibleString_Integration(t *testing.T) {
	// Test that FlexibleString works in a real struct
	type testStruct struct {
		StatusCode FlexibleString `json:"status_code"`
	}

	t.Run("unmarshal string value", func(t *testing.T) {
		data := []byte(`{"status_code": "200"}`)
		var result testStruct
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if result.StatusCode != "200" {
			t.Errorf("expected '200', got %q", result.StatusCode)
		}
	})

	t.Run("unmarshal integer value", func(t *testing.T) {
		data := []byte(`{"status_code": 404}`)
		var result testStruct
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if result.StatusCode != "404" {
			t.Errorf("expected '404', got %q", result.StatusCode)
		}
	})

	t.Run("unmarshal float value", func(t *testing.T) {
		data := []byte(`{"status_code": 200.5}`)
		var result testStruct
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if result.StatusCode != "200.5" {
			t.Errorf("expected '200.5', got %q", result.StatusCode)
		}
	})

	t.Run("oversized input rejected", func(t *testing.T) {
		oversized := `"` + strings.Repeat("a", 200) + `"`
		var fs FlexibleString
		err := json.Unmarshal([]byte(oversized), &fs)
		if err == nil {
			t.Error("expected error for oversized FlexibleString input, got nil")
		}
		if !strings.Contains(err.Error(), "exceeds maximum size") {
			t.Errorf("expected size error, got: %v", err)
		}
	})
}

func TestCreateMonitorRequest_Validate(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		req := CreateMonitorRequest{Name: "test", URL: "https://example.com"}
		if err := req.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("name too long", func(t *testing.T) {
		req := CreateMonitorRequest{Name: strings.Repeat("a", 256), URL: "https://example.com"}
		if err := req.Validate(); err == nil {
			t.Error("expected validation error for oversized name")
		}
	})

	t.Run("url too long", func(t *testing.T) {
		req := CreateMonitorRequest{Name: "test", URL: strings.Repeat("a", 2049)}
		if err := req.Validate(); err == nil {
			t.Error("expected validation error for oversized URL")
		}
	})
}

func TestCreateIncidentRequest_Validate(t *testing.T) {
	t.Run("title too long", func(t *testing.T) {
		req := CreateIncidentRequest{
			Title: LocalizedText{En: strings.Repeat("x", 256)},
			Text:  LocalizedText{En: "short"},
		}
		if err := req.Validate(); err == nil {
			t.Error("expected validation error for oversized title")
		}
	})

	t.Run("text too long", func(t *testing.T) {
		req := CreateIncidentRequest{
			Title: LocalizedText{En: "short"},
			Text:  LocalizedText{En: strings.Repeat("x", 10001)},
		}
		if err := req.Validate(); err == nil {
			t.Error("expected validation error for oversized text")
		}
	})
}
