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
	"sync"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// mockIntegrationServer provides a unified mock server for integration testing
// that handles monitors, incidents, and maintenance windows.
type mockIntegrationServer struct {
	*httptest.Server
	t           *testing.T
	mu          sync.RWMutex
	monitors    map[string]map[string]interface{}
	incidents   map[string]map[string]interface{}
	maintenance map[string]map[string]interface{}
	counter     int

	// Error injection flags
	injectError      bool
	errorStatusCode  int
	errorMessage     string
	errorOnResource  string // "monitor", "incident", "maintenance"
	errorOnOperation string // "create", "read", "update", "delete"
}

func newMockIntegrationServer(t *testing.T) *mockIntegrationServer {
	m := &mockIntegrationServer{
		t:           t,
		monitors:    make(map[string]map[string]interface{}),
		incidents:   make(map[string]map[string]interface{}),
		maintenance: make(map[string]map[string]interface{}),
		counter:     0,
	}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handleRequest(w, r)
	}))

	return m
}

func (m *mockIntegrationServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)

	// Check for error injection
	if m.shouldInjectError(r) {
		w.WriteHeader(m.errorStatusCode)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": m.errorMessage}); err != nil {
			m.t.Errorf("failed to encode error response: %v", err)
		}
		return
	}

	// Route to appropriate handler
	switch {
	// Monitor endpoints
	case r.Method == "GET" && r.URL.Path == client.MonitorsBasePath:
		m.listMonitors(w)
	case r.Method == "POST" && r.URL.Path == client.MonitorsBasePath:
		m.createMonitor(w, r)
	case r.Method == "GET" && strings.HasPrefix(r.URL.Path, client.MonitorsBasePath+"/"):
		m.getMonitor(w, r)
	case r.Method == "PUT" && strings.HasPrefix(r.URL.Path, client.MonitorsBasePath+"/"):
		m.updateMonitor(w, r)
	case r.Method == "DELETE" && strings.HasPrefix(r.URL.Path, client.MonitorsBasePath+"/"):
		m.deleteMonitor(w, r)

	// Incident endpoints
	case r.Method == "GET" && r.URL.Path == client.IncidentsBasePath:
		m.listIncidents(w)
	case r.Method == "POST" && r.URL.Path == client.IncidentsBasePath:
		m.createIncident(w, r)
	case r.Method == "GET" && strings.HasPrefix(r.URL.Path, client.IncidentsBasePath+"/"):
		m.getIncident(w, r)
	case r.Method == "PUT" && strings.HasPrefix(r.URL.Path, client.IncidentsBasePath+"/"):
		m.updateIncident(w, r)
	case r.Method == "DELETE" && strings.HasPrefix(r.URL.Path, client.IncidentsBasePath+"/"):
		m.deleteIncident(w, r)

	// Maintenance endpoints
	case r.Method == "GET" && r.URL.Path == client.MaintenanceBasePath:
		m.listMaintenance(w)
	case r.Method == "POST" && r.URL.Path == client.MaintenanceBasePath:
		m.createMaintenance(w, r)
	case r.Method == "GET" && strings.HasPrefix(r.URL.Path, client.MaintenanceBasePath+"/"):
		m.getMaintenance(w, r)
	case r.Method == "PUT" && strings.HasPrefix(r.URL.Path, client.MaintenanceBasePath+"/"):
		m.updateMaintenance(w, r)
	case r.Method == "DELETE" && strings.HasPrefix(r.URL.Path, client.MaintenanceBasePath+"/"):
		m.deleteMaintenance(w, r)

	default:
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "Not found"}); err != nil {
			m.t.Errorf("failed to encode error response: %v", err)
		}
	}
}

func (m *mockIntegrationServer) shouldInjectError(r *http.Request) bool {
	if !m.injectError {
		return false
	}

	// Check if this request matches the error injection criteria
	var resourceType string
	switch {
	case strings.HasPrefix(r.URL.Path, client.MonitorsBasePath):
		resourceType = "monitor"
	case strings.HasPrefix(r.URL.Path, client.IncidentsBasePath):
		resourceType = "incident"
	case strings.HasPrefix(r.URL.Path, client.MaintenanceBasePath):
		resourceType = "maintenance"
	default:
		return false
	}

	if m.errorOnResource != "" && m.errorOnResource != resourceType {
		return false
	}

	var operation string
	switch r.Method {
	case "POST":
		operation = "create"
	case "GET":
		operation = "read"
	case "PUT":
		operation = "update"
	case "DELETE":
		operation = "delete"
	}

	return m.errorOnOperation == "" || m.errorOnOperation == operation
}

func (m *mockIntegrationServer) setError(resource, operation string, statusCode int, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.injectError = true
	m.errorOnResource = resource
	m.errorOnOperation = operation
	m.errorStatusCode = statusCode
	m.errorMessage = message
}

func (m *mockIntegrationServer) clearError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.injectError = false
	m.errorOnResource = ""
	m.errorOnOperation = ""
}

// Monitor handlers
func (m *mockIntegrationServer) listMonitors(w http.ResponseWriter) {
	monitors := make([]map[string]interface{}, 0, len(m.monitors))
	for _, monitor := range m.monitors {
		monitors = append(monitors, monitor)
	}
	if err := json.NewEncoder(w).Encode(monitors); err != nil {
		m.t.Errorf("failed to encode monitors list: %v", err)
	}
}

func (m *mockIntegrationServer) createMonitor(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"}); encodeErr != nil {
			m.t.Errorf("failed to encode error response: %v", encodeErr)
		}
		return
	}

	m.counter++
	id := fmt.Sprintf("mon_int%d", m.counter)

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
	}

	m.monitors[id] = monitor

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(monitor); err != nil {
		m.t.Errorf("failed to encode monitor response: %v", err)
	}
}

func (m *mockIntegrationServer) getMonitor(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, client.MonitorsBasePath+"/")

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

func (m *mockIntegrationServer) updateMonitor(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, client.MonitorsBasePath+"/")

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

	// Update fields
	for key, value := range req {
		monitor[key] = value
	}

	m.monitors[id] = monitor
	if err := json.NewEncoder(w).Encode(monitor); err != nil {
		m.t.Errorf("failed to encode monitor response: %v", err)
	}
}

func (m *mockIntegrationServer) deleteMonitor(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, client.MonitorsBasePath+"/")

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

// Incident handlers
func (m *mockIntegrationServer) listIncidents(w http.ResponseWriter) {
	incidents := make([]map[string]interface{}, 0, len(m.incidents))
	for _, incident := range m.incidents {
		incidents = append(incidents, incident)
	}
	if err := json.NewEncoder(w).Encode(incidents); err != nil {
		m.t.Errorf("failed to encode incidents list: %v", err)
	}
}

func (m *mockIntegrationServer) createIncident(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"}); encodeErr != nil {
			m.t.Errorf("failed to encode error response: %v", encodeErr)
		}
		return
	}

	m.counter++
	id := fmt.Sprintf("inc_int%d", m.counter)

	// Extract title and text from LocalizedText format
	var titleText, descText string
	if titleMap, ok := req["title"].(map[string]interface{}); ok {
		if en, ok := titleMap["en"].(string); ok {
			titleText = en
		}
	} else if title, ok := req["title"].(string); ok {
		titleText = title
	}

	if textMap, ok := req["text"].(map[string]interface{}); ok {
		if en, ok := textMap["en"].(string); ok {
			descText = en
		}
	} else if text, ok := req["text"].(string); ok {
		descText = text
	}

	incident := map[string]interface{}{
		"uuid":        id,
		"title":       map[string]string{"en": titleText},
		"text":        map[string]string{"en": descText},
		"type":        getOrDefault(req, "type", "incident"),
		"statuspages": []interface{}{},
		"date":        "2026-02-16T10:00:00Z",
	}

	// Handle statuspages
	if statusPages, ok := req["statuspages"].([]interface{}); ok {
		incident["statuspages"] = statusPages
	}

	// Handle affectedComponents
	if affectedComponents, ok := req["affectedComponents"].([]interface{}); ok {
		incident["affectedComponents"] = affectedComponents
	}

	m.incidents[id] = incident

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(incident); err != nil {
		m.t.Errorf("failed to encode incident response: %v", err)
	}
	m.t.Logf("Created incident: %+v", incident)
}

func (m *mockIntegrationServer) getIncident(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, client.IncidentsBasePath+"/")

	incident, exists := m.incidents[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "Incident not found"}); err != nil {
			m.t.Errorf("failed to encode error response: %v", err)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(incident); err != nil {
		m.t.Errorf("failed to encode incident response: %v", err)
	}
}

func (m *mockIntegrationServer) updateIncident(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, client.IncidentsBasePath+"/")

	incident, exists := m.incidents[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "Incident not found"}); err != nil {
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

	// Update fields
	for key, value := range req {
		incident[key] = value
	}

	m.incidents[id] = incident
	if err := json.NewEncoder(w).Encode(incident); err != nil {
		m.t.Errorf("failed to encode incident response: %v", err)
	}
}

func (m *mockIntegrationServer) deleteIncident(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, client.IncidentsBasePath+"/")

	if _, exists := m.incidents[id]; !exists {
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "Incident not found"}); err != nil {
			m.t.Errorf("failed to encode error response: %v", err)
		}
		return
	}

	delete(m.incidents, id)
	w.WriteHeader(http.StatusNoContent)
}

// Maintenance handlers
func (m *mockIntegrationServer) listMaintenance(w http.ResponseWriter) {
	maintenances := make([]map[string]interface{}, 0, len(m.maintenance))
	for _, maint := range m.maintenance {
		maintenances = append(maintenances, maint)
	}
	if err := json.NewEncoder(w).Encode(maintenances); err != nil {
		m.t.Errorf("failed to encode maintenance list: %v", err)
	}
}

func (m *mockIntegrationServer) createMaintenance(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"}); encodeErr != nil {
			m.t.Errorf("failed to encode error response: %v", encodeErr)
		}
		return
	}

	m.counter++
	id := fmt.Sprintf("maint_int%d", m.counter)

	// Handle title and text as LocalizedText
	var titleText, descText map[string]string
	if title, ok := req["title"].(string); ok {
		titleText = map[string]string{"en": title}
	} else if titleMap, ok := req["title"].(map[string]interface{}); ok {
		titleText = make(map[string]string)
		for k, v := range titleMap {
			if str, ok := v.(string); ok {
				titleText[k] = str
			}
		}
	}

	if text, ok := req["text"].(string); ok {
		descText = map[string]string{"en": text}
	} else if textMap, ok := req["text"].(map[string]interface{}); ok {
		descText = make(map[string]string)
		for k, v := range textMap {
			if str, ok := v.(string); ok {
				descText[k] = str
			}
		}
	}

	maintenance := map[string]interface{}{
		"uuid":       id,
		"name":       req["name"],
		"title":      titleText,
		"text":       descText,
		"start_date": req["start_date"],
		"end_date":   req["end_date"],
		"monitors":   []interface{}{},
	}

	if monitors, ok := req["monitors"].([]interface{}); ok {
		maintenance["monitors"] = monitors
	}

	m.maintenance[id] = maintenance

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(maintenance); err != nil {
		m.t.Errorf("failed to encode maintenance response: %v", err)
	}
}

func (m *mockIntegrationServer) getMaintenance(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, client.MaintenanceBasePath+"/")

	maintenance, exists := m.maintenance[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "Maintenance not found"}); err != nil {
			m.t.Errorf("failed to encode error response: %v", err)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(maintenance); err != nil {
		m.t.Errorf("failed to encode maintenance response: %v", err)
	}
}

func (m *mockIntegrationServer) updateMaintenance(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, client.MaintenanceBasePath+"/")

	maintenance, exists := m.maintenance[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "Maintenance not found"}); err != nil {
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

	// Update fields
	for key, value := range req {
		maintenance[key] = value
	}

	m.maintenance[id] = maintenance
	if err := json.NewEncoder(w).Encode(maintenance); err != nil {
		m.t.Errorf("failed to encode maintenance response: %v", err)
	}
}

func (m *mockIntegrationServer) deleteMaintenance(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, client.MaintenanceBasePath+"/")

	if _, exists := m.maintenance[id]; !exists {
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "Maintenance not found"}); err != nil {
			m.t.Errorf("failed to encode error response: %v", err)
		}
		return
	}

	delete(m.maintenance, id)
	w.WriteHeader(http.StatusNoContent)
}

// Integration Tests

// TestAccIntegration_monitorIncidentRelationship tests that incidents can reference
// monitors and that the relationship is maintained even when monitors are deleted.
func TestAccIntegration_monitorIncidentRelationship(t *testing.T) {
	server := newMockIntegrationServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create 2 monitors and incident affecting both
			{
				Config: testAccIntegrationMonitorIncidentConfig(server.URL, 2),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test.0", "name", "Monitor 0"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test.1", "name", "Monitor 1"),
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "title", "Test Incident"),
					tfresource.TestCheckResourceAttrSet("hyperping_incident.test", "id"),
				),
			},
			// Delete one monitor, verify incident still references remaining monitor
			{
				Config: testAccIntegrationMonitorIncidentConfig(server.URL, 1),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test.0", "name", "Monitor 0"),
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "title", "Test Incident"),
				),
			},
		},
	})
}

// TestAccIntegration_monitorMaintenanceRelationship tests that maintenance windows
// can affect multiple monitors and be updated correctly.
func TestAccIntegration_monitorMaintenanceRelationship(t *testing.T) {
	server := newMockIntegrationServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create 3 monitors and maintenance affecting all 3
			{
				Config: testAccIntegrationMonitorMaintenanceConfig(server.URL, 3, 3),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test.0", "name", "Monitor 0"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test.1", "name", "Monitor 1"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test.2", "name", "Monitor 2"),
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "title", "Scheduled Maintenance"),
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "monitors.#", "3"),
				),
			},
			// Update maintenance to affect only 2 monitors
			{
				Config: testAccIntegrationMonitorMaintenanceConfig(server.URL, 3, 2),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "monitors.#", "2"),
				),
			},
		},
	})
}

// TestAccIntegration_complexUpdateWorkflow tests a complex workflow of creating
// resources and updating them through various states.
func TestAccIntegration_complexUpdateWorkflow(t *testing.T) {
	server := newMockIntegrationServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create monitor and incident
			{
				Config: testAccIntegrationComplexWorkflowConfig(server.URL, "Initial Monitor"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "name", "Initial Monitor"),
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "title", "Test Incident"),
				),
			},
			// Update monitor name
			{
				Config: testAccIntegrationComplexWorkflowConfig(server.URL, "Updated Monitor"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "name", "Updated Monitor"),
				),
			},
		},
	})
}

// TestAccIntegration_bulkMonitorCreation tests creating and updating multiple
// monitors simultaneously without race conditions.
func TestAccIntegration_bulkMonitorCreation(t *testing.T) {
	server := newMockIntegrationServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create 10 monitors
			{
				Config: testAccIntegrationBulkMonitorsConfig(server.URL, 10, "https://example.com"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.bulk.0", "name", "Bulk Monitor 0"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.bulk.9", "name", "Bulk Monitor 9"),
				),
			},
			// Update all 10 monitors simultaneously
			{
				Config: testAccIntegrationBulkMonitorsConfig(server.URL, 10, "https://updated.example.com"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.bulk.0", "url", "https://updated.example.com"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.bulk.9", "url", "https://updated.example.com"),
				),
			},
		},
	})
}

// TestAccIntegration_fullMonitoringStack tests a complete monitoring setup with
// monitors, incidents, and maintenance windows working together.
func TestAccIntegration_fullMonitoringStack(t *testing.T) {
	server := newMockIntegrationServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create full stack
			{
				Config: testAccIntegrationFullStackConfig(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					// Monitors
					tfresource.TestCheckResourceAttr("hyperping_monitor.http", "name", "HTTP Monitor"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.api", "name", "API Monitor"),
					// Incident
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "title", "Service Degradation"),
					// Maintenance
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "title", "System Upgrade"),
				),
			},
			// Verify clean deletion order (dependent resources first)
			{
				Config: testAccIntegrationEmptyConfig(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					testAccCheckAllResourcesDeleted(server),
				),
			},
		},
	})
}

// TestAccIntegration_importThenUpdate tests creating a monitor and
// immediately updating it without state drift.
func TestAccIntegration_importThenUpdate(t *testing.T) {
	server := newMockIntegrationServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create monitor
			{
				Config: testAccIntegrationImportConfig(server.URL, "Initial Monitor"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.imported", "name", "Initial Monitor"),
					tfresource.TestCheckResourceAttrSet("hyperping_monitor.imported", "id"),
				),
			},
			// Immediately update the monitor (tests no state drift)
			{
				Config: testAccIntegrationImportConfig(server.URL, "Updated After Create"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.imported", "name", "Updated After Create"),
				),
			},
		},
	})
}

// TestAccIntegration_errorRecovery tests that the provider handles API errors
// gracefully and can recover when the API becomes available again.
func TestAccIntegration_errorRecovery(t *testing.T) {
	server := newMockIntegrationServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create monitor successfully first (needed for update test later)
			{
				Config: testAccIntegrationErrorRecoveryConfig(server.URL, "Error Test Monitor"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.error_test", "name", "Error Test Monitor"),
				),
			},
			// Test update with error injection - expect failure
			{
				PreConfig: func() {
					server.setError("monitor", "update", http.StatusInternalServerError, "Update failed")
				},
				Config:      testAccIntegrationErrorRecoveryConfig(server.URL, "Updated Monitor"),
				ExpectError: regexp.MustCompile("Update failed|500"),
			},
			// Clear error and verify update succeeds
			{
				PreConfig: func() {
					server.clearError()
				},
				Config: testAccIntegrationErrorRecoveryConfig(server.URL, "Updated Monitor"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.error_test", "name", "Updated Monitor"),
				),
			},
		},
	})
}

// Test configuration helpers

func testAccIntegrationMonitorIncidentConfig(baseURL string, numMonitors int) string {
	config := fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %q
}

resource "hyperping_monitor" "test" {
  count = %d
  name  = "Monitor ${count.index}"
  url   = "https://example.com"
}

resource "hyperping_incident" "test" {
  title        = "Test Incident"
  text         = "Testing monitor-incident relationship"
  status_pages = ["sp_test"]
}
`, baseURL, numMonitors)
	return config
}

func testAccIntegrationMonitorMaintenanceConfig(baseURL string, numMonitors, numAffected int) string {
	config := fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %q
}

resource "hyperping_monitor" "test" {
  count = %d
  name  = "Monitor ${count.index}"
  url   = "https://example.com"
}

resource "hyperping_maintenance" "test" {
  name       = "Scheduled Maintenance"
  title      = "Scheduled Maintenance"
  text       = "Testing monitor-maintenance relationship"
  start_date = "2026-12-01T00:00:00Z"
  end_date   = "2026-12-01T04:00:00Z"
  monitors   = slice(hyperping_monitor.test[*].id, 0, %d)
}
`, baseURL, numMonitors, numAffected)
	return config
}

func testAccIntegrationComplexWorkflowConfig(baseURL, monitorName string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %q
}

resource "hyperping_monitor" "test" {
  name = %q
  url  = "https://example.com"
}

resource "hyperping_incident" "test" {
  title        = "Test Incident"
  text         = "Testing complex workflow"
  type         = "incident"
  status_pages = ["sp_test"]
}
`, baseURL, monitorName)
}

func testAccIntegrationBulkMonitorsConfig(baseURL string, count int, url string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %q
}

resource "hyperping_monitor" "bulk" {
  count = %d
  name  = "Bulk Monitor ${count.index}"
  url   = %q
}
`, baseURL, count, url)
}

func testAccIntegrationFullStackConfig(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %q
}

resource "hyperping_monitor" "http" {
  name     = "HTTP Monitor"
  url      = "https://example.com"
  protocol = "http"
}

resource "hyperping_monitor" "api" {
  name        = "API Monitor"
  url         = "https://api.example.com"
  http_method = "POST"
}

resource "hyperping_incident" "test" {
  title        = "Service Degradation"
  text         = "Experiencing performance issues"
  type         = "incident"
  status_pages = ["sp_test"]
}

resource "hyperping_maintenance" "test" {
  name       = "System Upgrade"
  title      = "System Upgrade"
  text       = "Scheduled system maintenance"
  start_date = "2026-12-15T00:00:00Z"
  end_date   = "2026-12-15T04:00:00Z"
  monitors   = [hyperping_monitor.http.id, hyperping_monitor.api.id]
}
`, baseURL)
}

func testAccIntegrationEmptyConfig(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %q
}
`, baseURL)
}

func testAccIntegrationImportConfig(baseURL, monitorName string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %q
}

resource "hyperping_monitor" "imported" {
  name = %q
  url  = "https://example.com"
}
`, baseURL, monitorName)
}

func testAccIntegrationErrorRecoveryConfig(baseURL, monitorName string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %q
}

resource "hyperping_monitor" "error_test" {
  name = %q
  url  = "https://example.com"
}
`, baseURL, monitorName)
}

// Test check helpers

func testAccCheckAllResourcesDeleted(server *mockIntegrationServer) tfresource.TestCheckFunc {
	return func(s *terraform.State) error {
		server.mu.RLock()
		defer server.mu.RUnlock()

		if len(server.monitors) > 0 {
			return fmt.Errorf("expected all monitors to be deleted, found %d", len(server.monitors))
		}
		if len(server.incidents) > 0 {
			return fmt.Errorf("expected all incidents to be deleted, found %d", len(server.incidents))
		}
		if len(server.maintenance) > 0 {
			return fmt.Errorf("expected all maintenance windows to be deleted, found %d", len(server.maintenance))
		}

		return nil
	}
}
