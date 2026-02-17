// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestMonitorResource_ConfigureWrongType(t *testing.T) {
	r := &MonitorResource{}

	req := resource.ConfigureRequest{
		ProviderData: "wrong type - should be *client.Client",
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("Expected error when provider data is wrong type")
	}

	// Verify error message contains expected text
	found := false
	for _, d := range resp.Diagnostics.Errors() {
		if d.Summary() == "Unexpected Resource Configure Type" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'Unexpected Resource Configure Type' error")
	}
}

func TestMonitorResource_ConfigureNilProviderData(t *testing.T) {
	r := &MonitorResource{}

	req := resource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	// Should not error, just return early
	if resp.Diagnostics.HasError() {
		t.Error("Expected no error when provider data is nil")
	}
}

func TestMonitorResource_ConfigureValidClient(t *testing.T) {
	r := &MonitorResource{}

	// Create a real client
	c := client.NewClient("test_api_key")

	req := resource.ConfigureRequest{
		ProviderData: c,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	// Should not error
	if resp.Diagnostics.HasError() {
		t.Errorf("Unexpected error: %v", resp.Diagnostics)
	}

	// Client should be set
	if r.client == nil {
		t.Error("Expected client to be set")
	}
}

func TestMonitorResource_Metadata(t *testing.T) {
	r := &MonitorResource{}

	req := resource.MetadataRequest{
		ProviderTypeName: "hyperping",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "hyperping_monitor" {
		t.Errorf("Expected type name 'hyperping_monitor', got '%s'", resp.TypeName)
	}
}

func TestMonitorResource_Schema(t *testing.T) {
	r := &MonitorResource{}

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	// Verify essential attributes exist
	requiredAttrs := []string{"id", "name", "url", "protocol", "http_method", "check_frequency"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing '%s' attribute", attr)
		}
	}

	// Verify new optional attributes exist (Phase 1 additions)
	newAttrs := []string{"port", "alerts_wait", "escalation_policy", "required_keyword"}
	for _, attr := range newAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing new '%s' attribute", attr)
		}
	}
}

func TestAllowedRegions(t *testing.T) {
	// Verify client.AllowedRegions contains expected values
	// From official API documentation (8 regions)
	expectedRegions := []string{
		// Europe
		"london", "frankfurt",
		// Asia Pacific
		"singapore", "sydney", "tokyo",
		// Americas
		"virginia", "saopaulo",
		// Middle East
		"bahrain",
	}

	if len(client.AllowedRegions) != len(expectedRegions) {
		t.Errorf("Expected %d regions, got %d", len(expectedRegions), len(client.AllowedRegions))
	}

	regionMap := make(map[string]bool)
	for _, r := range client.AllowedRegions {
		regionMap[r] = true
	}

	for _, expected := range expectedRegions {
		if !regionMap[expected] {
			t.Errorf("Expected region %q not found in client.AllowedRegions", expected)
		}
	}
}

// monitorToModelCase defines a single test scenario for mapMonitorToModel.
type monitorToModelCase struct {
	name    string
	monitor *client.Monitor
	verify  func(*testing.T, *MonitorResourceModel)
}

func buildMonitorToModelCases() []monitorToModelCase {
	port := 8080
	policy := "policy_abc123"
	keyword := "HEALTHY"

	return []monitorToModelCase{
		{
			name: "all fields populated",
			monitor: &client.Monitor{
				UUID:               "mon-123",
				Name:               "Test Monitor",
				URL:                "https://example.com",
				Protocol:           "http",
				HTTPMethod:         "GET",
				CheckFrequency:     60,
				ExpectedStatusCode: "200",
				FollowRedirects:    true,
				Paused:             false,
				Regions:            []string{"london", "frankfurt"},
				RequestHeaders:     []client.RequestHeader{{Name: "X-Custom", Value: "value"}},
				RequestBody:        "test body",
			},
			verify: func(t *testing.T, m *MonitorResourceModel) {
				t.Helper()
				if m.ID.ValueString() != "mon-123" {
					t.Errorf("expected ID 'mon-123', got %s", m.ID.ValueString())
				}
				if m.RequestBody.ValueString() != "test body" {
					t.Errorf("expected body 'test body', got %s", m.RequestBody.ValueString())
				}
			},
		},
		{
			name: "empty body",
			monitor: &client.Monitor{
				UUID:               "mon-456",
				Name:               "No Body",
				URL:                "https://example.com",
				Protocol:           "http",
				HTTPMethod:         "GET",
				CheckFrequency:     60,
				ExpectedStatusCode: "200",
				FollowRedirects:    true,
				RequestBody:        "",
			},
			verify: func(t *testing.T, m *MonitorResourceModel) {
				t.Helper()
				if !m.RequestBody.IsNull() {
					t.Error("expected RequestBody to be null for empty string")
				}
			},
		},
		{
			name: "empty regions and headers",
			monitor: &client.Monitor{
				UUID:               "mon-000",
				Name:               "Empty Collections",
				URL:                "https://example.com",
				Protocol:           "http",
				HTTPMethod:         "GET",
				CheckFrequency:     60,
				ExpectedStatusCode: "200",
				FollowRedirects:    true,
				Regions:            []string{},
				RequestHeaders:     []client.RequestHeader{},
			},
			verify: func(t *testing.T, m *MonitorResourceModel) {
				t.Helper()
				if !m.Regions.IsNull() {
					t.Error("expected Regions to be null for empty slice")
				}
				if !m.RequestHeaders.IsNull() {
					t.Error("expected RequestHeaders to be null for empty slice")
				}
			},
		},
		{
			name: "new optional fields populated",
			monitor: &client.Monitor{
				UUID:               "mon-new",
				Name:               "Monitor with New Fields",
				URL:                "https://example.com:8080",
				Protocol:           "port",
				HTTPMethod:         "GET",
				CheckFrequency:     60,
				ExpectedStatusCode: "200",
				FollowRedirects:    true,
				Paused:             false,
				Port:               &port,
				AlertsWait:         30,
				EscalationPolicy:   &policy,
				RequiredKeyword:    &keyword,
			},
			verify: verifyNewFieldsPopulated,
		},
		{
			name: "new optional fields null when not set",
			monitor: &client.Monitor{
				UUID:               "mon-null",
				Name:               "Monitor without New Fields",
				URL:                "https://example.com",
				Protocol:           "http",
				HTTPMethod:         "GET",
				CheckFrequency:     60,
				ExpectedStatusCode: "200",
				FollowRedirects:    true,
				Port:               nil,
				AlertsWait:         0,
				EscalationPolicy:   nil,
				RequiredKeyword:    nil,
			},
			verify: verifyNewFieldsNull,
		},
	}
}

// verifyNewFieldsPopulated checks that new optional fields are correctly set.
func verifyNewFieldsPopulated(t *testing.T, m *MonitorResourceModel) {
	t.Helper()
	if m.Port.ValueInt64() != 8080 {
		t.Errorf("expected Port 8080, got %d", m.Port.ValueInt64())
	}
	if m.AlertsWait.ValueInt64() != 30 {
		t.Errorf("expected AlertsWait 30, got %d", m.AlertsWait.ValueInt64())
	}
	if m.EscalationPolicy.ValueString() != "policy_abc123" {
		t.Errorf("expected EscalationPolicy 'policy_abc123', got %s", m.EscalationPolicy.ValueString())
	}
	if m.RequiredKeyword.ValueString() != "HEALTHY" {
		t.Errorf("expected RequiredKeyword 'HEALTHY', got %s", m.RequiredKeyword.ValueString())
	}
}

// verifyNewFieldsNull checks that new optional fields are null when not set.
func verifyNewFieldsNull(t *testing.T, m *MonitorResourceModel) {
	t.Helper()
	if !m.Port.IsNull() {
		t.Error("expected Port to be null when not set")
	}
	if !m.AlertsWait.IsNull() {
		t.Error("expected AlertsWait to be null when 0")
	}
	if !m.EscalationPolicy.IsNull() {
		t.Error("expected EscalationPolicy to be null when not set")
	}
	if !m.RequiredKeyword.IsNull() {
		t.Error("expected RequiredKeyword to be null when not set")
	}
}

func TestMonitorResource_mapMonitorToModel(t *testing.T) {
	r := &MonitorResource{}

	for _, tt := range buildMonitorToModelCases() {
		t.Run(tt.name, func(t *testing.T) {
			model := &MonitorResourceModel{}
			diags := &diag.Diagnostics{}
			r.mapMonitorToModel(tt.monitor, model, diags)

			if diags.HasError() {
				t.Errorf("unexpected error: %v", diags)
				return
			}
			tt.verify(t, model)
		})
	}
}
