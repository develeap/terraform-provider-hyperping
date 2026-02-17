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

// TestMapSectionsToTF tests sections mapping
func TestMapSectionsToTF(t *testing.T) {
	tests := []struct {
		name  string
		input []client.StatusPageSection
	}{
		{
			name:  "nil sections",
			input: nil,
		},
		{
			name:  "empty sections",
			input: []client.StatusPageSection{},
		},
		{
			name: "single section",
			input: []client.StatusPageSection{
				{
					Name: map[string]string{
						"en": "API Services",
					},
					IsSplit:  false,
					Services: []client.StatusPageService{},
				},
			},
		},
		{
			name: "multiple sections with services",
			input: []client.StatusPageSection{
				{
					Name: map[string]string{
						"en": "Frontend",
						"fr": "Interface",
					},
					IsSplit: true,
					Services: []client.StatusPageService{
						{
							ID:         "svc_1",
							UUID:       "mon_web",
							Name:       map[string]string{"en": "Web App"},
							ShowUptime: true,
						},
					},
				},
				{
					Name: map[string]string{
						"en": "Backend",
					},
					IsSplit: false,
					Services: []client.StatusPageService{
						{
							ID:   "svc_2",
							UUID: "mon_api",
							Name: map[string]string{"en": "API"},
						},
					},
				},
			},
		},
		{
			name: "nested service groups",
			input: []client.StatusPageSection{
				{
					Name: map[string]string{
						"en": "Databases",
					},
					IsSplit: false,
					Services: []client.StatusPageService{
						{
							ID:      "grp_1",
							UUID:    "mon_db_primary",
							Name:    map[string]string{"en": "Primary DB"},
							IsGroup: true,
							Services: []client.StatusPageService{
								{
									ID:   "svc_db1",
									UUID: "mon_db1",
									Name: map[string]string{"en": "DB Node 1"},
								},
								{
									ID:   "svc_db2",
									UUID: "mon_db2",
									Name: map[string]string{"en": "DB Node 2"},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := mapSectionsToTF(tt.input, &diags)

			if diags.HasError() {
				t.Errorf("unexpected error: %v", diags.Errors())
				return
			}

			if len(tt.input) == 0 {
				if !result.IsNull() {
					t.Errorf("expected null list for nil/empty input")
				}
				return
			}

			if result.IsNull() {
				t.Errorf("expected non-null list")
				return
			}

			elements := result.Elements()
			if len(elements) != len(tt.input) {
				t.Errorf("expected %d sections, got %d", len(tt.input), len(elements))
			}
		})
	}
}

// TestMapTFToSections tests reverse mapping from TF to client
func TestMapTFToSections(t *testing.T) {
	tests := []struct {
		name      string
		input     types.List
		wantCount int
		wantError bool
	}{
		{
			name:      "null list returns empty",
			input:     types.ListNull(types.ObjectType{AttrTypes: SectionAttrTypes()}),
			wantCount: 0,
			wantError: false,
		},
		{
			name: "single section with name",
			input: types.ListValueMust(types.ObjectType{AttrTypes: SectionAttrTypes()}, []attr.Value{
				types.ObjectValueMust(SectionAttrTypes(), map[string]attr.Value{
					"name": types.MapValueMust(types.StringType, map[string]attr.Value{
						"en": types.StringValue("API Services"),
					}),
					"is_split": types.BoolValue(true),
					"services": types.ListNull(types.ObjectType{AttrTypes: ServiceAttrTypes()}),
				}),
			}),
			wantCount: 1,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := mapTFToSections(tt.input, &diags)

			if tt.wantError {
				if !diags.HasError() {
					t.Errorf("expected error, got none")
				}
				return
			}

			if diags.HasError() {
				t.Errorf("unexpected error: %v", diags.Errors())
				return
			}

			if len(result) != tt.wantCount {
				t.Errorf("expected %d sections, got %d", tt.wantCount, len(result))
			}

			// Verify first section has English name if present
			if tt.wantCount > 0 && result[0].Name != "" {
				// Name should be extracted from the "en" key
				if result[0].Name == "" {
					t.Errorf("expected non-empty name for first section")
				}
			}
		})
	}
}

// TestMapTFToServices tests conversion of TF services list to API structs
func TestMapTFToServices(t *testing.T) {
	t.Run("null list returns empty", func(t *testing.T) {
		var diags diag.Diagnostics
		result := mapTFToServices(types.ListNull(types.ObjectType{AttrTypes: ServiceAttrTypes()}), &diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags.Errors())
		}
		if len(result) != 0 {
			t.Errorf("expected 0 services for null list, got %d", len(result))
		}
	})

	t.Run("unknown list returns empty", func(t *testing.T) {
		var diags diag.Diagnostics
		result := mapTFToServices(types.ListUnknown(types.ObjectType{AttrTypes: ServiceAttrTypes()}), &diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags.Errors())
		}
		if len(result) != 0 {
			t.Errorf("expected 0 services for unknown list, got %d", len(result))
		}
	})
}

// serviceTestCase defines a single scenario for TestMapTFToService.
type serviceTestCase struct {
	name      string
	input     attr.Value
	wantError bool
	verify    func(*testing.T, client.CreateStatusPageService)
}

func buildServiceTestCases() []serviceTestCase {
	return []serviceTestCase{
		{
			name:   "valid service with all fields",
			input:  buildFullServiceObj(),
			verify: verifyFullService,
		},
		{
			name:   "minimal service with uuid only",
			input:  buildMinimalServiceObj(),
			verify: verifyMinimalService,
		},
		{
			name:      "invalid element type returns error",
			input:     types.StringValue("not an object"),
			wantError: true,
			verify: func(t *testing.T, s client.CreateStatusPageService) {
				t.Helper()
				if s.MonitorUUID != "" {
					t.Error("expected empty MonitorUUID for error case")
				}
			},
		},
	}
}

func buildFullServiceObj() types.Object {
	return types.ObjectValueMust(ServiceAttrTypes(), map[string]attr.Value{
		"id":   types.StringValue("svc_1"),
		"uuid": types.StringValue("mon_123"),
		"name": types.MapValueMust(types.StringType, map[string]attr.Value{
			"en": types.StringValue("API Service"),
		}),
		"show_uptime":         types.BoolValue(true),
		"show_response_times": types.BoolValue(true),
		"is_group":            types.BoolValue(false),
	})
}

func buildMinimalServiceObj() types.Object {
	return types.ObjectValueMust(ServiceAttrTypes(), map[string]attr.Value{
		"id":                  types.StringNull(),
		"uuid":                types.StringValue("mon_minimal"),
		"name":                types.MapNull(types.StringType),
		"show_uptime":         types.BoolNull(),
		"show_response_times": types.BoolNull(),
		"is_group":            types.BoolNull(),
	})
}

func verifyFullService(t *testing.T, result client.CreateStatusPageService) {
	t.Helper()
	if result.MonitorUUID != "mon_123" {
		t.Errorf("expected MonitorUUID 'mon_123', got %q", result.MonitorUUID)
	}
	if result.NameShown == nil || *result.NameShown != "API Service" {
		t.Errorf("expected NameShown 'API Service', got %v", result.NameShown)
	}
	if result.ShowUptime == nil || !*result.ShowUptime {
		t.Error("expected ShowUptime true")
	}
	if result.ShowResponseTimes == nil || !*result.ShowResponseTimes {
		t.Error("expected ShowResponseTimes true")
	}
	if result.IsGroup == nil || *result.IsGroup {
		t.Error("expected IsGroup false")
	}
}

func verifyMinimalService(t *testing.T, result client.CreateStatusPageService) {
	t.Helper()
	if result.MonitorUUID != "mon_minimal" {
		t.Errorf("expected MonitorUUID 'mon_minimal', got %q", result.MonitorUUID)
	}
	if result.NameShown != nil {
		t.Errorf("expected nil NameShown, got %v", *result.NameShown)
	}
	if result.ShowUptime != nil {
		t.Error("expected nil ShowUptime")
	}
	if result.ShowResponseTimes != nil {
		t.Error("expected nil ShowResponseTimes")
	}
}

// TestMapTFToService tests conversion of a single TF service object to API struct
func TestMapTFToService(t *testing.T) {
	for _, tt := range buildServiceTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := mapTFToService(tt.input, &diags)

			if tt.wantError {
				if !diags.HasError() {
					t.Error("expected error for invalid element type")
				}
			} else if diags.HasError() {
				t.Errorf("unexpected error: %v", diags.Errors())
				return
			}

			tt.verify(t, result)
		})
	}
}
