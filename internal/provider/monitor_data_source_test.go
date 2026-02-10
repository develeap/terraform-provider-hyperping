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

func TestAccMonitorDataSource_basic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)

		if r.Method == "GET" && r.URL.Path == client.MonitorsBasePath+"/mon-test-123" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"uuid":                 "mon-test-123",
				"name":                 "Test Monitor",
				"url":                  "https://example.com",
				"protocol":             "http",
				"http_method":          "GET",
				"check_frequency":      60,
				"expected_status_code": "200",
				"follow_redirects":     true,
				"paused":               false,
				"regions":              []string{"london", "virginia"},
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}))
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

data "hyperping_monitor" "test" {
  id = "mon-test-123"
}
`, server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_monitor.test", "id", "mon-test-123"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor.test", "name", "Test Monitor"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor.test", "url", "https://example.com"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor.test", "protocol", "http"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor.test", "http_method", "GET"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor.test", "check_frequency", "60"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor.test", "expected_status_code", "200"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor.test", "follow_redirects", "true"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor.test", "paused", "false"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor.test", "regions.#", "2"),
				),
			},
		},
	})
}

func TestAccMonitorDataSource_withAllFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)

		if r.Method == "GET" && r.URL.Path == client.MonitorsBasePath+"/mon-full-123" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"uuid":                 "mon-full-123",
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
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}))
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

data "hyperping_monitor" "test" {
  id = "mon-full-123"
}
`, server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_monitor.test", "id", "mon-full-123"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor.test", "name", "Full Monitor"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor.test", "http_method", "POST"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor.test", "check_frequency", "300"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor.test", "expected_status_code", "201"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor.test", "follow_redirects", "false"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor.test", "paused", "true"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor.test", "regions.#", "3"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor.test", "request_headers.#", "2"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor.test", "request_body", `{"check":"health"}`),
				),
			},
		},
	})
}

func TestAccMonitorDataSource_notFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Monitor not found"})
	}))
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

data "hyperping_monitor" "test" {
  id = "non-existent-id"
}
`, server.URL),
				ExpectError: regexp.MustCompile(`Could not read monitor`),
			},
		},
	})
}

func TestAccMonitorDataSource_serverError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
	}))
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

data "hyperping_monitor" "test" {
  id = "mon-error-123"
}
`, server.URL),
				ExpectError: regexp.MustCompile(`Could not read monitor`),
			},
		},
	})
}

// Unit tests

func TestMonitorDataSource_Metadata(t *testing.T) {
	d := &MonitorDataSource{}

	req := datasource.MetadataRequest{
		ProviderTypeName: "hyperping",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	if resp.TypeName != "hyperping_monitor" {
		t.Errorf("Expected type name 'hyperping_monitor', got '%s'", resp.TypeName)
	}
}

func TestMonitorDataSource_Schema(t *testing.T) {
	d := &MonitorDataSource{}

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	// Verify id attribute is required
	if _, ok := resp.Schema.Attributes["id"]; !ok {
		t.Error("Schema missing 'id' attribute")
	}

	// Verify computed attributes exist
	computedAttrs := []string{"name", "url", "protocol", "http_method", "check_frequency", "regions",
		"request_headers", "request_body", "expected_status_code", "follow_redirects", "paused"}
	for _, attr := range computedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing '%s' attribute", attr)
		}
	}
}

func TestMonitorDataSource_ConfigureWrongType(t *testing.T) {
	d := &MonitorDataSource{}

	req := datasource.ConfigureRequest{
		ProviderData: "wrong type - should be *client.Client",
	}
	resp := &datasource.ConfigureResponse{}

	d.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("Expected error when provider data is wrong type")
	}
}

func TestMonitorDataSource_ConfigureNilProviderData(t *testing.T) {
	d := &MonitorDataSource{}

	req := datasource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &datasource.ConfigureResponse{}

	d.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Error("Expected no error when provider data is nil")
	}
}

func TestMonitorDataSource_ConfigureValidClient(t *testing.T) {
	d := &MonitorDataSource{}

	c := client.NewClient("test_api_key")

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
}

func TestMonitorDataSource_mapMonitorToDataSourceModel(t *testing.T) {
	d := &MonitorDataSource{}

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

		model := &MonitorDataSourceModel{}
		diags := &diag.Diagnostics{}
		d.mapMonitorToDataSourceModel(monitor, model, diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags)
		}
		if model.ID.ValueString() != "mon-123" {
			t.Errorf("expected ID 'mon-123', got %s", model.ID.ValueString())
		}
		if model.Name.ValueString() != "Test Monitor" {
			t.Errorf("expected name 'Test Monitor', got %s", model.Name.ValueString())
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

		model := &MonitorDataSourceModel{}
		diags := &diag.Diagnostics{}
		d.mapMonitorToDataSourceModel(monitor, model, diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags)
		}
		if !model.RequestBody.IsNull() {
			t.Error("expected RequestBody to be null for empty string")
		}
	})

	t.Run("empty collections", func(t *testing.T) {
		monitor := &client.Monitor{
			UUID:               "mon-789",
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

		model := &MonitorDataSourceModel{}
		diags := &diag.Diagnostics{}
		d.mapMonitorToDataSourceModel(monitor, model, diags)

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

func TestNewMonitorDataSource(t *testing.T) {
	ds := NewMonitorDataSource()
	if ds == nil {
		t.Fatal("NewMonitorDataSource returned nil")
	}
	if _, ok := ds.(*MonitorDataSource); !ok {
		t.Errorf("expected *MonitorDataSource, got %T", ds)
	}
}
