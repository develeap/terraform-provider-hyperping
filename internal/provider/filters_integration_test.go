// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// mockFilterServer is a multi-endpoint mock server used across filter integration tests.
type mockFilterServer struct {
	*httptest.Server
	t                  *testing.T
	monitors           []map[string]interface{}
	incidents          []map[string]interface{}
	maintenanceWindows []map[string]interface{}
	healthchecks       []map[string]interface{}
	outages            []map[string]interface{}
	statusPages        []map[string]interface{}
}

func newMockFilterServer(t *testing.T) *mockFilterServer {
	t.Helper()

	m := &mockFilterServer{t: t}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handleRequest(w, r)
	}))

	return m
}

func (m *mockFilterServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)

	switch {
	case r.Method == http.MethodGet && r.URL.Path == client.MonitorsBasePath:
		json.NewEncoder(w).Encode(m.monitors)
	case r.Method == http.MethodGet && r.URL.Path == client.IncidentsBasePath:
		json.NewEncoder(w).Encode(m.incidents)
	case r.Method == http.MethodGet && r.URL.Path == client.MaintenanceBasePath:
		json.NewEncoder(w).Encode(m.maintenanceWindows)
	case r.Method == http.MethodGet && r.URL.Path == client.HealthchecksBasePath:
		json.NewEncoder(w).Encode(m.healthchecks)
	case r.Method == http.MethodGet && r.URL.Path == client.OutagesBasePath:
		json.NewEncoder(w).Encode(m.outages)
	case r.Method == http.MethodGet && r.URL.Path == client.StatuspagesBasePath:
		m.listStatusPages(w)
	default:
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}
}

func (m *mockFilterServer) listStatusPages(w http.ResponseWriter) {
	if m.statusPages == nil {
		m.statusPages = []map[string]interface{}{}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"statuspages":   m.statusPages,
		"total":         len(m.statusPages),
		"has_next_page": false,
	})
}

// filterProviderConfig returns a HCL provider block pointing to the mock server.
func filterProviderConfig(serverURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "hp_test"
  base_url = %q
}
`, serverURL)
}

// TestAccDataSourceFilters_multipleFiltersMonitors tests combining multiple filter types on monitors.
func TestAccDataSourceFilters_multipleFiltersMonitors(t *testing.T) {
	server := newMockFilterServer(t)
	defer server.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	server.monitors = []map[string]interface{}{
		{
			"uuid": "mon-prod-1", "name": "[PROD]-API", "url": "https://example.com",
			"protocol": "https", "http_method": "GET", "check_frequency": 60,
			"expected_status_code": "200", "follow_redirects": true, "paused": false,
			"regions": []string{"london"}, "createdAt": now,
		},
		{
			"uuid": "mon-prod-2", "name": "[PROD]-Web", "url": "https://example2.com",
			"protocol": "https", "http_method": "GET", "check_frequency": 60,
			"expected_status_code": "200", "follow_redirects": true, "paused": false,
			"regions": []string{"london"}, "createdAt": now,
		},
		{
			// Should be filtered out: wrong protocol
			"uuid": "mon-dev-1", "name": "[PROD]-Dev", "url": "http://dev.example.com",
			"protocol": "http", "http_method": "GET", "check_frequency": 60,
			"expected_status_code": "200", "follow_redirects": true, "paused": false,
			"regions": []string{"london"}, "createdAt": now,
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: filterProviderConfig(server.URL) + `
data "hyperping_monitors" "filtered" {
  filter = {
    name_regex = "\\[PROD\\]-.*"
    protocol   = "https"
    paused     = false
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.filtered", "monitors.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_multipleFiltersIncidents tests combining multiple filter types on incidents.
func TestAccDataSourceFilters_multipleFiltersIncidents(t *testing.T) {
	server := newMockFilterServer(t)
	defer server.Close()

	server.incidents = []map[string]interface{}{
		{
			"uuid":        "inci-1",
			"title":       map[string]interface{}{"en": "API Outage"},
			"text":        map[string]interface{}{"en": "API is down"},
			"type":        "outage",
			"statuspages": []string{},
			"date":        "2026-01-01T00:00:00Z",
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: filterProviderConfig(server.URL) + `
data "hyperping_incidents" "filtered" {
  filter = {
    name_regex = "API.*"
    status     = "outage"
    severity   = "major"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_incidents.filtered", "incidents.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_emptyFilter tests that empty filter returns all resources.
func TestAccDataSourceFilters_emptyFilter(t *testing.T) {
	server := newMockFilterServer(t)
	defer server.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	server.monitors = []map[string]interface{}{
		{
			"uuid": "mon-1", "name": "Monitor One", "url": "https://example.com",
			"protocol": "http", "http_method": "GET", "check_frequency": 60,
			"expected_status_code": "200", "follow_redirects": true, "paused": false,
			"regions": []string{"london"}, "createdAt": now,
		},
		{
			"uuid": "mon-2", "name": "Monitor Two", "url": "https://example2.com",
			"protocol": "https", "http_method": "GET", "check_frequency": 60,
			"expected_status_code": "200", "follow_redirects": true, "paused": false,
			"regions": []string{"london"}, "createdAt": now,
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: filterProviderConfig(server.URL) + `
data "hyperping_monitors" "all" {
  filter = {}
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.all", "monitors.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_noMatches tests filter with no matches returns empty list.
func TestAccDataSourceFilters_noMatches(t *testing.T) {
	server := newMockFilterServer(t)
	defer server.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	server.monitors = []map[string]interface{}{
		{
			"uuid": "mon-1", "name": "Monitor One", "url": "https://example.com",
			"protocol": "http", "http_method": "GET", "check_frequency": 60,
			"expected_status_code": "200", "follow_redirects": true, "paused": false,
			"regions": []string{"london"}, "createdAt": now,
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: filterProviderConfig(server.URL) + `
data "hyperping_monitors" "no_matches" {
  filter = {
    name_regex = "^THIS_PATTERN_SHOULD_NOT_MATCH_ANYTHING_12345$"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.hyperping_monitors.no_matches", "monitors.#", "0"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_invalidRegex tests that invalid regex returns error.
func TestAccDataSourceFilters_invalidRegex(t *testing.T) {
	server := newMockFilterServer(t)
	defer server.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	server.monitors = []map[string]interface{}{
		{
			"uuid": "mon-1", "name": "Monitor One", "url": "https://example.com",
			"protocol": "http", "http_method": "GET", "check_frequency": 60,
			"expected_status_code": "200", "follow_redirects": true, "paused": false,
			"regions": []string{"london"}, "createdAt": now,
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: filterProviderConfig(server.URL) + `
data "hyperping_monitors" "invalid" {
  filter = {
    name_regex = "[invalid(regex"
  }
}
`,
				ExpectError: regexp.MustCompile("Invalid filter regex|Invalid name_regex|Failed to compile name_regex pattern|error parsing regexp"),
			},
		},
	})
}

// TestAccDataSourceFilters_crossDataSource tests filtering across related data sources.
func TestAccDataSourceFilters_crossDataSource(t *testing.T) {
	server := newMockFilterServer(t)
	defer server.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	server.monitors = []map[string]interface{}{
		{
			"uuid": "mon-prod-1", "name": "[PROD]-API", "url": "https://example.com",
			"protocol": "https", "http_method": "GET", "check_frequency": 60,
			"expected_status_code": "200", "follow_redirects": true, "paused": false,
			"regions": []string{"london"}, "createdAt": now,
		},
	}
	server.incidents = []map[string]interface{}{
		{
			"uuid":        "inci-major-1",
			"title":       map[string]interface{}{"en": "Critical Incident"},
			"text":        map[string]interface{}{"en": "Something is down"},
			"type":        "outage",
			"statuspages": []string{},
			"date":        "2026-01-01T00:00:00Z",
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: filterProviderConfig(server.URL) + `
data "hyperping_monitors" "production" {
  filter = {
    name_regex = "\\[PROD\\]-.*"
  }
}

data "hyperping_incidents" "production_incidents" {
  filter = {
    severity = "major"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.production", "monitors.#"),
					resource.TestCheckResourceAttrSet("data.hyperping_incidents.production_incidents", "incidents.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_regexCaseSensitive tests regex is case-sensitive.
func TestAccDataSourceFilters_regexCaseSensitive(t *testing.T) {
	server := newMockFilterServer(t)
	defer server.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	server.monitors = []map[string]interface{}{
		{
			"uuid": "mon-prod-1", "name": "[PROD]-API", "url": "https://example.com",
			"protocol": "https", "http_method": "GET", "check_frequency": 60,
			"expected_status_code": "200", "follow_redirects": true, "paused": false,
			"regions": []string{"london"}, "createdAt": now,
		},
		{
			"uuid": "mon-prod-lower", "name": "[prod]-API", "url": "https://example2.com",
			"protocol": "https", "http_method": "GET", "check_frequency": 60,
			"expected_status_code": "200", "follow_redirects": true, "paused": false,
			"regions": []string{"london"}, "createdAt": now,
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: filterProviderConfig(server.URL) + `
data "hyperping_monitors" "uppercase" {
  filter = {
    name_regex = "\\[PROD\\]-.*"
  }
}

data "hyperping_monitors" "lowercase" {
  filter = {
    name_regex = "\\[prod\\]-.*"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.uppercase", "monitors.#"),
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.lowercase", "monitors.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_booleanFilter tests boolean filter behavior.
func TestAccDataSourceFilters_booleanFilter(t *testing.T) {
	server := newMockFilterServer(t)
	defer server.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	server.monitors = []map[string]interface{}{
		{
			"uuid": "mon-active", "name": "Active Monitor", "url": "https://example.com",
			"protocol": "https", "http_method": "GET", "check_frequency": 60,
			"expected_status_code": "200", "follow_redirects": true, "paused": false,
			"regions": []string{"london"}, "createdAt": now,
		},
		{
			"uuid": "mon-paused", "name": "Paused Monitor", "url": "https://paused.example.com",
			"protocol": "https", "http_method": "GET", "check_frequency": 60,
			"expected_status_code": "200", "follow_redirects": true, "paused": true,
			"regions": []string{"london"}, "createdAt": now,
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: filterProviderConfig(server.URL) + `
data "hyperping_monitors" "active" {
  filter = {
    paused = false
  }
}

data "hyperping_monitors" "paused" {
  filter = {
    paused = true
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.active", "monitors.#"),
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.paused", "monitors.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_exactMatchFilter tests exact string matching.
func TestAccDataSourceFilters_exactMatchFilter(t *testing.T) {
	server := newMockFilterServer(t)
	defer server.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	server.monitors = []map[string]interface{}{
		{
			"uuid": "mon-https-1", "name": "HTTPS Monitor", "url": "https://example.com",
			"protocol": "https", "http_method": "GET", "check_frequency": 60,
			"expected_status_code": "200", "follow_redirects": true, "paused": false,
			"regions": []string{"london"}, "createdAt": now,
		},
		{
			"uuid": "mon-http-1", "name": "HTTP Monitor", "url": "http://example.com",
			"protocol": "http", "http_method": "GET", "check_frequency": 60,
			"expected_status_code": "200", "follow_redirects": true, "paused": false,
			"regions": []string{"london"}, "createdAt": now,
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: filterProviderConfig(server.URL) + `
data "hyperping_monitors" "https_only" {
  filter = {
    protocol = "https"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.https_only", "monitors.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_maintenanceWindows tests maintenance window filtering.
func TestAccDataSourceFilters_maintenanceWindows(t *testing.T) {
	server := newMockFilterServer(t)
	defer server.Close()

	future := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	futureEnd := time.Now().UTC().Add(26 * time.Hour).Format(time.RFC3339)

	server.maintenanceWindows = []map[string]interface{}{
		{
			"uuid":       "mw-1",
			"name":       "Scheduled Maintenance",
			"title":      map[string]interface{}{"en": "Scheduled Maintenance"},
			"text":       map[string]interface{}{"en": "Planned downtime"},
			"start_date": future,
			"end_date":   futureEnd,
			"monitors":   []string{},
			"status":     "scheduled",
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: filterProviderConfig(server.URL) + `
data "hyperping_maintenance_windows" "scheduled" {
  filter = {
    status = "scheduled"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_maintenance_windows.scheduled", "maintenance_windows.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_healthchecks tests healthcheck filtering.
func TestAccDataSourceFilters_healthchecks(t *testing.T) {
	server := newMockFilterServer(t)
	defer server.Close()

	server.healthchecks = []map[string]interface{}{
		{
			"uuid":             "hc-health-1",
			"name":             "API health check",
			"periodValue":      60,
			"periodType":       "seconds",
			"period":           60,
			"gracePeriodValue": 300,
			"gracePeriodType":  "seconds",
			"gracePeriod":      300,
			"isPaused":         false,
			"isDown":           false,
			"pingUrl":          "https://ping.hyperping.io/hc-health-1",
			"createdAt":        "2026-01-01T00:00:00Z",
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: filterProviderConfig(server.URL) + `
data "hyperping_healthchecks" "filtered" {
  filter = {
    name_regex = ".*health.*"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_healthchecks.filtered", "healthchecks.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_outages tests outage filtering.
func TestAccDataSourceFilters_outages(t *testing.T) {
	server := newMockFilterServer(t)
	defer server.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	server.outages = []map[string]interface{}{
		{
			"uuid":             "out-1",
			"startDate":        now,
			"outageType":       "manual",
			"isResolved":       true,
			"durationMs":       60000,
			"detectedLocation": "london",
			"monitor": map[string]interface{}{
				"uuid":     "mon-prod-1",
				"name":     "[PROD]-API",
				"url":      "https://example.com",
				"protocol": "https",
			},
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: filterProviderConfig(server.URL) + `
data "hyperping_outages" "filtered" {
  filter = {
    name_regex = "\\[PROD\\]-.*"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_outages.filtered", "outages.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_statusPages tests status page filtering.
func TestAccDataSourceFilters_statusPages(t *testing.T) {
	server := newMockFilterServer(t)
	defer server.Close()

	server.statusPages = []map[string]interface{}{
		{
			"uuid":               "sp_prod1",
			"name":               "Production Status",
			"hostedsubdomain":    "prod-status",
			"url":                "https://prod-status.hyperping.app",
			"password_protected": false,
			"hostname":           nil,
			"settings": map[string]interface{}{
				"name":             "Production Settings",
				"languages":        []string{"en"},
				"default_language": "en",
				"theme":            "light",
			},
			"sections": []interface{}{},
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: filterProviderConfig(server.URL) + `
data "hyperping_statuspages" "filtered" {
  filter = {
    name_regex = "Production.*"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_statuspages.filtered", "statuspages.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_complexRegexPatterns tests various regex patterns.
func TestAccDataSourceFilters_complexRegexPatterns(t *testing.T) {
	server := newMockFilterServer(t)
	defer server.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	server.monitors = []map[string]interface{}{
		{
			"uuid": "mon-prod-api", "name": "[PROD]-API", "url": "https://example.com",
			"protocol": "https", "http_method": "GET", "check_frequency": 60,
			"expected_status_code": "200", "follow_redirects": true, "paused": false,
			"regions": []string{"london"}, "createdAt": now,
		},
		{
			"uuid": "mon-staging-web", "name": "[STAGING]-Web", "url": "https://staging.example.com",
			"protocol": "https", "http_method": "GET", "check_frequency": 60,
			"expected_status_code": "200", "follow_redirects": true, "paused": false,
			"regions": []string{"london"}, "createdAt": now,
		},
		{
			"uuid": "mon-service-api", "name": "Service-API", "url": "https://service.example.com",
			"protocol": "https", "http_method": "GET", "check_frequency": 60,
			"expected_status_code": "200", "follow_redirects": true, "paused": false,
			"regions": []string{"london"}, "createdAt": now,
		},
		{
			"uuid": "mon-digits-v2", "name": "Monitor v2", "url": "https://v2.example.com",
			"protocol": "https", "http_method": "GET", "check_frequency": 60,
			"expected_status_code": "200", "follow_redirects": true, "paused": false,
			"regions": []string{"london"}, "createdAt": now,
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: filterProviderConfig(server.URL) + `
data "hyperping_monitors" "env_prefix" {
  filter = {
    name_regex = "^\\[(PROD|STAGING)\\]-.*"
  }
}

data "hyperping_monitors" "api_suffix" {
  filter = {
    name_regex = ".*-API$"
  }
}

data "hyperping_monitors" "with_digits" {
  filter = {
    name_regex = ".*\\d+.*"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.env_prefix", "monitors.#"),
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.api_suffix", "monitors.#"),
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.with_digits", "monitors.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_nullVsUnknownHandling tests null and unknown value handling.
func TestAccDataSourceFilters_nullVsUnknownHandling(t *testing.T) {
	server := newMockFilterServer(t)
	defer server.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	server.monitors = []map[string]interface{}{
		{
			"uuid": "mon-https-partial", "name": "HTTPS Monitor", "url": "https://example.com",
			"protocol": "https", "http_method": "GET", "check_frequency": 60,
			"expected_status_code": "200", "follow_redirects": true, "paused": false,
			"regions": []string{"london"}, "createdAt": now,
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: filterProviderConfig(server.URL) + `
data "hyperping_monitors" "partial_filter" {
  filter = {
    protocol = "https"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.partial_filter", "monitors.#"),
				),
			},
		},
	})
}

// TestAccDataSourceFilters_realWorldScenarios tests realistic use cases.
func TestAccDataSourceFilters_realWorldScenarios(t *testing.T) {
	server := newMockFilterServer(t)
	defer server.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	future := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	futureEnd := time.Now().UTC().Add(26 * time.Hour).Format(time.RFC3339)

	server.monitors = []map[string]interface{}{
		{
			"uuid": "mon-prod-https", "name": "[PROD]-API", "url": "https://example.com",
			"protocol": "https", "http_method": "GET", "check_frequency": 60,
			"expected_status_code": "200", "follow_redirects": true, "paused": false,
			"regions": []string{"london"}, "createdAt": now,
		},
	}
	server.incidents = []map[string]interface{}{
		{
			"uuid":        "inci-critical-1",
			"title":       map[string]interface{}{"en": "Critical Outage"},
			"text":        map[string]interface{}{"en": "Service unavailable"},
			"type":        "critical",
			"statuspages": []string{},
			"date":        "2026-01-01T00:00:00Z",
		},
	}
	server.maintenanceWindows = []map[string]interface{}{
		{
			"uuid":       "mw-upcoming",
			"name":       "Upcoming Maintenance",
			"title":      map[string]interface{}{"en": "Upcoming Maintenance"},
			"text":       map[string]interface{}{"en": "Scheduled downtime"},
			"start_date": future,
			"end_date":   futureEnd,
			"monitors":   []string{},
			"status":     "scheduled",
		},
	}
	server.outages = []map[string]interface{}{
		{
			"uuid":             "out-api-1",
			"startDate":        now,
			"outageType":       "manual",
			"isResolved":       true,
			"durationMs":       30000,
			"detectedLocation": "london",
			"monitor": map[string]interface{}{
				"uuid":     "mon-api-1",
				"name":     "API Monitor",
				"url":      "https://api.example.com",
				"protocol": "https",
			},
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: filterProviderConfig(server.URL) + `
data "hyperping_monitors" "prod_https_active" {
  filter = {
    name_regex = "\\[PROD\\]-.*"
    protocol   = "https"
    paused     = false
  }
}

data "hyperping_incidents" "critical_active" {
  filter = {
    severity = "critical"
  }
}

data "hyperping_maintenance_windows" "upcoming" {
  filter = {
    status = "scheduled"
  }
}

data "hyperping_outages" "api_outages" {
  filter = {
    name_regex = ".*API.*"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hyperping_monitors.prod_https_active", "monitors.#"),
					resource.TestCheckResourceAttrSet("data.hyperping_incidents.critical_active", "incidents.#"),
					resource.TestCheckResourceAttrSet("data.hyperping_maintenance_windows.upcoming", "maintenance_windows.#"),
					resource.TestCheckResourceAttrSet("data.hyperping_outages.api_outages", "outages.#"),
				),
			},
		},
	})
}
