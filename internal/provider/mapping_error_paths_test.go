// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestMapTFListToRequestHeaders_ErrorPaths tests error handling scenarios
// to increase coverage from 71.4% to 100%
func TestMapTFListToRequestHeaders_ErrorPaths(t *testing.T) {
	t.Run("invalid element type - not an object", func(t *testing.T) {
		var diags diag.Diagnostics

		// Create a list with a string element instead of object
		invalidList, _ := types.ListValue(
			types.StringType, // Wrong type - should be ObjectType
			[]attr.Value{types.StringValue("invalid")},
		)

		// This will cause a type assertion failure in the function
		// Note: We can't easily create this scenario without breaking type safety,
		// but we can test by creating a list with the wrong element type
		result := mapTFListToRequestHeaders(invalidList, &diags)

		// Function should handle gracefully
		if len(result) != 0 {
			t.Errorf("expected empty result for invalid list, got %d headers", len(result))
		}
	})

	t.Run("invalid name field - not a string", func(t *testing.T) {
		var diags diag.Diagnostics

		// Create header object with non-string name field
		headerObj, _ := types.ObjectValue(
			map[string]attr.Type{
				"name":  types.Int64Type, // Wrong type
				"value": types.StringType,
			},
			map[string]attr.Value{
				"name":  types.Int64Value(123), // Invalid
				"value": types.StringValue("test-value"),
			},
		)

		headersList, _ := types.ListValue(
			types.ObjectType{AttrTypes: map[string]attr.Type{
				"name":  types.Int64Type,
				"value": types.StringType,
			}},
			[]attr.Value{headerObj},
		)

		result := mapTFListToRequestHeaders(headersList, &diags)

		if !diags.HasError() {
			t.Error("expected diagnostic error for invalid name field type")
		}
		if len(result) != 0 {
			t.Errorf("expected empty result when name field is invalid, got %d headers", len(result))
		}

		// Check error message
		errs := diags.Errors()
		found := false
		for _, err := range errs {
			if err.Summary() == "Invalid header name" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected 'Invalid header name' error")
		}
	})

	t.Run("invalid value field - not a string", func(t *testing.T) {
		var diags diag.Diagnostics

		// Create header object with non-string value field
		headerObj, _ := types.ObjectValue(
			map[string]attr.Type{
				"name":  types.StringType,
				"value": types.BoolType, // Wrong type
			},
			map[string]attr.Value{
				"name":  types.StringValue("X-Test"),
				"value": types.BoolValue(true), // Invalid
			},
		)

		headersList, _ := types.ListValue(
			types.ObjectType{AttrTypes: map[string]attr.Type{
				"name":  types.StringType,
				"value": types.BoolType,
			}},
			[]attr.Value{headerObj},
		)

		result := mapTFListToRequestHeaders(headersList, &diags)

		if !diags.HasError() {
			t.Error("expected diagnostic error for invalid value field type")
		}
		if len(result) != 0 {
			t.Errorf("expected empty result when value field is invalid, got %d headers", len(result))
		}

		// Check error message
		errs := diags.Errors()
		found := false
		for _, err := range errs {
			if err.Summary() == "Invalid header value" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected 'Invalid header value' error")
		}
	})

	t.Run("valid headers processed correctly", func(t *testing.T) {
		var diags diag.Diagnostics

		validHeader, _ := types.ObjectValue(RequestHeaderAttrTypes(), map[string]attr.Value{
			"name":  types.StringValue("X-Valid"),
			"value": types.StringValue("valid-value"),
		})

		headersList, _ := types.ListValue(
			types.ObjectType{AttrTypes: RequestHeaderAttrTypes()},
			[]attr.Value{validHeader},
		)

		result := mapTFListToRequestHeaders(headersList, &diags)

		if diags.HasError() {
			t.Errorf("unexpected error for valid headers: %v", diags.Errors())
		}
		if len(result) != 1 {
			t.Errorf("expected 1 header, got %d", len(result))
		}
		if len(result) > 0 && result[0].Name != "X-Valid" {
			t.Errorf("expected header name 'X-Valid', got '%s'", result[0].Name)
		}
	})

	t.Run("headers with both fields null skipped", func(t *testing.T) {
		var diags diag.Diagnostics

		// Header with both fields null should be skipped (not appended)
		headerWithNulls, _ := types.ObjectValue(RequestHeaderAttrTypes(), map[string]attr.Value{
			"name":  types.StringNull(),
			"value": types.StringNull(),
		})

		validHeader, _ := types.ObjectValue(RequestHeaderAttrTypes(), map[string]attr.Value{
			"name":  types.StringValue("X-Valid"),
			"value": types.StringValue("valid"),
		})

		headersList, _ := types.ListValue(
			types.ObjectType{AttrTypes: RequestHeaderAttrTypes()},
			[]attr.Value{headerWithNulls, validHeader},
		)

		result := mapTFListToRequestHeaders(headersList, &diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags.Errors())
		}
		// Only valid header should be in result
		if len(result) != 1 {
			t.Errorf("expected 1 header (null fields skipped), got %d", len(result))
		}
		if len(result) > 0 && result[0].Name != "X-Valid" {
			t.Errorf("expected 'X-Valid', got '%s'", result[0].Name)
		}
	})
}
