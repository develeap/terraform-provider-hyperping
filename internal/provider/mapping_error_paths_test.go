// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// headerScenario defines a single scenario for header mapping tests.
type headerScenario struct {
	name           string
	input          types.List
	wantErr        bool
	wantErrSummary string
	wantCount      int
	wantFirstName  string
}

// TestMapTFListToRequestHeaders_ErrorPaths tests error handling scenarios
// to increase coverage from 71.4% to 100%
func TestMapTFListToRequestHeaders_ErrorPaths(t *testing.T) {
	tests := buildHeaderScenarios()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := mapTFListToRequestHeaders(tt.input, &diags)

			assertHeaderDiagnostics(t, diags, tt.wantErr, tt.wantErrSummary)
			assertHeaderCount(t, result, tt.wantCount)
			assertFirstHeaderName(t, result, tt.wantFirstName)
		})
	}
}

// buildHeaderScenarios constructs all test scenarios for header mapping.
func buildHeaderScenarios() []headerScenario {
	return []headerScenario{
		{
			name:           "invalid element type - not an object",
			input:          buildStringTypeList(),
			wantErr:        true,
			wantErrSummary: "Invalid header element",
			wantCount:      0,
		},
		{
			name:           "invalid name field - not a string",
			input:          buildInvalidNameFieldList(),
			wantErr:        true,
			wantErrSummary: "Invalid header name",
			wantCount:      0,
		},
		{
			name:           "invalid value field - not a string",
			input:          buildInvalidValueFieldList(),
			wantErr:        true,
			wantErrSummary: "Invalid header value",
			wantCount:      0,
		},
		{
			name:          "valid headers processed correctly",
			input:         buildValidHeaderList(),
			wantErr:       false,
			wantCount:     1,
			wantFirstName: "X-Valid",
		},
		{
			name:          "headers with both fields null skipped",
			input:         buildNullAndValidHeaderList(),
			wantErr:       false,
			wantCount:     1,
			wantFirstName: "X-Valid",
		},
	}
}

func buildStringTypeList() types.List {
	l, _ := types.ListValue(types.StringType, []attr.Value{types.StringValue("invalid")})
	return l
}

func buildInvalidNameFieldList() types.List {
	attrTypes := map[string]attr.Type{"name": types.Int64Type, "value": types.StringType}
	obj, _ := types.ObjectValue(attrTypes, map[string]attr.Value{
		"name":  types.Int64Value(123),
		"value": types.StringValue("test-value"),
	})
	l, _ := types.ListValue(types.ObjectType{AttrTypes: attrTypes}, []attr.Value{obj})
	return l
}

func buildInvalidValueFieldList() types.List {
	attrTypes := map[string]attr.Type{"name": types.StringType, "value": types.BoolType}
	obj, _ := types.ObjectValue(attrTypes, map[string]attr.Value{
		"name":  types.StringValue("X-Test"),
		"value": types.BoolValue(true),
	})
	l, _ := types.ListValue(types.ObjectType{AttrTypes: attrTypes}, []attr.Value{obj})
	return l
}

func buildValidHeaderList() types.List {
	h, _ := types.ObjectValue(RequestHeaderAttrTypes(), map[string]attr.Value{
		"name":  types.StringValue("X-Valid"),
		"value": types.StringValue("valid-value"),
	})
	l, _ := types.ListValue(types.ObjectType{AttrTypes: RequestHeaderAttrTypes()}, []attr.Value{h})
	return l
}

func buildNullAndValidHeaderList() types.List {
	nullHeader, _ := types.ObjectValue(RequestHeaderAttrTypes(), map[string]attr.Value{
		"name":  types.StringNull(),
		"value": types.StringNull(),
	})
	validHeader, _ := types.ObjectValue(RequestHeaderAttrTypes(), map[string]attr.Value{
		"name":  types.StringValue("X-Valid"),
		"value": types.StringValue("valid"),
	})
	l, _ := types.ListValue(
		types.ObjectType{AttrTypes: RequestHeaderAttrTypes()},
		[]attr.Value{nullHeader, validHeader},
	)
	return l
}

// assertHeaderDiagnostics checks error presence and optional summary match.
func assertHeaderDiagnostics(t *testing.T, diags diag.Diagnostics, wantErr bool, wantErrSummary string) {
	t.Helper()
	if wantErr && !diags.HasError() {
		t.Error("expected diagnostic error, got none")
		return
	}
	if !wantErr && diags.HasError() {
		t.Errorf("unexpected diagnostic error: %v", diags.Errors())
		return
	}
	if wantErrSummary == "" {
		return
	}
	assertDiagSummaryFound(t, diags, wantErrSummary)
}

// assertDiagSummaryFound verifies that at least one error has the expected summary.
func assertDiagSummaryFound(t *testing.T, diags diag.Diagnostics, wantSummary string) {
	t.Helper()
	for _, err := range diags.Errors() {
		if err.Summary() == wantSummary {
			return
		}
	}
	t.Errorf("expected error summary %q not found in: %v", wantSummary, diags.Errors())
}

// assertHeaderCount checks the number of headers in the result.
func assertHeaderCount(t *testing.T, result []client.RequestHeader, wantCount int) {
	t.Helper()
	if len(result) != wantCount {
		t.Errorf("expected %d headers, got %d", wantCount, len(result))
	}
}

// assertFirstHeaderName checks the name of the first header when a name is expected.
func assertFirstHeaderName(t *testing.T, result []client.RequestHeader, wantName string) {
	t.Helper()
	if wantName == "" || len(result) == 0 {
		return
	}
	if result[0].Name != wantName {
		t.Errorf("expected first header name %q, got %q", wantName, result[0].Name)
	}
}
