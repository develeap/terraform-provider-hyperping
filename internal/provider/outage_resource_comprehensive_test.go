// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// TestAccOutageResource_escalationPolicy tests the escalation_policy_uuid field
// which has 0% coverage. This test ensures:
// 1. Outage can be created with escalation_policy_uuid
// 2. Field persists correctly in state
// 3. Field can be updated (requires recreation due to ForceNew)
// 4. Field can be set to null/cleared
func TestAccOutageResource_escalationPolicy(t *testing.T) {
	server := newMockOutageServerWithEscalation(t)
	defer server.Close()

	now := time.Now().UTC()
	startDate := now.Add(-1 * time.Hour).Format(time.RFC3339)
	endDate := now.Format(time.RFC3339)

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create with escalation_policy_uuid
			{
				Config: testAccOutageResourceConfig_withEscalationPolicy(server.URL, startDate, endDate, "ep_test_123"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "escalation_policy_uuid", "ep_test_123"),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "monitor_uuid", "mon_escalation_test"),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "status_code", "503"),
				),
			},
			// Update escalation_policy_uuid (triggers ForceNew)
			{
				Config: testAccOutageResourceConfig_withEscalationPolicy(server.URL, startDate, endDate, "ep_test_456"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "escalation_policy_uuid", "ep_test_456"),
				),
			},
			// Clear escalation_policy_uuid (omit from config)
			{
				Config: testAccOutageResourceConfig_withoutEscalationPolicy(server.URL, startDate, endDate),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "monitor_uuid", "mon_escalation_test"),
					tfresource.TestCheckNoResourceAttr("hyperping_outage.test", "escalation_policy_uuid"),
				),
			},
		},
	})
}

// TestAccOutageResource_statusCodeRanges tests various HTTP status code ranges
// to ensure proper handling across all valid status code categories.
// Tests: 1xx (informational), 2xx (success), 3xx (redirect), 4xx (client error), 5xx (server error)
func TestAccOutageResource_statusCodeRanges(t *testing.T) {
	server := newMockOutageServerWithStatusCodes(t)
	defer server.Close()

	now := time.Now().UTC()
	startDate := now.Add(-1 * time.Hour).Format(time.RFC3339)
	endDate := now.Format(time.RFC3339)

	statusCodeTests := []struct {
		code        int
		description string
	}{
		{100, "Continue"},
		{101, "Switching Protocols"},
		{102, "Processing"},
		{200, "OK"},
		{201, "Created"},
		{204, "No Content"},
		{301, "Moved Permanently"},
		{302, "Found"},
		{307, "Temporary Redirect"},
		{400, "Bad Request"},
		{401, "Unauthorized"},
		{403, "Forbidden"},
		{404, "Not Found"},
		{500, "Internal Server Error"},
		{502, "Bad Gateway"},
		{503, "Service Unavailable"},
		{504, "Gateway Timeout"},
	}

	steps := make([]tfresource.TestStep, 0, len(statusCodeTests))

	for _, tc := range statusCodeTests {
		steps = append(steps, tfresource.TestStep{
			Config: testAccOutageResourceConfig_withStatusCode(server.URL, startDate, endDate, tc.code, tc.description),
			Check: tfresource.ComposeAggregateTestCheckFunc(
				tfresource.TestCheckResourceAttr("hyperping_outage.test", "status_code", fmt.Sprintf("%d", tc.code)),
				tfresource.TestCheckResourceAttr("hyperping_outage.test", "description", tc.description),
			),
		})
	}

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    steps,
	})
}

// TestAccOutageResource_computedFields verifies that computed fields are correctly
// populated from the API response. Tests:
// - duration_ms (may be null for ongoing outages)
// - detected_location (should always be set)
// - outage_type (should be "manual" for created outages)
// - is_resolved (false for ongoing, true for resolved)
func TestAccOutageResource_computedFields(t *testing.T) {
	server := newMockOutageServerWithComputed(t)
	defer server.Close()

	now := time.Now().UTC()
	startDate := now.Add(-2 * time.Hour).Format(time.RFC3339)
	endDate := now.Add(-1 * time.Hour).Format(time.RFC3339)

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Ongoing outage (no end_date)
			{
				Config: testAccOutageResourceConfig_computedOngoing(server.URL, startDate),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "outage_type", "manual"),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "is_resolved", "false"),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "detected_location", "us-east-1"),
					tfresource.TestCheckResourceAttrSet("hyperping_outage.test", "duration_ms"),
				),
			},
			// Resolved outage (with end_date)
			{
				Config: testAccOutageResourceConfig_computedResolved(server.URL, startDate, endDate),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "outage_type", "manual"),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "is_resolved", "true"),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "detected_location", "eu-west-1"),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "duration_ms", "3600000"), // 1 hour in ms
				),
			},
		},
	})
}

// TestAccOutageResource_nestedComputedObjects verifies that nested computed objects
// (monitor and acknowledged_by) are correctly mapped from the API response.
// Tests:
// - monitor object with uuid, name, url, protocol
// - acknowledged_by object (may be null if not acknowledged)
func TestAccOutageResource_nestedComputedObjects(t *testing.T) {
	server := newMockOutageServerWithNested(t)
	defer server.Close()

	now := time.Now().UTC()
	startDate := now.Add(-1 * time.Hour).Format(time.RFC3339)
	endDate := now.Format(time.RFC3339)

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Outage with monitor object
			{
				Config: testAccOutageResourceConfig_nestedObjects(server.URL, startDate, endDate),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					// Verify monitor nested object
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "monitor.uuid", "mon_nested_test"),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "monitor.name", "Production API"),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "monitor.url", "https://api.production.com"),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "monitor.protocol", "https"),
					// acknowledged_by should be null for new outages
					tfresource.TestCheckNoResourceAttr("hyperping_outage.test", "acknowledged_by"),
				),
			},
			// Test with acknowledged outage
			{
				Config: testAccOutageResourceConfig_acknowledged(server.URL, startDate, endDate),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					// Verify acknowledged_by nested object
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "acknowledged_by.uuid", "user_123"),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "acknowledged_by.email", "admin@example.com"),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "acknowledged_by.name", "Admin User"),
				),
			},
		},
	})
}

// Config helpers

func testAccOutageResourceConfig_withEscalationPolicy(baseURL, startDate, endDate, escalationPolicyUUID string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_outage" "test" {
  monitor_uuid            = "mon_escalation_test"
  start_date              = %[2]q
  end_date                = %[3]q
  status_code             = 503
  description             = "Service unavailable with escalation"
  escalation_policy_uuid  = %[4]q
}
`, baseURL, startDate, endDate, escalationPolicyUUID)
}

func testAccOutageResourceConfig_withoutEscalationPolicy(baseURL, startDate, endDate string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_outage" "test" {
  monitor_uuid = "mon_escalation_test"
  start_date   = %[2]q
  end_date     = %[3]q
  status_code  = 503
  description  = "Service unavailable without escalation"
}
`, baseURL, startDate, endDate)
}

func testAccOutageResourceConfig_withStatusCode(baseURL, startDate, endDate string, statusCode int, description string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_outage" "test" {
  monitor_uuid = "mon_status_test"
  start_date   = %[2]q
  end_date     = %[3]q
  status_code  = %[4]d
  description  = %[5]q
}
`, baseURL, startDate, endDate, statusCode, description)
}

func testAccOutageResourceConfig_computedOngoing(baseURL, startDate string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_outage" "test" {
  monitor_uuid = "mon_computed_test"
  start_date   = %[2]q
  status_code  = 500
  description  = "Ongoing outage for computed field testing"
}
`, baseURL, startDate)
}

func testAccOutageResourceConfig_computedResolved(baseURL, startDate, endDate string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_outage" "test" {
  monitor_uuid = "mon_computed_test"
  start_date   = %[2]q
  end_date     = %[3]q
  status_code  = 500
  description  = "Resolved outage for computed field testing"
}
`, baseURL, startDate, endDate)
}

func testAccOutageResourceConfig_nestedObjects(baseURL, startDate, endDate string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_outage" "test" {
  monitor_uuid = "mon_nested_test"
  start_date   = %[2]q
  end_date     = %[3]q
  status_code  = 503
  description  = "Testing nested monitor object"
}
`, baseURL, startDate, endDate)
}

func testAccOutageResourceConfig_acknowledged(baseURL, startDate, endDate string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_outage" "test" {
  monitor_uuid = "mon_acknowledged_test"
  start_date   = %[2]q
  end_date     = %[3]q
  status_code  = 500
  description  = "Testing acknowledged_by object"
}
`, baseURL, startDate, endDate)
}

// Mock server implementations

type mockOutageServerWithEscalation struct {
	*httptest.Server
	t       *testing.T
	outages map[string]map[string]interface{}
	counter int
}

func newMockOutageServerWithEscalation(t *testing.T) *mockOutageServerWithEscalation {
	m := &mockOutageServerWithEscalation{
		t:       t,
		outages: make(map[string]map[string]interface{}),
		counter: 0,
	}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handleRequest(w, r)
	}))

	return m
}

func (m *mockOutageServerWithEscalation) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)
	basePath := client.OutagesBasePath
	basePathWithSlash := basePath + "/"

	switch {
	case r.Method == "POST" && r.URL.Path == basePath:
		m.createOutage(w, r)
	case r.Method == "GET" && strings.HasPrefix(r.URL.Path, basePathWithSlash):
		m.getOutage(w, r)
	case r.Method == "DELETE" && strings.HasPrefix(r.URL.Path, basePathWithSlash):
		m.deleteOutage(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}
}

func (m *mockOutageServerWithEscalation) createOutage(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	m.counter++
	id := fmt.Sprintf("outage_esc_%d", m.counter)

	outage := map[string]interface{}{
		"uuid":             id,
		"startDate":        req["startDate"],
		"outageType":       "manual",
		"isResolved":       false,
		"durationMs":       0,
		"detectedLocation": "test-location",
		"monitor": map[string]interface{}{
			"uuid":     req["monitorUuid"],
			"name":     "Test Monitor",
			"url":      "https://example.com",
			"protocol": "https",
		},
	}

	if endDate, ok := req["endDate"].(string); ok {
		outage["endDate"] = endDate
		outage["isResolved"] = true
	}

	if statusCode, ok := req["statusCode"].(float64); ok {
		outage["statusCode"] = int(statusCode)
	}

	if description, ok := req["description"].(string); ok {
		outage["description"] = description
	}

	// Handle escalation_policy_uuid
	if escalationPolicyUUID, ok := req["escalationPolicyUuid"].(string); ok && escalationPolicyUUID != "" {
		outage["escalationPolicy"] = map[string]interface{}{
			"uuid":         escalationPolicyUUID,
			"name":         "Test Escalation Policy",
			"alertedSteps": 0,
			"totalSteps":   3,
		}
	}

	m.outages[id] = outage

	response := map[string]interface{}{
		"message": "Incident created",
		"outage":  outage,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (m *mockOutageServerWithEscalation) getOutage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, client.OutagesBasePath+"/")

	outage, exists := m.outages[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Outage not found"})
		return
	}

	response := map[string]interface{}{
		"outage": outage,
	}
	json.NewEncoder(w).Encode(response)
}

func (m *mockOutageServerWithEscalation) deleteOutage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, client.OutagesBasePath+"/")

	if _, exists := m.outages[id]; !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Outage not found"})
		return
	}

	delete(m.outages, id)
	w.WriteHeader(http.StatusNoContent)
}

type mockOutageServerWithStatusCodes struct {
	*httptest.Server
	t       *testing.T
	outages map[string]map[string]interface{}
	counter int
}

func newMockOutageServerWithStatusCodes(t *testing.T) *mockOutageServerWithStatusCodes {
	m := &mockOutageServerWithStatusCodes{
		t:       t,
		outages: make(map[string]map[string]interface{}),
		counter: 0,
	}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handleRequest(w, r)
	}))

	return m
}

func (m *mockOutageServerWithStatusCodes) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)
	basePath := client.OutagesBasePath
	basePathWithSlash := basePath + "/"

	switch {
	case r.Method == "POST" && r.URL.Path == basePath:
		m.createOutage(w, r)
	case r.Method == "GET" && strings.HasPrefix(r.URL.Path, basePathWithSlash):
		m.getOutage(w, r)
	case r.Method == "DELETE" && strings.HasPrefix(r.URL.Path, basePathWithSlash):
		m.deleteOutage(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}
}

func (m *mockOutageServerWithStatusCodes) createOutage(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	m.counter++
	id := fmt.Sprintf("outage_sc_%d", m.counter)

	outage := map[string]interface{}{
		"uuid":             id,
		"startDate":        req["startDate"],
		"outageType":       "manual",
		"isResolved":       true,
		"durationMs":       3600000,
		"detectedLocation": "test-location",
		"monitor": map[string]interface{}{
			"uuid":     req["monitorUuid"],
			"name":     "Test Monitor",
			"url":      "https://example.com",
			"protocol": "https",
		},
	}

	if endDate, ok := req["endDate"].(string); ok {
		outage["endDate"] = endDate
	}

	if statusCode, ok := req["statusCode"].(float64); ok {
		outage["statusCode"] = int(statusCode)
	}

	if description, ok := req["description"].(string); ok {
		outage["description"] = description
	}

	m.outages[id] = outage

	response := map[string]interface{}{
		"message": "Incident created",
		"outage":  outage,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (m *mockOutageServerWithStatusCodes) getOutage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, client.OutagesBasePath+"/")

	outage, exists := m.outages[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Outage not found"})
		return
	}

	response := map[string]interface{}{
		"outage": outage,
	}
	json.NewEncoder(w).Encode(response)
}

func (m *mockOutageServerWithStatusCodes) deleteOutage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, client.OutagesBasePath+"/")

	if _, exists := m.outages[id]; !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Outage not found"})
		return
	}

	delete(m.outages, id)
	w.WriteHeader(http.StatusNoContent)
}

type mockOutageServerWithComputed struct {
	*httptest.Server
	t       *testing.T
	outages map[string]map[string]interface{}
	counter int
}

func newMockOutageServerWithComputed(t *testing.T) *mockOutageServerWithComputed {
	m := &mockOutageServerWithComputed{
		t:       t,
		outages: make(map[string]map[string]interface{}),
		counter: 0,
	}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handleRequest(w, r)
	}))

	return m
}

func (m *mockOutageServerWithComputed) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)
	basePath := client.OutagesBasePath
	basePathWithSlash := basePath + "/"

	switch {
	case r.Method == "POST" && r.URL.Path == basePath:
		m.createOutage(w, r)
	case r.Method == "GET" && strings.HasPrefix(r.URL.Path, basePathWithSlash):
		m.getOutage(w, r)
	case r.Method == "DELETE" && strings.HasPrefix(r.URL.Path, basePathWithSlash):
		m.deleteOutage(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}
}

func (m *mockOutageServerWithComputed) createOutage(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	m.counter++
	id := fmt.Sprintf("outage_comp_%d", m.counter)

	hasEndDate := false
	if _, ok := req["endDate"].(string); ok {
		hasEndDate = true
	}

	location := "us-east-1"
	durationMs := 1800000 // 30 minutes for ongoing
	isResolved := false

	if hasEndDate {
		location = "eu-west-1"
		durationMs = 3600000 // 1 hour for resolved
		isResolved = true
	}

	outage := map[string]interface{}{
		"uuid":             id,
		"startDate":        req["startDate"],
		"outageType":       "manual",
		"isResolved":       isResolved,
		"durationMs":       durationMs,
		"detectedLocation": location,
		"monitor": map[string]interface{}{
			"uuid":     req["monitorUuid"],
			"name":     "Test Monitor",
			"url":      "https://example.com",
			"protocol": "https",
		},
	}

	if endDate, ok := req["endDate"].(string); ok {
		outage["endDate"] = endDate
	}

	if statusCode, ok := req["statusCode"].(float64); ok {
		outage["statusCode"] = int(statusCode)
	}

	if description, ok := req["description"].(string); ok {
		outage["description"] = description
	}

	m.outages[id] = outage

	response := map[string]interface{}{
		"message": "Incident created",
		"outage":  outage,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (m *mockOutageServerWithComputed) getOutage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, client.OutagesBasePath+"/")

	outage, exists := m.outages[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Outage not found"})
		return
	}

	response := map[string]interface{}{
		"outage": outage,
	}
	json.NewEncoder(w).Encode(response)
}

func (m *mockOutageServerWithComputed) deleteOutage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, client.OutagesBasePath+"/")

	if _, exists := m.outages[id]; !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Outage not found"})
		return
	}

	delete(m.outages, id)
	w.WriteHeader(http.StatusNoContent)
}

type mockOutageServerWithNested struct {
	*httptest.Server
	t       *testing.T
	outages map[string]map[string]interface{}
	counter int
}

func newMockOutageServerWithNested(t *testing.T) *mockOutageServerWithNested {
	m := &mockOutageServerWithNested{
		t:       t,
		outages: make(map[string]map[string]interface{}),
		counter: 0,
	}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handleRequest(w, r)
	}))

	return m
}

func (m *mockOutageServerWithNested) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)
	basePath := client.OutagesBasePath
	basePathWithSlash := basePath + "/"

	switch {
	case r.Method == "POST" && r.URL.Path == basePath:
		m.createOutage(w, r)
	case r.Method == "GET" && strings.HasPrefix(r.URL.Path, basePathWithSlash):
		m.getOutage(w, r)
	case r.Method == "DELETE" && strings.HasPrefix(r.URL.Path, basePathWithSlash):
		m.deleteOutage(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}
}

func (m *mockOutageServerWithNested) createOutage(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	m.counter++
	id := fmt.Sprintf("outage_nested_%d", m.counter)

	monitorUUID := req["monitorUuid"].(string)
	isAcknowledged := monitorUUID == "mon_acknowledged_test"

	outage := map[string]interface{}{
		"uuid":             id,
		"startDate":        req["startDate"],
		"outageType":       "manual",
		"isResolved":       true,
		"durationMs":       3600000,
		"detectedLocation": "test-location",
		"monitor": map[string]interface{}{
			"uuid":     monitorUUID,
			"name":     "Production API",
			"url":      "https://api.production.com",
			"protocol": "https",
		},
	}

	if endDate, ok := req["endDate"].(string); ok {
		outage["endDate"] = endDate
	}

	if statusCode, ok := req["statusCode"].(float64); ok {
		outage["statusCode"] = int(statusCode)
	}

	if description, ok := req["description"].(string); ok {
		outage["description"] = description
	}

	// Add acknowledged_by for acknowledged outages
	if isAcknowledged {
		outage["acknowledgedBy"] = map[string]interface{}{
			"uuid":  "user_123",
			"email": "admin@example.com",
			"name":  "Admin User",
		}
	}

	m.outages[id] = outage

	response := map[string]interface{}{
		"message": "Incident created",
		"outage":  outage,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (m *mockOutageServerWithNested) getOutage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, client.OutagesBasePath+"/")

	outage, exists := m.outages[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Outage not found"})
		return
	}

	response := map[string]interface{}{
		"outage": outage,
	}
	json.NewEncoder(w).Encode(response)
}

func (m *mockOutageServerWithNested) deleteOutage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, client.OutagesBasePath+"/")

	if _, exists := m.outages[id]; !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Outage not found"})
		return
	}

	delete(m.outages, id)
	w.WriteHeader(http.StatusNoContent)
}
