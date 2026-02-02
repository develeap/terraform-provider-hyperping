// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestNewOutagesDataSource(t *testing.T) {
	ds := NewOutagesDataSource()
	if ds == nil {
		t.Fatal("NewOutagesDataSource returned nil")
	}
	if _, ok := ds.(*OutagesDataSource); !ok {
		t.Errorf("expected *OutagesDataSource, got %T", ds)
	}
}

func TestOutagesDataSource_Metadata(t *testing.T) {
	d := &OutagesDataSource{}

	req := datasource.MetadataRequest{
		ProviderTypeName: "hyperping",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	if resp.TypeName != "hyperping_outages" {
		t.Errorf("Expected type name 'hyperping_outages', got '%s'", resp.TypeName)
	}
}

func TestOutagesDataSource_Schema(t *testing.T) {
	d := &OutagesDataSource{}

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	if _, ok := resp.Schema.Attributes["outages"]; !ok {
		t.Error("Schema missing 'outages' attribute")
	}
}

func TestOutagesDataSource_Configure(t *testing.T) {
	t.Run("valid client", func(t *testing.T) {
		d := &OutagesDataSource{}
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
		d := &OutagesDataSource{}

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
		d := &OutagesDataSource{}

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

func TestOutagesDataSource_mapOutageToDataModel(t *testing.T) {
	d := &OutagesDataSource{}

	t.Run("all fields populated", func(t *testing.T) {
		endDate := "2026-02-01T15:00:00Z"

		outage := &client.Outage{
			UUID:      "outage-full",
			StartDate: "2026-02-01T14:00:00Z",
			EndDate:   &endDate,
			Monitor: client.MonitorReference{
				UUID:     "mon-full",
				Name:     "Full Monitor",
				URL:      "https://full.example.com",
				Protocol: "https",
			},
			StatusCode:       502,
			Description:      "Bad Gateway",
			OutageType:       "automatic",
			IsResolved:       true,
			DurationMs:       3600000,
			DetectedLocation: "frankfurt",
			AcknowledgedBy: &client.AcknowledgedByUser{
				UUID:  "user-full",
				Email: "user@example.com",
				Name:  "Test User",
			},
		}

		model := &OutageDataModel{}
		resp := &datasource.ReadResponse{}
		d.mapOutageToDataModel(outage, model, &resp.Diagnostics)

		if resp.Diagnostics.HasError() {
			t.Fatalf("Unexpected error: %v", resp.Diagnostics)
		}

		if model.ID.ValueString() != "outage-full" {
			t.Errorf("Expected ID 'outage-full', got %s", model.ID.ValueString())
		}
		if model.StatusCode.ValueInt64() != 502 {
			t.Errorf("Expected StatusCode 502, got %d", model.StatusCode.ValueInt64())
		}
	})

	t.Run("minimal fields", func(t *testing.T) {
		outage := &client.Outage{
			UUID:      "outage-min",
			StartDate: "2026-02-05T10:00:00Z",
			EndDate:   nil,
			Monitor: client.MonitorReference{
				UUID:     "mon-min",
				Name:     "Min",
				URL:      "https://min.com",
				Protocol: "http",
			},
			StatusCode:       404,
			Description:      "Not Found",
			OutageType:       "manual",
			IsResolved:       false,
			DurationMs:       1000,
			DetectedLocation: "sydney",
			AcknowledgedBy:   nil,
		}

		model := &OutageDataModel{}
		resp := &datasource.ReadResponse{}
		d.mapOutageToDataModel(outage, model, &resp.Diagnostics)

		if resp.Diagnostics.HasError() {
			t.Fatalf("Unexpected error: %v", resp.Diagnostics)
		}

		if !model.EndDate.IsNull() {
			t.Error("Expected EndDate to be null for nil value")
		}
	})
}
