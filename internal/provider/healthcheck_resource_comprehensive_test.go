// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// TestAccHealthcheckResource_cronSchedule tests cron-based scheduling.
// Verifies that cron expressions persist and can be updated.
func TestAccHealthcheckResource_cronSchedule(t *testing.T) {
	server := newMockCronHealthcheckServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Test 1: Create with cron schedule (every 6 hours)
			{
				Config: testAccHealthcheckResourceConfig_cronSchedule(server.URL, "0 */6 * * *", "UTC"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "name", "cron-test"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "cron", "0 */6 * * *"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "timezone", "UTC"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "period", "21600"), // 6 hours
					tfresource.TestCheckResourceAttrSet("hyperping_healthcheck.test", "id"),
					tfresource.TestCheckResourceAttrSet("hyperping_healthcheck.test", "ping_url"),
				),
			},
			// Test 2: Update cron to daily at midnight
			{
				Config: testAccHealthcheckResourceConfig_cronSchedule(server.URL, "0 0 * * *", "UTC"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "cron", "0 0 * * *"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "period", "86400"), // Daily
				),
			},
		},
	})
}

// TestAccHealthcheckResource_cronWithTimezone tests cron with various timezones.
// Verifies that timezone changes persist correctly.
func TestAccHealthcheckResource_cronWithTimezone(t *testing.T) {
	server := newMockCronHealthcheckServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create with America/New_York timezone
			{
				Config: testAccHealthcheckResourceConfig_cronSchedule(server.URL, "0 * * * *", "America/New_York"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "cron", "0 * * * *"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "timezone", "America/New_York"),
				),
			},
			// Update to Europe/London
			{
				Config: testAccHealthcheckResourceConfig_cronSchedule(server.URL, "0 * * * *", "Europe/London"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "timezone", "Europe/London"),
				),
			},
			// Update to Asia/Tokyo
			{
				Config: testAccHealthcheckResourceConfig_cronSchedule(server.URL, "0 * * * *", "Asia/Tokyo"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "timezone", "Asia/Tokyo"),
				),
			},
			// Update to Australia/Sydney
			{
				Config: testAccHealthcheckResourceConfig_cronSchedule(server.URL, "0 * * * *", "Australia/Sydney"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "timezone", "Australia/Sydney"),
				),
			},
			// Update to America/Los_Angeles
			{
				Config: testAccHealthcheckResourceConfig_cronSchedule(server.URL, "0 * * * *", "America/Los_Angeles"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "timezone", "America/Los_Angeles"),
				),
			},
		},
	})
}

// TestAccHealthcheckResource_cronVsPeriodExclusive tests mutual exclusivity validation.
// Verifies that providing both cron and period_value triggers an error.
func TestAccHealthcheckResource_cronVsPeriodExclusive(t *testing.T) {
	server := newMockCronHealthcheckServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccHealthcheckResourceConfig_cronAndPeriod(server.URL),
				ExpectError: regexp.MustCompile(`(cron|period).*(cron|period)`),
			},
		},
	})
}

// TestAccHealthcheckResource_computedFields tests computed field verification.
// Verifies that is_down, last_ping, and created_at are properly set by the API.
func TestAccHealthcheckResource_computedFields(t *testing.T) {
	server := newMockCronHealthcheckServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccHealthcheckResourceConfig_cronSchedule(server.URL, "0 0 * * *", "UTC"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					// Verify is_down is set (should be false for new healthcheck)
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "is_down", "false"),
					// Verify last_ping is set (may be empty initially, but field should exist)
					tfresource.TestCheckResourceAttrSet("hyperping_healthcheck.test", "last_ping"),
					// Verify created_at is set and matches ISO 8601 format
					tfresource.TestCheckResourceAttrSet("hyperping_healthcheck.test", "created_at"),
					tfresource.TestMatchResourceAttr("hyperping_healthcheck.test", "created_at", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`)),
				),
			},
		},
	})
}

// TestAccHealthcheckResource_cronEdgeCases tests various cron expressions.
// Verifies that different cron patterns are accepted and persist correctly.
func TestAccHealthcheckResource_cronEdgeCases(t *testing.T) {
	server := newMockCronHealthcheckServer(t)
	defer server.Close()

	testCases := []struct {
		cron        string
		description string
		period      string
	}{
		{"* * * * *", "every minute", "60"},
		{"0 0 1 * *", "first day of month", "2592000"},
		{"0 0 * * 0", "every Sunday", "604800"},
		{"*/15 * * * *", "every 15 minutes", "900"},
		{"0 */2 * * *", "every 2 hours", "7200"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			tfresource.Test(t, tfresource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []tfresource.TestStep{
					{
						Config: testAccHealthcheckResourceConfig_cronSchedule(server.URL, tc.cron, "UTC"),
						Check: tfresource.ComposeAggregateTestCheckFunc(
							tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "cron", tc.cron),
							tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "period", tc.period),
							tfresource.TestCheckResourceAttrSet("hyperping_healthcheck.test", "id"),
						),
					},
				},
			})
		})
	}
}

// Configuration helpers

func testAccHealthcheckResourceConfig_cronSchedule(baseURL, cron, timezone string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_healthcheck" "test" {
  name               = "cron-test"
  cron               = %[2]q
  timezone           = %[3]q
  grace_period_value = 30
  grace_period_type  = "minutes"
}
`, baseURL, cron, timezone)
}

func testAccHealthcheckResourceConfig_cronAndPeriod(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_healthcheck" "test" {
  name               = "invalid-config"
  cron               = "0 * * * *"
  timezone           = "UTC"
  period_value       = 60
  period_type        = "minutes"
  grace_period_value = 15
  grace_period_type  = "minutes"
}
`, baseURL)
}

// Mock server implementation for cron-based healthchecks

type mockCronHealthcheckServer struct {
	*httptest.Server
	t            *testing.T
	healthchecks map[string]map[string]interface{}
	counter      int
}

func newMockCronHealthcheckServer(t *testing.T) *mockCronHealthcheckServer {
	m := &mockCronHealthcheckServer{
		t:            t,
		healthchecks: make(map[string]map[string]interface{}),
		counter:      0,
	}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handleRequest(w, r)
	}))

	return m
}

func (m *mockCronHealthcheckServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)
	basePath := client.HealthchecksBasePath
	basePathWithSlash := basePath + "/"

	switch {
	case r.Method == "GET" && r.URL.Path == basePath:
		m.listHealthchecks(w)
	case r.Method == "POST" && r.URL.Path == basePath:
		m.createHealthcheck(w, r)
	case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/pause"):
		m.pauseHealthcheck(w, r)
	case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/resume"):
		m.resumeHealthcheck(w, r)
	case r.Method == "GET" && strings.HasPrefix(r.URL.Path, basePathWithSlash):
		m.getHealthcheck(w, r)
	case r.Method == "PUT" && strings.HasPrefix(r.URL.Path, basePathWithSlash):
		m.updateHealthcheck(w, r)
	case r.Method == "DELETE" && strings.HasPrefix(r.URL.Path, basePathWithSlash):
		m.deleteHealthcheck(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}
}

func (m *mockCronHealthcheckServer) listHealthchecks(w http.ResponseWriter) {
	healthchecks := make([]map[string]interface{}, 0, len(m.healthchecks))
	for _, hc := range m.healthchecks {
		healthchecks = append(healthchecks, hc)
	}
	json.NewEncoder(w).Encode(healthchecks)
}

func (m *mockCronHealthcheckServer) createHealthcheck(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	m.counter++
	id := fmt.Sprintf("hc_%d", m.counter)

	// Extract cron/timezone fields
	cron, _ := req["cron"].(string)
	timezone, _ := req["timezone"].(string)
	gracePeriodValue := getOrDefaultInt(req, "grace_period_value", 300)
	gracePeriodType := getOrDefaultString(req, "grace_period_type")

	// Calculate period from cron expression
	period := calculatePeriodFromCronExpr(cron)
	gracePeriod := calculateSeconds(gracePeriodValue, gracePeriodType)

	// API returns camelCase
	healthcheck := map[string]interface{}{
		"uuid":             id,
		"name":             req["name"],
		"pingUrl":          fmt.Sprintf("https://hb.tinyping.io/%s", id),
		"cron":             cron,
		"timezone":         timezone,
		"period":           period,
		"gracePeriod":      gracePeriod,
		"gracePeriodValue": gracePeriodValue,
		"gracePeriodType":  gracePeriodType,
		"isPaused":         false,
		"isDown":           false,
		"lastPing":         "2026-02-16T10:30:00Z",
		"createdAt":        "2026-02-16T10:00:00Z",
	}

	m.healthchecks[id] = healthcheck

	response := map[string]interface{}{
		"message":     "Healthcheck created successfully",
		"healthcheck": healthcheck,
	}
	json.NewEncoder(w).Encode(response)
}

func (m *mockCronHealthcheckServer) getHealthcheck(w http.ResponseWriter, r *http.Request) {
	basePath := client.HealthchecksBasePath
	id := strings.TrimPrefix(r.URL.Path, basePath+"/")

	healthcheck, ok := m.healthchecks[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Healthcheck not found"})
		return
	}

	response := map[string]interface{}{
		"healthcheck": healthcheck,
	}
	json.NewEncoder(w).Encode(response)
}

func (m *mockCronHealthcheckServer) updateHealthcheck(w http.ResponseWriter, r *http.Request) {
	basePath := client.HealthchecksBasePath
	id := strings.TrimPrefix(r.URL.Path, basePath+"/")

	healthcheck, ok := m.healthchecks[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Healthcheck not found"})
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	// Update fields
	if name, ok := req["name"]; ok {
		healthcheck["name"] = name
	}
	if cron, ok := req["cron"]; ok {
		healthcheck["cron"] = cron
		// Recalculate period when cron changes
		if cronStr, ok := cron.(string); ok {
			healthcheck["period"] = calculatePeriodFromCronExpr(cronStr)
		}
	}
	if timezone, ok := req["timezone"]; ok {
		healthcheck["timezone"] = timezone
	}
	if gpv, ok := req["grace_period_value"]; ok {
		healthcheck["gracePeriodValue"] = gpv
	}
	if gpt, ok := req["grace_period_type"]; ok {
		healthcheck["gracePeriodType"] = gpt
	}

	response := map[string]interface{}{
		"message":     "Healthcheck updated successfully",
		"healthcheck": healthcheck,
	}
	json.NewEncoder(w).Encode(response)
}

func (m *mockCronHealthcheckServer) deleteHealthcheck(w http.ResponseWriter, r *http.Request) {
	basePath := client.HealthchecksBasePath
	id := strings.TrimPrefix(r.URL.Path, basePath+"/")

	if _, ok := m.healthchecks[id]; !ok {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Healthcheck not found"})
		return
	}

	delete(m.healthchecks, id)
	w.WriteHeader(http.StatusNoContent)
}

func (m *mockCronHealthcheckServer) pauseHealthcheck(w http.ResponseWriter, r *http.Request) {
	basePath := client.HealthchecksBasePath
	path := strings.TrimPrefix(r.URL.Path, basePath+"/")
	id := strings.TrimSuffix(path, "/pause")

	healthcheck, ok := m.healthchecks[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Healthcheck not found"})
		return
	}

	healthcheck["isPaused"] = true

	response := map[string]interface{}{
		"message": "Healthcheck paused",
		"uuid":    id,
	}
	json.NewEncoder(w).Encode(response)
}

func (m *mockCronHealthcheckServer) resumeHealthcheck(w http.ResponseWriter, r *http.Request) {
	basePath := client.HealthchecksBasePath
	path := strings.TrimPrefix(r.URL.Path, basePath+"/")
	id := strings.TrimSuffix(path, "/resume")

	healthcheck, ok := m.healthchecks[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Healthcheck not found"})
		return
	}

	healthcheck["isPaused"] = false

	response := map[string]interface{}{
		"message": "Healthcheck resumed",
		"uuid":    id,
	}
	json.NewEncoder(w).Encode(response)
}

// Helper function for calculating period from cron expression (local to this file)

func calculatePeriodFromCronExpr(cron string) int {
	// Simplified: return reasonable defaults based on common patterns
	switch cron {
	case "* * * * *":
		return 60 // Every minute
	case "*/15 * * * *":
		return 900 // Every 15 minutes
	case "0 * * * *":
		return 3600 // Hourly
	case "0 */2 * * *":
		return 7200 // Every 2 hours
	case "0 */6 * * *":
		return 21600 // Every 6 hours
	case "0 0 * * *":
		return 86400 // Daily
	case "0 0 1 * *":
		return 2592000 // Monthly (30 days)
	case "0 0 * * 0":
		return 604800 // Weekly
	default:
		return 3600 // Default to hourly
	}
}
