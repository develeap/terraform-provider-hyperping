// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestNewIncidentsDataSource(t *testing.T) {
	ds := NewIncidentsDataSource()
	if ds == nil {
		t.Fatal("NewIncidentsDataSource returned nil")
	}
	if _, ok := ds.(*IncidentsDataSource); !ok {
		t.Errorf("expected *IncidentsDataSource, got %T", ds)
	}
}

func TestIncidentsDataSource_Metadata(t *testing.T) {
	d := &IncidentsDataSource{}

	req := datasource.MetadataRequest{
		ProviderTypeName: "hyperping",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	if resp.TypeName != "hyperping_incidents" {
		t.Errorf("Expected type name 'hyperping_incidents', got '%s'", resp.TypeName)
	}
}

func TestIncidentsDataSource_Schema(t *testing.T) {
	d := &IncidentsDataSource{}

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	if _, ok := resp.Schema.Attributes["incidents"]; !ok {
		t.Error("Schema missing 'incidents' attribute")
	}
}

func TestIncidentsDataSource_Configure(t *testing.T) {
	t.Run("valid client", func(t *testing.T) {
		d := &IncidentsDataSource{}
		c := &client.Client{}

		req := datasource.ConfigureRequest{
			ProviderData: c,
		}
		resp := &datasource.ConfigureResponse{}

		d.Configure(context.Background(), req, resp)

		if resp.Diagnostics.HasError() {
			t.Errorf("Unexpected error: %v", resp.Diagnostics)
		}

		if d.client == nil {
			t.Error("Expected client to be set")
		}
	})

	t.Run("nil provider data", func(t *testing.T) {
		d := &IncidentsDataSource{}

		req := datasource.ConfigureRequest{
			ProviderData: nil,
		}
		resp := &datasource.ConfigureResponse{}

		d.Configure(context.Background(), req, resp)

		if resp.Diagnostics.HasError() {
			t.Error("Expected no error when provider data is nil")
		}
	})

	t.Run("wrong type", func(t *testing.T) {
		d := &IncidentsDataSource{}

		req := datasource.ConfigureRequest{
			ProviderData: "wrong type",
		}
		resp := &datasource.ConfigureResponse{}

		d.Configure(context.Background(), req, resp)

		if !resp.Diagnostics.HasError() {
			t.Fatal("Expected error when provider data is wrong type")
		}
	})
}

func TestIncidentsDataSource_mapIncidentToDataModel(t *testing.T) {
	d := &IncidentsDataSource{}

	t.Run("all fields populated", func(t *testing.T) {
		inc := &client.Incident{
			UUID: "inc-full",
			Title: client.LocalizedText{
				En: "Complete Incident",
			},
			Text: client.LocalizedText{
				En: "Full description",
			},
			Type:               "outage",
			Date:               "2026-01-20T12:00:00Z",
			AffectedComponents: []string{"comp-a", "comp-b"},
			StatusPages:        []string{"page-x"},
			Updates: []client.IncidentUpdate{
				{
					UUID: "upd-a",
					Date: "2026-01-20T12:15:00Z",
					Text: client.LocalizedText{
						En: "First update",
					},
					Type: "investigating",
				},
			},
		}

		model := &IncidentDataModel{}
		resp := &datasource.ReadResponse{}
		d.mapIncidentToDataModel(inc, model, &resp.Diagnostics)

		if resp.Diagnostics.HasError() {
			t.Fatalf("Unexpected error: %v", resp.Diagnostics)
		}

		if model.ID.ValueString() != "inc-full" {
			t.Errorf("Expected ID 'inc-full', got %s", model.ID.ValueString())
		}
		if model.Title.ValueString() != "Complete Incident" {
			t.Errorf("Expected title 'Complete Incident', got %s", model.Title.ValueString())
		}
	})

	t.Run("minimal fields", func(t *testing.T) {
		inc := &client.Incident{
			UUID: "inc-min",
			Title: client.LocalizedText{
				En: "Minimal",
			},
			Text: client.LocalizedText{
				En: "Short",
			},
			Type: "incident",
		}

		model := &IncidentDataModel{}
		resp := &datasource.ReadResponse{}
		d.mapIncidentToDataModel(inc, model, &resp.Diagnostics)

		if resp.Diagnostics.HasError() {
			t.Fatalf("Unexpected error: %v", resp.Diagnostics)
		}

		if !model.Date.IsNull() {
			t.Error("Expected Date to be null for empty value")
		}
	})
}

func TestIncidentUpdateAttrTypes(t *testing.T) {
	attrTypes := IncidentUpdateAttrTypes()

	expectedKeys := []string{"id", "date", "text", "type"}
	for _, key := range expectedKeys {
		if _, ok := attrTypes[key]; !ok {
			t.Errorf("Expected attribute type for key '%s'", key)
		}
	}
}
