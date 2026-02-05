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
	// MaintenanceWindow (single), MaintenanceWindows (list), MonitorReport,
	// Outage (single), Outages (list), Healthcheck (single), Healthchecks (list),
	// StatusPage (single), StatusPages (list), StatusPageSubscribers
	if len(dataSources) != 14 {
		t.Errorf("expected 14 data sources, got %d", len(dataSources))
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

// Unit tests for TFLogAdapter
func TestNewTFLogAdapter(t *testing.T) {
	adapter := NewTFLogAdapter()
	if adapter == nil {
		t.Fatal("NewTFLogAdapter returned nil")
	}
}

func TestTFLogAdapter_Debug(t *testing.T) {
	adapter := NewTFLogAdapter()
	// The Debug method calls tflog.Debug which requires a proper context
	// In unit tests without Terraform plugin framework context, this will be a no-op
	// but we still verify it doesn't panic
	ctx := context.Background()
	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}
	// Should not panic
	adapter.Debug(ctx, "test message", fields)
	adapter.Debug(ctx, "message without fields", nil)
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
