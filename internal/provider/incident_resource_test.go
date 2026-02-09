// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestAccIncidentResource_basic(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create and Read testing
			{
				Config: testAccIncidentResourceConfig_basic(server.URL, "Test Incident"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "title", "Test Incident"),
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "text", "Something went wrong"),
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "type", "incident"),
					tfresource.TestCheckResourceAttrSet("hyperping_incident.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "hyperping_incident.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccIncidentResourceConfig_basic(server.URL, "Updated Incident"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "title", "Updated Incident"),
				),
			},
		},
	})
}

func TestAccIncidentResource_full(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccIncidentResourceConfig_full(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "title", "Major Outage"),
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "text", "We are experiencing a major outage"),
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "type", "outage"),
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "status_pages.#", "1"),
				),
			},
		},
	})
}

func TestAccIncidentResource_typeChange(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Start with incident type
			{
				Config: testAccIncidentResourceConfig_withType(server.URL, "incident"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "type", "incident"),
				),
			},
			// Change to outage type
			{
				Config: testAccIncidentResourceConfig_withType(server.URL, "outage"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "type", "outage"),
				),
			},
		},
	})
}

func TestAccIncidentResource_disappears(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccIncidentResourceConfig_basic(server.URL, "disappear-test"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttrSet("hyperping_incident.test", "id"),
					testAccCheckIncidentDisappears(server),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIncidentResource_createError(t *testing.T) {
	server := newMockIncidentServerWithErrors(t)
	defer server.Close()

	server.setCreateError(true)

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccIncidentResourceConfig_basic(server.URL, "error-test"),
				ExpectError: regexp.MustCompile(`Failed to Create Incident`),
			},
		},
	})
}

func TestAccIncidentResource_updateError(t *testing.T) {
	server := newMockIncidentServerWithErrors(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccIncidentResourceConfig_basic(server.URL, "update-error-test"),
			},
			{
				PreConfig:   func() { server.setUpdateError(true) },
				Config:      testAccIncidentResourceConfig_basic(server.URL, "updated-title"),
				ExpectError: regexp.MustCompile(`Failed to Update Incident`),
			},
		},
	})
}

func TestAccIncidentResource_readError(t *testing.T) {
	server := newMockIncidentServerWithErrors(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccIncidentResourceConfig_basic(server.URL, "read-error-test"),
			},
			{
				PreConfig:          func() { server.setReadError(true) },
				Config:             testAccIncidentResourceConfig_basic(server.URL, "read-error-test"),
				ExpectError:        regexp.MustCompile(`Could not read incident`),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIncidentResource_readAfterCreateError(t *testing.T) {
	server := newMockIncidentServerWithErrors(t)
	defer server.Close()

	// Enable read-after-create error: POST succeeds but GET fails
	server.setReadAfterCreateError(true)

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccIncidentResourceConfig_basic(server.URL, "read-after-create-error-test"),
				ExpectError: regexp.MustCompile(`Incident Created But Read Failed`),
			},
		},
	})
}

// Unit tests

func TestIncidentResource_ConfigureWrongType(t *testing.T) {
	r := &IncidentResource{}

	req := resource.ConfigureRequest{
		ProviderData: "wrong type",
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("Expected error when provider data is wrong type")
	}
}

func TestIncidentResource_ConfigureNilProviderData(t *testing.T) {
	r := &IncidentResource{}

	req := resource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Error("Expected no error when provider data is nil")
	}
}

func TestIncidentResource_ConfigureValidClient(t *testing.T) {
	r := &IncidentResource{}

	// Create a real client
	c := client.NewClient("test_api_key")

	req := resource.ConfigureRequest{
		ProviderData: c,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	// Should not error
	if resp.Diagnostics.HasError() {
		t.Errorf("Unexpected error: %v", resp.Diagnostics)
	}

	// Client should be set
	if r.client == nil {
		t.Error("Expected client to be set")
	}
}

func TestIncidentResource_Metadata(t *testing.T) {
	r := &IncidentResource{}

	req := resource.MetadataRequest{
		ProviderTypeName: "hyperping",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "hyperping_incident" {
		t.Errorf("Expected type name 'hyperping_incident', got '%s'", resp.TypeName)
	}
}

func TestIncidentResource_Schema(t *testing.T) {
	r := &IncidentResource{}

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	// Verify essential attributes exist
	requiredAttrs := []string{"id", "title", "text", "type", "status_pages"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing '%s' attribute", attr)
		}
	}
}

func TestIncidentResource_mapIncidentToModel(t *testing.T) {
	r := &IncidentResource{}

	t.Run("all fields populated", func(t *testing.T) {
		incident := &client.Incident{
			UUID:               "inci_123",
			Title:              client.LocalizedText{En: "Test Incident"},
			Text:               client.LocalizedText{En: "Test description"},
			Type:               "outage",
			AffectedComponents: []string{"comp-1", "comp-2"},
			StatusPages:        []string{"sp-1", "sp-2"},
			Date:               "2025-01-15T10:00:00.000Z",
		}

		model := &IncidentResourceModel{}
		diags := &diag.Diagnostics{}
		r.mapIncidentToModel(incident, model, diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags)
		}
		if model.ID.ValueString() != "inci_123" {
			t.Errorf("expected ID 'inci_123', got %s", model.ID.ValueString())
		}
		if model.Title.ValueString() != "Test Incident" {
			t.Errorf("expected title 'Test Incident', got %s", model.Title.ValueString())
		}
		if model.Text.ValueString() != "Test description" {
			t.Errorf("expected text 'Test description', got %s", model.Text.ValueString())
		}
		if model.Type.ValueString() != "outage" {
			t.Errorf("expected type 'outage', got %s", model.Type.ValueString())
		}
	})

	t.Run("empty affected components", func(t *testing.T) {
		incident := &client.Incident{
			UUID:               "inci_456",
			Title:              client.LocalizedText{En: "No Components"},
			Text:               client.LocalizedText{En: "No affected components"},
			Type:               "incident",
			AffectedComponents: []string{},
			StatusPages:        []string{"sp-1"},
		}

		model := &IncidentResourceModel{}
		diags := &diag.Diagnostics{}
		r.mapIncidentToModel(incident, model, diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags)
		}
		if !model.AffectedComponents.IsNull() {
			t.Error("expected AffectedComponents to be null for empty slice")
		}
	})

	t.Run("empty date", func(t *testing.T) {
		incident := &client.Incident{
			UUID:        "inci_789",
			Title:       client.LocalizedText{En: "No Date"},
			Text:        client.LocalizedText{En: "No date set"},
			Type:        "incident",
			StatusPages: []string{"sp-1"},
			Date:        "",
		}

		model := &IncidentResourceModel{}
		diags := &diag.Diagnostics{}
		r.mapIncidentToModel(incident, model, diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags)
		}
		if !model.Date.IsNull() {
			t.Error("expected Date to be null for empty string")
		}
	})
}

func TestAccIncidentResource_withAffectedComponents(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccIncidentResourceConfig_withAffectedComponents(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "title", "Component Test"),
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "affected_components.#", "2"),
				),
			},
		},
	})
}

// Helper functions

func testAccIncidentResourceConfig_withAffectedComponents(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_incident" "test" {
  title               = "Component Test"
  text                = "Testing with affected components"
  status_pages        = ["sp_main"]
  affected_components = ["comp_123", "comp_456"]
}
`, baseURL)
}

func testAccIncidentResourceConfig_basic(baseURL, title string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_incident" "test" {
  title        = %[2]q
  text         = "Something went wrong"
  status_pages = ["sp_main"]
}
`, baseURL, title)
}

func testAccIncidentResourceConfig_full(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_incident" "test" {
  title        = "Major Outage"
  text         = "We are experiencing a major outage"
  type         = "outage"
  status_pages = ["sp_main"]
}
`, baseURL)
}

func testAccIncidentResourceConfig_withType(baseURL, incidentType string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_incident" "test" {
  title        = "Type Test"
  text         = "Testing type changes"
  type         = %[2]q
  status_pages = ["sp_main"]
}
`, baseURL, incidentType)
}

func testAccCheckIncidentDisappears(server *mockIncidentServer) tfresource.TestCheckFunc {
	return func(s *terraform.State) error {
		server.deleteAllIncidents()
		return nil
	}
}

// Mock server implementation

type mockIncidentServer struct {
	*httptest.Server
	t         *testing.T
	incidents map[string]map[string]interface{}
	counter   int
}

func newMockIncidentServer(t *testing.T) *mockIncidentServer {
	m := &mockIncidentServer{
		t:         t,
		incidents: make(map[string]map[string]interface{}),
		counter:   0,
	}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handleRequest(w, r)
	}))

	return m
}

func (m *mockIncidentServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)

	switch {
	case r.Method == "GET" && r.URL.Path == client.IncidentsBasePath:
		m.listIncidents(w)
	case r.Method == "POST" && r.URL.Path == client.IncidentsBasePath:
		m.createIncident(w, r)
	case r.Method == "POST" && strings.Contains(r.URL.Path, client.IncidentsBasePath+"/") && strings.HasSuffix(r.URL.Path, "/updates"):
		m.addIncidentUpdate(w, r)
	case r.Method == "GET" && len(r.URL.Path) > len(client.IncidentsBasePath+"/"):
		m.getIncident(w, r)
	case r.Method == "PUT" && len(r.URL.Path) > len(client.IncidentsBasePath+"/"):
		m.updateIncident(w, r)
	case r.Method == "DELETE" && len(r.URL.Path) > len(client.IncidentsBasePath+"/"):
		m.deleteIncident(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}
}

func (m *mockIncidentServer) listIncidents(w http.ResponseWriter) {
	incidents := make([]map[string]interface{}, 0, len(m.incidents))
	for _, incident := range m.incidents {
		incidents = append(incidents, incident)
	}
	json.NewEncoder(w).Encode(incidents)
}

func (m *mockIncidentServer) createIncident(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	m.counter++
	id := fmt.Sprintf("inci_%d", m.counter)

	// Extract title and text from localized objects
	titleObj, ok := req["title"].(map[string]interface{})
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid title format"})
		return
	}
	textObj, ok := req["text"].(map[string]interface{})
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid text format"})
		return
	}

	incident := map[string]interface{}{
		"uuid":        id,
		"title":       titleObj,
		"text":        textObj,
		"type":        getOrDefault(req, "type", "incident"),
		"statuspages": req["statuspages"],
		"date":        "2025-01-15T10:00:00.000Z",
	}

	if affectedComponents, ok := req["affectedComponents"].([]interface{}); ok {
		incident["affectedComponents"] = affectedComponents
	}

	m.incidents[id] = incident

	// Return ONLY UUID (simulating real Hyperping API behavior)
	// The provider must do a GET to fetch the full incident
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"uuid": id})
}

func (m *mockIncidentServer) getIncident(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len(client.IncidentsBasePath+"/"):]

	incident, exists := m.incidents[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Incident not found"})
		return
	}

	json.NewEncoder(w).Encode(incident)
}

func (m *mockIncidentServer) updateIncident(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len(client.IncidentsBasePath+"/"):]

	incident, exists := m.incidents[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Incident not found"})
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	// Update fields if present
	if title, ok := req["title"].(map[string]interface{}); ok {
		incident["title"] = title
	}
	if text, ok := req["text"].(map[string]interface{}); ok {
		incident["text"] = text
	}
	if incType, ok := req["type"].(string); ok {
		incident["type"] = incType
	}
	if statusPages, ok := req["statuspages"].([]interface{}); ok {
		incident["statuspages"] = statusPages
	}
	if components, ok := req["affectedComponents"].([]interface{}); ok {
		incident["affectedComponents"] = components
	}

	m.incidents[id] = incident
	json.NewEncoder(w).Encode(incident)
}

func (m *mockIncidentServer) deleteIncident(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len(client.IncidentsBasePath+"/"):]

	if _, exists := m.incidents[id]; !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Incident not found"})
		return
	}

	delete(m.incidents, id)
	w.WriteHeader(http.StatusNoContent)
}

func (m *mockIncidentServer) deleteAllIncidents() {
	m.incidents = make(map[string]map[string]interface{})
}

func (m *mockIncidentServer) addIncidentUpdate(w http.ResponseWriter, r *http.Request) {
	// Extract incident ID from path like client.IncidentsBasePath+"/"{id}/updates"
	path := r.URL.Path[len(client.IncidentsBasePath+"/"):]
	incidentID := path[:len(path)-len("/updates")]

	incident, exists := m.incidents[incidentID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Incident not found"})
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	// Generate update ID
	m.counter++
	updateID := fmt.Sprintf("upd_%d", m.counter)

	update := map[string]interface{}{
		"uuid": updateID,
		"text": req["text"],
		"type": req["type"],
		"date": "2026-01-15T10:00:00.000Z",
	}

	// Add update to incident
	updates, ok := incident["updates"].([]interface{})
	if !ok {
		updates = []interface{}{}
	}
	updates = append(updates, update)
	incident["updates"] = updates
	m.incidents[incidentID] = incident

	w.WriteHeader(http.StatusCreated)
	// Return full incident (API behavior)
	if err := json.NewEncoder(w).Encode(incident); err != nil {
		m.t.Errorf("failed to encode response: %v", err)
	}
}

// Mock server with error injection

type mockIncidentServerWithErrors struct {
	*mockIncidentServer
	createError          bool
	readError            bool
	updateError          bool
	deleteError          bool
	readAfterCreateError bool
	createCalled         bool
}

func newMockIncidentServerWithErrors(t *testing.T) *mockIncidentServerWithErrors {
	m := &mockIncidentServerWithErrors{
		mockIncidentServer: &mockIncidentServer{
			t:         t,
			incidents: make(map[string]map[string]interface{}),
			counter:   0,
		},
	}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handleRequestWithErrors(w, r)
	}))

	return m
}

func (m *mockIncidentServerWithErrors) setCreateError(v bool)          { m.createError = v }
func (m *mockIncidentServerWithErrors) setReadError(v bool)            { m.readError = v }
func (m *mockIncidentServerWithErrors) setUpdateError(v bool)          { m.updateError = v }
func (m *mockIncidentServerWithErrors) setReadAfterCreateError(v bool) { m.readAfterCreateError = v }

func (m *mockIncidentServerWithErrors) handleRequestWithErrors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(client.HeaderContentType, client.ContentTypeJSON)

	switch {
	case r.Method == "POST" && r.URL.Path == client.IncidentsBasePath:
		if m.createError {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
			return
		}
		m.createCalled = true
		m.createIncident(w, r)

	case r.Method == "POST" && strings.Contains(r.URL.Path, client.IncidentsBasePath+"/") && strings.HasSuffix(r.URL.Path, "/updates"):
		if m.updateError {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
			return
		}
		m.addIncidentUpdate(w, r)

	case r.Method == "GET" && len(r.URL.Path) > len(client.IncidentsBasePath+"/"):
		if m.readError || (m.readAfterCreateError && m.createCalled) {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
			return
		}
		m.getIncident(w, r)

	case r.Method == "PUT" && len(r.URL.Path) > len(client.IncidentsBasePath+"/"):
		if m.updateError {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
			return
		}
		m.updateIncident(w, r)

	case r.Method == "DELETE" && len(r.URL.Path) > len(client.IncidentsBasePath+"/"):
		if m.deleteError {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
			return
		}
		m.deleteIncident(w, r)

	default:
		m.handleRequest(w, r)
	}
}
