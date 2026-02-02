// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestNewMaintenanceWindowDataSource(t *testing.T) {
	ds := NewMaintenanceWindowDataSource()
	if ds == nil {
		t.Fatal("NewMaintenanceWindowDataSource returned nil")
	}
	if _, ok := ds.(*MaintenanceWindowDataSource); !ok {
		t.Errorf("expected *MaintenanceWindowDataSource, got %T", ds)
	}
}

func TestMaintenanceWindowDataSource_Metadata(t *testing.T) {
	d := &MaintenanceWindowDataSource{}

	req := datasource.MetadataRequest{
		ProviderTypeName: "hyperping",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	if resp.TypeName != "hyperping_maintenance_window" {
		t.Errorf("Expected type name 'hyperping_maintenance_window', got '%s'", resp.TypeName)
	}
}

func TestMaintenanceWindowDataSource_Schema(t *testing.T) {
	d := &MaintenanceWindowDataSource{}

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	if _, ok := resp.Schema.Attributes["id"]; !ok {
		t.Error("Schema missing 'id' attribute")
	}

	computedAttrs := []string{
		"name", "title", "text", "start_date", "end_date",
		"timezone", "monitors", "status_pages", "created_by",
	}
	for _, attr := range computedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing '%s' attribute", attr)
		}
	}
}

func TestMaintenanceWindowDataSource_Configure(t *testing.T) {
	t.Run("valid client", func(t *testing.T) {
		d := &MaintenanceWindowDataSource{}
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
		d := &MaintenanceWindowDataSource{}

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
		d := &MaintenanceWindowDataSource{}

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

func TestMaintenanceWindowDataSource_mapMaintenanceToDataSourceModel(t *testing.T) {
	d := &MaintenanceWindowDataSource{}

	t.Run("all fields populated", func(t *testing.T) {
		startDate := "2026-02-01T02:00:00Z"
		endDate := "2026-02-01T04:00:00Z"

		maint := &client.Maintenance{
			UUID: "maint-123",
			Name: "Database Upgrade",
			Title: client.LocalizedText{
				En: "Scheduled Maintenance",
			},
			Text: client.LocalizedText{
				En: "We will be upgrading the database",
			},
			StartDate:   &startDate,
			EndDate:     &endDate,
			Timezone:    "UTC",
			Monitors:    []string{"mon-1", "mon-2"},
			StatusPages: []string{"page-1"},
			CreatedBy:   "admin@example.com",
		}

		model := &MaintenanceWindowDataSourceModel{}
		resp := &datasource.ReadResponse{}
		d.mapMaintenanceToDataSourceModel(maint, model, &resp.Diagnostics)

		if resp.Diagnostics.HasError() {
			t.Fatalf("Unexpected error: %v", resp.Diagnostics)
		}

		if model.ID.ValueString() != "maint-123" {
			t.Errorf("Expected ID 'maint-123', got %s", model.ID.ValueString())
		}
		if model.Name.ValueString() != "Database Upgrade" {
			t.Errorf("Expected name 'Database Upgrade', got %s", model.Name.ValueString())
		}
		if model.Title.ValueString() != "Scheduled Maintenance" {
			t.Errorf("Expected title 'Scheduled Maintenance', got %s", model.Title.ValueString())
		}
	})

	t.Run("minimal fields with nil dates", func(t *testing.T) {
		maint := &client.Maintenance{
			UUID: "maint-min",
			Name: "Minimal Maintenance",
			Title: client.LocalizedText{
				En: "Brief",
			},
			Text: client.LocalizedText{
				En: "Short",
			},
			StartDate: nil,
			EndDate:   nil,
			Timezone:  "America/New_York",
			CreatedBy: "user@example.com",
		}

		model := &MaintenanceWindowDataSourceModel{}
		resp := &datasource.ReadResponse{}
		d.mapMaintenanceToDataSourceModel(maint, model, &resp.Diagnostics)

		if resp.Diagnostics.HasError() {
			t.Fatalf("Unexpected error: %v", resp.Diagnostics)
		}

		if !model.StartDate.IsNull() {
			t.Error("Expected StartDate to be null when nil in response")
		}
		if !model.EndDate.IsNull() {
			t.Error("Expected EndDate to be null when nil in response")
		}
	})
}
