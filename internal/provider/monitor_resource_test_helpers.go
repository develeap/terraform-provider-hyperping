// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Test configuration helpers (shared across test files)

func testAccMonitorResourceConfigBasic(baseURL, name string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name = %[2]q
  url  = "https://example.com"
}
`, baseURL, name)
}

func testAccMonitorResourceConfigFull(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name                 = "full-monitor"
  url                  = "https://api.example.com/health"
  protocol             = "http"
  http_method          = "POST"
  check_frequency      = 300
  expected_status_code = "201"
  follow_redirects     = false
  regions              = ["london", "virginia"]
  request_headers = [
    { name = "Content-Type", value = "application/json" },
    { name = "X-Custom", value = "value" }
  ]
  request_body = jsonencode({
    key = "value"
  })
}
`, baseURL)
}

func testAccMonitorResourceConfigPaused(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name   = "paused-monitor"
  url    = "https://example.com"
  paused = true
}
`, baseURL)
}

func testAccCheckMonitorDisappears(server *mockHyperpingServer) tfresource.TestCheckFunc {
	return func(s *terraform.State) error {
		server.deleteAllMonitors()
		return nil
	}
}

func testAccMonitorResourceConfigUpdateAll(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name                 = "updated-all-fields"
  url                  = "https://updated.example.com"
  http_method          = "PUT"
  check_frequency      = 120
  expected_status_code = "204"
  follow_redirects     = false
}
`, baseURL)
}

func testAccMonitorResourceConfigWithHeaders(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name = "headers-test"
  url  = "https://example.com"
  request_headers = [
    { name = "X-First", value = "value1" },
    { name = "X-Second", value = "value2" }
  ]
}
`, baseURL)
}

func testAccMonitorResourceConfigWithUpdatedHeaders(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name = "headers-test"
  url  = "https://example.com"
  request_headers = [
    { name = "X-New", value = "newvalue" }
  ]
}
`, baseURL)
}

func testAccMonitorResourceConfigWithBody(baseURL, body string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name         = "body-test"
  url          = "https://example.com"
  http_method  = "POST"
  request_body = %[2]q
}
`, baseURL, body)
}

func testAccMonitorResourceConfigWithRegions(baseURL, regions string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name    = "regions-test"
  url     = "https://example.com"
  regions = %[2]s
}
`, baseURL, regions)
}

func testAccMonitorResourceConfigWithPaused(baseURL, name string, paused bool) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name   = %[2]q
  url    = "https://example.com"
  paused = %[3]t
}
`, baseURL, name, paused)
}

// Error handling tests

func testAccMonitorResourceConfigAllOptional(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name         = "remove-optional-test"
  url          = "https://example.com"
  http_method  = "POST"
  regions      = ["london", "virginia"]
  request_headers = [
    { name = "Content-Type", value = "application/json" }
  ]
  request_body = jsonencode({
    test = "data"
  })
}
`, baseURL)
}

// Mock server implementation

type mockHyperpingServer struct {
	*httptest.Server
	t        *testing.T
	monitors map[string]map[string]interface{}
	counter  int
}

func newMockHyperpingServer(t *testing.T) *mockHyperpingServer {
	m := &mockHyperpingServer{
		t:        t,
		monitors: make(map[string]map[string]interface{}),
		counter:  0,
	}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handleRequest(w, r)
	}))

	return m
}

// Mock server helpers

func (m *mockHyperpingServer) createTestMonitor(id, name string) {
	m.monitors[id] = map[string]interface{}{
		"monitorUuid":     id,
		"name":            name,
		"url":             "https://example.com",
		"method":          "GET",
		"frequency":       60,
		"timeout":         10,
		"expectedStatus":  200,
		"followRedirects": true,
		"paused":          false,
		"down":            false,
		"regions":         []string{"london", "frankfurt"},
	}
}

// Mock server with error injection

type mockHyperpingServerWithErrors struct {
	*mockHyperpingServer
	createError bool
	readError   bool
	updateError bool
	deleteError bool
	pauseError  bool
}

func newMockHyperpingServerWithErrors(t *testing.T) *mockHyperpingServerWithErrors {
	m := &mockHyperpingServerWithErrors{
		mockHyperpingServer: &mockHyperpingServer{
			t:        t,
			monitors: make(map[string]map[string]interface{}),
			counter:  0,
		},
	}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handleRequestWithErrors(w, r)
	}))

	return m
}

func (m *mockHyperpingServerWithErrors) setCreateError(v bool) { m.createError = v }
func (m *mockHyperpingServerWithErrors) setReadError(v bool)   { m.readError = v }
func (m *mockHyperpingServerWithErrors) setUpdateError(v bool) { m.updateError = v }
func (m *mockHyperpingServerWithErrors) setDeleteError(v bool) { m.deleteError = v }
func (m *mockHyperpingServerWithErrors) setPauseError(v bool)  { m.pauseError = v }

func (m *mockHyperpingServerWithErrors) handleRequestWithErrors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == "POST" && r.URL.Path == "/v1/monitors":
		if m.createError {
			w.WriteHeader(http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"}); err != nil {
				m.t.Errorf("failed to encode error response: %v", err)
			}
			return
		}
		m.createMonitor(w, r)

	case r.Method == "GET" && len(r.URL.Path) > len("/v1/monitors/"):
		if m.readError {
			w.WriteHeader(http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"}); err != nil {
				m.t.Errorf("failed to encode error response: %v", err)
			}
			return
		}
		m.getMonitor(w, r)

	case r.Method == "PUT" && len(r.URL.Path) > len("/v1/monitors/"):
		if m.updateError {
			w.WriteHeader(http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"}); err != nil {
				m.t.Errorf("failed to encode error response: %v", err)
			}
			return
		}
		// Check if this is a pause operation
		if m.pauseError {
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				m.t.Errorf("failed to decode request body: %v", err)
				return
			}
			if paused, ok := req["paused"].(bool); ok && paused {
				w.WriteHeader(http.StatusInternalServerError)
				if err := json.NewEncoder(w).Encode(map[string]string{"error": "Failed to pause"}); err != nil {
					m.t.Errorf("failed to encode error response: %v", err)
				}
				return
			}
		}
		m.updateMonitor(w, r)

	case r.Method == "DELETE" && len(r.URL.Path) > len("/v1/monitors/"):
		if m.deleteError {
			w.WriteHeader(http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"}); err != nil {
				m.t.Errorf("failed to encode error response: %v", err)
			}
			return
		}
		m.deleteMonitor(w, r)

	default:
		m.handleRequest(w, r)
	}
}

func (m *mockHyperpingServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == "GET" && r.URL.Path == "/v1/monitors":
		m.listMonitors(w)
	case r.Method == "POST" && r.URL.Path == "/v1/monitors":
		m.createMonitor(w, r)
	case r.Method == "GET" && len(r.URL.Path) > len("/v1/monitors/"):
		m.getMonitor(w, r)
	case r.Method == "PUT" && len(r.URL.Path) > len("/v1/monitors/"):
		m.updateMonitor(w, r)
	case r.Method == "DELETE" && len(r.URL.Path) > len("/v1/monitors/"):
		m.deleteMonitor(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "Not found"}); err != nil {
			m.t.Errorf("failed to encode error response: %v", err)
		}
	}
}

func (m *mockHyperpingServer) listMonitors(w http.ResponseWriter) {
	monitors := make([]map[string]interface{}, 0, len(m.monitors))
	for _, monitor := range m.monitors {
		monitors = append(monitors, monitor)
	}
	if err := json.NewEncoder(w).Encode(monitors); err != nil {
		m.t.Errorf("failed to encode monitors list: %v", err)
	}
}

func (m *mockHyperpingServer) createMonitor(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"}); encodeErr != nil {
			m.t.Errorf("failed to encode error response: %v", encodeErr)
		}
		return
	}

	m.counter++
	id := fmt.Sprintf("mon_mock%d", m.counter)

	monitor := map[string]interface{}{
		"uuid":                 id,
		"name":                 req["name"],
		"url":                  req["url"],
		"protocol":             getOrDefault(req, "protocol", "http"),
		"http_method":          getOrDefault(req, "http_method", "GET"),
		"check_frequency":      getOrDefaultInt(req, "check_frequency", 60),
		"expected_status_code": getOrDefault(req, "expected_status_code", "2xx"),
		"follow_redirects":     getOrDefaultBool(req, "follow_redirects", true),
		"paused":               getOrDefaultBool(req, "paused", false),
		"regions":              getOrDefaultSlice(req, "regions", []string{"london", "frankfurt"}),
		"request_headers":      []interface{}{},
		"request_body":         "",
	}

	if headers, ok := req["request_headers"].([]interface{}); ok {
		monitor["request_headers"] = headers
	}

	if body, ok := req["request_body"].(string); ok {
		monitor["request_body"] = body
	}

	if port, ok := req["port"].(float64); ok {
		monitor["port"] = int(port)
	}

	if alertsWait, ok := req["alerts_wait"].(float64); ok {
		monitor["alerts_wait"] = int(alertsWait)
	}

	if escalationPolicy, ok := req["escalation_policy"].(string); ok {
		monitor["escalation_policy"] = escalationPolicy
	}

	if requiredKeyword, ok := req["required_keyword"].(string); ok {
		monitor["required_keyword"] = requiredKeyword
	}

	m.monitors[id] = monitor

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(monitor); err != nil {
		m.t.Errorf("failed to encode monitor response: %v", err)
	}
}

func (m *mockHyperpingServer) getMonitor(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/v1/monitors/"):]

	monitor, exists := m.monitors[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "Monitor not found"}); err != nil {
			m.t.Errorf("failed to encode error response: %v", err)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(monitor); err != nil {
		m.t.Errorf("failed to encode monitor response: %v", err)
	}
}

func (m *mockHyperpingServer) updateMonitor(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/v1/monitors/"):]

	monitor, exists := m.monitors[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "Monitor not found"}); err != nil {
			m.t.Errorf("failed to encode error response: %v", err)
		}
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"}); encodeErr != nil {
			m.t.Errorf("failed to encode error response: %v", encodeErr)
		}
		return
	}

	// Update fields if present
	for key, value := range req {
		switch key {
		case "name", "url", "protocol", "http_method", "request_body", "expected_status_code", "escalation_policy", "required_keyword":
			if v, ok := value.(string); ok {
				monitor[key] = v
			}
		case "check_frequency", "port", "alerts_wait":
			if v, ok := value.(float64); ok {
				monitor[key] = int(v)
			}
		case "follow_redirects", "paused":
			if v, ok := value.(bool); ok {
				monitor[key] = v
			}
		case "regions":
			if v, ok := value.([]interface{}); ok {
				regions := make([]string, len(v))
				for i, region := range v {
					if str, ok := region.(string); ok {
						regions[i] = str
					}
				}
				monitor[key] = regions
			} else if value == nil {
				delete(monitor, key)
			}
		case "request_headers":
			if v, ok := value.([]interface{}); ok {
				if len(v) == 0 {
					monitor[key] = []interface{}{}
				} else {
					monitor[key] = v
				}
			} else if value == nil {
				monitor[key] = []interface{}{}
			}
		}
	}

	m.monitors[id] = monitor
	if err := json.NewEncoder(w).Encode(monitor); err != nil {
		m.t.Errorf("failed to encode monitor response: %v", err)
	}
}

func (m *mockHyperpingServer) deleteMonitor(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/v1/monitors/"):]

	if _, exists := m.monitors[id]; !exists {
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "Monitor not found"}); err != nil {
			m.t.Errorf("failed to encode error response: %v", err)
		}
		return
	}

	delete(m.monitors, id)
	w.WriteHeader(http.StatusNoContent)
}

func (m *mockHyperpingServer) deleteAllMonitors() {
	m.monitors = make(map[string]map[string]interface{})
}

// Helper functions for defaults
func getOrDefault(m map[string]interface{}, key, defaultVal string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return defaultVal
}

func getOrDefaultInt(m map[string]interface{}, key string, defaultVal int) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return defaultVal
}

func getOrDefaultBool(m map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return defaultVal
}

func getOrDefaultSlice(m map[string]interface{}, key string, defaultVal []string) []string {
	if v, ok := m[key].([]interface{}); ok {
		result := make([]string, len(v))
		for i, item := range v {
			if str, ok := item.(string); ok {
				result[i] = str
			}
		}
		return result
	}
	return defaultVal
}
