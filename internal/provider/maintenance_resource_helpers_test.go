// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// maintenanceTestFixture represents a maintenance window for testing
type maintenanceTestFixture struct {
	UUID                string
	Name                string
	Title               string
	Text                string
	StartDate           string
	EndDate             string
	Monitors            []string
	StatusPages         []string
	NotificationOption  string
	NotificationMinutes int
}

// toAPIResponse converts fixture to API response format
func (f *maintenanceTestFixture) toAPIResponse() map[string]interface{} {
	resp := map[string]interface{}{
		"uuid":       f.UUID,
		"name":       f.Name,
		"start_date": f.StartDate,
		"end_date":   f.EndDate,
	}

	if f.Title != "" {
		resp["title"] = map[string]interface{}{"en": f.Title}
	}
	if f.Text != "" {
		resp["text"] = map[string]interface{}{"en": f.Text}
	}
	if len(f.Monitors) > 0 {
		resp["monitors"] = f.Monitors
	}
	if len(f.StatusPages) > 0 {
		resp["statuspages"] = f.StatusPages
	}
	if f.NotificationOption != "" {
		resp["notificationOption"] = f.NotificationOption
	} else {
		// Default to "scheduled" if not set
		resp["notificationOption"] = "scheduled"
	}
	if f.NotificationMinutes > 0 {
		resp["notificationMinutes"] = f.NotificationMinutes
	} else {
		// Default to 60 if not set
		resp["notificationMinutes"] = 60
	}

	return resp
}

// generateMaintenanceTimeRange creates valid start and end times for testing
func generateMaintenanceTimeRange(hoursInFuture, durationHours int) (string, string) {
	now := time.Now().UTC()
	start := now.Add(time.Duration(hoursInFuture) * time.Hour).Truncate(time.Second)
	end := start.Add(time.Duration(durationHours) * time.Hour)
	return start.Format(time.RFC3339), end.Format(time.RFC3339)
}

// newMaintenanceMockServer creates a test HTTP server for maintenance operations
type maintenanceMockServer struct {
	fixtures      map[string]*maintenanceTestFixture
	deleted       map[string]bool
	createHandler func(w http.ResponseWriter, r *http.Request)
	updateHandler func(w http.ResponseWriter, r *http.Request)
}

func newMaintenanceMockServer() *maintenanceMockServer {
	return &maintenanceMockServer{
		fixtures: make(map[string]*maintenanceTestFixture),
		deleted:  make(map[string]bool),
	}
}

func (m *maintenanceMockServer) addFixture(f *maintenanceTestFixture) {
	m.fixtures[f.UUID] = f
}

func (m *maintenanceMockServer) setDeleted(uuid string, deleted bool) {
	m.deleted[uuid] = deleted
}

// handleCreate handles POST requests to create a maintenance window.
func (m *maintenanceMockServer) handleCreate(w http.ResponseWriter, r *http.Request) {
	if m.createHandler != nil {
		m.createHandler(w, r)
		return
	}
	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)
	for _, f := range m.fixtures {
		if nameVal, ok := req["name"].(string); ok && nameVal == f.Name {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{"uuid": f.UUID})
			return
		}
	}
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{"error": "fixture not found"})
}

// applyFixtureUpdate applies request fields onto the fixture.
func applyFixtureUpdate(f *maintenanceTestFixture, req map[string]interface{}) {
	if name, ok := req["name"].(string); ok {
		f.Name = name
	}
	if titleMap, ok := req["title"].(map[string]interface{}); ok {
		if en, ok := titleMap["en"].(string); ok {
			f.Title = en
		}
	}
	if textMap, ok := req["text"].(map[string]interface{}); ok {
		if en, ok := textMap["en"].(string); ok {
			f.Text = en
		}
	}
	if startDate, ok := req["start_date"].(string); ok {
		f.StartDate = startDate
	}
	if endDate, ok := req["end_date"].(string); ok {
		f.EndDate = endDate
	}
	if notifOpt, ok := req["notificationOption"].(string); ok {
		f.NotificationOption = notifOpt
	}
	if notifMins, ok := req["notificationMinutes"].(float64); ok {
		f.NotificationMinutes = int(notifMins)
	}
}

// handleUpdate handles PUT requests to update a maintenance window.
func (m *maintenanceMockServer) handleUpdate(w http.ResponseWriter, r *http.Request) {
	if m.updateHandler != nil {
		m.updateHandler(w, r)
		return
	}
	for uuid, f := range m.fixtures {
		if r.URL.Path != client.MaintenanceBasePath+"/"+uuid {
			continue
		}
		if m.deleted[uuid] {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
			return
		}
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		applyFixtureUpdate(f, req)
		json.NewEncoder(w).Encode(f.toAPIResponse())
		return
	}
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
}

// handleGet handles GET requests for a single maintenance window.
func (m *maintenanceMockServer) handleGet(w http.ResponseWriter, r *http.Request) {
	for uuid, f := range m.fixtures {
		if r.URL.Path != client.MaintenanceBasePath+"/"+uuid {
			continue
		}
		if m.deleted[uuid] {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
			return
		}
		json.NewEncoder(w).Encode(f.toAPIResponse())
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

// handleDelete handles DELETE requests for a maintenance window.
func (m *maintenanceMockServer) handleDelete(w http.ResponseWriter, r *http.Request) {
	for uuid := range m.fixtures {
		if r.URL.Path == client.MaintenanceBasePath+"/"+uuid {
			m.deleted[uuid] = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func (m *maintenanceMockServer) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)

		switch r.Method {
		case http.MethodPost:
			m.handleCreate(w, r)
		case http.MethodPut:
			m.handleUpdate(w, r)
		case http.MethodGet:
			m.handleGet(w, r)
		case http.MethodDelete:
			m.handleDelete(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

// newSimpleMaintenanceServer creates a basic mock server for simple tests
func newSimpleMaintenanceServer(fixture *maintenanceTestFixture) *httptest.Server {
	mock := newMaintenanceMockServer()
	mock.addFixture(fixture)
	return httptest.NewServer(mock.handler())
}

// newMaintenanceServerWithCustomCreate creates a server with custom create handling
func newMaintenanceServerWithCustomCreate(fixture *maintenanceTestFixture, createFn func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	mock := newMaintenanceMockServer()
	mock.addFixture(fixture)
	mock.createHandler = createFn
	return httptest.NewServer(mock.handler())
}

// newMaintenanceServerWithCustomUpdate creates a server with custom update handling
func newMaintenanceServerWithCustomUpdate(fixture *maintenanceTestFixture, updateFn func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	mock := newMaintenanceMockServer()
	mock.addFixture(fixture)
	mock.updateHandler = updateFn
	return httptest.NewServer(mock.handler())
}

// generateMaintenanceConfig generates Terraform config for maintenance resource
func generateMaintenanceConfig(serverURL, name, title, text, startDate, endDate string, monitors []string) string {
	config := fmt.Sprintf(`
provider "hyperping" {
  api_key  = "hp_test_key"
  base_url = %q
}

resource "hyperping_maintenance" "test" {
  name       = %q
  title      = %q
  text       = %q
  start_date = %q
  end_date   = %q
  monitors   = [`, serverURL, name, title, text, startDate, endDate)

	for i, mon := range monitors {
		if i > 0 {
			config += ", "
		}
		config += fmt.Sprintf("%q", mon)
	}
	config += "]\n}\n"
	return config
}

// generateMaintenanceConfigFull generates Terraform config with all optional fields
func generateMaintenanceConfigFull(serverURL, name, title, text, startDate, endDate string, monitors, statusPages []string, notificationOption string, notificationMinutes int) string {
	config := fmt.Sprintf(`
provider "hyperping" {
  api_key  = "hp_test_key"
  base_url = %q
}

resource "hyperping_maintenance" "test" {
  name                 = %q
  title                = %q
  text                 = %q
  start_date           = %q
  end_date             = %q
  monitors             = [`, serverURL, name, title, text, startDate, endDate)

	for i, mon := range monitors {
		if i > 0 {
			config += ", "
		}
		config += fmt.Sprintf("%q", mon)
	}
	config += "]\n"

	if len(statusPages) > 0 {
		config += "  status_pages         = ["
		for i, sp := range statusPages {
			if i > 0 {
				config += ", "
			}
			config += fmt.Sprintf("%q", sp)
		}
		config += "]\n"
	}

	if notificationOption != "" {
		config += fmt.Sprintf("  notification_option  = %q\n", notificationOption)
	}
	if notificationMinutes > 0 {
		config += fmt.Sprintf("  notification_minutes = %d\n", notificationMinutes)
	}

	config += "}\n"
	return config
}

// generateMaintenanceConfigMinimal generates minimal Terraform config (name and dates only)
func generateMaintenanceConfigMinimal(serverURL, name, startDate, endDate string, monitors []string) string {
	config := fmt.Sprintf(`
provider "hyperping" {
  api_key  = "hp_test_key"
  base_url = %q
}

resource "hyperping_maintenance" "test" {
  name       = %q
  start_date = %q
  end_date   = %q
  monitors   = [`, serverURL, name, startDate, endDate)

	for i, mon := range monitors {
		if i > 0 {
			config += ", "
		}
		config += fmt.Sprintf("%q", mon)
	}
	config += "]\n}\n"
	return config
}

// generateProviderOnlyConfig generates config with provider but no resources
func generateProviderOnlyConfig(serverURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "hp_test_key"
  base_url = %q
}
`, serverURL)
}

// newErrorServer creates a server that always returns errors
func newErrorServer(statusCode int, errorMessage string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]string{"error": errorMessage})
	}))
}

// newMaintenanceServerWithDisappear creates a server that simulates resource disappearing
func newMaintenanceServerWithDisappear(fixture *maintenanceTestFixture) (*httptest.Server, *bool) {
	deleted := false
	mock := newMaintenanceMockServer()
	mock.addFixture(fixture)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)

		switch {
		case r.Method == http.MethodPost && r.URL.Path == client.MaintenanceBasePath:
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{"uuid": fixture.UUID})
		case r.Method == http.MethodGet && r.URL.Path == client.MaintenanceBasePath+"/"+fixture.UUID:
			if deleted {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
				return
			}
			json.NewEncoder(w).Encode(fixture.toAPIResponse())
		case r.Method == http.MethodDelete && r.URL.Path == client.MaintenanceBasePath+"/"+fixture.UUID:
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	return server, &deleted
}

// newMaintenanceServerWithUpdateCapture creates a server that captures updates
func newMaintenanceServerWithUpdateCapture(fixture *maintenanceTestFixture) (*httptest.Server, *map[string]interface{}) {
	maintenance := fixture.toAPIResponse()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)

		switch {
		case r.Method == http.MethodPost && r.URL.Path == client.MaintenanceBasePath:
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{"uuid": fixture.UUID})
		case r.Method == http.MethodPut && r.URL.Path == client.MaintenanceBasePath+"/"+fixture.UUID:
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)
			// Update maintenance with new values
			for k, v := range req {
				maintenance[k] = v
			}
			json.NewEncoder(w).Encode(maintenance)
		case r.Method == http.MethodGet && r.URL.Path == client.MaintenanceBasePath+"/"+fixture.UUID:
			json.NewEncoder(w).Encode(maintenance)
		case r.Method == http.MethodDelete && r.URL.Path == client.MaintenanceBasePath+"/"+fixture.UUID:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	return server, &maintenance
}

// newMaintenanceServerReadAfterCreateError creates a server where POST succeeds but GET fails
func newMaintenanceServerReadAfterCreateError(uuid string) *httptest.Server {
	createCalled := false

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)

		switch {
		case r.Method == http.MethodPost && r.URL.Path == client.MaintenanceBasePath:
			createCalled = true
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{"uuid": uuid})
		case r.Method == http.MethodGet && createCalled:
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

// newMaintenanceServerDeleteNotFound creates a server where DELETE returns 404
func newMaintenanceServerDeleteNotFound(fixture *maintenanceTestFixture) *httptest.Server {
	deleted := false

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)

		switch {
		case r.Method == http.MethodPost && r.URL.Path == client.MaintenanceBasePath:
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{"uuid": fixture.UUID})
		case r.Method == http.MethodGet && r.URL.Path == client.MaintenanceBasePath+"/"+fixture.UUID:
			if deleted {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
				return
			}
			json.NewEncoder(w).Encode(fixture.toAPIResponse())
		case r.Method == http.MethodDelete && r.URL.Path == client.MaintenanceBasePath+"/"+fixture.UUID:
			// Return not found - simulates already deleted resource
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
			deleted = true
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}
