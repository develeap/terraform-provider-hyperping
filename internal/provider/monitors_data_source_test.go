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

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestAccMonitorsDataSource_basic(t *testing.T) {
	server := newMockHyperpingServerForDataSource(t)
	defer server.Close()

	// Pre-create some monitors
	server.createTestMonitor("mon-1", "Monitor One")
	server.createTestMonitor("mon-2", "Monitor Two")

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitorsDataSourceConfig(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_monitors.all", "monitors.#", "2"),
					// Don't check specific order since map iteration is not deterministic
				),
			},
		},
	})
}

func TestAccMonitorsDataSource_empty(t *testing.T) {
	server := newMockHyperpingServerForDataSource(t)
	defer server.Close()

	// No monitors created - should return empty list

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitorsDataSourceConfig(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_monitors.all", "monitors.#", "0"),
				),
			},
		},
	})
}

func TestAccMonitorsDataSource_withAllFields(t *testing.T) {
	server := newMockHyperpingServerForDataSource(t)
	defer server.Close()

	// Create a monitor with all fields populated
	server.createFullMonitor()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitorsDataSourceConfig(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_monitors.all", "monitors.#", "1"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitors.all", "monitors.0.id", "full-monitor-id"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitors.all", "monitors.0.name", "Full Monitor"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitors.all", "monitors.0.url", "https://api.example.com/health"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitors.all", "monitors.0.protocol", "http"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitors.all", "monitors.0.http_method", "POST"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitors.all", "monitors.0.check_frequency", "300"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitors.all", "monitors.0.expected_status_code", "201"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitors.all", "monitors.0.follow_redirects", "false"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitors.all", "monitors.0.paused", "true"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitors.all", "monitors.0.regions.#", "3"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitors.all", "monitors.0.request_headers.#", "2"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitors.all", "monitors.0.request_body", `{"check":"health"}`),
				),
			},
		},
	})
}

func TestAccMonitorsDataSource_readError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
	}))
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccMonitorsDataSourceConfig(server.URL),
				ExpectError: regexp.MustCompile(`Could not list monitors`),
			},
		},
	})
}

func TestAccMonitorsDataSource_manyMonitors(t *testing.T) {
	server := newMockHyperpingServerForDataSource(t)
	defer server.Close()

	// Create 10 monitors
	for i := 1; i <= 10; i++ {
		server.createTestMonitor(fmt.Sprintf("mon-%d", i), fmt.Sprintf("Monitor %d", i))
	}

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitorsDataSourceConfig(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_monitors.all", "monitors.#", "10"),
				),
			},
		},
	})
}

// Unit tests for data source

func TestMonitorsDataSource_Metadata(t *testing.T) {
	d := &MonitorsDataSource{}

	req := datasource.MetadataRequest{
		ProviderTypeName: "hyperping",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	if resp.TypeName != "hyperping_monitors" {
		t.Errorf("Expected type name 'hyperping_monitors', got '%s'", resp.TypeName)
	}
}

func TestMonitorsDataSource_Schema(t *testing.T) {
	d := &MonitorsDataSource{}

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	// Verify monitors attribute exists
	if _, ok := resp.Schema.Attributes["monitors"]; !ok {
		t.Error("Schema missing 'monitors' attribute")
	}
}

func TestMonitorsDataSource_ConfigureWrongType(t *testing.T) {
	d := &MonitorsDataSource{}

	req := datasource.ConfigureRequest{
		ProviderData: "wrong type - should be *client.Client",
	}
	resp := &datasource.ConfigureResponse{}

	d.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("Expected error when provider data is wrong type")
	}

	// Verify error message contains expected text
	found := false
	for _, diag := range resp.Diagnostics.Errors() {
		if diag.Summary() == "Unexpected Data Source Configure Type" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'Unexpected Data Source Configure Type' error")
	}
}

func TestMonitorsDataSource_ConfigureNilProviderData(t *testing.T) {
	d := &MonitorsDataSource{}

	req := datasource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &datasource.ConfigureResponse{}

	d.Configure(context.Background(), req, resp)

	// Should not error, just return early
	if resp.Diagnostics.HasError() {
		t.Error("Expected no error when provider data is nil")
	}
}

func TestMonitorsDataSource_ConfigureValidClient(t *testing.T) {
	d := &MonitorsDataSource{}

	// Create a real client
	c := client.NewClient("test_api_key")

	req := datasource.ConfigureRequest{
		ProviderData: c,
	}
	resp := &datasource.ConfigureResponse{}

	d.Configure(context.Background(), req, resp)

	// Should not error
	if resp.Diagnostics.HasError() {
		t.Errorf("Unexpected error: %v", resp.Diagnostics)
	}

	// Client should be set
	if d.client == nil {
		t.Error("Expected client to be set")
	}
}

// Helper functions

func testAccMonitorsDataSourceConfig(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

data "hyperping_monitors" "all" {
}
`, baseURL)
}

// Mock server for data source tests

type mockHyperpingServerForDS struct {
	*httptest.Server
	t        *testing.T
	monitors map[string]map[string]interface{}
}

func newMockHyperpingServerForDataSource(t *testing.T) *mockHyperpingServerForDS {
	m := &mockHyperpingServerForDS{
		t:        t,
		monitors: make(map[string]map[string]interface{}),
	}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handleRequest(w, r)
	}))

	return m
}

func (m *mockHyperpingServerForDS) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)

	switch {
	case r.Method == "GET" && r.URL.Path == client.MonitorsBasePath:
		m.listMonitors(w)
	default:
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}
}

func (m *mockHyperpingServerForDS) listMonitors(w http.ResponseWriter) {
	monitors := make([]map[string]interface{}, 0, len(m.monitors))
	for _, monitor := range m.monitors {
		monitors = append(monitors, monitor)
	}
	json.NewEncoder(w).Encode(monitors)
}

func (m *mockHyperpingServerForDS) createTestMonitor(id, name string) {
	m.monitors[id] = map[string]interface{}{
		"uuid":                 id,
		"name":                 name,
		"url":                  "https://example.com",
		"protocol":             "http",
		"http_method":          "GET",
		"check_frequency":      60,
		"expected_status_code": "200",
		"follow_redirects":     true,
		"paused":               false,
		"regions":              []string{"london", "frankfurt"},
	}
}

func (m *mockHyperpingServerForDS) createFullMonitor() {
	m.monitors["full-monitor-id"] = map[string]interface{}{
		"uuid":                 "full-monitor-id",
		"name":                 "Full Monitor",
		"url":                  "https://api.example.com/health",
		"protocol":             "http",
		"http_method":          "POST",
		"check_frequency":      300,
		"expected_status_code": "201",
		"follow_redirects":     false,
		"paused":               true,
		"regions":              []string{"london", "virginia", "singapore"},
		"request_headers": []map[string]interface{}{
			{"name": "Content-Type", "value": "application/json"},
			{"name": "Authorization", "value": "Bearer token"},
		},
		"request_body": `{"check":"health"}`,
	}
}

// HCL config helpers for filter acceptance tests

func testAccMonitorsDataSourceConfig_withStatusFilter(serverURL, status string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "hp_test_key"
  base_url = %q
}

data "hyperping_monitors" "filtered" {
  filter = {
    status = %q
  }
}
`, serverURL, status)
}

func testAccMonitorsDataSourceConfig_withProjectUUIDFilter(serverURL, uuid string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "hp_test_key"
  base_url = %q
}

data "hyperping_monitors" "filtered" {
  filter = {
    project_uuid = %q
  }
}
`, serverURL, uuid)
}

func testAccMonitorsDataSourceConfig_withCombinedFilter(serverURL, status, uuid string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "hp_test_key"
  base_url = %q
}

data "hyperping_monitors" "filtered" {
  filter = {
    status       = %q
    project_uuid = %q
  }
}
`, serverURL, status, uuid)
}

func TestAccMonitorsDataSource_filterByStatus(t *testing.T) {
	server := newMockHyperpingServerForDataSource(t)
	defer server.Close()

	server.monitors["mon_1"] = map[string]interface{}{
		"uuid": "mon_1", "url": "https://example.com", "name": "Monitor 1",
		"protocol": "https", "status": "up", "projectUuid": "proj_abc",
		"paused": false, "check_frequency": 60,
	}
	server.monitors["mon_2"] = map[string]interface{}{
		"uuid": "mon_2", "url": "https://api.example.com", "name": "Monitor 2",
		"protocol": "https", "status": "up", "projectUuid": "proj_xyz",
		"paused": false, "check_frequency": 60,
	}
	server.monitors["mon_3"] = map[string]interface{}{
		"uuid": "mon_3", "url": "https://down.example.com", "name": "Monitor 3",
		"protocol": "https", "status": "down", "projectUuid": "proj_abc",
		"paused": false, "check_frequency": 60,
	}

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitorsDataSourceConfig_withStatusFilter(server.URL, "up"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_monitors.filtered", "monitors.#", "2"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitors.filtered", "monitors.0.status", "up"),
				),
			},
		},
	})
}

func TestAccMonitorsDataSource_filterByProjectUUID(t *testing.T) {
	server := newMockHyperpingServerForDataSource(t)
	defer server.Close()

	server.monitors["mon_1"] = map[string]interface{}{
		"uuid": "mon_1", "url": "https://example.com", "name": "Monitor 1",
		"protocol": "https", "status": "up", "projectUuid": "proj_abc",
		"paused": false, "check_frequency": 60,
	}
	server.monitors["mon_2"] = map[string]interface{}{
		"uuid": "mon_2", "url": "https://api.example.com", "name": "Monitor 2",
		"protocol": "https", "status": "up", "projectUuid": "proj_abc",
		"paused": false, "check_frequency": 60,
	}
	server.monitors["mon_3"] = map[string]interface{}{
		"uuid": "mon_3", "url": "https://other.example.com", "name": "Monitor 3",
		"protocol": "https", "status": "up", "projectUuid": "proj_xyz",
		"paused": false, "check_frequency": 60,
	}

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitorsDataSourceConfig_withProjectUUIDFilter(server.URL, "proj_abc"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_monitors.filtered", "monitors.#", "2"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitors.filtered", "monitors.0.project_uuid", "proj_abc"),
				),
			},
		},
	})
}

func TestAccMonitorsDataSource_filterByCombined(t *testing.T) {
	server := newMockHyperpingServerForDataSource(t)
	defer server.Close()

	server.monitors["mon_1"] = map[string]interface{}{
		"uuid": "mon_1", "url": "https://example.com", "name": "Monitor 1",
		"protocol": "https", "status": "up", "projectUuid": "proj_abc",
		"paused": false, "check_frequency": 60,
	}
	server.monitors["mon_2"] = map[string]interface{}{
		"uuid": "mon_2", "url": "https://api.example.com", "name": "Monitor 2",
		"protocol": "https", "status": "down", "projectUuid": "proj_abc",
		"paused": false, "check_frequency": 60,
	}
	server.monitors["mon_3"] = map[string]interface{}{
		"uuid": "mon_3", "url": "https://other.example.com", "name": "Monitor 3",
		"protocol": "https", "status": "up", "projectUuid": "proj_xyz",
		"paused": false, "check_frequency": 60,
	}

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitorsDataSourceConfig_withCombinedFilter(server.URL, "up", "proj_abc"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_monitors.filtered", "monitors.#", "1"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitors.filtered", "monitors.0.status", "up"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitors.filtered", "monitors.0.project_uuid", "proj_abc"),
				),
			},
		},
	})
}

func TestMonitorsDataSource_mapMonitorToDataModel(t *testing.T) {
	d := &MonitorsDataSource{}

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

		model := &MonitorDataModel{}
		diags := &diag.Diagnostics{}
		d.mapMonitorToDataModel(monitor, model, diags)

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

	t.Run("empty request body", func(t *testing.T) {
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

		model := &MonitorDataModel{}
		diags := &diag.Diagnostics{}
		d.mapMonitorToDataModel(monitor, model, diags)

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
			Regions:            []string{},
			RequestHeaders:     []client.RequestHeader{},
		}

		model := &MonitorDataModel{}
		diags := &diag.Diagnostics{}
		d.mapMonitorToDataModel(monitor, model, diags)

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
}
