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
	"time"
)

func TestClient_ListMonitors_DirectArray(t *testing.T) {
	monitors := []Monitor{
		{UUID: "mon_001", Name: "Monitor 1", URL: "https://example1.com", Protocol: "http"},
		{UUID: "mon_002", Name: "Monitor 2", URL: "https://example2.com", Protocol: "http"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != MonitorsBasePath {
			t.Errorf("expected path %s, got %s", MonitorsBasePath, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(monitors)
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	result, err := c.ListMonitors(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 monitors, got %d", len(result))
	}
	if result[0].UUID != "mon_001" {
		t.Errorf("expected UUID 'mon_001', got %q", result[0].UUID)
	}
}

func TestClient_ListMonitors_WrappedMonitors(t *testing.T) {
	response := map[string]interface{}{
		"monitors": []Monitor{
			{UUID: "mon_001", Name: "Monitor 1", URL: "https://example1.com", Protocol: "http"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	result, err := c.ListMonitors(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 monitor, got %d", len(result))
	}
}

func TestClient_ListMonitors_WrappedData(t *testing.T) {
	response := map[string]interface{}{
		"data": []Monitor{
			{UUID: "mon_001", Name: "Monitor 1", URL: "https://example1.com", Protocol: "http"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	result, err := c.ListMonitors(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 monitor, got %d", len(result))
	}
}

func TestClient_ListMonitors_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]Monitor{})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	result, err := c.ListMonitors(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 monitors, got %d", len(result))
	}
}

func TestClient_GetMonitor_Success(t *testing.T) {
	monitor := Monitor{
		UUID:           "mon_test",
		Name:           "Test Monitor",
		URL:            "https://example.com",
		Protocol:       "http",
		HTTPMethod:     "GET",
		CheckFrequency: 60,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != MonitorsBasePath+"/mon_test" {
			t.Errorf("expected path %s/mon_test, got %s", MonitorsBasePath, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(monitor)
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	result, err := c.GetMonitor(context.Background(), "mon_test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.UUID != "mon_test" {
		t.Errorf("expected UUID 'mon_test', got %q", result.UUID)
	}
	if result.Name != "Test Monitor" {
		t.Errorf("expected Name 'Test Monitor', got %q", result.Name)
	}
}

func TestClient_GetMonitor_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	_, err := c.GetMonitor(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestClient_CreateMonitor_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != MonitorsBasePath {
			t.Errorf("expected path %s, got %s", MonitorsBasePath, r.URL.Path)
		}

		var req CreateMonitorRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if req.Name != "New Monitor" {
			t.Errorf("expected Name 'New Monitor', got %q", req.Name)
		}
		if req.URL != "https://example.com" {
			t.Errorf("expected URL 'https://example.com', got %q", req.URL)
		}

		response := Monitor{
			UUID:           "mon_new",
			Name:           req.Name,
			URL:            req.URL,
			Protocol:       req.Protocol,
			HTTPMethod:     "GET",
			CheckFrequency: 60,
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	result, err := c.CreateMonitor(context.Background(), CreateMonitorRequest{
		Name:     "New Monitor",
		URL:      "https://example.com",
		Protocol: "http",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.UUID != "mon_new" {
		t.Errorf("expected UUID 'mon_new', got %q", result.UUID)
	}
}

func TestClient_CreateMonitor_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Validation failed",
			"details": []map[string]string{
				{"field": "url", "message": "Invalid URL"},
			},
		})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	_, err := c.CreateMonitor(context.Background(), CreateMonitorRequest{
		Name:     "Test",
		URL:      "invalid",
		Protocol: "http",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsValidation(err) {
		t.Errorf("expected validation error, got %v", err)
	}
}

func TestClient_UpdateMonitor_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != MonitorsBasePath+"/mon_test" {
			t.Errorf("expected path %s/mon_test, got %s", MonitorsBasePath, r.URL.Path)
		}

		var req UpdateMonitorRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if req.Name == nil || *req.Name != "Updated Name" {
			t.Errorf("expected Name 'Updated Name', got %v", req.Name)
		}

		response := Monitor{
			UUID:           "mon_test",
			Name:           *req.Name,
			URL:            "https://example.com",
			Protocol:       "http",
			HTTPMethod:     "GET",
			CheckFrequency: 60,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	name := "Updated Name"
	result, err := c.UpdateMonitor(context.Background(), "mon_test", UpdateMonitorRequest{
		Name: &name,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Name != "Updated Name" {
		t.Errorf("expected Name 'Updated Name', got %q", result.Name)
	}
}

func TestClient_DeleteMonitor_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != MonitorsBasePath+"/mon_test" {
			t.Errorf("expected path %s/mon_test, got %s", MonitorsBasePath, r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	err := c.DeleteMonitor(context.Background(), "mon_test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestClient_DeleteMonitor_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	err := c.DeleteMonitor(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestClient_PauseMonitor(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		var req UpdateMonitorRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if req.Paused == nil || *req.Paused != true {
			t.Errorf("expected Paused to be true, got %v", req.Paused)
		}

		response := Monitor{
			UUID:     "mon_test",
			Name:     "Test Monitor",
			Protocol: "http",
			Paused:   true,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	result, err := c.PauseMonitor(context.Background(), "mon_test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Paused {
		t.Error("expected monitor to be paused")
	}
}

func TestClient_ResumeMonitor(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		var req UpdateMonitorRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if req.Paused == nil || *req.Paused != false {
			t.Errorf("expected Paused to be false, got %v", req.Paused)
		}

		response := Monitor{
			UUID:     "mon_test",
			Name:     "Test Monitor",
			Protocol: "http",
			Paused:   false,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	result, err := c.ResumeMonitor(context.Background(), "mon_test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Paused {
		t.Error("expected monitor to not be paused")
	}
}

// Additional comprehensive tests

func TestClient_ListMonitors_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid API key"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	_, err := c.ListMonitors(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsUnauthorized(err) {
		t.Errorf("expected unauthorized error, got %v", err)
	}
}

func TestClient_ListMonitors_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Internal error"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	_, err := c.ListMonitors(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsServerError(err) {
		t.Errorf("expected server error, got %v", err)
	}
}

func TestClient_ListMonitors_InvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not valid json at all"))
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	_, err := c.ListMonitors(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestClient_ListMonitors_EmptyWrappedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"other_field": "value",
		})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	result, err := c.ListMonitors(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 monitors for empty wrapped response, got %d", len(result))
	}
}

func TestClient_GetMonitor_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	_, err := c.GetMonitor(context.Background(), "mon_test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsServerError(err) {
		t.Errorf("expected server error, got %v", err)
	}
}

func TestClient_GetMonitor_FullResponse(t *testing.T) {
	monitor := Monitor{
		UUID:               "mon_test_123",
		Name:               "Production Monitor",
		URL:                "https://api.example.com/health",
		Protocol:           "http",
		HTTPMethod:         "GET",
		CheckFrequency:     60,
		Regions:            []string{"london", "frankfurt", "virginia"},
		RequestHeaders:     []RequestHeader{{Name: "X-Custom", Value: "value"}},
		ExpectedStatusCode: "2xx",
		FollowRedirects:    true,
		Paused:             false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(monitor)
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	result, err := c.GetMonitor(context.Background(), "mon_test_123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.UUID != monitor.UUID {
		t.Errorf("UUID = %q, expected %q", result.UUID, monitor.UUID)
	}
	if result.Name != monitor.Name {
		t.Errorf("Name = %q, expected %q", result.Name, monitor.Name)
	}
	if result.URL != monitor.URL {
		t.Errorf("URL = %q, expected %q", result.URL, monitor.URL)
	}
	if result.HTTPMethod != monitor.HTTPMethod {
		t.Errorf("HTTPMethod = %q, expected %q", result.HTTPMethod, monitor.HTTPMethod)
	}
	if result.CheckFrequency != monitor.CheckFrequency {
		t.Errorf("CheckFrequency = %d, expected %d", result.CheckFrequency, monitor.CheckFrequency)
	}
	if len(result.Regions) != len(monitor.Regions) {
		t.Errorf("len(Regions) = %d, expected %d", len(result.Regions), len(monitor.Regions))
	}
	if result.ExpectedStatusCode != monitor.ExpectedStatusCode {
		t.Errorf("ExpectedStatusCode = %q, expected %q", result.ExpectedStatusCode, monitor.ExpectedStatusCode)
	}
	if result.FollowRedirects != monitor.FollowRedirects {
		t.Errorf("FollowRedirects = %v, expected %v", result.FollowRedirects, monitor.FollowRedirects)
	}
	if len(result.RequestHeaders) != 1 || result.RequestHeaders[0].Name != "X-Custom" {
		t.Errorf("RequestHeaders not properly set")
	}
}

func TestClient_CreateMonitor_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	_, err := c.CreateMonitor(context.Background(), CreateMonitorRequest{
		Name:     "Test",
		URL:      "https://example.com",
		Protocol: "http",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsServerError(err) {
		t.Errorf("expected server error, got %v", err)
	}
}

// assertCreateMonitorRequest validates the core fields of a decoded CreateMonitorRequest.
func assertCreateMonitorRequest(t *testing.T, req CreateMonitorRequest, expectedBody string) {
	t.Helper()
	if req.Name != "Full Monitor" {
		t.Errorf("Name = %q, expected 'Full Monitor'", req.Name)
	}
	if req.URL != "https://api.example.com" {
		t.Errorf("URL = %q, expected 'https://api.example.com'", req.URL)
	}
	if req.Protocol != "http" {
		t.Errorf("Protocol = %q, expected 'http'", req.Protocol)
	}
	if req.HTTPMethod != "POST" {
		t.Errorf("HTTPMethod = %q, expected 'POST'", req.HTTPMethod)
	}
	if req.CheckFrequency != 120 {
		t.Errorf("CheckFrequency = %d, expected 120", req.CheckFrequency)
	}
	assertCreateMonitorCollections(t, req, expectedBody)
}

// assertCreateMonitorCollections validates collections and optional fields of a CreateMonitorRequest.
func assertCreateMonitorCollections(t *testing.T, req CreateMonitorRequest, expectedBody string) {
	t.Helper()
	if len(req.Regions) != 2 {
		t.Errorf("len(Regions) = %d, expected 2", len(req.Regions))
	}
	if req.ExpectedStatusCode != "201" {
		t.Errorf("ExpectedStatusCode = %q, expected '201'", req.ExpectedStatusCode)
	}
	if req.FollowRedirects == nil || !*req.FollowRedirects {
		t.Errorf("FollowRedirects = %v, expected true", req.FollowRedirects)
	}
	if req.RequestBody == nil || *req.RequestBody != expectedBody {
		t.Errorf("RequestBody = %v, expected %q", req.RequestBody, expectedBody)
	}
	if len(req.RequestHeaders) != 1 || req.RequestHeaders[0].Name != "Content-Type" {
		t.Errorf("RequestHeaders not properly set")
	}
}

func TestClient_CreateMonitor_FullRequest(t *testing.T) {
	followRedirects := true
	body := `{"key": "value"}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateMonitorRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		assertCreateMonitorRequest(t, req, body)

		response := Monitor{
			UUID:               "mon_new",
			Name:               req.Name,
			URL:                req.URL,
			Protocol:           req.Protocol,
			HTTPMethod:         req.HTTPMethod,
			CheckFrequency:     req.CheckFrequency,
			Regions:            req.Regions,
			ExpectedStatusCode: FlexibleString(req.ExpectedStatusCode),
			FollowRedirects:    *req.FollowRedirects,
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	result, err := c.CreateMonitor(context.Background(), CreateMonitorRequest{
		Name:               "Full Monitor",
		URL:                "https://api.example.com",
		Protocol:           "http",
		HTTPMethod:         "POST",
		CheckFrequency:     120,
		Regions:            []string{"london", "virginia"},
		RequestHeaders:     []RequestHeader{{Name: "Content-Type", Value: "application/json"}},
		RequestBody:        &body,
		ExpectedStatusCode: "201",
		FollowRedirects:    &followRedirects,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.UUID != "mon_new" {
		t.Errorf("UUID = %q, expected 'mon_new'", result.UUID)
	}
}

func TestClient_UpdateMonitor_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	name := "Updated"
	_, err := c.UpdateMonitor(context.Background(), "mon_test", UpdateMonitorRequest{
		Name: &name,
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsServerError(err) {
		t.Errorf("expected server error, got %v", err)
	}
}

func TestClient_UpdateMonitor_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Monitor not found"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	name := "Updated"
	_, err := c.UpdateMonitor(context.Background(), "nonexistent", UpdateMonitorRequest{
		Name: &name,
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestClient_UpdateMonitor_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Validation failed",
			"details": []map[string]string{
				{"field": "check_frequency", "message": "Invalid frequency value"},
			},
		})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	freq := 999
	_, err := c.UpdateMonitor(context.Background(), "mon_test", UpdateMonitorRequest{
		CheckFrequency: &freq,
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsValidation(err) {
		t.Errorf("expected validation error, got %v", err)
	}
}

func TestClient_UpdateMonitor_FullRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req UpdateMonitorRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		// Verify all fields are present
		if req.Name == nil {
			t.Error("Name should not be nil")
		}
		if req.URL == nil {
			t.Error("URL should not be nil")
		}
		if req.HTTPMethod == nil {
			t.Error("HTTPMethod should not be nil")
		}
		if req.CheckFrequency == nil {
			t.Error("CheckFrequency should not be nil")
		}
		if req.ExpectedStatusCode == nil {
			t.Error("ExpectedStatusCode should not be nil")
		}
		if req.FollowRedirects == nil {
			t.Error("FollowRedirects should not be nil")
		}
		if req.Paused == nil {
			t.Error("Paused should not be nil")
		}

		response := Monitor{
			UUID:               "mon_test",
			Name:               *req.Name,
			URL:                *req.URL,
			Protocol:           "http",
			HTTPMethod:         *req.HTTPMethod,
			CheckFrequency:     *req.CheckFrequency,
			ExpectedStatusCode: FlexibleString(*req.ExpectedStatusCode),
			FollowRedirects:    *req.FollowRedirects,
			Paused:             *req.Paused,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	name := "Updated Name"
	url := "https://updated.example.com"
	method := "POST"
	frequency := 300
	expectedStatus := "201"
	followRedirects := false
	paused := true

	result, err := c.UpdateMonitor(context.Background(), "mon_test", UpdateMonitorRequest{
		Name:               &name,
		URL:                &url,
		HTTPMethod:         &method,
		CheckFrequency:     &frequency,
		ExpectedStatusCode: &expectedStatus,
		FollowRedirects:    &followRedirects,
		Paused:             &paused,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Name != name {
		t.Errorf("Name = %q, expected %q", result.Name, name)
	}
	if result.Paused != paused {
		t.Errorf("Paused = %v, expected %v", result.Paused, paused)
	}
}

func TestClient_DeleteMonitor_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	err := c.DeleteMonitor(context.Background(), "mon_test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsServerError(err) {
		t.Errorf("expected server error, got %v", err)
	}
}

func TestClient_DeleteMonitor_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid API key"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	err := c.DeleteMonitor(context.Background(), "mon_test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsUnauthorized(err) {
		t.Errorf("expected unauthorized error, got %v", err)
	}
}

func TestClient_PauseMonitor_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Monitor not found"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	_, err := c.PauseMonitor(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestClient_ResumeMonitor_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Monitor not found"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	_, err := c.ResumeMonitor(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestParseMonitorListResponse_InvalidJSON(t *testing.T) {
	raw := []byte(`{invalid json}`)
	_, err := parseMonitorListResponse(raw)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestClient_ListMonitors_NonArrayNonObjectResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`"just a string"`))
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	_, err := c.ListMonitors(context.Background())
	if err == nil {
		t.Fatal("expected error for non-array/non-object JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse monitors response") {
		t.Errorf("expected parse error, got %v", err)
	}
}

func TestClient_ListMonitors_NumberResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`12345`))
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	_, err := c.ListMonitors(context.Background())
	if err == nil {
		t.Fatal("expected error for number JSON, got nil")
	}
}

func TestClient_ListMonitors_NullResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`null`))
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	result, err := c.ListMonitors(context.Background())
	if err != nil {
		t.Fatalf("unexpected error for null JSON: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 monitors for null response, got %d", len(result))
	}
}

func TestClient_ListMonitors_BooleanResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`true`))
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	_, err := c.ListMonitors(context.Background())
	if err == nil {
		t.Fatal("expected error for boolean JSON, got nil")
	}
}

func TestParseMonitorListResponse_ArrayWithMultipleMonitors(t *testing.T) {
	raw := []byte(`[
		{"uuid": "mon_001", "name": "Monitor 1", "url": "https://example1.com"},
		{"uuid": "mon_002", "name": "Monitor 2", "url": "https://example2.com"},
		{"uuid": "mon_003", "name": "Monitor 3", "url": "https://example3.com"}
	]`)

	monitors, err := parseMonitorListResponse(raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(monitors) != 3 {
		t.Errorf("expected 3 monitors, got %d", len(monitors))
	}
}

func TestParseMonitorListResponse_EmptyObject(t *testing.T) {
	raw := []byte(`{}`)

	monitors, err := parseMonitorListResponse(raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(monitors) != 0 {
		t.Errorf("expected 0 monitors, got %d", len(monitors))
	}
}

func TestClient_ListMonitors_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			return
		case <-time.After(100 * time.Millisecond):
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]Monitor{})
		}
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.ListMonitors(ctx)
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
}

func TestClient_CreateMonitor_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			return
		case <-time.After(100 * time.Millisecond):
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(Monitor{UUID: "mon_new"})
		}
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.CreateMonitor(ctx, CreateMonitorRequest{Name: "Test", URL: "https://example.com", Protocol: "http"})
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
}

func TestMonitor_JSONMarshaling(t *testing.T) {
	body := `{"test": "data"}`
	monitor := Monitor{
		UUID:               "mon_test",
		Name:               "Test Monitor",
		URL:                "https://example.com",
		Protocol:           "http",
		HTTPMethod:         "POST",
		CheckFrequency:     60,
		Regions:            []string{"london", "frankfurt"},
		RequestHeaders:     []RequestHeader{{Name: "X-Custom", Value: "value"}},
		RequestBody:        body,
		ExpectedStatusCode: "2xx",
		FollowRedirects:    true,
		Paused:             false,
	}

	// Marshal to JSON
	data, err := json.Marshal(monitor)
	if err != nil {
		t.Fatalf("failed to marshal monitor: %v", err)
	}

	// Unmarshal back
	var decoded Monitor
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal monitor: %v", err)
	}

	// Verify fields
	if decoded.UUID != monitor.UUID {
		t.Errorf("UUID = %q, expected %q", decoded.UUID, monitor.UUID)
	}
	if decoded.RequestBody != body {
		t.Errorf("RequestBody = %q, expected %q", decoded.RequestBody, body)
	}
	if len(decoded.Regions) != 2 {
		t.Errorf("len(Regions) = %d, expected 2", len(decoded.Regions))
	}
}

func TestCreateMonitorRequest_JSONMarshaling(t *testing.T) {
	followRedirects := true
	body := `{"key": "value"}`

	req := CreateMonitorRequest{
		Name:               "Test",
		URL:                "https://example.com",
		Protocol:           "http",
		HTTPMethod:         "POST",
		CheckFrequency:     120,
		Regions:            []string{"london"},
		RequestHeaders:     []RequestHeader{{Name: "Auth", Value: "Bearer token"}},
		RequestBody:        &body,
		ExpectedStatusCode: "201",
		FollowRedirects:    &followRedirects,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	// Verify expected fields are present with correct names
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded["name"] != "Test" {
		t.Errorf("name = %v, expected 'Test'", decoded["name"])
	}
	if decoded["http_method"] != "POST" {
		t.Errorf("http_method = %v, expected 'POST'", decoded["http_method"])
	}
	if decoded["follow_redirects"] != true {
		t.Errorf("follow_redirects = %v, expected true", decoded["follow_redirects"])
	}
}

func TestClient_ListMonitors_NewFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"uuid":"mon_abc","url":"https://example.com","name":"Test Monitor",` +
			`"protocol":"https","status":"up","ssl_expiration":30,"projectUuid":"proj_abc123",` +
			`"paused":false,"check_frequency":60}]`))
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))
	monitors, err := c.ListMonitors(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(monitors) != 1 {
		t.Fatalf("expected 1 monitor, got %d", len(monitors))
	}
	if monitors[0].Status != "up" {
		t.Errorf("expected status 'up', got %q", monitors[0].Status)
	}
	if monitors[0].SSLExpiration == nil {
		t.Fatal("expected SSLExpiration to be non-nil")
	}
	if *monitors[0].SSLExpiration != 30 {
		t.Errorf("expected ssl_expiration 30, got %d", *monitors[0].SSLExpiration)
	}
	if monitors[0].ProjectUUID != "proj_abc123" {
		t.Errorf("expected projectUuid 'proj_abc123', got %q", monitors[0].ProjectUUID)
	}
}

func TestClient_ListMonitors_NilSSLExpiration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"uuid":"mon_abc","url":"https://example.com","name":"Test Monitor",` +
			`"protocol":"https","status":"up","projectUuid":"proj_abc123",` +
			`"paused":false,"check_frequency":60}]`))
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))
	monitors, err := c.ListMonitors(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(monitors) != 1 {
		t.Fatalf("expected 1 monitor, got %d", len(monitors))
	}
	if monitors[0].SSLExpiration != nil {
		t.Errorf("expected SSLExpiration nil, got %v", *monitors[0].SSLExpiration)
	}
}

func TestClient_ListMonitors_EmptyProjectUUID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"uuid":"mon_abc","url":"https://example.com","name":"Test Monitor",` +
			`"protocol":"https","status":"up","projectUuid":"",` +
			`"paused":false,"check_frequency":60}]`))
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))
	monitors, err := c.ListMonitors(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(monitors) != 1 {
		t.Fatalf("expected 1 monitor, got %d", len(monitors))
	}
	if monitors[0].ProjectUUID != "" {
		t.Errorf("expected empty ProjectUUID, got %q", monitors[0].ProjectUUID)
	}
}

func TestUpdateMonitorRequest_JSONMarshaling_OmitsNilFields(t *testing.T) {
	name := "Updated"
	req := UpdateMonitorRequest{
		Name: &name,
		// All other fields are nil
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded["name"] != "Updated" {
		t.Errorf("name = %v, expected 'Updated'", decoded["name"])
	}

	// These fields should not be present (omitempty)
	if _, exists := decoded["url"]; exists {
		t.Error("url should not be present when nil")
	}
	if _, exists := decoded["http_method"]; exists {
		t.Error("http_method should not be present when nil")
	}
	if _, exists := decoded["check_frequency"]; exists {
		t.Error("check_frequency should not be present when nil")
	}
}
