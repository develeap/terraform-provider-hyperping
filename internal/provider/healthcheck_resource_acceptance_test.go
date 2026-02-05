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
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccHealthcheckResource_basic(t *testing.T) {
	server := newMockHealthcheckServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create and Read testing
			{
				Config: testAccHealthcheckResourceConfig_basic(server.URL, "test-healthcheck"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "name", "test-healthcheck"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "period_value", "60"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "period_type", "seconds"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "grace_period_value", "300"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "grace_period_type", "seconds"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "period", "60"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "grace_period", "300"),
					tfresource.TestCheckResourceAttrSet("hyperping_healthcheck.test", "id"),
					tfresource.TestCheckResourceAttrSet("hyperping_healthcheck.test", "ping_url"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "hyperping_healthcheck.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccHealthcheckResourceConfig_basic(server.URL, "updated-healthcheck"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "name", "updated-healthcheck"),
				),
			},
		},
	})
}

func TestAccHealthcheckResource_full(t *testing.T) {
	server := newMockHealthcheckServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccHealthcheckResourceConfig_full(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "name", "full-healthcheck"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "period_value", "300"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "period_type", "seconds"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "grace_period_value", "600"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "grace_period_type", "seconds"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "period", "300"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "grace_period", "600"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "is_paused", "false"),
					tfresource.TestCheckResourceAttrSet("hyperping_healthcheck.test", "escalation_policy"),
				),
			},
		},
	})
}

func TestAccHealthcheckResource_pause(t *testing.T) {
	server := newMockHealthcheckServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create unpaused
			{
				Config: testAccHealthcheckResourceConfig_withPaused(server.URL, false),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "is_paused", "false"),
				),
			},
			// Pause
			{
				Config: testAccHealthcheckResourceConfig_withPaused(server.URL, true),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "is_paused", "true"),
				),
			},
			// Unpause
			{
				Config: testAccHealthcheckResourceConfig_withPaused(server.URL, false),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "is_paused", "false"),
				),
			},
		},
	})
}

func TestAccHealthcheckResource_disappears(t *testing.T) {
	server := newMockHealthcheckServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccHealthcheckResourceConfig_basic(server.URL, "disappear-test"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttrSet("hyperping_healthcheck.test", "id"),
					testAccCheckHealthcheckDisappears(server),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccHealthcheckResource_updateAll(t *testing.T) {
	server := newMockHealthcheckServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create with minimal config
			{
				Config: testAccHealthcheckResourceConfig_basic(server.URL, "update-test"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "name", "update-test"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "period_value", "60"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "period_type", "seconds"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "grace_period_value", "300"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "grace_period_type", "seconds"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "period", "60"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "grace_period", "300"),
				),
			},
			// Update all fields
			{
				Config: testAccHealthcheckResourceConfig_updateAll(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "name", "updated-all"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "period_value", "180"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "period_type", "seconds"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "grace_period_value", "900"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "grace_period_type", "seconds"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "period", "180"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "grace_period", "900"),
				),
			},
		},
	})
}

func TestAccHealthcheckResource_createError(t *testing.T) {
	server := newMockHealthcheckServerWithErrors(t)
	defer server.Close()

	server.setCreateError(true)

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccHealthcheckResourceConfig_basic(server.URL, "error-test"),
				ExpectError: regexp.MustCompile(`Could not create healthcheck`),
			},
		},
	})
}

// Helper functions

func testAccHealthcheckResourceConfig_basic(baseURL, name string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_healthcheck" "test" {
  name               = %[2]q
  period_value       = 60
  period_type        = "seconds"
  grace_period_value = 300
  grace_period_type  = "seconds"
}
`, baseURL, name)
}

func testAccHealthcheckResourceConfig_full(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_healthcheck" "test" {
  name               = "full-healthcheck"
  period_value       = 300
  period_type        = "seconds"
  grace_period_value = 600
  grace_period_type  = "seconds"
  is_paused          = false
  escalation_policy  = "ep_test123"
}
`, baseURL)
}

func testAccHealthcheckResourceConfig_withPaused(baseURL string, paused bool) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_healthcheck" "test" {
  name               = "pause-test"
  period_value       = 60
  period_type        = "seconds"
  grace_period_value = 300
  grace_period_type  = "seconds"
  is_paused          = %[2]t
}
`, baseURL, paused)
}

func testAccHealthcheckResourceConfig_updateAll(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_healthcheck" "test" {
  name               = "updated-all"
  period_value       = 180
  period_type        = "seconds"
  grace_period_value = 900
  grace_period_type  = "seconds"
}
`, baseURL)
}

func testAccCheckHealthcheckDisappears(server *mockHealthcheckServer) tfresource.TestCheckFunc {
	return func(s *terraform.State) error {
		server.deleteAllHealthchecks()
		return nil
	}
}

// Mock server implementation

type mockHealthcheckServer struct {
	*httptest.Server
	t            *testing.T
	healthchecks map[string]map[string]interface{}
	counter      int
}

func newMockHealthcheckServer(t *testing.T) *mockHealthcheckServer {
	m := &mockHealthcheckServer{
		t:            t,
		healthchecks: make(map[string]map[string]interface{}),
		counter:      0,
	}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handleRequest(w, r)
	}))

	return m
}

func (m *mockHealthcheckServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == "GET" && r.URL.Path == "/v2/healthchecks":
		m.listHealthchecks(w)
	case r.Method == "POST" && r.URL.Path == "/v2/healthchecks":
		m.createHealthcheck(w, r)
	case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/pause"):
		m.pauseHealthcheck(w, r)
	case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/resume"):
		m.resumeHealthcheck(w, r)
	case r.Method == "GET" && len(r.URL.Path) > len("/v2/healthchecks/"):
		m.getHealthcheck(w, r)
	case r.Method == "PUT" && len(r.URL.Path) > len("/v2/healthchecks/"):
		m.updateHealthcheck(w, r)
	case r.Method == "DELETE" && len(r.URL.Path) > len("/v2/healthchecks/"):
		m.deleteHealthcheck(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}
}

func (m *mockHealthcheckServer) listHealthchecks(w http.ResponseWriter) {
	healthchecks := make([]map[string]interface{}, 0, len(m.healthchecks))
	for _, hc := range m.healthchecks {
		healthchecks = append(healthchecks, hc)
	}
	json.NewEncoder(w).Encode(healthchecks)
}

func (m *mockHealthcheckServer) createHealthcheck(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	m.counter++
	id := fmt.Sprintf("hc_%d", m.counter)

	// Extract period and grace values (API receives snake_case)
	periodValue := getOrDefaultInt(req, "period_value", 60)
	periodType := getOrDefaultString(req, "period_type")
	gracePeriodValue := getOrDefaultInt(req, "grace_period_value", 300)
	gracePeriodType := getOrDefaultString(req, "grace_period_type")

	// Calculate period and gracePeriod in seconds
	period := calculateSeconds(periodValue, periodType)
	gracePeriod := calculateSeconds(gracePeriodValue, gracePeriodType)

	// API returns camelCase
	healthcheck := map[string]interface{}{
		"uuid":             id,
		"name":             req["name"],
		"periodValue":      periodValue,
		"periodType":       periodType,
		"period":           period,
		"gracePeriodValue": gracePeriodValue,
		"gracePeriodType":  gracePeriodType,
		"gracePeriod":      gracePeriod,
		"isPaused":         getOrDefaultBool(req, "is_paused", false),
		"isDown":           false,
		"pingUrl":          fmt.Sprintf("https://ping.hyperping.io/%s", id),
		"createdAt":        "2026-01-01T00:00:00Z",
	}

	if ep, ok := req["escalation_policy"].(string); ok && ep != "" {
		healthcheck["escalationPolicy"] = map[string]interface{}{
			"uuid": ep,
		}
	}

	m.healthchecks[id] = healthcheck

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(healthcheck)
}

func (m *mockHealthcheckServer) getHealthcheck(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/v2/healthchecks/"):]

	healthcheck, exists := m.healthchecks[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Healthcheck not found"})
		return
	}

	json.NewEncoder(w).Encode(healthcheck)
}

func (m *mockHealthcheckServer) updateHealthcheck(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/v2/healthchecks/"):]

	healthcheck, exists := m.healthchecks[id]
	if !exists {
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

	// Update fields if present (API receives snake_case)
	if name, ok := req["name"].(string); ok {
		healthcheck["name"] = name
	}
	if periodValue, ok := req["period_value"].(float64); ok {
		healthcheck["periodValue"] = int(periodValue)
		periodType := getOrDefaultString(healthcheck, "periodType")
		healthcheck["period"] = calculateSeconds(int(periodValue), periodType)
	}
	if periodType, ok := req["period_type"].(string); ok {
		healthcheck["periodType"] = periodType
		periodValue := getOrDefaultInt(healthcheck, "periodValue", 60)
		healthcheck["period"] = calculateSeconds(periodValue, periodType)
	}
	if gracePeriodValue, ok := req["grace_period_value"].(float64); ok {
		healthcheck["gracePeriodValue"] = int(gracePeriodValue)
		gracePeriodType := getOrDefaultString(healthcheck, "gracePeriodType")
		healthcheck["gracePeriod"] = calculateSeconds(int(gracePeriodValue), gracePeriodType)
	}
	if gracePeriodType, ok := req["grace_period_type"].(string); ok {
		healthcheck["gracePeriodType"] = gracePeriodType
		gracePeriodValue := getOrDefaultInt(healthcheck, "gracePeriodValue", 300)
		healthcheck["gracePeriod"] = calculateSeconds(gracePeriodValue, gracePeriodType)
	}
	if ep, ok := req["escalation_policy"].(string); ok {
		if ep != "" {
			healthcheck["escalationPolicy"] = map[string]interface{}{
				"uuid": ep,
			}
		} else {
			delete(healthcheck, "escalationPolicy")
		}
	}

	m.healthchecks[id] = healthcheck
	json.NewEncoder(w).Encode(healthcheck)
}

func (m *mockHealthcheckServer) deleteHealthcheck(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/v2/healthchecks/"):]

	if _, exists := m.healthchecks[id]; !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Healthcheck not found"})
		return
	}

	delete(m.healthchecks, id)
	w.WriteHeader(http.StatusNoContent)
}

func (m *mockHealthcheckServer) pauseHealthcheck(w http.ResponseWriter, r *http.Request) {
	// Extract UUID from path: /v2/healthchecks/{uuid}/pause
	path := strings.TrimPrefix(r.URL.Path, "/v2/healthchecks/")
	id := strings.TrimSuffix(path, "/pause")

	healthcheck, exists := m.healthchecks[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Healthcheck not found"})
		return
	}

	healthcheck["isPaused"] = true
	m.healthchecks[id] = healthcheck

	json.NewEncoder(w).Encode(map[string]string{"message": "paused", "uuid": id})
}

func (m *mockHealthcheckServer) resumeHealthcheck(w http.ResponseWriter, r *http.Request) {
	// Extract UUID from path: /v2/healthchecks/{uuid}/resume
	path := strings.TrimPrefix(r.URL.Path, "/v2/healthchecks/")
	id := strings.TrimSuffix(path, "/resume")

	healthcheck, exists := m.healthchecks[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Healthcheck not found"})
		return
	}

	healthcheck["isPaused"] = false
	m.healthchecks[id] = healthcheck

	json.NewEncoder(w).Encode(map[string]string{"message": "resumed", "uuid": id})
}

func (m *mockHealthcheckServer) deleteAllHealthchecks() {
	m.healthchecks = make(map[string]map[string]interface{})
}

// Mock server with error injection

type mockHealthcheckServerWithErrors struct {
	*mockHealthcheckServer
	createError bool
	readError   bool
	updateError bool
	deleteError bool
}

func newMockHealthcheckServerWithErrors(t *testing.T) *mockHealthcheckServerWithErrors {
	m := &mockHealthcheckServerWithErrors{
		mockHealthcheckServer: &mockHealthcheckServer{
			t:            t,
			healthchecks: make(map[string]map[string]interface{}),
			counter:      0,
		},
	}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handleRequestWithErrors(w, r)
	}))

	return m
}

func (m *mockHealthcheckServerWithErrors) setCreateError(v bool) { m.createError = v }
func (m *mockHealthcheckServerWithErrors) setReadError(v bool)   { m.readError = v }
func (m *mockHealthcheckServerWithErrors) setUpdateError(v bool) { m.updateError = v }
func (m *mockHealthcheckServerWithErrors) setDeleteError(v bool) { m.deleteError = v }

func (m *mockHealthcheckServerWithErrors) handleRequestWithErrors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == "POST" && r.URL.Path == "/v2/healthchecks":
		if m.createError {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
			return
		}
		m.createHealthcheck(w, r)

	case r.Method == "GET" && len(r.URL.Path) > len("/v2/healthchecks/"):
		if m.readError {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
			return
		}
		m.getHealthcheck(w, r)

	case r.Method == "PUT" && len(r.URL.Path) > len("/v2/healthchecks/"):
		if m.updateError {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
			return
		}
		m.updateHealthcheck(w, r)

	case r.Method == "DELETE" && len(r.URL.Path) > len("/v2/healthchecks/"):
		if m.deleteError {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
			return
		}
		m.deleteHealthcheck(w, r)

	default:
		m.handleRequest(w, r)
	}
}

// Helper functions for mock server

func getOrDefaultString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return "seconds"
}

func calculateSeconds(value int, unit string) int {
	switch unit {
	case "seconds":
		return value
	case "minutes":
		return value * 60
	case "hours":
		return value * 3600
	case "days":
		return value * 86400
	default:
		return value
	}
}
