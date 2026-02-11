// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestAccMaintenanceResource_basic(t *testing.T) {
	startStr, endStr := generateMaintenanceTimeRange(24, 2)

	fixture := &maintenanceTestFixture{
		UUID:      "mw_test_123",
		Name:      "Test Maintenance",
		Title:     "Test Maintenance Title",
		Text:      "Test maintenance description",
		StartDate: startStr,
		EndDate:   endStr,
		Monitors:  []string{"mon_123"},
	}

	server := newSimpleMaintenanceServer(fixture)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			{
				Config: generateMaintenanceConfig(server.URL, fixture.Name, fixture.Title, fixture.Text, startStr, endStr, fixture.Monitors),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "id", "mw_test_123"),
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "name", "Test Maintenance"),
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "title", "Test Maintenance Title"),
				),
			},
		},
	})
}

func TestAccMaintenanceResource_full(t *testing.T) {
	startStr, endStr := generateMaintenanceTimeRange(48, 4)

	fixture := &maintenanceTestFixture{
		UUID:                "mw_full_123",
		Name:                "Full Maintenance",
		Title:               "Full Maintenance Title",
		Text:                "Complete maintenance description",
		StartDate:           startStr,
		EndDate:             endStr,
		Monitors:            []string{"mon_123", "mon_456"},
		StatusPages:         []string{"sp_main"},
		NotificationOption:  "scheduled",
		NotificationMinutes: 120,
	}

	server := newSimpleMaintenanceServer(fixture)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			{
				Config: generateMaintenanceConfigFull(server.URL, fixture.Name, fixture.Title, fixture.Text, startStr, endStr, fixture.Monitors, fixture.StatusPages, fixture.NotificationOption, fixture.NotificationMinutes),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "id", "mw_full_123"),
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "name", "Full Maintenance"),
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "notification_option", "scheduled"),
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "notification_minutes", "120"),
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "monitors.#", "2"),
				),
			},
		},
	})
}

func TestAccMaintenanceResource_timeUpdate(t *testing.T) {
	startStr, endStr := generateMaintenanceTimeRange(24, 2)
	newStartStr, newEndStr := generateMaintenanceTimeRange(48, 4)

	fixture := &maintenanceTestFixture{
		UUID:      "mw_time_123",
		Name:      "Time Update Test",
		Title:     "Time Update Test",
		Text:      "Testing time updates",
		StartDate: startStr,
		EndDate:   endStr,
		Monitors:  []string{"mon_123"},
	}

	server, _ := newMaintenanceServerWithUpdateCapture(fixture)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			{
				Config: generateMaintenanceConfig(server.URL, fixture.Name, fixture.Title, fixture.Text, startStr, endStr, fixture.Monitors),
				Check:  tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "name", "Time Update Test"),
			},
			{
				Config: generateMaintenanceConfig(server.URL, fixture.Name, fixture.Title, fixture.Text, newStartStr, newEndStr, fixture.Monitors),
				Check:  tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "name", "Time Update Test"),
			},
		},
	})
}

func TestAccMaintenanceResource_disappears(t *testing.T) {
	startStr, endStr := generateMaintenanceTimeRange(24, 2)

	fixture := &maintenanceTestFixture{
		UUID:      "mw_disappear_123",
		Name:      "Disappearing Maintenance",
		Title:     "Disappearing Title",
		Text:      "This will disappear",
		StartDate: startStr,
		EndDate:   endStr,
		Monitors:  []string{"mon_123"},
	}

	server, deleted := newMaintenanceServerWithDisappear(fixture)
	defer server.Close()

	config := generateMaintenanceConfig(server.URL, fixture.Name, fixture.Title, fixture.Text, startStr, endStr, fixture.Monitors)

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			{
				Config: config,
				Check:  tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "id", "mw_disappear_123"),
			},
			{
				PreConfig: func() {
					*deleted = true
				},
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccMaintenanceResource_createError(t *testing.T) {
	startStr, endStr := generateMaintenanceTimeRange(24, 2)

	server := newErrorServer(http.StatusInternalServerError, "internal error")
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			{
				Config:      generateMaintenanceConfigMinimal(server.URL, "Error Test", startStr, endStr, []string{"mon_123"}),
				ExpectError: regexp.MustCompile("Failed to Create Maintenance Window"),
			},
		},
	})
}

func TestAccMaintenanceResource_readAfterCreateError(t *testing.T) {
	startStr, endStr := generateMaintenanceTimeRange(24, 2)

	server := newMaintenanceServerReadAfterCreateError("mw_read_error_123")
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			{
				Config:      generateMaintenanceConfigMinimal(server.URL, "Read After Create Error Test", startStr, endStr, []string{"mon_123"}),
				ExpectError: regexp.MustCompile("Maintenance Window Created But Read Failed"),
			},
		},
	})
}

func TestAccMaintenanceResource_invalidTimeRange(t *testing.T) {
	start, _ := generateMaintenanceTimeRange(24, 2)
	_, end := generateMaintenanceTimeRange(2, 2) // End before start

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			{
				Config:      generateMaintenanceConfigMinimal("http://localhost:9999", "Invalid Time", start, end, []string{"mon_123"}),
				ExpectError: regexp.MustCompile("end_date must be after start_date"),
			},
		},
	})
}

func TestAccMaintenanceResource_invalidTimeFormat(t *testing.T) {
	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			{
				Config: `
provider "hyperping" {
  api_key  = "hp_test_key"
  base_url = "http://localhost:9999"
}

resource "hyperping_maintenance" "test" {
  name       = "Invalid Time Format"
  start_date = "not-a-valid-time"
  end_date   = "also-not-valid"
  monitors   = ["mon_123"]
}
`,
				ExpectError: regexp.MustCompile("Invalid Start Date"),
			},
		},
	})
}

// Unit tests

func TestMaintenanceResource_ConfigureWrongType(t *testing.T) {
	r := &MaintenanceResource{}

	resp := &frameworkresource.ConfigureResponse{}
	r.Configure(context.Background(), frameworkresource.ConfigureRequest{
		ProviderData: "wrong type",
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong type")
	}
}

func TestMaintenanceResource_ConfigureNilProviderData(t *testing.T) {
	r := &MaintenanceResource{}

	resp := &frameworkresource.ConfigureResponse{}
	r.Configure(context.Background(), frameworkresource.ConfigureRequest{
		ProviderData: nil,
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}
}

func TestMaintenanceResource_ConfigureValidClient(t *testing.T) {
	r := &MaintenanceResource{}

	// Create a real client
	c := client.NewClient("test_api_key")

	resp := &frameworkresource.ConfigureResponse{}
	r.Configure(context.Background(), frameworkresource.ConfigureRequest{
		ProviderData: c,
	}, resp)

	// Should not error
	if resp.Diagnostics.HasError() {
		t.Errorf("Unexpected error: %v", resp.Diagnostics)
	}

	// Client should be set
	if r.client == nil {
		t.Error("Expected client to be set")
	}
}

func TestMaintenanceResource_Metadata(t *testing.T) {
	r := &MaintenanceResource{}

	resp := &frameworkresource.MetadataResponse{}
	r.Metadata(context.Background(), frameworkresource.MetadataRequest{
		ProviderTypeName: "hyperping",
	}, resp)

	if resp.TypeName != "hyperping_maintenance" {
		t.Errorf("expected type name 'hyperping_maintenance', got %s", resp.TypeName)
	}
}

func TestMaintenanceResource_Schema(t *testing.T) {
	r := &MaintenanceResource{}

	resp := &frameworkresource.SchemaResponse{}
	r.Schema(context.Background(), frameworkresource.SchemaRequest{}, resp)

	// Check required attributes
	requiredAttrs := []string{"name", "start_date", "end_date", "monitors"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute %s in schema", attr)
		}
	}

	// Check optional attributes
	optionalAttrs := []string{"title", "text", "status_pages", "notification_option", "notification_minutes"}
	for _, attr := range optionalAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute %s in schema", attr)
		}
	}

	// Check computed attribute
	if _, ok := resp.Schema.Attributes["id"]; !ok {
		t.Error("expected computed attribute id in schema")
	}
}

func TestMaintenanceResource_mapMaintenanceToModel(t *testing.T) {
	r := &MaintenanceResource{}

	t.Run("all fields populated", func(t *testing.T) {
		startDate := "2025-12-20T02:00:00.000Z"
		endDate := "2025-12-20T06:00:00.000Z"
		notifyMinutes := 60
		maintenance := &client.Maintenance{
			UUID:                "mw_123",
			Name:                "Test Maintenance",
			Title:               client.LocalizedText{En: "Test Title"},
			Text:                client.LocalizedText{En: "Test Description"},
			StartDate:           &startDate,
			EndDate:             &endDate,
			Monitors:            []string{"mon-1", "mon-2"},
			StatusPages:         []string{"sp-1"},
			NotificationOption:  "scheduled",
			NotificationMinutes: &notifyMinutes,
		}

		model := &MaintenanceResourceModel{}
		diags := &diag.Diagnostics{}
		r.mapMaintenanceToModel(maintenance, model, diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags)
		}
		if model.ID.ValueString() != "mw_123" {
			t.Errorf("expected ID 'mw_123', got %s", model.ID.ValueString())
		}
		if model.Name.ValueString() != "Test Maintenance" {
			t.Errorf("expected name 'Test Maintenance', got %s", model.Name.ValueString())
		}
		if model.Title.ValueString() != "Test Title" {
			t.Errorf("expected title 'Test Title', got %s", model.Title.ValueString())
		}
		if model.StartDate.IsNull() {
			t.Error("expected StartDate to be set")
		}
	})

	t.Run("null title and text", func(t *testing.T) {
		startDate := "2025-12-20T02:00:00.000Z"
		endDate := "2025-12-20T06:00:00.000Z"
		maintenance := &client.Maintenance{
			UUID:      "mw_456",
			Name:      "No Title",
			Title:     client.LocalizedText{}, // Empty
			Text:      client.LocalizedText{}, // Empty
			StartDate: &startDate,
			EndDate:   &endDate,
			Monitors:  []string{"mon-1"},
		}

		model := &MaintenanceResourceModel{}
		diags := &diag.Diagnostics{}
		r.mapMaintenanceToModel(maintenance, model, diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags)
		}
		if !model.Title.IsNull() {
			t.Error("expected Title to be null")
		}
		if !model.Text.IsNull() {
			t.Error("expected Text to be null")
		}
	})

	t.Run("empty monitors", func(t *testing.T) {
		startDate := "2025-12-20T02:00:00.000Z"
		endDate := "2025-12-20T06:00:00.000Z"
		maintenance := &client.Maintenance{
			UUID:      "mw_789",
			Name:      "Empty Monitors",
			StartDate: &startDate,
			EndDate:   &endDate,
			Monitors:  []string{}, // Empty slice
		}

		model := &MaintenanceResourceModel{}
		diags := &diag.Diagnostics{}
		r.mapMaintenanceToModel(maintenance, model, diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags)
		}
		if !model.Monitors.IsNull() {
			t.Error("expected Monitors to be null for empty slice")
		}
	})

	t.Run("null dates", func(t *testing.T) {
		maintenance := &client.Maintenance{
			UUID:      "mw_000",
			Name:      "Null Dates",
			StartDate: nil,
			EndDate:   nil,
			Monitors:  []string{"mon-1"},
		}

		model := &MaintenanceResourceModel{}
		diags := &diag.Diagnostics{}
		r.mapMaintenanceToModel(maintenance, model, diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags)
		}
		if !model.StartDate.IsNull() {
			t.Error("expected StartDate to be null")
		}
		if !model.EndDate.IsNull() {
			t.Error("expected EndDate to be null")
		}
	})
}

func TestAccMaintenanceResource_withMonitors(t *testing.T) {
	startStr, endStr := generateMaintenanceTimeRange(24, 2)

	fixture := &maintenanceTestFixture{
		UUID:      "mw_monitors_123",
		Name:      "Monitor Test",
		Title:     "Monitor Test Title",
		Text:      "Test with monitors",
		StartDate: startStr,
		EndDate:   endStr,
		Monitors:  []string{"mon_123", "mon_456"},
	}

	server := newSimpleMaintenanceServer(fixture)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			{
				Config: generateMaintenanceConfig(server.URL, fixture.Name, fixture.Title, fixture.Text, startStr, endStr, fixture.Monitors),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "name", "Monitor Test"),
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "monitors.#", "2"),
				),
			},
		},
	})
}

func TestAccMaintenanceResource_import(t *testing.T) {
	startStr, endStr := generateMaintenanceTimeRange(24, 2)

	fixture := &maintenanceTestFixture{
		UUID:      "mw_import_123",
		Name:      "Import Test",
		Title:     "Import Test Title",
		Text:      "Import test description",
		StartDate: startStr,
		EndDate:   endStr,
		Monitors:  []string{"mon_123"},
	}

	server := newSimpleMaintenanceServer(fixture)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			{
				Config: generateMaintenanceConfig(server.URL, fixture.Name, fixture.Title, fixture.Text, startStr, endStr, fixture.Monitors),
			},
			{
				ResourceName:            "hyperping_maintenance.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"notification_option", "notification_minutes"},
			},
		},
	})
}

func TestAccMaintenanceResource_deleteNotFound(t *testing.T) {
	startStr, endStr := generateMaintenanceTimeRange(24, 2)

	fixture := &maintenanceTestFixture{
		UUID:      "mw_delete_nf_123",
		Name:      "Delete Not Found Test",
		Title:     "Delete Test Title",
		Text:      "Delete not found test",
		StartDate: startStr,
		EndDate:   endStr,
		Monitors:  []string{"mon_123"},
	}

	server := newMaintenanceServerDeleteNotFound(fixture)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			{
				Config: generateMaintenanceConfig(server.URL, fixture.Name, fixture.Title, fixture.Text, startStr, endStr, fixture.Monitors),
			},
			{
				Config:  generateProviderOnlyConfig(server.URL),
				Destroy: true,
			},
		},
	})
}
