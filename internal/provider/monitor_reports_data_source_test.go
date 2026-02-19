// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestNewMonitorReportsDataSource(t *testing.T) {
	ds := NewMonitorReportsDataSource()
	if ds == nil {
		t.Fatal("NewMonitorReportsDataSource returned nil")
	}
	if _, ok := ds.(*MonitorReportsDataSource); !ok {
		t.Errorf("expected *MonitorReportsDataSource, got %T", ds)
	}
}

func TestMonitorReportsDataSource_Metadata(t *testing.T) {
	d := &MonitorReportsDataSource{}

	req := datasource.MetadataRequest{
		ProviderTypeName: "hyperping",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	if resp.TypeName != "hyperping_monitor_reports" {
		t.Errorf("Expected type name 'hyperping_monitor_reports', got '%s'", resp.TypeName)
	}
}

func TestMonitorReportsDataSource_Schema(t *testing.T) {
	d := &MonitorReportsDataSource{}

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	expectedAttrs := []string{"from", "to", "monitors"}
	for _, attr := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing '%s' attribute", attr)
		}
	}
}

func TestMonitorReportsDataSource_Configure(t *testing.T) {
	t.Run("valid client", func(t *testing.T) {
		d := &MonitorReportsDataSource{}
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
		d := &MonitorReportsDataSource{}

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
		d := &MonitorReportsDataSource{}

		req := datasource.ConfigureRequest{
			ProviderData: "wrong type",
		}
		resp := &datasource.ConfigureResponse{}

		d.Configure(context.Background(), req, resp)

		if !resp.Diagnostics.HasError() {
			t.Fatal("Expected error when provider data is wrong type")
		}

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
	})
}

func TestMonitorReportsDataSource_mapReportsToListItems(t *testing.T) {
	t.Run("single report with all fields", func(t *testing.T) {
		reports := []client.MonitorReport{
			{
				UUID:          "mon_test001",
				Name:          "Website",
				Protocol:      "http",
				SLA:           99.95,
				MTTR:          120,
				MTTRFormatted: "2min 0s",
				Outages: client.OutageStats{
					Count:                  1,
					TotalDowntime:          120,
					TotalDowntimeFormatted: "2min 0s",
					LongestOutage:          120,
					LongestOutageFormatted: "2min 0s",
				},
			},
		}

		items := mapReportsToListItems(reports)

		if len(items) != 1 {
			t.Fatalf("Expected 1 item, got %d", len(items))
		}

		item := items[0]
		if item.ID.ValueString() != "mon_test001" {
			t.Errorf("Expected ID 'mon_test001', got %s", item.ID.ValueString())
		}
		if item.Name.ValueString() != "Website" {
			t.Errorf("Expected Name 'Website', got %s", item.Name.ValueString())
		}
		if item.Protocol.ValueString() != "http" {
			t.Errorf("Expected Protocol 'http', got %s", item.Protocol.ValueString())
		}
		if item.SLA.ValueFloat64() != 99.95 {
			t.Errorf("Expected SLA 99.95, got %f", item.SLA.ValueFloat64())
		}
		if item.MTTR.ValueInt64() != 120 {
			t.Errorf("Expected MTTR 120, got %d", item.MTTR.ValueInt64())
		}
		if item.MTTRFormatted.ValueString() != "2min 0s" {
			t.Errorf("Expected MTTRFormatted '2min 0s', got %s", item.MTTRFormatted.ValueString())
		}
		if item.OutageCount.ValueInt64() != 1 {
			t.Errorf("Expected OutageCount 1, got %d", item.OutageCount.ValueInt64())
		}
		if item.TotalDowntime.ValueInt64() != 120 {
			t.Errorf("Expected TotalDowntime 120, got %d", item.TotalDowntime.ValueInt64())
		}
		if item.TotalDowntimeFormatted.ValueString() != "2min 0s" {
			t.Errorf("Expected TotalDowntimeFormatted '2min 0s', got %s", item.TotalDowntimeFormatted.ValueString())
		}
	})

	t.Run("empty reports slice", func(t *testing.T) {
		items := mapReportsToListItems([]client.MonitorReport{})
		if len(items) != 0 {
			t.Errorf("Expected 0 items, got %d", len(items))
		}
	})

	t.Run("zero MTTR for monitor with no outages", func(t *testing.T) {
		reports := []client.MonitorReport{
			{
				UUID:     "mon_no_outages",
				Name:     "Healthy Monitor",
				Protocol: "http",
				SLA:      100.0,
				MTTR:     0,
				Outages: client.OutageStats{
					Count:         0,
					TotalDowntime: 0,
				},
			},
		}

		items := mapReportsToListItems(reports)
		if len(items) != 1 {
			t.Fatalf("Expected 1 item, got %d", len(items))
		}
		if items[0].MTTR.ValueInt64() != 0 {
			t.Errorf("Expected MTTR 0, got %d", items[0].MTTR.ValueInt64())
		}
		if items[0].OutageCount.ValueInt64() != 0 {
			t.Errorf("Expected OutageCount 0, got %d", items[0].OutageCount.ValueInt64())
		}
	})
}

func TestMonitorReportsAttrTypes(t *testing.T) {
	attrTypes := monitorReportListItemAttrTypes()

	expectedKeys := []string{
		"id", "name", "protocol", "sla", "mttr", "mttr_formatted",
		"outage_count", "total_downtime", "total_downtime_formatted",
	}

	if len(attrTypes) != len(expectedKeys) {
		t.Errorf("Expected %d attribute types, got %d", len(expectedKeys), len(attrTypes))
	}

	for _, key := range expectedKeys {
		if _, ok := attrTypes[key]; !ok {
			t.Errorf("Missing expected attribute type: %s", key)
		}
	}
}

// Acceptance tests with mock HTTP server

func TestAccMonitorReportsDataSource_basic(t *testing.T) {
	server := newMockMonitorReportsServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitorReportsDataSourceConfig(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_monitor_reports.test", "monitors.#", "1"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor_reports.test", "monitors.0.id", "mon_test001"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor_reports.test", "monitors.0.name", "Website"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor_reports.test", "monitors.0.sla", "99.95"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor_reports.test", "monitors.0.outage_count", "1"),
				),
			},
		},
	})
}

func TestAccMonitorReportsDataSource_empty(t *testing.T) {
	server := newMockMonitorReportsServerEmpty(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitorReportsDataSourceConfig(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_monitor_reports.test", "monitors.#", "0"),
				),
			},
		},
	})
}

func testAccMonitorReportsDataSourceConfig(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

data "hyperping_monitor_reports" "test" {
}
`, baseURL)
}

func testAccMonitorReportsDataSourceConfig_withDateRange(baseURL, from, to string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

data "hyperping_monitor_reports" "ranged" {
  from = %[2]q
  to   = %[3]q
}
`, baseURL, from, to)
}

func TestAccMonitorReportsDataSource_withDateRange(t *testing.T) {
	var capturedFrom, capturedTo string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedFrom = r.URL.Query().Get("from")
		capturedTo = r.URL.Query().Get("to")
		w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)
		payload := map[string]interface{}{
			"period": map[string]string{
				"from": "2026-01-01T00:00:00Z",
				"to":   "2026-01-31T23:59:59Z",
			},
			"monitors": []map[string]interface{}{
				{
					"uuid":          "mon_range001",
					"name":          "API Check",
					"protocol":      "https",
					"sla":           99.95,
					"mttr":          300,
					"mttrFormatted": "5min",
					"outages": map[string]interface{}{
						"count":                  2,
						"totalDowntime":          600,
						"totalDowntimeFormatted": "10min",
						"longestOutage":          300,
						"longestOutageFormatted": "5min",
						"details":                []interface{}{},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(payload) //nolint:errcheck
	}))
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitorReportsDataSourceConfig_withDateRange(server.URL, "2026-01-01T00:00:00Z", "2026-01-31T23:59:59Z"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_monitor_reports.ranged", "monitors.#", "1"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor_reports.ranged", "monitors.0.name", "API Check"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor_reports.ranged", "monitors.0.sla", "99.95"),
					func(_ *terraform.State) error {
						if capturedFrom != "2026-01-01T00:00:00Z" {
							return fmt.Errorf("expected from=2026-01-01T00:00:00Z, got %q", capturedFrom)
						}
						if capturedTo != "2026-01-31T23:59:59Z" {
							return fmt.Errorf("expected to=2026-01-31T23:59:59Z, got %q", capturedTo)
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccMonitorReportsDataSource_noDateRange(t *testing.T) {
	var capturedFrom, capturedTo string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedFrom = r.URL.Query().Get("from")
		capturedTo = r.URL.Query().Get("to")
		w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)
		payload := map[string]interface{}{
			"period": map[string]string{
				"from": "",
				"to":   "",
			},
			"monitors": []map[string]interface{}{
				{
					"uuid":     "mon_ndr001",
					"name":     "Health Check",
					"protocol": "https",
					"sla":      100.0,
					"outages": map[string]interface{}{
						"count":                  0,
						"totalDowntime":          0,
						"totalDowntimeFormatted": "",
						"longestOutage":          0,
						"longestOutageFormatted": "",
						"details":                []interface{}{},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(payload) //nolint:errcheck
	}))
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitorReportsDataSourceConfig(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_monitor_reports.test", "monitors.#", "1"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor_reports.test", "monitors.0.name", "Health Check"),
					func(_ *terraform.State) error {
						if capturedFrom != "" {
							return fmt.Errorf("expected empty from param, got %q", capturedFrom)
						}
						if capturedTo != "" {
							return fmt.Errorf("expected empty to param, got %q", capturedTo)
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccMonitorReportsDataSource_outageFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)
		payload := map[string]interface{}{
			"period": map[string]string{
				"from": "2026-01-01T00:00:00Z",
				"to":   "2026-01-31T23:59:59Z",
			},
			"monitors": []map[string]interface{}{
				{
					"uuid":          "mon_outage001",
					"name":          "Outage Monitor",
					"protocol":      "http",
					"sla":           98.5,
					"mttr":          600,
					"mttrFormatted": "10min",
					"outages": map[string]interface{}{
						"count":                  3,
						"totalDowntime":          1800,
						"totalDowntimeFormatted": "30min",
						"longestOutage":          900,
						"longestOutageFormatted": "15min",
						"details":                []interface{}{},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(payload) //nolint:errcheck
	}))
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitorReportsDataSourceConfig(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_monitor_reports.test", "monitors.#", "1"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor_reports.test", "monitors.0.sla", "98.5"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor_reports.test", "monitors.0.outage_count", "3"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor_reports.test", "monitors.0.total_downtime", "1800"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor_reports.test", "monitors.0.total_downtime_formatted", "30min"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor_reports.test", "monitors.0.mttr", "600"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitor_reports.test", "monitors.0.mttr_formatted", "10min"),
				),
			},
		},
	})
}

func newMockMonitorReportsServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != client.ReportsBasePath {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "not found"}) //nolint:errcheck
			return
		}
		w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)
		payload := map[string]interface{}{
			"period": map[string]string{
				"from": "2025-01-01T00:00:00Z",
				"to":   "2025-01-08T00:00:00Z",
			},
			"monitors": []map[string]interface{}{
				{
					"uuid":          "mon_test001",
					"name":          "Website",
					"protocol":      "http",
					"sla":           99.95,
					"mttr":          120,
					"mttrFormatted": "2min 0s",
					"outages": map[string]interface{}{
						"count":                  1,
						"totalDowntime":          120,
						"totalDowntimeFormatted": "2min 0s",
						"longestOutage":          120,
						"longestOutageFormatted": "2min 0s",
						"details":                []interface{}{},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(payload) //nolint:errcheck
	}))
}

func newMockMonitorReportsServerEmpty(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != client.ReportsBasePath {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "not found"}) //nolint:errcheck
			return
		}
		w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)
		payload := map[string]interface{}{
			"period": map[string]string{
				"from": "2025-01-01T00:00:00Z",
				"to":   "2025-01-08T00:00:00Z",
			},
			"monitors": []interface{}{},
		}
		json.NewEncoder(w).Encode(payload) //nolint:errcheck
	}))
}
