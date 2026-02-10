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

func TestMonitorResource_mapMonitorToModel(t *testing.T) {
	r := &MonitorResource{}

	t.Run("all fields populated", func(t *testing.T) {
		monitor := &client.Monitor{
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
			RequestHeaders: []client.RequestHeader{
				{Name: "X-Custom", Value: "value"},
			},
			RequestBody: "test body",
		}

		model := &MonitorResourceModel{}
		diags := &diag.Diagnostics{}
		r.mapMonitorToModel(monitor, model, diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags)
		}
		if model.ID.ValueString() != "mon-123" {
			t.Errorf("expected ID 'mon-123', got %s", model.ID.ValueString())
		}
		if model.RequestBody.ValueString() != "test body" {
			t.Errorf("expected body 'test body', got %s", model.RequestBody.ValueString())
		}
	})

	t.Run("empty body", func(t *testing.T) {
		monitor := &client.Monitor{
			UUID:               "mon-456",
			Name:               "No Body",
			URL:                "https://example.com",
			Protocol:           "http",
			HTTPMethod:         "GET",
			CheckFrequency:     60,
			ExpectedStatusCode: "200",
			FollowRedirects:    true,
			RequestBody:        "", // Empty body
		}

		model := &MonitorResourceModel{}
		diags := &diag.Diagnostics{}
		r.mapMonitorToModel(monitor, model, diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags)
		}
		if !model.RequestBody.IsNull() {
			t.Error("expected RequestBody to be null for empty string")
		}
	})

	t.Run("empty regions and headers", func(t *testing.T) {
		monitor := &client.Monitor{
			UUID:               "mon-000",
			Name:               "Empty Collections",
			URL:                "https://example.com",
			Protocol:           "http",
			HTTPMethod:         "GET",
			CheckFrequency:     60,
			ExpectedStatusCode: "200",
			FollowRedirects:    true,
			Regions:            []string{},               // Empty slice
			RequestHeaders:     []client.RequestHeader{}, // Empty slice
		}

		model := &MonitorResourceModel{}
		diags := &diag.Diagnostics{}
		r.mapMonitorToModel(monitor, model, diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags)
		}
		if !model.Regions.IsNull() {
			t.Error("expected Regions to be null for empty slice")
		}
		if !model.RequestHeaders.IsNull() {
			t.Error("expected RequestHeaders to be null for empty slice")
		}
	})

	t.Run("new optional fields populated", func(t *testing.T) {
		port := 8080
		policy := "policy_abc123"
		keyword := "HEALTHY"
		monitor := &client.Monitor{
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
		}

		model := &MonitorResourceModel{}
		diags := &diag.Diagnostics{}
		r.mapMonitorToModel(monitor, model, diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags)
		}
		if model.Port.ValueInt64() != 8080 {
			t.Errorf("expected Port 8080, got %d", model.Port.ValueInt64())
		}
		if model.AlertsWait.ValueInt64() != 30 {
			t.Errorf("expected AlertsWait 30, got %d", model.AlertsWait.ValueInt64())
		}
		if model.EscalationPolicy.ValueString() != "policy_abc123" {
			t.Errorf("expected EscalationPolicy 'policy_abc123', got %s", model.EscalationPolicy.ValueString())
		}
		if model.RequiredKeyword.ValueString() != "HEALTHY" {
			t.Errorf("expected RequiredKeyword 'HEALTHY', got %s", model.RequiredKeyword.ValueString())
		}
	})

	t.Run("new optional fields null when not set", func(t *testing.T) {
		monitor := &client.Monitor{
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
		}

		model := &MonitorResourceModel{}
		diags := &diag.Diagnostics{}
		r.mapMonitorToModel(monitor, model, diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags)
		}
		if !model.Port.IsNull() {
			t.Error("expected Port to be null when not set")
		}
		if !model.AlertsWait.IsNull() {
			t.Error("expected AlertsWait to be null when 0")
		}
		if !model.EscalationPolicy.IsNull() {
			t.Error("expected EscalationPolicy to be null when not set")
		}
		if !model.RequiredKeyword.IsNull() {
			t.Error("expected RequiredKeyword to be null when not set")
		}
	})
}
