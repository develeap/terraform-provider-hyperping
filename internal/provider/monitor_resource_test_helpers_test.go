// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	hyperping "github.com/develeap/hyperping-go"
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

func testAccMonitorResourceConfigWithAuthHeader(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name = "auth-header-test"
  url  = "https://example.com"
  request_headers = [
    { name = "Authorization", value = "Basic dXNlcjpwYXNzd29yZA==" },
    { name = "Cookie",        value = "session=abc123" }
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

// recordedRequest stores a snapshot of an HTTP request for later assertions.
type recordedRequest struct {
	Method string
	Path   string
	Body   map[string]interface{} // decoded JSON body (nil for GET/DELETE)
}

// Mock server implementation

type mockHyperpingServer struct {
	*httptest.Server
	t        *testing.T
	mu       sync.RWMutex
	monitors map[string]map[string]interface{}
	counter  int
	requests []recordedRequest // append-only log of all requests
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

// recordRequest snapshots the incoming request and restores r.Body for downstream handlers.
func (m *mockHyperpingServer) recordRequest(r *http.Request) {
	rec := recordedRequest{Method: r.Method, Path: r.URL.Path}
	if r.Body != nil && r.Method != "GET" && r.Method != "DELETE" {
		bodyBytes, err := io.ReadAll(r.Body)
		if err == nil && len(bodyBytes) > 0 {
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			var decoded map[string]interface{}
			if json.Unmarshal(bodyBytes, &decoded) == nil {
				rec.Body = decoded
			}
		} else if err == nil {
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
	}
	m.mu.Lock()
	m.requests = append(m.requests, rec)
	m.mu.Unlock()
}

// getRequests returns a copy of all recorded requests (thread-safe).
func (m *mockHyperpingServer) getRequests() []recordedRequest {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]recordedRequest, len(m.requests))
	copy(out, m.requests)
	return out
}

// lastRequest returns the most recent recorded request, or nil if none.
func (m *mockHyperpingServer) lastRequest() *recordedRequest {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.requests) == 0 {
		return nil
	}
	r := m.requests[len(m.requests)-1]
	return &r
}

func (m *mockHyperpingServer) createTestMonitor(id, name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.monitors[id] = map[string]interface{}{
		"uuid":            id,
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
	createError atomic.Bool
	readError   atomic.Bool
	updateError atomic.Bool
	deleteError atomic.Bool
	pauseError  atomic.Bool
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

func (m *mockHyperpingServerWithErrors) setCreateError(v bool) { m.createError.Store(v) }
func (m *mockHyperpingServerWithErrors) setReadError(v bool)   { m.readError.Store(v) }
func (m *mockHyperpingServerWithErrors) setUpdateError(v bool) { m.updateError.Store(v) }
func (m *mockHyperpingServerWithErrors) setDeleteError(v bool) { m.deleteError.Store(v) }
func (m *mockHyperpingServerWithErrors) setPauseError(v bool)  { m.pauseError.Store(v) }

func (m *mockHyperpingServerWithErrors) writeInternalError(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": msg}); err != nil {
		m.t.Errorf("failed to encode error response: %v", err)
	}
}

// handlePauseError checks whether the request is a pause operation and,
// if pauseError is set, writes an error response. Returns true if an error was written.
// The request body is buffered and restored so subsequent handlers can read it.
func (m *mockHyperpingServerWithErrors) handlePauseError(w http.ResponseWriter, r *http.Request) bool {
	if !m.pauseError.Load() {
		return false
	}
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		m.t.Errorf("failed to read request body: %v", err)
		return true
	}
	// Restore the body for subsequent readers
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	var req map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		m.t.Errorf("failed to decode request body: %v", err)
		return true
	}
	if paused, ok := req["paused"].(bool); ok && paused {
		m.writeInternalError(w, "Failed to pause")
		return true
	}
	return false
}

func (m *mockHyperpingServerWithErrors) handleRequestWithErrors(w http.ResponseWriter, r *http.Request) {
	m.recordRequest(r)
	w.Header().Set(hyperping.HeaderContentType, hyperping.ContentTypeJSON)

	isMonitorPath := len(r.URL.Path) > len(hyperping.MonitorsBasePath+"/")

	switch {
	case r.Method == "POST" && r.URL.Path == hyperping.MonitorsBasePath:
		if m.createError.Load() {
			m.writeInternalError(w, "Internal server error")
			return
		}
		m.createMonitor(w, r)

	case r.Method == "GET" && isMonitorPath:
		if m.readError.Load() {
			m.writeInternalError(w, "Internal server error")
			return
		}
		m.getMonitor(w, r)

	case r.Method == "PUT" && isMonitorPath:
		if m.updateError.Load() {
			m.writeInternalError(w, "Internal server error")
			return
		}
		if m.handlePauseError(w, r) {
			return
		}
		m.updateMonitor(w, r)

	case r.Method == "DELETE" && isMonitorPath:
		if m.deleteError.Load() {
			m.writeInternalError(w, "Internal server error")
			return
		}
		m.deleteMonitor(w, r)

	default:
		m.handleRequest(w, r)
	}
}

func (m *mockHyperpingServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	m.recordRequest(r)
	w.Header().Set(hyperping.HeaderContentType, hyperping.ContentTypeJSON)

	switch {
	case r.Method == "GET" && r.URL.Path == hyperping.MonitorsBasePath:
		m.listMonitors(w)
	case r.Method == "POST" && r.URL.Path == hyperping.MonitorsBasePath:
		m.createMonitor(w, r)
	case r.Method == "GET" && len(r.URL.Path) > len(hyperping.MonitorsBasePath+"/"):
		m.getMonitor(w, r)
	case r.Method == "PUT" && len(r.URL.Path) > len(hyperping.MonitorsBasePath+"/"):
		m.updateMonitor(w, r)
	case r.Method == "DELETE" && len(r.URL.Path) > len(hyperping.MonitorsBasePath+"/"):
		m.deleteMonitor(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "Not found"}); err != nil {
			m.t.Errorf("failed to encode error response: %v", err)
		}
	}
}

func (m *mockHyperpingServer) listMonitors(w http.ResponseWriter) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Collect UUIDs and sort for deterministic ordering across Go map iterations.
	uuids := make([]string, 0, len(m.monitors))
	for id := range m.monitors {
		uuids = append(uuids, id)
	}
	sort.Strings(uuids)

	monitors := make([]map[string]interface{}, 0, len(m.monitors))
	for _, id := range uuids {
		monitors = append(monitors, m.monitors[id])
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

	m.mu.Lock()
	defer m.mu.Unlock()

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
		"status":               "up",
		"ssl_expiration":       90,
		"projectUuid":          "proj_test123",
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

	if ep, ok := req["escalation_policy"].(string); ok && ep != "" && ep != "none" {
		monitor["escalation_policy"] = map[string]interface{}{
			"uuid": ep,
			"name": "Mock Policy",
		}
	} else {
		monitor["escalation_policy"] = nil
	}

	if requiredKeyword, ok := req["required_keyword"].(string); ok {
		monitor["required_keyword"] = requiredKeyword
	}

	// Handle DNS fields — only store for DNS protocol monitors
	protocol := getOrDefault(req, "protocol", "http")
	if protocol == "dns" {
		if drt, ok := req["dns_record_type"].(string); ok && drt != "" {
			monitor["dns_record_type"] = drt
		} else {
			monitor["dns_record_type"] = "A" // API default
		}
		if ns, ok := req["dns_nameserver"].(string); ok {
			monitor["dns_nameserver"] = ns
		}
		if ea, ok := req["dns_expected_answer"].(string); ok {
			monitor["dns_expected_answer"] = ea
		}
	}
	// For non-DNS monitors, dns_record_type is intentionally NOT stored —
	// the real API ignores it and returns null.

	m.monitors[id] = monitor

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(monitor); err != nil {
		m.t.Errorf("failed to encode monitor response: %v", err)
	}
}

func (m *mockHyperpingServer) getMonitor(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, hyperping.MonitorsBasePath+"/")

	m.mu.RLock()
	defer m.mu.RUnlock()

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

// applyStringField sets monitor[key] = value if value is a string.
func applyStringField(monitor map[string]interface{}, key string, value interface{}) {
	if v, ok := value.(string); ok {
		monitor[key] = v
	}
}

// applyIntField sets monitor[key] = int(value) if value is a float64 (JSON number).
func applyIntField(monitor map[string]interface{}, key string, value interface{}) {
	if v, ok := value.(float64); ok {
		monitor[key] = int(v)
	}
}

// applyBoolField sets monitor[key] = value if value is a bool.
func applyBoolField(monitor map[string]interface{}, key string, value interface{}) {
	if v, ok := value.(bool); ok {
		monitor[key] = v
	}
}

// applyRegionsField converts a JSON []interface{} of strings into []string and stores it.
func applyRegionsField(monitor map[string]interface{}, key string, value interface{}) {
	if value == nil {
		delete(monitor, key)
		return
	}
	v, ok := value.([]interface{})
	if !ok {
		return
	}
	regions := make([]string, len(v))
	for i, region := range v {
		if str, ok := region.(string); ok {
			regions[i] = str
		}
	}
	monitor[key] = regions
}

// applyHeadersField stores request_headers as an empty or populated []interface{}.
func applyHeadersField(monitor map[string]interface{}, key string, value interface{}) {
	if value == nil {
		monitor[key] = []interface{}{}
		return
	}
	if v, ok := value.([]interface{}); ok {
		monitor[key] = v
	}
}

// stringFields are monitor fields that map directly from JSON strings.
var monitorStringFields = map[string]bool{
	"name": true, "url": true, "protocol": true, "http_method": true,
	"request_body": true, "expected_status_code": true,
	"required_keyword": true,
	"status":           true, "projectUuid": true,
}

// dnsStringFields are DNS-specific fields that should only be stored for DNS monitors.
var dnsStringFields = map[string]bool{
	"dns_record_type": true, "dns_nameserver": true, "dns_expected_answer": true,
}

// intFields are monitor fields that map from JSON numbers.
var monitorIntFields = map[string]bool{
	"check_frequency": true, "port": true, "alerts_wait": true, "ssl_expiration": true,
}

// boolFields are monitor fields that map from JSON booleans.
var monitorBoolFields = map[string]bool{
	"follow_redirects": true, "paused": true,
}

// applyMonitorField applies a single field from the request map to the monitor map.
func applyMonitorField(monitor map[string]interface{}, key string, value interface{}) {
	switch {
	case monitorStringFields[key]:
		applyStringField(monitor, key, value)
	case monitorIntFields[key]:
		applyIntField(monitor, key, value)
	case monitorBoolFields[key]:
		applyBoolField(monitor, key, value)
	case dnsStringFields[key]:
		// DNS fields are only stored for DNS monitors, matching real API behavior.
		// The real API ignores dns_record_type on non-DNS monitors.
		if protocol, ok := monitor["protocol"].(string); ok && protocol == "dns" {
			applyStringField(monitor, key, value)
		}
	case key == "regions":
		applyRegionsField(monitor, key, value)
	case key == "request_headers":
		applyHeadersField(monitor, key, value)
	case key == "escalation_policy":
		// Store as object to match real API GET response shape.
		// "none" is the sentinel value to unlink; treat same as empty/nil.
		if ep, ok := value.(string); ok && ep != "" && ep != "none" {
			monitor[key] = map[string]interface{}{
				"uuid": ep,
				"name": "Mock Policy",
			}
		} else {
			monitor[key] = nil
		}
	}
}

func (m *mockHyperpingServer) updateMonitor(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, hyperping.MonitorsBasePath+"/")

	m.mu.Lock()
	defer m.mu.Unlock()

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

	// Apply protocol first so DNS field logic in applyMonitorField sees the new protocol.
	if proto, ok := req["protocol"]; ok {
		applyMonitorField(monitor, "protocol", proto)
	}

	for key, value := range req {
		if key == "protocol" {
			continue // already applied above
		}
		applyMonitorField(monitor, key, value)
	}

	// When protocol changes away from DNS, clear DNS-specific fields (matches real API behavior).
	if newProtocol, ok := req["protocol"].(string); ok && newProtocol != "dns" {
		delete(monitor, "dns_record_type")
		delete(monitor, "dns_nameserver")
		delete(monitor, "dns_expected_answer")
	}

	m.monitors[id] = monitor
	if err := json.NewEncoder(w).Encode(monitor); err != nil {
		m.t.Errorf("failed to encode monitor response: %v", err)
	}
}

func (m *mockHyperpingServer) deleteMonitor(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, hyperping.MonitorsBasePath+"/")

	m.mu.Lock()
	defer m.mu.Unlock()

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
	m.mu.Lock()
	defer m.mu.Unlock()
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

func testAccMonitorResourceConfigWithRequiredKeyword(baseURL, keyword string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name             = "keyword-test"
  url              = "https://api.example.com/health"
  http_method      = "POST"
  request_body     = jsonencode({ service = "test" })
  required_keyword = %[2]q
}
`, baseURL, keyword)
}

// testAccProviderConfig returns a provider configuration block pointing at the given baseURL.
// Used by acceptance tests that spin up a mock Hyperping server.
func testAccProviderConfig(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}
`, baseURL)
}

// tfInt formats an integer value for use in HCL configuration strings.
func tfInt(val int) string {
	return fmt.Sprintf("%d", val)
}

// DNS monitor test configs

func testAccMonitorResourceConfigDNSBasic(baseURL string) string {
	return testAccProviderConfig(baseURL) + `
resource "hyperping_monitor" "test" {
  name     = "dns-basic"
  url      = "example.com"
  protocol = "dns"
}
`
}

func testAccMonitorResourceConfigDNSFull(baseURL string) string {
	return testAccProviderConfig(baseURL) + `
resource "hyperping_monitor" "test" {
  name                = "dns-full"
  url                 = "example.com"
  protocol            = "dns"
  dns_record_type     = "CNAME"
  dns_nameserver      = "8.8.8.8"
  dns_expected_answer = "www.example.com"
  check_frequency     = 300
  regions             = ["london", "virginia"]
}
`
}

func testAccMonitorResourceConfigDNSUpdate(baseURL string) string {
	return testAccProviderConfig(baseURL) + `
resource "hyperping_monitor" "test" {
  name                = "dns-updated"
  url                 = "example.com"
  protocol            = "dns"
  dns_record_type     = "MX"
  dns_nameserver      = "1.1.1.1"
  dns_expected_answer = "mail.example.com"
  check_frequency     = 600
}
`
}

func testAccMonitorResourceConfigDNSRecordTypeOnly(baseURL string) string {
	return testAccProviderConfig(baseURL) + `
resource "hyperping_monitor" "test" {
  name            = "dns-record-type-only"
  url             = "example.com"
  protocol        = "dns"
  dns_record_type = "AAAA"
}
`
}

func testAccMonitorResourceConfigDNSWithRecordType(baseURL, recordType string) string {
	return testAccProviderConfig(baseURL) + fmt.Sprintf(`
resource "hyperping_monitor" "test" {
  name            = "dns-record-type-test"
  url             = "example.com"
  protocol        = "dns"
  dns_record_type = %q
}
`, recordType)
}

func testAccMonitorResourceConfigSwitchDNSToHTTP(baseURL string) string {
	return testAccProviderConfig(baseURL) + `
resource "hyperping_monitor" "test" {
  name     = "dns-switched-to-http"
  url      = "https://example.com"
  protocol = "http"
}
`
}
