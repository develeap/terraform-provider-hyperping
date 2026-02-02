// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestNewIncidentDataSource(t *testing.T) {
	ds := NewIncidentDataSource()
	if ds == nil {
		t.Fatal("NewIncidentDataSource returned nil")
	}
	if _, ok := ds.(*IncidentDataSource); !ok {
		t.Errorf("expected *IncidentDataSource, got %T", ds)
	}
}

func TestIncidentDataSource_Metadata(t *testing.T) {
	d := &IncidentDataSource{}

	req := datasource.MetadataRequest{
		ProviderTypeName: "hyperping",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	if resp.TypeName != "hyperping_incident" {
		t.Errorf("Expected type name 'hyperping_incident', got '%s'", resp.TypeName)
	}
}

func TestIncidentDataSource_Schema(t *testing.T) {
	d := &IncidentDataSource{}

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	if _, ok := resp.Schema.Attributes["id"]; !ok {
		t.Error("Schema missing 'id' attribute")
	}

	computedAttrs := []string{
		"title", "text", "type", "date", "affected_components", "status_pages", "updates",
	}
	for _, attr := range computedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing '%s' attribute", attr)
		}
	}
}

func TestIncidentDataSource_Configure(t *testing.T) {
	t.Run("valid client", func(t *testing.T) {
		d := &IncidentDataSource{}
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
		d := &IncidentDataSource{}

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
		d := &IncidentDataSource{}

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

func TestIncidentDataSource_mapIncidentToDataSourceModel(t *testing.T) {
	d := &IncidentDataSource{}

	t.Run("all fields populated with updates", func(t *testing.T) {
		inc := &client.Incident{
			UUID: "inc-123",
			Title: client.LocalizedText{
				En: "Service Degradation",
			},
			Text: client.LocalizedText{
				En: "We are experiencing slow response times",
			},
			Type:               "incident",
			Date:               "2026-01-15T10:00:00Z",
			AffectedComponents: []string{"comp-1", "comp-2"},
			StatusPages:        []string{"page-1"},
			Updates: []client.IncidentUpdate{
				{
					UUID: "upd-1",
					Date: "2026-01-15T10:30:00Z",
					Text: client.LocalizedText{
						En: "Issue identified",
					},
					Type: "identified",
				},
			},
		}

		model := &IncidentDataSourceModel{}
		resp := &datasource.ReadResponse{}
		d.mapIncidentToDataSourceModel(inc, model, &resp.Diagnostics)

		if resp.Diagnostics.HasError() {
			t.Fatalf("Unexpected error: %v", resp.Diagnostics)
		}

		if model.ID.ValueString() != "inc-123" {
			t.Errorf("Expected ID 'inc-123', got %s", model.ID.ValueString())
		}
		if model.Title.ValueString() != "Service Degradation" {
			t.Errorf("Expected title 'Service Degradation', got %s", model.Title.ValueString())
		}
		if model.Type.ValueString() != "incident" {
			t.Errorf("Expected type 'incident', got %s", model.Type.ValueString())
		}
	})

	t.Run("minimal fields no updates", func(t *testing.T) {
		inc := &client.Incident{
			UUID: "inc-min",
			Title: client.LocalizedText{
				En: "Minor Issue",
			},
			Text: client.LocalizedText{
				En: "Brief description",
			},
			Type:    "outage",
			Updates: []client.IncidentUpdate{},
		}

		model := &IncidentDataSourceModel{}
		resp := &datasource.ReadResponse{}
		d.mapIncidentToDataSourceModel(inc, model, &resp.Diagnostics)

		if resp.Diagnostics.HasError() {
			t.Fatalf("Unexpected error: %v", resp.Diagnostics)
		}

		if !model.Date.IsNull() {
			t.Error("Expected Date to be null for empty value")
		}
		if !model.Updates.IsNull() {
			t.Error("Expected Updates to be null for empty slice")
		}
	})
}
