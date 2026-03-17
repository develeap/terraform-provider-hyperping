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

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// TestMonitorValidateConfig_SchemaConsistency verifies that all attributes
// accessed by ValidateConfig exist in the real MonitorResource schema with
// the expected types. This catches drift between the test fixtures and the
// real schema (e.g., a renamed or removed attribute).
func TestMonitorValidateConfig_SchemaConsistency(t *testing.T) {
	r := &MonitorResource{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)
	realSchema := schemaResp.Schema

	// Every attribute that ValidateConfig reads via path.Root must be present.
	type attrExpectation struct {
		name     string
		wantType string // Go type name from the schema attribute
	}

	expectations := []attrExpectation{
		{"protocol", "schema.StringAttribute"},
		{"url", "schema.StringAttribute"},
		{"http_method", "schema.StringAttribute"},
		{"expected_status_code", "schema.StringAttribute"},
		{"follow_redirects", "schema.BoolAttribute"},
		{"request_headers", "schema.ListNestedAttribute"},
		{"request_body", "schema.StringAttribute"},
		{"required_keyword", "schema.StringAttribute"},
		{"port", "schema.Int64Attribute"},
		{"dns_record_type", "schema.StringAttribute"},
		{"dns_nameserver", "schema.StringAttribute"},
		{"dns_expected_answer", "schema.StringAttribute"},
	}

	for _, exp := range expectations {
		attr, ok := realSchema.Attributes[exp.name]
		if !ok {
			t.Errorf("validation references attribute %q but it does not exist in the real schema", exp.name)
			continue
		}
		gotType := fmt.Sprintf("%T", attr)
		if gotType != exp.wantType {
			t.Errorf("attribute %q: expected type %s, got %s", exp.name, exp.wantType, gotType)
		}
	}
}

// monitorConfigBuilder constructs tftypes.Value objects for ValidateConfig unit tests.
// Each field mirrors a monitor schema attribute; nil means "leave null".
type monitorConfigBuilder struct {
	protocol          interface{} // string, nil (null), or tftypes.UnknownValue
	url               *string
	httpMethod        *string
	expectedStatus    *string
	followRedirects   *bool
	requestBody       *string
	requiredKeyword   *string
	port              *int64
	dnsRecordType     *string
	dnsNameserver     *string
	dnsExpectedAnswer *string
	requestHeaders    []map[string]string // nil = null, non-nil = set list
}

// buildConfigValue converts the builder into a tftypes.Value matching the monitor schema.
func (b *monitorConfigBuilder) buildConfigValue(s schema.Schema) tftypes.Value {
	attrTypes := make(map[string]tftypes.Type)
	for name, attr := range s.Attributes {
		attrTypes[name] = attr.GetType().TerraformType(context.Background())
	}
	objType := tftypes.Object{AttributeTypes: attrTypes}

	vals := make(map[string]tftypes.Value)
	for name, attrType := range attrTypes {
		vals[name] = tftypes.NewValue(attrType, nil) // null by default
	}

	// Protocol
	switch v := b.protocol.(type) {
	case string:
		vals["protocol"] = tftypes.NewValue(tftypes.String, v)
	case nil:
		vals["protocol"] = tftypes.NewValue(tftypes.String, nil)
	default:
		vals["protocol"] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	}

	// Required fields always need values.
	vals["name"] = tftypes.NewValue(tftypes.String, "test-monitor")

	if b.url != nil {
		vals["url"] = tftypes.NewValue(tftypes.String, *b.url)
	} else {
		vals["url"] = tftypes.NewValue(tftypes.String, "https://example.com")
	}

	setStringAttr(vals, "http_method", b.httpMethod)
	setStringAttr(vals, "expected_status_code", b.expectedStatus)
	setStringAttr(vals, "request_body", b.requestBody)
	setStringAttr(vals, "required_keyword", b.requiredKeyword)
	setStringAttr(vals, "dns_record_type", b.dnsRecordType)
	setStringAttr(vals, "dns_nameserver", b.dnsNameserver)
	setStringAttr(vals, "dns_expected_answer", b.dnsExpectedAnswer)

	if b.followRedirects != nil {
		vals["follow_redirects"] = tftypes.NewValue(tftypes.Bool, *b.followRedirects)
	}

	if b.port != nil {
		vals["port"] = tftypes.NewValue(tftypes.Number, *b.port)
	}

	if b.requestHeaders != nil {
		headerObjType := attrTypes["request_headers"].(tftypes.List).ElementType.(tftypes.Object)
		headerVals := make([]tftypes.Value, 0, len(b.requestHeaders))
		for _, h := range b.requestHeaders {
			headerVals = append(headerVals, tftypes.NewValue(headerObjType, map[string]tftypes.Value{
				"name":  tftypes.NewValue(tftypes.String, h["name"]),
				"value": tftypes.NewValue(tftypes.String, h["value"]),
			}))
		}
		vals["request_headers"] = tftypes.NewValue(attrTypes["request_headers"], headerVals)
	}

	return tftypes.NewValue(objType, vals)
}

func setStringAttr(vals map[string]tftypes.Value, name string, val *string) {
	if val != nil {
		vals[name] = tftypes.NewValue(tftypes.String, *val)
	}
}

// Helper to create *string from literal.
func strPtr(s string) *string { return &s }

// Helper to create *bool from literal.
func boolPtr(b bool) *bool { return &b }

// Helper to create *int64 from literal.
func int64Ptr(n int64) *int64 { return &n }

// runValidateConfig is a test helper that invokes ValidateConfig on MonitorResource.
func runValidateConfig(t *testing.T, b *monitorConfigBuilder) *resource.ValidateConfigResponse {
	t.Helper()

	r := &MonitorResource{}
	ctx := context.Background()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	config := tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    b.buildConfigValue(schemaResp.Schema),
	}

	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)

	return resp
}

// hasErrorOnPath returns true if diagnostics contain an error referencing the given attribute.
func hasErrorOnPath(resp *resource.ValidateConfigResponse, attrName string) bool {
	for _, d := range resp.Diagnostics.Errors() {
		if containsAttrRef(d.Detail(), attrName) {
			return true
		}
	}
	return false
}

func containsAttrRef(detail, attrName string) bool {
	for i := range detail {
		if i+len(attrName) <= len(detail) && detail[i:i+len(attrName)] == attrName {
			return true
		}
	}
	return false
}

func TestValidateConfig_HTTPProtocolValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		builder *monitorConfigBuilder
	}{
		{
			name: "http minimal config",
			builder: &monitorConfigBuilder{
				protocol: "http",
			},
		},
		{
			name: "http with all HTTP-only fields set",
			builder: &monitorConfigBuilder{
				protocol:        "http",
				httpMethod:      strPtr("POST"),
				expectedStatus:  strPtr("201"),
				followRedirects: boolPtr(false),
				requestBody:     strPtr(`{"check":"health"}`),
				requiredKeyword: strPtr("OK"),
				requestHeaders:  []map[string]string{{"name": "X-Test", "value": "val"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resp := runValidateConfig(t, tt.builder)
			if resp.Diagnostics.HasError() {
				t.Errorf("expected no errors, got: %v", resp.Diagnostics)
			}
		})
	}
}

func TestValidateConfig_HTTPProtocolRejectsPort(t *testing.T) {
	t.Parallel()

	resp := runValidateConfig(t, &monitorConfigBuilder{
		protocol: "http",
		port:     int64Ptr(443),
	})

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error when port is set for http protocol")
	}

	foundPortError := false
	for _, d := range resp.Diagnostics.Errors() {
		if containsAttrRef(d.Detail(), "port") {
			foundPortError = true
			break
		}
	}
	if !foundPortError {
		t.Error("expected error to reference 'port' attribute")
	}
}

func TestValidateConfig_ICMPProtocol(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		builder       *monitorConfigBuilder
		wantError     bool
		errorContains string
	}{
		{
			name: "icmp minimal valid config",
			builder: &monitorConfigBuilder{
				protocol: "icmp",
			},
			wantError: false,
		},
		{
			name: "icmp rejects http_method",
			builder: &monitorConfigBuilder{
				protocol:   "icmp",
				httpMethod: strPtr("POST"),
			},
			wantError:     true,
			errorContains: "http_method",
		},
		{
			name: "icmp rejects expected_status_code",
			builder: &monitorConfigBuilder{
				protocol:       "icmp",
				expectedStatus: strPtr("200"),
			},
			wantError:     true,
			errorContains: "expected_status_code",
		},
		{
			name: "icmp rejects follow_redirects",
			builder: &monitorConfigBuilder{
				protocol:        "icmp",
				followRedirects: boolPtr(false),
			},
			wantError:     true,
			errorContains: "follow_redirects",
		},
		{
			name: "icmp rejects request_headers",
			builder: &monitorConfigBuilder{
				protocol:       "icmp",
				requestHeaders: []map[string]string{{"name": "X-Test", "value": "val"}},
			},
			wantError:     true,
			errorContains: "request_headers",
		},
		{
			name: "icmp rejects request_body",
			builder: &monitorConfigBuilder{
				protocol:    "icmp",
				requestBody: strPtr("test"),
			},
			wantError:     true,
			errorContains: "request_body",
		},
		{
			name: "icmp rejects required_keyword",
			builder: &monitorConfigBuilder{
				protocol:        "icmp",
				requiredKeyword: strPtr("HEALTHY"),
			},
			wantError:     true,
			errorContains: "required_keyword",
		},
		{
			name: "icmp rejects port",
			builder: &monitorConfigBuilder{
				protocol: "icmp",
				port:     int64Ptr(443),
			},
			wantError:     true,
			errorContains: "port",
		},
		{
			name: "icmp rejects dns_record_type",
			builder: &monitorConfigBuilder{
				protocol:      "icmp",
				dnsRecordType: strPtr("A"),
			},
			wantError:     true,
			errorContains: "dns_record_type",
		},
		{
			name: "icmp rejects dns_nameserver",
			builder: &monitorConfigBuilder{
				protocol:      "icmp",
				dnsNameserver: strPtr("8.8.8.8"),
			},
			wantError:     true,
			errorContains: "dns_nameserver",
		},
		{
			name: "icmp rejects dns_expected_answer",
			builder: &monitorConfigBuilder{
				protocol:          "icmp",
				dnsExpectedAnswer: strPtr("1.2.3.4"),
			},
			wantError:     true,
			errorContains: "dns_expected_answer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resp := runValidateConfig(t, tt.builder)

			if tt.wantError && !resp.Diagnostics.HasError() {
				t.Fatalf("expected error containing %q, got none", tt.errorContains)
			}
			if !tt.wantError && resp.Diagnostics.HasError() {
				t.Fatalf("expected no error, got: %v", resp.Diagnostics)
			}
			if tt.wantError && tt.errorContains != "" {
				found := false
				for _, d := range resp.Diagnostics.Errors() {
					if containsAttrRef(d.Detail(), tt.errorContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error referencing %q, got: %v", tt.errorContains, resp.Diagnostics)
				}
			}
		})
	}
}

func TestValidateConfig_PortProtocol(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		builder       *monitorConfigBuilder
		wantError     bool
		errorContains string
	}{
		{
			name: "port valid config with port set",
			builder: &monitorConfigBuilder{
				protocol: "port",
				port:     int64Ptr(5432),
			},
			wantError: false,
		},
		{
			name: "port requires port field",
			builder: &monitorConfigBuilder{
				protocol: "port",
			},
			wantError:     true,
			errorContains: "port",
		},
		{
			name: "port rejects http_method",
			builder: &monitorConfigBuilder{
				protocol:   "port",
				port:       int64Ptr(5432),
				httpMethod: strPtr("POST"),
			},
			wantError:     true,
			errorContains: "http_method",
		},
		{
			name: "port rejects expected_status_code",
			builder: &monitorConfigBuilder{
				protocol:       "port",
				port:           int64Ptr(5432),
				expectedStatus: strPtr("200"),
			},
			wantError:     true,
			errorContains: "expected_status_code",
		},
		{
			name: "port rejects follow_redirects",
			builder: &monitorConfigBuilder{
				protocol:        "port",
				port:            int64Ptr(5432),
				followRedirects: boolPtr(true),
			},
			wantError:     true,
			errorContains: "follow_redirects",
		},
		{
			name: "port rejects request_body",
			builder: &monitorConfigBuilder{
				protocol:    "port",
				port:        int64Ptr(5432),
				requestBody: strPtr("body"),
			},
			wantError:     true,
			errorContains: "request_body",
		},
		{
			name: "port rejects required_keyword",
			builder: &monitorConfigBuilder{
				protocol:        "port",
				port:            int64Ptr(5432),
				requiredKeyword: strPtr("HEALTHY"),
			},
			wantError:     true,
			errorContains: "required_keyword",
		},
		{
			name: "port rejects request_headers",
			builder: &monitorConfigBuilder{
				protocol:       "port",
				port:           int64Ptr(5432),
				requestHeaders: []map[string]string{{"name": "X-Test", "value": "val"}},
			},
			wantError:     true,
			errorContains: "request_headers",
		},
		{
			name: "port rejects dns_record_type",
			builder: &monitorConfigBuilder{
				protocol:      "port",
				port:          int64Ptr(5432),
				dnsRecordType: strPtr("A"),
			},
			wantError:     true,
			errorContains: "dns_record_type",
		},
		{
			name: "port rejects dns_nameserver",
			builder: &monitorConfigBuilder{
				protocol:      "port",
				port:          int64Ptr(5432),
				dnsNameserver: strPtr("8.8.8.8"),
			},
			wantError:     true,
			errorContains: "dns_nameserver",
		},
		{
			name: "port rejects dns_expected_answer",
			builder: &monitorConfigBuilder{
				protocol:          "port",
				port:              int64Ptr(5432),
				dnsExpectedAnswer: strPtr("1.2.3.4"),
			},
			wantError:     true,
			errorContains: "dns_expected_answer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resp := runValidateConfig(t, tt.builder)

			if tt.wantError && !resp.Diagnostics.HasError() {
				t.Fatalf("expected error containing %q, got none", tt.errorContains)
			}
			if !tt.wantError && resp.Diagnostics.HasError() {
				t.Fatalf("expected no error, got: %v", resp.Diagnostics)
			}
			if tt.wantError && tt.errorContains != "" {
				found := false
				for _, d := range resp.Diagnostics.Errors() {
					if containsAttrRef(d.Detail(), tt.errorContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error referencing %q, got: %v", tt.errorContains, resp.Diagnostics)
				}
			}
		})
	}
}

func TestValidateConfig_DNSProtocol(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		builder       *monitorConfigBuilder
		wantError     bool
		errorContains string
	}{
		{
			name: "dns valid config with all DNS fields",
			builder: &monitorConfigBuilder{
				protocol:          "dns",
				url:               strPtr("example.com"),
				dnsRecordType:     strPtr("A"),
				dnsNameserver:     strPtr("8.8.8.8"),
				dnsExpectedAnswer: strPtr("93.184.216.34"),
			},
			wantError: false,
		},
		{
			name: "dns minimal config",
			builder: &monitorConfigBuilder{
				protocol: "dns",
				url:      strPtr("example.com"),
			},
			wantError: false,
		},
		{
			name: "dns rejects port",
			builder: &monitorConfigBuilder{
				protocol: "dns",
				url:      strPtr("example.com"),
				port:     int64Ptr(53),
			},
			wantError:     true,
			errorContains: "port",
		},
		{
			name: "dns rejects http_method",
			builder: &monitorConfigBuilder{
				protocol:   "dns",
				url:        strPtr("example.com"),
				httpMethod: strPtr("GET"),
			},
			wantError:     true,
			errorContains: "http_method",
		},
		{
			name: "dns rejects expected_status_code",
			builder: &monitorConfigBuilder{
				protocol:       "dns",
				url:            strPtr("example.com"),
				expectedStatus: strPtr("200"),
			},
			wantError:     true,
			errorContains: "expected_status_code",
		},
		{
			name: "dns rejects follow_redirects",
			builder: &monitorConfigBuilder{
				protocol:        "dns",
				url:             strPtr("example.com"),
				followRedirects: boolPtr(true),
			},
			wantError:     true,
			errorContains: "follow_redirects",
		},
		{
			name: "dns rejects request_body",
			builder: &monitorConfigBuilder{
				protocol:    "dns",
				url:         strPtr("example.com"),
				requestBody: strPtr("body"),
			},
			wantError:     true,
			errorContains: "request_body",
		},
		{
			name: "dns rejects required_keyword",
			builder: &monitorConfigBuilder{
				protocol:        "dns",
				url:             strPtr("example.com"),
				requiredKeyword: strPtr("HEALTHY"),
			},
			wantError:     true,
			errorContains: "required_keyword",
		},
		{
			name: "dns rejects request_headers",
			builder: &monitorConfigBuilder{
				protocol:       "dns",
				url:            strPtr("example.com"),
				requestHeaders: []map[string]string{{"name": "X-Test", "value": "val"}},
			},
			wantError:     true,
			errorContains: "request_headers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resp := runValidateConfig(t, tt.builder)

			if tt.wantError && !resp.Diagnostics.HasError() {
				t.Fatalf("expected error containing %q, got none", tt.errorContains)
			}
			if !tt.wantError && resp.Diagnostics.HasError() {
				t.Fatalf("expected no error, got: %v", resp.Diagnostics)
			}
			if tt.wantError && tt.errorContains != "" {
				found := false
				for _, d := range resp.Diagnostics.Errors() {
					if containsAttrRef(d.Detail(), tt.errorContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error referencing %q, got: %v", tt.errorContains, resp.Diagnostics)
				}
			}
		})
	}
}

func TestValidateConfig_ProtocolSkipsValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		protocol interface{}
	}{
		{
			name:     "unknown protocol skips validation",
			protocol: tftypes.UnknownValue,
		},
		{
			name:     "null protocol skips validation",
			protocol: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Set http_method which would normally error for non-HTTP protocols.
			// With unknown/null protocol, validation should be skipped entirely.
			resp := runValidateConfig(t, &monitorConfigBuilder{
				protocol:   tt.protocol,
				httpMethod: strPtr("POST"),
				port:       int64Ptr(8080),
			})

			if resp.Diagnostics.HasError() {
				t.Errorf("expected no errors when protocol is %v, got: %v", tt.name, resp.Diagnostics)
			}
		})
	}
}

func TestValidateConfig_URLValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		protocol  string
		url       string
		wantError bool
	}{
		{
			name:      "http valid https URL",
			protocol:  "http",
			url:       "https://example.com",
			wantError: false,
		},
		{
			name:      "http valid http URL",
			protocol:  "http",
			url:       "http://example.com/health",
			wantError: false,
		},
		{
			name:      "http invalid bare domain",
			protocol:  "http",
			url:       "example.com",
			wantError: true,
		},
		{
			name:      "http invalid ftp scheme",
			protocol:  "http",
			url:       "ftp://example.com",
			wantError: true,
		},
		{
			name:      "http invalid empty URL",
			protocol:  "http",
			url:       "",
			wantError: true,
		},
		{
			name:      "icmp valid https URL",
			protocol:  "icmp",
			url:       "https://example.com",
			wantError: false,
		},
		{
			name:      "icmp invalid bare domain",
			protocol:  "icmp",
			url:       "example.com",
			wantError: true,
		},
		{
			name:      "port valid https URL",
			protocol:  "port",
			url:       "https://example.com",
			wantError: false,
		},
		{
			name:      "port invalid bare domain",
			protocol:  "port",
			url:       "example.com",
			wantError: true,
		},
		{
			name:      "dns bare domain allowed (no URL validation)",
			protocol:  "dns",
			url:       "example.com",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder := &monitorConfigBuilder{
				protocol: tt.protocol,
				url:      strPtr(tt.url),
			}

			// Port protocol requires port to be set to avoid a separate error.
			if tt.protocol == "port" {
				builder.port = int64Ptr(5432)
			}

			resp := runValidateConfig(t, builder)

			hasURLError := false
			for _, d := range resp.Diagnostics.Errors() {
				if d.Summary() == "Invalid URL Format" {
					hasURLError = true
					break
				}
			}

			if tt.wantError && !hasURLError {
				t.Errorf("expected URL validation error, got none (diagnostics: %v)", resp.Diagnostics)
			}
			if !tt.wantError && hasURLError {
				t.Errorf("expected no URL validation error, got one (diagnostics: %v)", resp.Diagnostics)
			}
		})
	}
}

func TestValidateConfig_MultipleErrors(t *testing.T) {
	t.Parallel()

	// ICMP with multiple HTTP-only fields set should produce at least one error.
	// Note: the check helpers use resp.Diagnostics.HasError() as an early-return
	// guard, which means once the first field produces an error, subsequent checks
	// within the same validation function bail out. This test verifies that at
	// least the first invalid field is caught.
	resp := runValidateConfig(t, &monitorConfigBuilder{
		protocol:       "icmp",
		httpMethod:     strPtr("POST"),
		expectedStatus: strPtr("200"),
		port:           int64Ptr(443),
		dnsRecordType:  strPtr("A"),
	})

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected at least one error, got none")
	}

	// The first field checked by validateNonHTTPProtocol is http_method.
	if !hasErrorOnPath(resp, "http_method") {
		t.Errorf("expected error referencing 'http_method', got: %v", resp.Diagnostics)
	}
}

func TestValidateConfig_HTTPRejectsDNSFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		builder       *monitorConfigBuilder
		errorContains string
	}{
		{
			name: "http rejects dns_record_type",
			builder: &monitorConfigBuilder{
				protocol:      "http",
				dnsRecordType: strPtr("A"),
			},
			errorContains: "dns_record_type",
		},
		{
			name: "http rejects dns_nameserver",
			builder: &monitorConfigBuilder{
				protocol:      "http",
				dnsNameserver: strPtr("8.8.8.8"),
			},
			errorContains: "dns_nameserver",
		},
		{
			name: "http rejects dns_expected_answer",
			builder: &monitorConfigBuilder{
				protocol:          "http",
				dnsExpectedAnswer: strPtr("1.2.3.4"),
			},
			errorContains: "dns_expected_answer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resp := runValidateConfig(t, tt.builder)

			if !resp.Diagnostics.HasError() {
				t.Fatalf("expected error containing %q, got none", tt.errorContains)
			}

			found := false
			for _, d := range resp.Diagnostics.Errors() {
				if containsAttrRef(d.Detail(), tt.errorContains) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected error referencing %q, got: %v", tt.errorContains, resp.Diagnostics)
			}
		})
	}
}

func TestValidateConfig_ErrorMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		builder         *monitorConfigBuilder
		expectedSummary string
		detailSubstring string
	}{
		{
			name: "port missing produces Missing Required Attribute",
			builder: &monitorConfigBuilder{
				protocol: "port",
			},
			expectedSummary: "Missing Required Attribute",
			detailSubstring: "port is required",
		},
		{
			name: "http with port produces Invalid Attribute Combination",
			builder: &monitorConfigBuilder{
				protocol: "http",
				port:     int64Ptr(443),
			},
			expectedSummary: "Invalid Attribute Combination",
			detailSubstring: "port is not valid",
		},
		{
			name: "icmp http_method produces Invalid Attribute Combination",
			builder: &monitorConfigBuilder{
				protocol:   "icmp",
				httpMethod: strPtr("POST"),
			},
			expectedSummary: "Invalid Attribute Combination",
			detailSubstring: "http_method is only valid for HTTP monitors",
		},
		{
			name: "dns port produces Invalid Attribute Combination",
			builder: &monitorConfigBuilder{
				protocol: "dns",
				url:      strPtr("example.com"),
				port:     int64Ptr(53),
			},
			expectedSummary: "Invalid Attribute Combination",
			detailSubstring: "port is not valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resp := runValidateConfig(t, tt.builder)

			if !resp.Diagnostics.HasError() {
				t.Fatal("expected an error, got none")
			}

			foundSummary := false
			foundDetail := false
			for _, d := range resp.Diagnostics.Errors() {
				if d.Summary() == tt.expectedSummary {
					foundSummary = true
				}
				if containsAttrRef(d.Detail(), tt.detailSubstring) {
					foundDetail = true
				}
			}

			if !foundSummary {
				t.Errorf("expected error with summary %q, got: %v", tt.expectedSummary, resp.Diagnostics)
			}
			if !foundDetail {
				t.Errorf("expected error detail containing %q, got: %v", tt.detailSubstring, resp.Diagnostics)
			}
		})
	}
}

func TestValidateConfig_UnrecognizedProtocolNoValidation(t *testing.T) {
	t.Parallel()

	// A known string protocol that doesn't match any case in the switch
	// should fall through without errors (forward-compatible with new protocols).
	resp := runValidateConfig(t, &monitorConfigBuilder{
		protocol:   "tcp",
		httpMethod: strPtr("POST"),
		port:       int64Ptr(8080),
	})

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no errors for unrecognized protocol, got: %v", resp.Diagnostics)
	}
}

func TestValidateConfig_NullURL(t *testing.T) {
	t.Parallel()

	// When URL is null, validateURLIsHTTP should skip (not error).
	b := &monitorConfigBuilder{
		protocol: "http",
	}
	r := &MonitorResource{}
	ctx := context.Background()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	// Build config with URL explicitly null
	configValue := b.buildConfigValue(schemaResp.Schema)
	attrTypes := make(map[string]tftypes.Type)
	for name, attr := range schemaResp.Schema.Attributes {
		attrTypes[name] = attr.GetType().TerraformType(ctx)
	}

	// Override url to null
	vals := make(map[string]tftypes.Value)
	if err := configValue.As(&vals); err != nil {
		t.Fatalf("failed to extract values: %v", err)
	}
	vals["url"] = tftypes.NewValue(tftypes.String, nil)

	objType := tftypes.Object{AttributeTypes: attrTypes}
	config := tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    tftypes.NewValue(objType, vals),
	}

	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)

	// With null URL and HTTP protocol, validateURLIsHTTP should skip (null guard).
	// But validateHTTPProtocol and validateDNSFieldsNotSet should still run without error.
	if resp.Diagnostics.HasError() {
		t.Errorf("expected no URL validation error for null URL, got: %v", resp.Diagnostics)
	}
}

// newMinimalMockServer creates a lightweight mock HTTP server for acceptance tests
// that only need provider initialization (no full CRUD). Used by validation and
// data source tests across the package.
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
