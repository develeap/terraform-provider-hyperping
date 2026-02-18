// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Terraform type conversion helpers - reduces code duplication across resources

// isNullOrUnknown checks if an attribute value is null or unknown.
// This is a common pattern used throughout the codebase for conditional logic.
func isNullOrUnknown(v attr.Value) bool {
	return v.IsNull() || v.IsUnknown()
}

// tfStringToPtr converts a types.String to *string.
// Returns nil if the value is null or unknown, otherwise returns a pointer to the string value.
func tfStringToPtr(v types.String) *string {
	if isNullOrUnknown(v) {
		return nil
	}
	s := v.ValueString()
	return &s
}

// tfInt64ToPtr converts a types.Int64 to *int64.
// Returns nil if the value is null or unknown, otherwise returns a pointer to the int64 value.
func tfInt64ToPtr(v types.Int64) *int64 {
	if isNullOrUnknown(v) {
		return nil
	}
	i := v.ValueInt64()
	return &i
}

// tfIntToPtr converts a types.Int64 to *int.
// Returns nil if the value is null or unknown, otherwise returns a pointer to the int value.
func tfIntToPtr(v types.Int64) *int {
	if isNullOrUnknown(v) {
		return nil
	}
	i := int(v.ValueInt64())
	return &i
}

// tfBoolToPtr converts a types.Bool to *bool.
// Returns nil if the value is null or unknown, otherwise returns a pointer to the bool value.
func tfBoolToPtr(v types.Bool) *bool {
	if isNullOrUnknown(v) {
		return nil
	}
	b := v.ValueBool()
	return &b
}

// stringValueOrEmpty returns the string value or empty string if null/unknown.
func stringValueOrEmpty(v types.String) string {
	if isNullOrUnknown(v) {
		return ""
	}
	return v.ValueString()
}

// int64ValueOrZero returns the int64 value or 0 if null/unknown.
func int64ValueOrZero(v types.Int64) int64 {
	if isNullOrUnknown(v) {
		return 0
	}
	return v.ValueInt64()
}

// boolValueOrFalse returns the bool value or false if null/unknown.
func boolValueOrFalse(v types.Bool) bool {
	if isNullOrUnknown(v) {
		return false
	}
	return v.ValueBool()
}

// tfListToStringSlice converts a types.List to []string.
// Returns nil if the list is null or unknown.
// Appends diagnostics if conversion fails.
func tfListToStringSlice(ctx context.Context, list types.List, diags *diag.Diagnostics) []string {
	if isNullOrUnknown(list) {
		return nil
	}

	var result []string
	diags.Append(list.ElementsAs(ctx, &result, false)...)
	return result
}

// stringSliceToTFList converts []string to types.List.
// Returns null list if slice is nil or empty.
func stringSliceToTFList(slice []string) types.List {
	if len(slice) == 0 {
		return types.ListNull(types.StringType)
	}

	elements := make([]attr.Value, len(slice))
	for i, s := range slice {
		elements[i] = types.StringValue(s)
	}

	list, _ := types.ListValue(types.StringType, elements)
	return list
}

// stringPtrToTF converts *string to types.String.
// Returns null if pointer is nil, otherwise returns string value.
func stringPtrToTF(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

// int64PtrToTF converts *int64 to types.Int64.
// Returns null if pointer is nil, otherwise returns int64 value.
func int64PtrToTF(i *int64) types.Int64 {
	if i == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*i)
}

// stringOrNull returns types.String with value if string is non-empty, otherwise null.
func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
