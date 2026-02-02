// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestNewIncidentUpdateResource(t *testing.T) {
	r := NewIncidentUpdateResource()
	if r == nil {
		t.Error("Expected non-nil resource")
	}
}

func TestIncidentUpdateResource_Metadata(t *testing.T) {
	r := &IncidentUpdateResource{}

	req := resource.MetadataRequest{
		ProviderTypeName: "hyperping",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "hyperping_incident_update" {
		t.Errorf("Expected type name 'hyperping_incident_update', got '%s'", resp.TypeName)
	}
}

func TestIncidentUpdateResource_Schema(t *testing.T) {
	r := &IncidentUpdateResource{}

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	// Verify essential attributes exist
	requiredAttrs := []string{"id", "incident_id", "text", "type", "date"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing '%s' attribute", attr)
		}
	}
}

func TestIncidentUpdateResource_ConfigureWrongType(t *testing.T) {
	r := &IncidentUpdateResource{}

	req := resource.ConfigureRequest{
		ProviderData: "wrong type",
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("Expected error for wrong type")
	}
}

func TestIncidentUpdateResource_ConfigureNilProviderData(t *testing.T) {
	r := &IncidentUpdateResource{}

	req := resource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Error("Expected no error when provider data is nil")
	}
}

func TestIncidentUpdateResource_ConfigureValidClient(t *testing.T) {
	r := &IncidentUpdateResource{}

	c := client.NewClient("test_api_key")

	req := resource.ConfigureRequest{
		ProviderData: c,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Unexpected error: %v", resp.Diagnostics)
	}

	if r.client == nil {
		t.Error("Expected client to be set")
	}
}

func TestParseIncidentUpdateID(t *testing.T) {
	tests := []struct {
		name             string
		id               string
		expectedIncident string
		expectedUpdate   string
	}{
		{
			name:             "valid composite ID",
			id:               "inci_123/update_456",
			expectedIncident: "inci_123",
			expectedUpdate:   "update_456",
		},
		{
			name:             "ID with multiple slashes",
			id:               "inci_123/update_456/extra",
			expectedIncident: "inci_123",
			expectedUpdate:   "update_456/extra",
		},
		{
			name:             "empty ID",
			id:               "",
			expectedIncident: "",
			expectedUpdate:   "",
		},
		{
			name:             "no slash",
			id:               "invalid",
			expectedIncident: "",
			expectedUpdate:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			incidentID, updateID := parseIncidentUpdateID(tt.id)
			if incidentID != tt.expectedIncident {
				t.Errorf("Expected incident ID '%s', got '%s'", tt.expectedIncident, incidentID)
			}
			if updateID != tt.expectedUpdate {
				t.Errorf("Expected update ID '%s', got '%s'", tt.expectedUpdate, updateID)
			}
		})
	}
}
