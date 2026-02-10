// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_ListIncidents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != IncidentsBasePath {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get(HeaderAuthorization) != BearerPrefix+"test_key" {
			t.Error("Missing or invalid Authorization header")
		}

		incidents := []Incident{
			{UUID: "inci_001", Title: LocalizedText{En: "Incident One"}, Type: "outage"},
			{UUID: "inci_002", Title: LocalizedText{En: "Incident Two"}, Type: "incident"},
		}
		json.NewEncoder(w).Encode(incidents)
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	incidents, err := c.ListIncidents(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(incidents) != 2 {
		t.Errorf("expected 2 incidents, got %d", len(incidents))
	}
	if incidents[0].UUID != "inci_001" {
		t.Errorf("expected UUID 'inci_001', got '%s'", incidents[0].UUID)
	}
	if incidents[0].Type != "outage" {
		t.Errorf("expected Type 'outage', got '%s'", incidents[0].Type)
	}
}

func TestClient_ListIncidents_WrappedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := struct {
			Incidents []Incident `json:"incidents"`
		}{
			Incidents: []Incident{
				{UUID: "inci_001", Title: LocalizedText{En: "Wrapped Incident"}},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	incidents, err := c.ListIncidents(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(incidents) != 1 {
		t.Errorf("expected 1 incident, got %d", len(incidents))
	}
}

func TestClient_ListIncidents_DataWrappedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := struct {
			Data []Incident `json:"data"`
		}{
			Data: []Incident{
				{UUID: "inci_001", Title: LocalizedText{En: "Data Wrapped Incident"}},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	incidents, err := c.ListIncidents(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(incidents) != 1 {
		t.Errorf("expected 1 incident, got %d", len(incidents))
	}
}

func TestClient_ListIncidents_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]Incident{})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	incidents, err := c.ListIncidents(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(incidents) != 0 {
		t.Errorf("expected 0 incidents, got %d", len(incidents))
	}
}

func TestClient_ListIncidents_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	_, err := c.ListIncidents(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_GetIncident(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != IncidentsBasePath+"/inci_123" {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}

		incident := Incident{
			UUID:  "inci_123",
			Title: LocalizedText{En: "Test Incident"},
			Text:  LocalizedText{En: "We are investigating an issue."},
			Type:  "incident",
			Date:  "2025-01-15T10:30:00.000Z",
		}
		json.NewEncoder(w).Encode(incident)
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	incident, err := c.GetIncident(context.Background(), "inci_123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if incident.UUID != "inci_123" {
		t.Errorf("expected UUID 'inci_123', got '%s'", incident.UUID)
	}
	if incident.Title.En != "Test Incident" {
		t.Errorf("expected Title.En 'Test Incident', got '%s'", incident.Title.En)
	}
}

func TestClient_GetIncident_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Incident not found"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	_, err := c.GetIncident(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestClient_CreateIncident(t *testing.T) {
	var createReqReceived CreateIncidentRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle POST request - returns minimal response
		if r.Method == "POST" && r.URL.Path == IncidentsBasePath {
			json.NewDecoder(r.Body).Decode(&createReqReceived)

			if createReqReceived.Title.En != "New Incident" {
				t.Errorf("expected Title.En 'New Incident', got '%s'", createReqReceived.Title.En)
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Incident created successfully",
				"uuid":    "inci_new",
			})
			return
		}

		// Handle GET request - returns full incident
		if r.Method == "GET" && r.URL.Path == IncidentsBasePath+"/inci_new" {
			json.NewEncoder(w).Encode(Incident{
				UUID:        "inci_new",
				Title:       createReqReceived.Title,
				Text:        createReqReceived.Text,
				Type:        createReqReceived.Type,
				StatusPages: createReqReceived.StatusPages,
			})
			return
		}

		t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	incident, err := c.CreateIncident(context.Background(), CreateIncidentRequest{
		Title:       LocalizedText{En: "New Incident"},
		Text:        LocalizedText{En: "Something went wrong"},
		Type:        "outage",
		StatusPages: []string{"sp_main"},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if incident.UUID != "inci_new" {
		t.Errorf("expected UUID 'inci_new', got '%s'", incident.UUID)
	}
	if incident.Title.En != "New Incident" {
		t.Errorf("expected Title.En 'New Incident', got '%s'", incident.Title.En)
	}
}

func TestClient_CreateIncident_WithComponents(t *testing.T) {
	var createReqReceived CreateIncidentRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle POST request - returns minimal response
		if r.Method == "POST" && r.URL.Path == IncidentsBasePath {
			json.NewDecoder(r.Body).Decode(&createReqReceived)

			if len(createReqReceived.AffectedComponents) != 2 {
				t.Errorf("expected 2 affected components, got %d", len(createReqReceived.AffectedComponents))
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Incident created successfully",
				"uuid":    "inci_with_components",
			})
			return
		}

		// Handle GET request - returns full incident
		// Note: Real API returns empty affectedComponents, not the values from POST
		if r.Method == "GET" && r.URL.Path == IncidentsBasePath+"/inci_with_components" {
			json.NewEncoder(w).Encode(Incident{
				UUID:               "inci_with_components",
				Title:              createReqReceived.Title,
				AffectedComponents: []string{}, // Real API behavior: always returns empty
			})
			return
		}

		t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	incident, err := c.CreateIncident(context.Background(), CreateIncidentRequest{
		Title:              LocalizedText{En: "Incident with Components"},
		Text:               LocalizedText{En: "Affecting multiple services"},
		Type:               "incident",
		AffectedComponents: []string{"comp_api", "comp_web"},
		StatusPages:        []string{"sp_main"},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Note: The real Hyperping API returns empty affectedComponents on GET,
	// even if components were sent in the POST request. This matches actual API behavior
	// as verified in VCR cassette incident_crud.yaml.
	// The mock reflects this real API behavior.
	if incident.UUID != "inci_with_components" {
		t.Errorf("expected UUID 'inci_with_components', got '%s'", incident.UUID)
	}
	if incident.Title.En != "Incident with Components" {
		t.Errorf("expected Title.En 'Incident with Components', got '%s'", incident.Title.En)
	}
}

func TestClient_CreateIncident_ReadAfterCreateFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle POST request - returns minimal response
		if r.Method == "POST" && r.URL.Path == IncidentsBasePath {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Incident created successfully",
				"uuid":    "inci_created_but_not_readable",
			})
			return
		}

		// Handle GET request - simulates failure reading the newly created incident
		if r.Method == "GET" && r.URL.Path == IncidentsBasePath+"/inci_created_but_not_readable" {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
			return
		}

		t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	_, err := c.CreateIncident(context.Background(), CreateIncidentRequest{
		Title:       LocalizedText{En: "Test Incident"},
		Text:        LocalizedText{En: "Test"},
		Type:        "incident",
		StatusPages: []string{"sp_main"},
	})
	if err == nil {
		t.Error("expected error when read-after-create fails, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read incident after creation") {
		t.Errorf("expected 'failed to read incident after creation' error, got: %v", err)
	}
}

func TestClient_CreateIncident_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Validation failed",
			"details": []map[string]string{{"field": "title", "message": "is required"}},
		})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	_, err := c.CreateIncident(context.Background(), CreateIncidentRequest{})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_UpdateIncident(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" || r.URL.Path != IncidentsBasePath+"/inci_123" {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var req UpdateIncidentRequest
		json.NewDecoder(r.Body).Decode(&req)

		json.NewEncoder(w).Encode(Incident{
			UUID:  "inci_123",
			Title: *req.Title,
			Type:  *req.Type,
		})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	title := LocalizedText{En: "Updated Title"}
	incidentType := "outage"
	incident, err := c.UpdateIncident(context.Background(), "inci_123", UpdateIncidentRequest{
		Title: &title,
		Type:  &incidentType,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if incident.Title.En != "Updated Title" {
		t.Errorf("expected Title.En 'Updated Title', got '%s'", incident.Title.En)
	}
	if incident.Type != "outage" {
		t.Errorf("expected Type 'outage', got '%s'", incident.Type)
	}
}

func TestClient_UpdateIncident_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Incident not found"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	title := LocalizedText{En: "Updated"}
	_, err := c.UpdateIncident(context.Background(), "nonexistent", UpdateIncidentRequest{Title: &title})
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestClient_AddIncidentUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != IncidentsBasePath+"/inci_123/updates" {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var req AddIncidentUpdateRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Text.En != "We identified the issue" {
			t.Errorf("expected Text.En 'We identified the issue', got '%s'", req.Text.En)
		}

		json.NewEncoder(w).Encode(Incident{
			UUID: "inci_123",
			Updates: []IncidentUpdate{
				{UUID: "update_001", Text: req.Text, Type: req.Type},
			},
		})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	incident, err := c.AddIncidentUpdate(context.Background(), "inci_123", AddIncidentUpdateRequest{
		Text: LocalizedText{En: "We identified the issue"},
		Type: "identified",
		Date: "2025-01-15T11:00:00.000Z",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(incident.Updates) != 1 {
		t.Errorf("expected 1 update, got %d", len(incident.Updates))
	}
	if incident.Updates[0].Text.En != "We identified the issue" {
		t.Errorf("expected update text 'We identified the issue', got '%s'", incident.Updates[0].Text.En)
	}
}

func TestClient_ResolveIncident(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != IncidentsBasePath+"/inci_123/updates" {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var req AddIncidentUpdateRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Type != "resolved" {
			t.Errorf("expected Type 'resolved', got '%s'", req.Type)
		}

		json.NewEncoder(w).Encode(Incident{
			UUID: "inci_123",
			Updates: []IncidentUpdate{
				{UUID: "update_resolved", Text: req.Text, Type: "resolved"},
			},
		})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	incident, err := c.ResolveIncident(context.Background(), "inci_123", "The issue has been resolved.")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(incident.Updates) != 1 {
		t.Errorf("expected 1 update, got %d", len(incident.Updates))
	}
	if incident.Updates[0].Type != "resolved" {
		t.Errorf("expected update type 'resolved', got '%s'", incident.Updates[0].Type)
	}
}

func TestClient_DeleteIncident(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != IncidentsBasePath+"/inci_123" {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	err := c.DeleteIncident(context.Background(), "inci_123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestClient_DeleteIncident_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Incident not found"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	err := c.DeleteIncident(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestClient_DeleteIncident_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	err := c.DeleteIncident(context.Background(), "inci_123")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_ListIncidents_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	_, err := c.ListIncidents(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_AddIncidentUpdate_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "update failed"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	_, err := c.AddIncidentUpdate(context.Background(), "inci_123", AddIncidentUpdateRequest{
		Text: LocalizedText{En: "Test update"},
		Type: "identified",
		Date: "2025-01-15T11:00:00.000Z",
	})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_AddIncidentUpdate_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	_, err := c.AddIncidentUpdate(context.Background(), "nonexistent", AddIncidentUpdateRequest{
		Text: LocalizedText{En: "Test update"},
		Type: "identified",
		Date: "2025-01-15T11:00:00.000Z",
	})
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestClient_ListIncidents_EmptyWrappedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"incidents": []Incident{},
		})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	incidents, err := c.ListIncidents(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(incidents) != 0 {
		t.Errorf("expected 0 incidents, got %d", len(incidents))
	}
}

func TestClient_ListIncidents_EmptyObjectResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	incidents, err := c.ListIncidents(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(incidents) != 0 {
		t.Errorf("expected 0 incidents, got %d", len(incidents))
	}
}

func TestClient_ListIncidents_UnexpectedJSONType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HeaderContentType, ContentTypeJSON)
		w.Write([]byte(`"unexpected string response"`))
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	_, err := c.ListIncidents(context.Background())
	if err == nil {
		t.Fatal("expected error for unexpected JSON type, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse incidents response") {
		t.Errorf("expected parse error, got %v", err)
	}
}
