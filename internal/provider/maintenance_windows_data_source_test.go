// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestNewMaintenanceWindowsDataSource(t *testing.T) {
	ds := NewMaintenanceWindowsDataSource()
	if ds == nil {
		t.Fatal("NewMaintenanceWindowsDataSource returned nil")
	}
	if _, ok := ds.(*MaintenanceWindowsDataSource); !ok {
		t.Errorf("expected *MaintenanceWindowsDataSource, got %T", ds)
	}
}

func TestMaintenanceWindowsDataSource_Metadata(t *testing.T) {
	d := &MaintenanceWindowsDataSource{}

	req := datasource.MetadataRequest{
		ProviderTypeName: "hyperping",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	if resp.TypeName != "hyperping_maintenance_windows" {
		t.Errorf("Expected type name 'hyperping_maintenance_windows', got '%s'", resp.TypeName)
	}
}

func TestMaintenanceWindowsDataSource_Schema(t *testing.T) {
	d := &MaintenanceWindowsDataSource{}

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	if _, ok := resp.Schema.Attributes["maintenance_windows"]; !ok {
		t.Error("Schema missing 'maintenance_windows' attribute")
	}
}

func TestMaintenanceWindowsDataSource_Configure(t *testing.T) {
	t.Run("valid client", func(t *testing.T) {
		d := &MaintenanceWindowsDataSource{}
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
		d := &MaintenanceWindowsDataSource{}

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
		d := &MaintenanceWindowsDataSource{}

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

func TestMaintenanceWindowsDataSource_mapMaintenanceToDataModel(t *testing.T) {
	d := &MaintenanceWindowsDataSource{}

	t.Run("all fields populated", func(t *testing.T) {
		startDate := "2026-03-01T10:00:00Z"
		endDate := "2026-03-01T12:00:00Z"

		maint := &client.Maintenance{
			UUID: "maint-full",
			Name: "Complete Maintenance",
			Title: client.LocalizedText{
				En: "Full Maintenance Window",
			},
			Text: client.LocalizedText{
				En: "Detailed description",
			},
			StartDate:   &startDate,
			EndDate:     &endDate,
			Timezone:    "Europe/London",
			Monitors:    []string{"mon-a", "mon-b", "mon-c"},
			StatusPages: []string{"page-x", "page-y"},
			CreatedBy:   "admin@company.com",
		}

		model := &MaintenanceWindowDataModel{}
		resp := &datasource.ReadResponse{}
		d.mapMaintenanceToDataModel(maint, model, &resp.Diagnostics)

		if resp.Diagnostics.HasError() {
			t.Fatalf("Unexpected error: %v", resp.Diagnostics)
		}

		if model.ID.ValueString() != "maint-full" {
			t.Errorf("Expected ID 'maint-full', got %s", model.ID.ValueString())
		}
		if model.Name.ValueString() != "Complete Maintenance" {
			t.Errorf("Expected name 'Complete Maintenance', got %s", model.Name.ValueString())
		}
	})

	t.Run("minimal fields", func(t *testing.T) {
		maint := &client.Maintenance{
			UUID: "maint-min",
			Name: "Minimal",
			Title: client.LocalizedText{
				En: "Min",
			},
			Text: client.LocalizedText{
				En: "Short",
			},
			Timezone:  "UTC",
			CreatedBy: "user@test.com",
		}

		model := &MaintenanceWindowDataModel{}
		resp := &datasource.ReadResponse{}
		d.mapMaintenanceToDataModel(maint, model, &resp.Diagnostics)

		if resp.Diagnostics.HasError() {
			t.Fatalf("Unexpected error: %v", resp.Diagnostics)
		}

		if !model.StartDate.IsNull() {
			t.Error("Expected StartDate to be null for nil value")
		}
		if !model.EndDate.IsNull() {
			t.Error("Expected EndDate to be null for nil value")
		}
	})
}
