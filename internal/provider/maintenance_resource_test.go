// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestAccMaintenanceResource_basic(t *testing.T) {
	now := time.Now().UTC()
	start := now.Add(24 * time.Hour).Truncate(time.Second)
	end := start.Add(2 * time.Hour)
	startStr := start.Format(time.RFC3339)
	endStr := end.Format(time.RFC3339)

	maintenance := map[string]interface{}{
		"uuid":       "mw_test_123",
		"name":       "Test Maintenance",
		"title":      map[string]interface{}{"en": "Test Maintenance Title"},
		"text":       map[string]interface{}{"en": "Test maintenance description"},
		"start_date": startStr,
		"end_date":   endStr,
		"monitors":   []string{"mon_123"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/maintenance-windows":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(maintenance)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/maintenance-windows/mw_test_123":
			json.NewEncoder(w).Encode(maintenance)
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/maintenance-windows/mw_test_123":
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "hyperping" {
  api_key  = "hp_test_key"
  base_url = %q
}

resource "hyperping_maintenance" "test" {
  name       = "Test Maintenance"
  title      = "Test Maintenance Title"
  text       = "Test maintenance description"
  start_date = %q
  end_date   = %q
  monitors   = ["mon_123"]
}
`, server.URL, startStr, endStr),
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
	now := time.Now().UTC()
	start := now.Add(48 * time.Hour).Truncate(time.Second)
	end := start.Add(4 * time.Hour)
	startStr := start.Format(time.RFC3339)
	endStr := end.Format(time.RFC3339)

	maintenance := map[string]interface{}{
		"uuid":                "mw_full_123",
		"name":                "Full Maintenance",
		"title":               map[string]interface{}{"en": "Full Maintenance Title"},
		"text":                map[string]interface{}{"en": "Complete maintenance description"},
		"start_date":          startStr,
		"end_date":            endStr,
		"monitors":            []string{"mon_123", "mon_456"},
		"statuspages":         []string{"sp_main"},
		"notificationOption":  "scheduled",
		"notificationMinutes": 120,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/maintenance-windows":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(maintenance)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/maintenance-windows/mw_full_123":
			json.NewEncoder(w).Encode(maintenance)
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/maintenance-windows/mw_full_123":
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "hyperping" {
  api_key  = "hp_test_key"
  base_url = %q
}

resource "hyperping_maintenance" "test" {
  name                 = "Full Maintenance"
  title                = "Full Maintenance Title"
  text                 = "Complete maintenance description"
  start_date           = %q
  end_date             = %q
  monitors             = ["mon_123", "mon_456"]
  status_pages         = ["sp_main"]
  notification_option  = "scheduled"
  notification_minutes = 120
}
`, server.URL, startStr, endStr),
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
	now := time.Now().UTC()
	start := now.Add(24 * time.Hour).Truncate(time.Second)
	end := start.Add(2 * time.Hour)
	newStart := now.Add(48 * time.Hour).Truncate(time.Second)
	newEnd := newStart.Add(4 * time.Hour)
	startStr := start.Format(time.RFC3339)
	endStr := end.Format(time.RFC3339)

	maintenance := map[string]interface{}{
		"uuid":       "mw_time_123",
		"name":       "Time Update Test",
		"title":      map[string]interface{}{"en": "Time Update Test"},
		"text":       map[string]interface{}{"en": "Testing time updates"},
		"start_date": startStr,
		"end_date":   endStr,
		"monitors":   []string{"mon_123"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/maintenance-windows":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(maintenance)
		case r.Method == http.MethodPut && r.URL.Path == "/v1/maintenance-windows/mw_time_123":
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)
			if sd, ok := req["start_date"].(string); ok {
				maintenance["start_date"] = sd
			}
			if ed, ok := req["end_date"].(string); ok {
				maintenance["end_date"] = ed
			}
			json.NewEncoder(w).Encode(maintenance)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/maintenance-windows/mw_time_123":
			json.NewEncoder(w).Encode(maintenance)
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/maintenance-windows/mw_time_123":
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "hyperping" {
  api_key  = "hp_test_key"
  base_url = %q
}

resource "hyperping_maintenance" "test" {
  name       = "Time Update Test"
  title      = "Time Update Test"
  text       = "Testing time updates"
  start_date = %q
  end_date   = %q
  monitors   = ["mon_123"]
}
`, server.URL, startStr, endStr),
				Check: tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "name", "Time Update Test"),
			},
			{
				Config: fmt.Sprintf(`
provider "hyperping" {
  api_key  = "hp_test_key"
  base_url = %q
}

resource "hyperping_maintenance" "test" {
  name       = "Time Update Test"
  title      = "Time Update Test"
  text       = "Testing time updates"
  start_date = %q
  end_date   = %q
  monitors   = ["mon_123"]
}
`, server.URL, newStart.Format(time.RFC3339), newEnd.Format(time.RFC3339)),
				Check: tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "name", "Time Update Test"),
			},
		},
	})
}

func TestAccMaintenanceResource_disappears(t *testing.T) {
	now := time.Now().UTC()
	start := now.Add(24 * time.Hour).Truncate(time.Second)
	end := start.Add(2 * time.Hour)
	startStr := start.Format(time.RFC3339)
	endStr := end.Format(time.RFC3339)

	maintenance := map[string]interface{}{
		"uuid":       "mw_disappear_123",
		"name":       "Disappearing Maintenance",
		"title":      map[string]interface{}{"en": "Disappearing Title"},
		"text":       map[string]interface{}{"en": "This will disappear"},
		"start_date": startStr,
		"end_date":   endStr,
		"monitors":   []string{"mon_123"},
	}
	deleted := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/maintenance-windows":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(maintenance)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/maintenance-windows/mw_disappear_123":
			if deleted {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
				return
			}
			json.NewEncoder(w).Encode(maintenance)
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/maintenance-windows/mw_disappear_123":
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "hyperping" {
  api_key  = "hp_test_key"
  base_url = %q
}

resource "hyperping_maintenance" "test" {
  name       = "Disappearing Maintenance"
  title      = "Disappearing Title"
  text       = "This will disappear"
  start_date = %q
  end_date   = %q
  monitors   = ["mon_123"]
}
`, server.URL, startStr, endStr),
				Check: tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "id", "mw_disappear_123"),
			},
			{
				PreConfig: func() {
					deleted = true
				},
				Config: fmt.Sprintf(`
provider "hyperping" {
  api_key  = "hp_test_key"
  base_url = %q
}

resource "hyperping_maintenance" "test" {
  name       = "Disappearing Maintenance"
  title      = "Disappearing Title"
  text       = "This will disappear"
  start_date = %q
  end_date   = %q
  monitors   = ["mon_123"]
}
`, server.URL, startStr, endStr),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccMaintenanceResource_createError(t *testing.T) {
	start := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)
	end := start.Add(2 * time.Hour)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
	}))
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "hyperping" {
  api_key  = "hp_test_key"
  base_url = %q
}

resource "hyperping_maintenance" "test" {
  name       = "Error Test"
  start_date = %q
  end_date   = %q
  monitors   = ["mon_123"]
}
`, server.URL, start.Format(time.RFC3339), end.Format(time.RFC3339)),
				ExpectError: regexp.MustCompile("Error creating maintenance window"),
			},
		},
	})
}

func TestAccMaintenanceResource_invalidTimeRange(t *testing.T) {
	now := time.Now().UTC()
	start := now.Add(24 * time.Hour).Truncate(time.Second)
	end := now.Add(2 * time.Hour) // End before start

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "hyperping" {
  api_key  = "hp_test_key"
  base_url = "http://localhost:9999"
}

resource "hyperping_maintenance" "test" {
  name       = "Invalid Time"
  start_date = %q
  end_date   = %q
  monitors   = ["mon_123"]
}
`, start.Format(time.RFC3339), end.Format(time.RFC3339)),
				ExpectError: regexp.MustCompile("end_date must be after start_date"),
			},
		},
	})
}

func TestAccMaintenanceResource_invalidTimeFormat(t *testing.T) {
	tfresource.Test(t, tfresource.TestCase{
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
	now := time.Now().UTC()
	start := now.Add(24 * time.Hour).Truncate(time.Second)
	end := start.Add(2 * time.Hour)
	startStr := start.Format(time.RFC3339)
	endStr := end.Format(time.RFC3339)

	maintenance := map[string]interface{}{
		"uuid":       "mw_monitors_123",
		"name":       "Monitor Test",
		"title":      map[string]interface{}{"en": "Monitor Test Title"},
		"text":       map[string]interface{}{"en": "Test with monitors"},
		"start_date": startStr,
		"end_date":   endStr,
		"monitors":   []string{"mon_123", "mon_456"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/maintenance-windows":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(maintenance)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/maintenance-windows/mw_monitors_123":
			json.NewEncoder(w).Encode(maintenance)
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/maintenance-windows/mw_monitors_123":
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "hyperping" {
  api_key  = "hp_test_key"
  base_url = %q
}

resource "hyperping_maintenance" "test" {
  name       = "Monitor Test"
  title      = "Monitor Test Title"
  text       = "Test with monitors"
  start_date = %q
  end_date   = %q
  monitors   = ["mon_123", "mon_456"]
}
`, server.URL, startStr, endStr),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "name", "Monitor Test"),
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "monitors.#", "2"),
				),
			},
		},
	})
}

func TestAccMaintenanceResource_import(t *testing.T) {
	now := time.Now().UTC()
	start := now.Add(24 * time.Hour).Truncate(time.Second)
	end := start.Add(2 * time.Hour)
	startStr := start.Format(time.RFC3339)
	endStr := end.Format(time.RFC3339)

	maintenance := map[string]interface{}{
		"uuid":       "mw_import_123",
		"name":       "Import Test",
		"title":      map[string]interface{}{"en": "Import Test Title"},
		"text":       map[string]interface{}{"en": "Import test description"},
		"start_date": startStr,
		"end_date":   endStr,
		"monitors":   []string{"mon_123"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/maintenance-windows":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(maintenance)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/maintenance-windows/mw_import_123":
			json.NewEncoder(w).Encode(maintenance)
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/maintenance-windows/mw_import_123":
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "hyperping" {
  api_key  = "hp_test_key"
  base_url = %q
}

resource "hyperping_maintenance" "test" {
  name       = "Import Test"
  title      = "Import Test Title"
  text       = "Import test description"
  start_date = %q
  end_date   = %q
  monitors   = ["mon_123"]
}
`, server.URL, startStr, endStr),
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
	now := time.Now().UTC()
	start := now.Add(24 * time.Hour).Truncate(time.Second)
	end := start.Add(2 * time.Hour)
	startStr := start.Format(time.RFC3339)
	endStr := end.Format(time.RFC3339)

	maintenance := map[string]interface{}{
		"uuid":       "mw_delete_nf_123",
		"name":       "Delete Not Found Test",
		"title":      map[string]interface{}{"en": "Delete Test Title"},
		"text":       map[string]interface{}{"en": "Delete not found test"},
		"start_date": startStr,
		"end_date":   endStr,
		"monitors":   []string{"mon_123"},
	}
	deleted := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/maintenance-windows":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(maintenance)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/maintenance-windows/mw_delete_nf_123":
			if deleted {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
				return
			}
			json.NewEncoder(w).Encode(maintenance)
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/maintenance-windows/mw_delete_nf_123":
			// Return not found - should not error since resource is already gone
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
			deleted = true
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "hyperping" {
  api_key  = "hp_test_key"
  base_url = %q
}

resource "hyperping_maintenance" "test" {
  name       = "Delete Not Found Test"
  title      = "Delete Test Title"
  text       = "Delete not found test"
  start_date = %q
  end_date   = %q
  monitors   = ["mon_123"]
}
`, server.URL, startStr, endStr),
			},
			{
				Config: fmt.Sprintf(`
provider "hyperping" {
  api_key  = "hp_test_key"
  base_url = %q
}
`, server.URL),
				Destroy: true,
				// Should succeed even though delete returns 404
			},
		},
	})
}
