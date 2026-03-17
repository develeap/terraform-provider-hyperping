// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestIsNullOrUnknown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value types.String
		want  bool
	}{
		{
			name:  "null value",
			value: types.StringNull(),
			want:  true,
		},
		{
			name:  "unknown value",
			value: types.StringUnknown(),
			want:  true,
		},
		{
			name:  "non-null value",
			value: types.StringValue("test"),
			want:  false,
		},
		{
			name:  "empty string value",
			value: types.StringValue(""),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isNullOrUnknown(tt.value)
			if got != tt.want {
				t.Errorf("isNullOrUnknown() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTfStringToPtr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value types.String
		want  *string
	}{
		{
			name:  "null value returns nil",
			value: types.StringNull(),
			want:  nil,
		},
		{
			name:  "unknown value returns nil",
			value: types.StringUnknown(),
			want:  nil,
		},
		{
			name:  "normal value returns pointer",
			value: types.StringValue("test"),
			want:  testStringPtr("test"),
		},
		{
			name:  "empty string returns pointer to empty",
			value: types.StringValue(""),
			want:  testStringPtr(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tfStringToPtr(tt.value)
			if tt.want == nil {
				if got != nil {
					t.Errorf("tfStringToPtr() = %v, want nil", *got)
				}
			} else {
				if got == nil {
					t.Fatalf("tfStringToPtr() = nil, want %q", *tt.want)
				}
				if *got != *tt.want {
					t.Errorf("tfStringToPtr() = %q, want %q", *got, *tt.want)
				}
			}
		})
	}
}

func TestTfIntToPtr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value types.Int64
		want  *int
	}{
		{
			name:  "null value returns nil",
			value: types.Int64Null(),
			want:  nil,
		},
		{
			name:  "unknown value returns nil",
			value: types.Int64Unknown(),
			want:  nil,
		},
		{
			name:  "normal value returns pointer",
			value: types.Int64Value(42),
			want:  testIntPtr(42),
		},
		{
			name:  "zero value returns pointer to zero",
			value: types.Int64Value(0),
			want:  testIntPtr(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tfIntToPtr(tt.value)
			if tt.want == nil {
				if got != nil {
					t.Errorf("tfIntToPtr() = %v, want nil", *got)
				}
			} else {
				if got == nil {
					t.Fatalf("tfIntToPtr() = nil, want %d", *tt.want)
				}
				if *got != *tt.want {
					t.Errorf("tfIntToPtr() = %d, want %d", *got, *tt.want)
				}
			}
		})
	}
}

func TestTfBoolToPtr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value types.Bool
		want  *bool
	}{
		{
			name:  "null value returns nil",
			value: types.BoolNull(),
			want:  nil,
		},
		{
			name:  "unknown value returns nil",
			value: types.BoolUnknown(),
			want:  nil,
		},
		{
			name:  "true value returns pointer",
			value: types.BoolValue(true),
			want:  testBoolPtr(true),
		},
		{
			name:  "false value returns pointer",
			value: types.BoolValue(false),
			want:  testBoolPtr(false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tfBoolToPtr(tt.value)
			if tt.want == nil {
				if got != nil {
					t.Errorf("tfBoolToPtr() = %v, want nil", *got)
				}
			} else {
				if got == nil {
					t.Fatalf("tfBoolToPtr() = nil, want %v", *tt.want)
				}
				if *got != *tt.want {
					t.Errorf("tfBoolToPtr() = %v, want %v", *got, *tt.want)
				}
			}
		})
	}
}

// Helper functions for tests
func testStringPtr(s string) *string {
	return &s
}

func testIntPtr(i int) *int {
	return &i
}

func testBoolPtr(b bool) *bool {
	return &b
}
