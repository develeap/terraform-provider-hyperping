// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestNewOutageDataSource(t *testing.T) {
	ds := NewOutageDataSource()
	if ds == nil {
		t.Fatal("NewOutageDataSource returned nil")
	}
	if _, ok := ds.(*OutageDataSource); !ok {
		t.Errorf("expected *OutageDataSource, got %T", ds)
	}
}

func TestOutageDataSource_Metadata(t *testing.T) {
	d := &OutageDataSource{}

	req := datasource.MetadataRequest{
		ProviderTypeName: "hyperping",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	if resp.TypeName != "hyperping_outage" {
		t.Errorf("Expected type name 'hyperping_outage', got '%s'", resp.TypeName)
	}
}

func TestOutageDataSource_Schema(t *testing.T) {
	d := &OutageDataSource{}

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	if _, ok := resp.Schema.Attributes["id"]; !ok {
		t.Error("Schema missing 'id' attribute")
	}

	computedAttrs := []string{
		"monitor_uuid", "start_date", "end_date", "status_code", "description",
		"outage_type", "is_resolved", "duration_ms", "detected_location",
		"monitor", "acknowledged_by",
	}
	for _, attr := range computedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing '%s' attribute", attr)
		}
	}
}

func TestOutageDataSource_Configure(t *testing.T) {
	t.Run("valid client", func(t *testing.T) {
		d := &OutageDataSource{}
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
		d := &OutageDataSource{}

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
		d := &OutageDataSource{}

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

func TestMapOutageToDataSourceModel(t *testing.T) {
	t.Run("all fields populated with acknowledged_by", func(t *testing.T) {
		endDate := "2026-01-15T11:00:00Z"

		outage := &client.Outage{
			UUID:      "outage-123",
			StartDate: "2026-01-15T10:00:00Z",
			EndDate:   &endDate,
			Monitor: client.MonitorReference{
				UUID:     "mon-123",
				Name:     "API Monitor",
				URL:      "https://api.example.com",
				Protocol: "http",
			},
			StatusCode:       503,
			Description:      "Service unavailable",
			OutageType:       "automatic",
			IsResolved:       true,
			DurationMs:       3600000,
			DetectedLocation: "london",
			AcknowledgedBy: &client.AcknowledgedByUser{
				UUID:  "user-123",
				Email: "admin@example.com",
				Name:  "Admin User",
			},
		}

		model := &OutageDataSourceModel{}
		resp := &datasource.ReadResponse{}
		mapOutageToDataSourceModel(outage, model, &resp.Diagnostics)

		if resp.Diagnostics.HasError() {
			t.Fatalf("Unexpected error: %v", resp.Diagnostics)
		}

		if model.ID.ValueString() != "outage-123" {
			t.Errorf("Expected ID 'outage-123', got %s", model.ID.ValueString())
		}
		if model.MonitorUUID.ValueString() != "mon-123" {
			t.Errorf("Expected MonitorUUID 'mon-123', got %s", model.MonitorUUID.ValueString())
		}
		if model.StatusCode.ValueInt64() != 503 {
			t.Errorf("Expected StatusCode 503, got %d", model.StatusCode.ValueInt64())
		}
		if model.IsResolved.ValueBool() != true {
			t.Error("Expected IsResolved to be true")
		}
	})

	t.Run("minimal fields without acknowledged_by", func(t *testing.T) {
		outage := &client.Outage{
			UUID:      "outage-min",
			StartDate: "2026-01-20T08:00:00Z",
			EndDate:   nil,
			Monitor: client.MonitorReference{
				UUID:     "mon-min",
				Name:     "Min Monitor",
				URL:      "https://example.com",
				Protocol: "https",
			},
			StatusCode:       500,
			Description:      "Error",
			OutageType:       "manual",
			IsResolved:       false,
			DurationMs:       0,
			DetectedLocation: "virginia",
			AcknowledgedBy:   nil,
		}

		model := &OutageDataSourceModel{}
		resp := &datasource.ReadResponse{}
		mapOutageToDataSourceModel(outage, model, &resp.Diagnostics)

		if resp.Diagnostics.HasError() {
			t.Fatalf("Unexpected error: %v", resp.Diagnostics)
		}

		if !model.EndDate.IsNull() {
			t.Error("Expected EndDate to be null when nil in response")
		}
	})
}
