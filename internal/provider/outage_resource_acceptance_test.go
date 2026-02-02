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

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccOutageResource_basic(t *testing.T) {
	server := newMockOutageServer(t)
	defer server.Close()

	now := time.Now().UTC()
	startDate := now.Add(-1 * time.Hour).Format(time.RFC3339)
	endDate := now.Format(time.RFC3339)

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create and Read testing
			{
				Config: testAccOutageResourceConfig_basic(server.URL, startDate, endDate),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "monitor_uuid", "mon_test123"),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "start_date", startDate),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "end_date", endDate),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "outage_type", "manual"),
					tfresource.TestCheckResourceAttrSet("hyperping_outage.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "hyperping_outage.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOutageResource_full(t *testing.T) {
	server := newMockOutageServer(t)
	defer server.Close()

	now := time.Now().UTC()
	startDate := now.Add(-2 * time.Hour).Format(time.RFC3339)
	endDate := now.Add(-1 * time.Hour).Format(time.RFC3339)

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccOutageResourceConfig_full(server.URL, startDate, endDate),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "monitor_uuid", "mon_test456"),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "start_date", startDate),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "end_date", endDate),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "status_code", "500"),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "description", "Server error"),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "outage_type", "manual"),
				),
			},
		},
	})
}

func TestAccOutageResource_ongoing(t *testing.T) {
	server := newMockOutageServer(t)
	defer server.Close()

	now := time.Now().UTC()
	startDate := now.Add(-1 * time.Hour).Format(time.RFC3339)

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccOutageResourceConfig_ongoing(server.URL, startDate),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "monitor_uuid", "mon_test789"),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "start_date", startDate),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "is_resolved", "false"),
				),
			},
		},
	})
}

func TestAccOutageResource_disappears(t *testing.T) {
	server := newMockOutageServer(t)
	defer server.Close()

	now := time.Now().UTC()
	startDate := now.Add(-1 * time.Hour).Format(time.RFC3339)
	endDate := now.Format(time.RFC3339)

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccOutageResourceConfig_basic(server.URL, startDate, endDate),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttrSet("hyperping_outage.test", "id"),
					testAccCheckOutageDisappears(server),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccOutageResource_createError(t *testing.T) {
	server := newMockOutageServerWithErrors(t)
	defer server.Close()

	server.setCreateError(true)

	now := time.Now().UTC()
	startDate := now.Add(-1 * time.Hour).Format(time.RFC3339)
	endDate := now.Format(time.RFC3339)

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccOutageResourceConfig_basic(server.URL, startDate, endDate),
				ExpectError: regexp.MustCompile(`Could not create outage`),
			},
		},
	})
}

// Helper functions

func testAccOutageResourceConfig_basic(baseURL, startDate, endDate string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_outage" "test" {
  monitor_uuid = "mon_test123"
  start_date   = %[2]q
  end_date     = %[3]q
  status_code  = 503
  description  = "Service unavailable"
}
`, baseURL, startDate, endDate)
}

func testAccOutageResourceConfig_full(baseURL, startDate, endDate string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_outage" "test" {
  monitor_uuid = "mon_test456"
  start_date   = %[2]q
  end_date     = %[3]q
  status_code  = 500
  description  = "Server error"
}
`, baseURL, startDate, endDate)
}

func testAccOutageResourceConfig_ongoing(baseURL, startDate string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_outage" "test" {
  monitor_uuid = "mon_test789"
  start_date   = %[2]q
  status_code  = 504
  description  = "Gateway timeout"
}
`, baseURL, startDate)
}

func testAccCheckOutageDisappears(server *mockOutageServer) tfresource.TestCheckFunc {
	return func(s *terraform.State) error {
		server.deleteAllOutages()
		return nil
	}
}

// Mock server implementation

type mockOutageServer struct {
	*httptest.Server
	t       *testing.T
	outages map[string]map[string]interface{}
	counter int
}

func newMockOutageServer(t *testing.T) *mockOutageServer {
	m := &mockOutageServer{
		t:       t,
		outages: make(map[string]map[string]interface{}),
		counter: 0,
	}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handleRequest(w, r)
	}))

	return m
}

func (m *mockOutageServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == "GET" && r.URL.Path == "/v1/outages":
		m.listOutages(w)
	case r.Method == "POST" && r.URL.Path == "/v1/outages":
		m.createOutage(w, r)
	case r.Method == "GET" && len(r.URL.Path) > len("/v1/outages/"):
		m.getOutage(w, r)
	case r.Method == "DELETE" && len(r.URL.Path) > len("/v1/outages/"):
		m.deleteOutage(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}
}

func (m *mockOutageServer) listOutages(w http.ResponseWriter) {
	outages := make([]map[string]interface{}, 0, len(m.outages))
	for _, outage := range m.outages {
		outages = append(outages, outage)
	}
	json.NewEncoder(w).Encode(outages)
}

func (m *mockOutageServer) createOutage(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	m.counter++
	id := fmt.Sprintf("outage_%d", m.counter)

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

	m.outages[id] = outage

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(outage)
}

func (m *mockOutageServer) getOutage(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/v1/outages/"):]

	outage, exists := m.outages[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Outage not found"})
		return
	}

	json.NewEncoder(w).Encode(outage)
}

func (m *mockOutageServer) deleteOutage(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/v1/outages/"):]

	if _, exists := m.outages[id]; !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Outage not found"})
		return
	}

	delete(m.outages, id)
	w.WriteHeader(http.StatusNoContent)
}

func (m *mockOutageServer) deleteAllOutages() {
	m.outages = make(map[string]map[string]interface{})
}

// Mock server with error injection

type mockOutageServerWithErrors struct {
	*mockOutageServer
	createError bool
	readError   bool
	deleteError bool
}

func newMockOutageServerWithErrors(t *testing.T) *mockOutageServerWithErrors {
	m := &mockOutageServerWithErrors{
		mockOutageServer: &mockOutageServer{
			t:       t,
			outages: make(map[string]map[string]interface{}),
			counter: 0,
		},
	}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handleRequestWithErrors(w, r)
	}))

	return m
}

func (m *mockOutageServerWithErrors) setCreateError(v bool) { m.createError = v }
func (m *mockOutageServerWithErrors) setReadError(v bool)   { m.readError = v }
func (m *mockOutageServerWithErrors) setDeleteError(v bool) { m.deleteError = v }

func (m *mockOutageServerWithErrors) handleRequestWithErrors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == "POST" && r.URL.Path == "/v1/outages":
		if m.createError {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
			return
		}
		m.createOutage(w, r)

	case r.Method == "GET" && len(r.URL.Path) > len("/v1/outages/"):
		if m.readError {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
			return
		}
		m.getOutage(w, r)

	case r.Method == "DELETE" && len(r.URL.Path) > len("/v1/outages/"):
		if m.deleteError {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
			return
		}
		m.deleteOutage(w, r)

	default:
		m.handleRequest(w, r)
	}
}
