// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"hyperping": providerserver.NewProtocol6WithError(New("test")()),
}

func TestNew(t *testing.T) {
	// Basic test that provider factory works
	factory := New("test")
	p := factory()
	if p == nil {
		t.Error("expected provider to be created")
	}
}

func TestProvider_Metadata(t *testing.T) {
	p := &HyperpingProvider{version: "1.0.0"}
	resp := &provider.MetadataResponse{}
	p.Metadata(context.Background(), provider.MetadataRequest{}, resp)

	if resp.TypeName != "hyperping" {
		t.Errorf("expected TypeName 'hyperping', got %s", resp.TypeName)
	}
	if resp.Version != "1.0.0" {
		t.Errorf("expected Version '1.0.0', got %s", resp.Version)
	}
}

func TestProvider_Schema(t *testing.T) {
	p := &HyperpingProvider{}
	resp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, resp)

	if _, ok := resp.Schema.Attributes["api_key"]; !ok {
		t.Error("expected api_key attribute in schema")
	}
	if _, ok := resp.Schema.Attributes["base_url"]; !ok {
		t.Error("expected base_url attribute in schema")
	}
}

func TestProvider_Resources(t *testing.T) {
	p := &HyperpingProvider{}
	resources := p.Resources(context.Background())

	// Monitor, Incident, IncidentUpdate, Maintenance, Outage, Healthcheck, StatusPage, StatusPageSubscriber
	if len(resources) != 8 {
		t.Errorf("expected 8 resources, got %d", len(resources))
	}
}

func TestProvider_DataSources(t *testing.T) {
	p := &HyperpingProvider{}
	dataSources := p.DataSources(context.Background())

	// Monitor (single), Monitors (list), Incident (single), Incidents (list),
	// MaintenanceWindow (single), MaintenanceWindows (list), MonitorReport, MonitorReports (list),
	// Outage (single), Outages (list), Healthcheck (single), Healthchecks (list),
	// StatusPage (single), StatusPages (list), StatusPageSubscribers
	if len(dataSources) != 15 {
		t.Errorf("expected 15 data sources, got %d", len(dataSources))
	}
}

func TestAccProvider_MissingAPIKey(t *testing.T) {
	// Save and unset the env var to test missing API key scenario
	originalKey := os.Getenv("HYPERPING_API_KEY")
	os.Unsetenv("HYPERPING_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("HYPERPING_API_KEY", originalKey)
		}
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "hyperping" {
  base_url = %q
}

data "hyperping_monitors" "all" {}
`, server.URL),
				ExpectError: regexp.MustCompile("Missing Hyperping API Key"),
			},
		},
	})
}

func TestAccProvider_WithBaseURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "hyperping" {
  api_key  = "hp_test_key"
  base_url = %q
}

data "hyperping_monitors" "all" {}
`, server.URL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hyperping_monitors.all", "monitors.#", "0"),
				),
			},
		},
	})
}

func TestIsAllowedBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		want    bool
	}{
		// Valid HTTPS Hyperping domains
		{
			name:    "valid - api.hyperping.io with https",
			baseURL: "https://api.hyperping.io",
			want:    true,
		},
		{
			name:    "valid - hyperping.io root with https",
			baseURL: "https://hyperping.io",
			want:    true,
		},
		{
			name:    "valid - subdomain of hyperping.io with https",
			baseURL: "https://staging.hyperping.io",
			want:    true,
		},
		{
			name:    "valid - nested subdomain with https",
			baseURL: "https://api.staging.hyperping.io",
			want:    true,
		},
		{
			name:    "valid - with path",
			baseURL: "https://api.hyperping.io/v1",
			want:    true,
		},
		{
			name:    "valid - with query string",
			baseURL: "https://api.hyperping.io?foo=bar",
			want:    true,
		},
		{
			name:    "valid - with fragment",
			baseURL: "https://api.hyperping.io#section",
			want:    true,
		},
		{
			name:    "valid - mixed case domain",
			baseURL: "https://API.HyperPing.io",
			want:    true,
		},

		// Localhost exemptions (HTTPS not required)
		{
			name:    "valid - localhost without port",
			baseURL: "http://localhost",
			want:    true,
		},
		{
			name:    "valid - localhost with port",
			baseURL: "http://localhost:8080",
			want:    true,
		},
		{
			name:    "valid - 127.0.0.1 without port",
			baseURL: "http://127.0.0.1",
			want:    true,
		},
		{
			name:    "valid - 127.0.0.1 with port",
			baseURL: "http://127.0.0.1:8080",
			want:    true,
		},
		{
			name:    "invalid - IPv6 localhost not supported",
			baseURL: "http://[::1]",
			want:    false,
		},
		{
			name:    "valid - localhost with https",
			baseURL: "https://localhost",
			want:    true,
		},
		{
			name:    "valid - 127.0.0.1 with https",
			baseURL: "https://127.0.0.1",
			want:    true,
		},

		// SECURITY: HTTP without HTTPS (VULN-016)
		{
			name:    "invalid - http without localhost",
			baseURL: "http://api.hyperping.io",
			want:    false,
		},
		{
			name:    "invalid - http hyperping.io root",
			baseURL: "http://hyperping.io",
			want:    false,
		},
		{
			name:    "invalid - http with subdomain",
			baseURL: "http://staging.hyperping.io",
			want:    false,
		},

		// SECURITY: Non-Hyperping domains (SSRF prevention)
		{
			name:    "invalid - example.com with https",
			baseURL: "https://example.com",
			want:    false,
		},
		{
			name:    "invalid - api.example.com",
			baseURL: "https://api.example.com",
			want:    false,
		},
		{
			name:    "invalid - looks like hyperping but different TLD",
			baseURL: "https://hyperping.com",
			want:    false,
		},
		{
			name:    "invalid - prefix match but different domain",
			baseURL: "https://hyperpingio.com",
			want:    false,
		},
		{
			name:    "invalid - suffix match but different domain",
			baseURL: "https://fakelyhyperping.io",
			want:    false,
		},

		// Edge cases
		{
			name:    "invalid - empty string",
			baseURL: "",
			want:    false,
		},
		{
			name:    "invalid - just domain without protocol",
			baseURL: "api.hyperping.io",
			want:    false,
		},
		{
			name:    "invalid - ftp protocol",
			baseURL: "ftp://api.hyperping.io",
			want:    false,
		},
		{
			name:    "invalid - data URL",
			baseURL: "data:text/plain,test",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAllowedBaseURL(tt.baseURL)
			if got != tt.want {
				t.Errorf("isAllowedBaseURL(%q) = %v, want %v", tt.baseURL, got, tt.want)
			}
		})
	}
}

// Unit tests for factory functions
func TestNewMonitorResource(t *testing.T) {
	r := NewMonitorResource()
	if r == nil {
		t.Fatal("NewMonitorResource returned nil")
	}
	// Verify it returns the correct type
	if _, ok := r.(*MonitorResource); !ok {
		t.Errorf("expected *MonitorResource, got %T", r)
	}
}

func TestNewIncidentResource(t *testing.T) {
	r := NewIncidentResource()
	if r == nil {
		t.Fatal("NewIncidentResource returned nil")
	}
	if _, ok := r.(*IncidentResource); !ok {
		t.Errorf("expected *IncidentResource, got %T", r)
	}
}

func TestNewMaintenanceResource(t *testing.T) {
	r := NewMaintenanceResource()
	if r == nil {
		t.Fatal("NewMaintenanceResource returned nil")
	}
	if _, ok := r.(*MaintenanceResource); !ok {
		t.Errorf("expected *MaintenanceResource, got %T", r)
	}
}

func TestNewMonitorsDataSource(t *testing.T) {
	ds := NewMonitorsDataSource()
	if ds == nil {
		t.Fatal("NewMonitorsDataSource returned nil")
	}
	if _, ok := ds.(*MonitorsDataSource); !ok {
		t.Errorf("expected *MonitorsDataSource, got %T", ds)
	}
}

// Unit tests for Provider.Configure
func TestProvider_Configure_WithAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	// Test Configure is called via acceptance test pattern since it requires
	// proper Terraform framework integration
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "hyperping" {
  api_key  = "hp_test_key"
  base_url = %q
}

data "hyperping_monitors" "all" {}
`, server.URL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hyperping_monitors.all", "monitors.#", "0"),
				),
			},
		},
	})
}

func TestProvider_Configure_EnvVar(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authorization header is present
		auth := r.Header.Get(client.HeaderAuthorization)
		if auth != "Bearer hp_env_test_key" {
			t.Errorf("expected Bearer hp_env_test_key, got %s", auth)
		}
		w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	// Set environment variable
	t.Setenv("HYPERPING_API_KEY", "hp_env_test_key")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "hyperping" {
  base_url = %q
}

data "hyperping_monitors" "all" {}
`, server.URL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hyperping_monitors.all", "monitors.#", "0"),
				),
			},
		},
	})
}
