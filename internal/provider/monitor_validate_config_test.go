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

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// T01: ICMP protocol rejects http_method
func TestAccMonitorResource_icmpRejectsHTTPMethod(t *testing.T) {
	server := newMinimalMockServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccMonitorConfig_icmpWithHTTPField(server.URL, `http_method = "POST"`),
				ExpectError: regexp.MustCompile(`(?i)http_method.*only valid.*http`),
			},
		},
	})
}

// T02: ICMP protocol rejects expected_status_code
func TestAccMonitorResource_icmpRejectsExpectedStatusCode(t *testing.T) {
	server := newMinimalMockServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccMonitorConfig_icmpWithHTTPField(server.URL, `expected_status_code = "200"`),
				ExpectError: regexp.MustCompile(`(?i)expected_status_code.*only valid.*http`),
			},
		},
	})
}

// T03: ICMP protocol rejects follow_redirects
func TestAccMonitorResource_icmpRejectsFollowRedirects(t *testing.T) {
	server := newMinimalMockServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccMonitorConfig_icmpWithHTTPField(server.URL, `follow_redirects = false`),
				ExpectError: regexp.MustCompile(`(?i)follow_redirects.*only.*http`),
			},
		},
	})
}

// T04: ICMP protocol rejects request_headers
func TestAccMonitorResource_icmpRejectsRequestHeaders(t *testing.T) {
	server := newMinimalMockServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccMonitorConfig_icmpWithRequestHeaders(server.URL),
				ExpectError: regexp.MustCompile(`(?i)request_headers.*only valid.*http`),
			},
		},
	})
}

// T05: ICMP protocol rejects request_body
func TestAccMonitorResource_icmpRejectsRequestBody(t *testing.T) {
	server := newMinimalMockServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccMonitorConfig_icmpWithHTTPField(server.URL, `request_body = "test"`),
				ExpectError: regexp.MustCompile(`(?i)request_body.*only valid.*http`),
			},
		},
	})
}

// T06: ICMP protocol rejects required_keyword
func TestAccMonitorResource_icmpRejectsRequiredKeyword(t *testing.T) {
	server := newMinimalMockServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccMonitorConfig_icmpWithHTTPField(server.URL, `required_keyword = "HEALTHY"`),
				ExpectError: regexp.MustCompile(`(?i)required_keyword.*only valid.*http`),
			},
		},
	})
}

// T07: ICMP protocol rejects port
func TestAccMonitorResource_icmpRejectsPort(t *testing.T) {
	server := newMinimalMockServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccMonitorConfig_icmpWithHTTPField(server.URL, `port = 443`),
				ExpectError: regexp.MustCompile(`(?i)port.*not valid.*icmp`),
			},
		},
	})
}

// T08: Port protocol rejects HTTP-only fields
func TestAccMonitorResource_portRejectsHTTPFields(t *testing.T) {
	server := newMinimalMockServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccMonitorConfig_portWithHTTPField(server.URL, `http_method = "POST"`),
				ExpectError: regexp.MustCompile(`(?i)http_method.*only valid.*http`),
			},
		},
	})
}

// T09: Port protocol requires port field
func TestAccMonitorResource_portRequiresPort(t *testing.T) {
	server := newMinimalMockServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccMonitorConfig_portWithoutPort(server.URL),
				ExpectError: regexp.MustCompile(`(?i)port.*required.*protocol.*port`),
			},
		},
	})
}

// T10: HTTP protocol rejects port field
func TestAccMonitorResource_httpRejectsPort(t *testing.T) {
	server := newMinimalMockServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccMonitorConfig_httpWithPort(server.URL),
				ExpectError: regexp.MustCompile(`(?i)port.*not valid.*http`),
			},
		},
	})
}

// T11: HTTP protocol accepts all HTTP fields without error
func TestAccMonitorResource_httpAcceptsAllHTTPFields(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitorConfig_httpFull(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "protocol", "http"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "http_method", "POST"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "expected_status_code", "201"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "follow_redirects", "false"),
					tfresource.TestCheckResourceAttrSet("hyperping_monitor.test", "id"),
				),
			},
		},
	})
}

// T12: Default values do not trigger cross-field errors for non-HTTP protocols
func TestAccMonitorResource_icmpDefaultsDoNotTriggerErrors(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitorConfig_icmpMinimal(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "protocol", "icmp"),
					tfresource.TestCheckResourceAttrSet("hyperping_monitor.test", "id"),
				),
			},
		},
	})
}

// T15: Port protocol rejects required_keyword
func TestAccMonitorResource_portRejectsRequiredKeyword(t *testing.T) {
	server := newMinimalMockServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccMonitorConfig_portWithHTTPField(server.URL, `required_keyword = "HEALTHY"`),
				ExpectError: regexp.MustCompile(`(?i)required_keyword.*only valid.*http`),
			},
		},
	})
}

// T16: Multiple invalid fields produce at least one error
func TestAccMonitorResource_multipleInvalidFields(t *testing.T) {
	server := newMinimalMockServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccMonitorConfig_icmpMultipleInvalid(server.URL),
				ExpectError: regexp.MustCompile(`(?i)(http_method|expected_status_code|port).*`),
			},
		},
	})
}

// T13: Unknown protocol skips validation (unit test)
func TestMonitorResource_ValidateConfig_UnknownProtocol(t *testing.T) {
	r := &MonitorResource{}
	ctx := context.Background()

	// Get the schema
	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	// Build a config where protocol is unknown and http_method is set
	config := tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    buildMonitorConfigValue(schemaResp.Schema, tftypes.UnknownValue, "POST"),
	}

	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Expected no errors when protocol is unknown, got: %v", resp.Diagnostics)
	}
}

// T14: Null protocol skips validation (unit test)
func TestMonitorResource_ValidateConfig_NullProtocol(t *testing.T) {
	r := &MonitorResource{}
	ctx := context.Background()

	// Get the schema
	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)

	// Build a config where protocol is null and http_method is set
	config := tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    buildMonitorConfigValue(schemaResp.Schema, nil, "POST"),
	}

	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Expected no errors when protocol is null, got: %v", resp.Diagnostics)
	}
}

// buildMonitorConfigValue constructs a tftypes.Value for the monitor schema.
// protocolVal: string for known, tftypes.UnknownValue for unknown, nil for null.
// httpMethod: set to a string value (or "" for null).
func buildMonitorConfigValue(s schema.Schema, protocolVal interface{}, httpMethod string) tftypes.Value {
	attrTypes := make(map[string]tftypes.Type)
	for name, attr := range s.Attributes {
		attrTypes[name] = attr.GetType().TerraformType(context.Background())
	}
	objType := tftypes.Object{AttributeTypes: attrTypes}

	vals := make(map[string]tftypes.Value)
	for name, attrType := range attrTypes {
		vals[name] = tftypes.NewValue(attrType, nil) // null by default
	}

	// Set protocol
	switch v := protocolVal.(type) {
	case string:
		vals["protocol"] = tftypes.NewValue(tftypes.String, v)
	case nil:
		vals["protocol"] = tftypes.NewValue(tftypes.String, nil)
	default:
		// tftypes.UnknownValue
		vals["protocol"] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	}

	// Set required fields with values
	vals["name"] = tftypes.NewValue(tftypes.String, "test")
	vals["url"] = tftypes.NewValue(tftypes.String, "https://example.com")

	// Set http_method if provided
	if httpMethod != "" {
		vals["http_method"] = tftypes.NewValue(tftypes.String, httpMethod)
	}

	return tftypes.NewValue(objType, vals)
}

// Helper: minimal mock server for validation tests (provider init only)
func newMinimalMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)
		switch {
		case r.Method == "GET" && r.URL.Path == client.MonitorsBasePath:
			json.NewEncoder(w).Encode([]interface{}{})
		case r.Method == "POST" && r.URL.Path == client.MonitorsBasePath:
			json.NewEncoder(w).Encode(map[string]interface{}{"uuid": "test-uuid"})
		case r.Method == "GET":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"uuid": "test-uuid", "name": "test", "url": "https://example.com",
				"protocol": "http", "http_method": "GET", "check_frequency": 60,
				"expected_status_code": "200", "follow_redirects": true, "paused": false,
				"regions": []string{"london"},
			})
		default:
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{})
		}
	}))
}

// Config helpers for cross-field validation tests

func testAccMonitorConfig_icmpWithHTTPField(baseURL, field string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name     = "Test ICMP Monitor"
  url      = "https://example.com"
  protocol = "icmp"
  %[2]s
}
`, baseURL, field)
}

func testAccMonitorConfig_icmpWithRequestHeaders(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name     = "Test ICMP Monitor"
  url      = "https://example.com"
  protocol = "icmp"
  request_headers = [
    {
      name  = "X-Test"
      value = "val"
    }
  ]
}
`, baseURL)
}

func testAccMonitorConfig_portWithHTTPField(baseURL, field string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name     = "Test Port Monitor"
  url      = "https://example.com"
  protocol = "port"
  port     = 5432
  %[2]s
}
`, baseURL, field)
}

func testAccMonitorConfig_portWithoutPort(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name     = "Test Port Monitor"
  url      = "https://example.com"
  protocol = "port"
}
`, baseURL)
}

func testAccMonitorConfig_httpWithPort(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name     = "Test HTTP Monitor"
  url      = "https://example.com"
  protocol = "http"
  port     = 443
}
`, baseURL)
}

func testAccMonitorConfig_httpFull(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name                 = "Test HTTP Full Monitor"
  url                  = "https://example.com/health"
  protocol             = "http"
  http_method          = "POST"
  expected_status_code = "201"
  follow_redirects     = false
  request_body         = "{\"check\":\"health\"}"
  required_keyword     = "OK"
  request_headers = [
    {
      name  = "Content-Type"
      value = "application/json"
    }
  ]
}
`, baseURL)
}

func testAccMonitorConfig_icmpMinimal(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name     = "Test ICMP Minimal"
  url      = "https://example.com"
  protocol = "icmp"
}
`, baseURL)
}

func testAccMonitorConfig_icmpMultipleInvalid(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name                = "Test ICMP Monitor"
  url                 = "https://example.com"
  protocol            = "icmp"
  http_method         = "POST"
  expected_status_code = "200"
  port                = 443
}
`, baseURL)
}
