// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestIsNullOrUnknown(t *testing.T) {
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
			got := isNullOrUnknown(tt.value)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTfStringToPtr(t *testing.T) {
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
			got := tfStringToPtr(tt.value)
			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.Equal(t, *tt.want, *got)
			}
		})
	}
}

func TestTfInt64ToPtr(t *testing.T) {
	tests := []struct {
		name  string
		value types.Int64
		want  *int64
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
			want:  testInt64Ptr(42),
		},
		{
			name:  "zero value returns pointer to zero",
			value: types.Int64Value(0),
			want:  testInt64Ptr(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tfInt64ToPtr(tt.value)
			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.Equal(t, *tt.want, *got)
			}
		})
	}
}

func TestTfIntToPtr(t *testing.T) {
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
			got := tfIntToPtr(tt.value)
			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.Equal(t, *tt.want, *got)
			}
		})
	}
}

func TestTfBoolToPtr(t *testing.T) {
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
			got := tfBoolToPtr(tt.value)
			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.Equal(t, *tt.want, *got)
			}
		})
	}
}

func TestStringValueOrEmpty(t *testing.T) {
	tests := []struct {
		name  string
		value types.String
		want  string
	}{
		{
			name:  "null returns empty",
			value: types.StringNull(),
			want:  "",
		},
		{
			name:  "unknown returns empty",
			value: types.StringUnknown(),
			want:  "",
		},
		{
			name:  "value returns value",
			value: types.StringValue("test"),
			want:  "test",
		},
		{
			name:  "empty value returns empty",
			value: types.StringValue(""),
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringValueOrEmpty(tt.value)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInt64ValueOrZero(t *testing.T) {
	tests := []struct {
		name  string
		value types.Int64
		want  int64
	}{
		{
			name:  "null returns zero",
			value: types.Int64Null(),
			want:  0,
		},
		{
			name:  "unknown returns zero",
			value: types.Int64Unknown(),
			want:  0,
		},
		{
			name:  "value returns value",
			value: types.Int64Value(42),
			want:  42,
		},
		{
			name:  "zero value returns zero",
			value: types.Int64Value(0),
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := int64ValueOrZero(tt.value)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBoolValueOrFalse(t *testing.T) {
	tests := []struct {
		name  string
		value types.Bool
		want  bool
	}{
		{
			name:  "null returns false",
			value: types.BoolNull(),
			want:  false,
		},
		{
			name:  "unknown returns false",
			value: types.BoolUnknown(),
			want:  false,
		},
		{
			name:  "true returns true",
			value: types.BoolValue(true),
			want:  true,
		},
		{
			name:  "false returns false",
			value: types.BoolValue(false),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := boolValueOrFalse(tt.value)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStringSliceToTFList(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		want  bool // whether result should be null
	}{
		{
			name:  "nil slice returns null",
			slice: nil,
			want:  true,
		},
		{
			name:  "empty slice returns null",
			slice: []string{},
			want:  true,
		},
		{
			name:  "non-empty slice returns list",
			slice: []string{"a", "b", "c"},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringSliceToTFList(tt.slice)
			assert.Equal(t, tt.want, got.IsNull())

			if !tt.want {
				var result []string
				diags := got.ElementsAs(context.Background(), &result, false)
				assert.False(t, diags.HasError())
				assert.Equal(t, tt.slice, result)
			}
		})
	}
}

func TestStringPtrToTF(t *testing.T) {
	tests := []struct {
		name string
		ptr  *string
		want bool // whether result should be null
	}{
		{
			name: "nil pointer returns null",
			ptr:  nil,
			want: true,
		},
		{
			name: "pointer to value returns value",
			ptr:  testStringPtr("test"),
			want: false,
		},
		{
			name: "pointer to empty string returns empty",
			ptr:  testStringPtr(""),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringPtrToTF(tt.ptr)
			assert.Equal(t, tt.want, got.IsNull())

			if !tt.want {
				assert.Equal(t, *tt.ptr, got.ValueString())
			}
		})
	}
}

func TestInt64PtrToTF(t *testing.T) {
	tests := []struct {
		name string
		ptr  *int64
		want bool // whether result should be null
	}{
		{
			name: "nil pointer returns null",
			ptr:  nil,
			want: true,
		},
		{
			name: "pointer to value returns value",
			ptr:  testInt64Ptr(42),
			want: false,
		},
		{
			name: "pointer to zero returns zero",
			ptr:  testInt64Ptr(0),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := int64PtrToTF(tt.ptr)
			assert.Equal(t, tt.want, got.IsNull())

			if !tt.want {
				assert.Equal(t, *tt.ptr, got.ValueInt64())
			}
		})
	}
}

func TestStringOrNull(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		isNull bool
	}{
		{
			name:   "empty string returns null",
			input:  "",
			isNull: true,
		},
		{
			name:   "non-empty string returns value",
			input:  "test",
			isNull: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringOrNull(tt.input)
			assert.Equal(t, tt.isNull, got.IsNull())

			if !tt.isNull {
				assert.Equal(t, tt.input, got.ValueString())
			}
		})
	}
}

func TestTfListToStringSlice(t *testing.T) {
	tests := []struct {
		name    string
		list    types.List
		want    []string
		wantNil bool
	}{
		{
			name:    "null list returns nil",
			list:    types.ListNull(types.StringType),
			want:    nil,
			wantNil: true,
		},
		{
			name:    "unknown list returns nil",
			list:    types.ListUnknown(types.StringType),
			want:    nil,
			wantNil: true,
		},
		{
			name:    "normal list returns slice",
			list:    stringSliceToTFList([]string{"a", "b", "c"}),
			want:    []string{"a", "b", "c"},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := &diag.Diagnostics{}
			got := tfListToStringSlice(context.Background(), tt.list, diags)
			assert.False(t, diags.HasError())

			if tt.wantNil {
				assert.Nil(t, got)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// Helper functions for tests - local to this file only
func testStringPtr(s string) *string {
	return &s
}

func testInt64Ptr(i int64) *int64 {
	return &i
}

func testIntPtr(i int) *int {
	return &i
}

func testBoolPtr(b bool) *bool {
	return &b
}
