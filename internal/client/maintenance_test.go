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

func TestClient_ListMaintenance(t *testing.T) {
	startDate := "2025-12-17T14:30:00.000Z"
	endDate := "2025-12-17T18:30:00.000Z"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != MaintenanceBasePath {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		maintenance := []Maintenance{
			{
				UUID:      "mw_123",
				Name:      "Database Maintenance",
				Title:     LocalizedText{En: "Planned Database Upgrade"},
				StartDate: &startDate,
				EndDate:   &endDate,
				Monitors:  []string{"mon_001"},
			},
			{
				UUID:     "mw_456",
				Name:     "Network Upgrade",
				Title:    LocalizedText{En: "Network Infrastructure Update"},
				Monitors: []string{"mon_002", "mon_003"},
			},
		}
		json.NewEncoder(w).Encode(maintenance)
	}))
	defer server.Close()

	client := NewClient("test_key", WithBaseURL(server.URL))
	maintenance, err := client.ListMaintenance(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(maintenance) != 2 {
		t.Errorf("expected 2 maintenance windows, got %d", len(maintenance))
	}
	if maintenance[0].Name != "Database Maintenance" {
		t.Errorf("expected name 'Database Maintenance', got %s", maintenance[0].Name)
	}
}

func TestClient_ListMaintenance_WrappedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"maintenanceWindows": []Maintenance{
				{UUID: "mw_123", Name: "Planned Downtime", Title: LocalizedText{En: "Planned Downtime"}},
			},
			"hasNextPage": false,
			"total":       1,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("test_key", WithBaseURL(server.URL))
	maintenance, err := client.ListMaintenance(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(maintenance) != 1 {
		t.Errorf("expected 1 maintenance window, got %d", len(maintenance))
	}
}

func TestClient_ListMaintenance_DataWrappedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": []Maintenance{
				{UUID: "mw_789", Name: "Server Migration", Title: LocalizedText{En: "Server Migration"}},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("test_key", WithBaseURL(server.URL))
	maintenance, err := client.ListMaintenance(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(maintenance) != 1 {
		t.Errorf("expected 1 maintenance window, got %d", len(maintenance))
	}
}

func TestClient_ListMaintenance_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer server.Close()

	client := NewClient("test_key", WithBaseURL(server.URL))
	maintenance, err := client.ListMaintenance(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(maintenance) != 0 {
		t.Errorf("expected 0 maintenance windows, got %d", len(maintenance))
	}
}

func TestClient_ListMaintenance_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
	}))
	defer server.Close()

	client := NewClient("test_key", WithBaseURL(server.URL))
	_, err := client.ListMaintenance(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetMaintenance(t *testing.T) {
	startDate := "2025-12-17T14:30:00.000Z"
	endDate := "2025-12-17T18:30:00.000Z"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != MaintenanceBasePath+"/mw_123" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		maintenance := Maintenance{
			UUID:      "mw_123",
			Name:      "Database Maintenance",
			Title:     LocalizedText{En: "Planned Database Upgrade"},
			StartDate: &startDate,
			EndDate:   &endDate,
			Monitors:  []string{"mon_123", "mon_456"},
		}
		json.NewEncoder(w).Encode(maintenance)
	}))
	defer server.Close()

	client := NewClient("test_key", WithBaseURL(server.URL))
	maintenance, err := client.GetMaintenance(context.Background(), "mw_123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if maintenance.UUID != "mw_123" {
		t.Errorf("expected UUID 'mw_123', got %s", maintenance.UUID)
	}
	if maintenance.Name != "Database Maintenance" {
		t.Errorf("expected name 'Database Maintenance', got %s", maintenance.Name)
	}
	if len(maintenance.Monitors) != 2 {
		t.Errorf("expected 2 monitors, got %d", len(maintenance.Monitors))
	}
}

func TestClient_GetMaintenance_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}))
	defer server.Close()

	client := NewClient("test_key", WithBaseURL(server.URL))
	_, err := client.GetMaintenance(context.Background(), "nonexistent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got: %v", err)
	}
}

func TestClient_CreateMaintenance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle POST - return only UUID
		if r.Method == http.MethodPost && r.URL.Path == MaintenanceBasePath {
			var req CreateMaintenanceRequest
			json.NewDecoder(r.Body).Decode(&req)

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"uuid": "mw_new"})
			return
		}

		// Handle GET - return full maintenance
		if r.Method == http.MethodGet && r.URL.Path == MaintenanceBasePath+"/mw_new" {
			startDate := "2025-12-20T02:00:00.000Z"
			endDate := "2025-12-20T06:00:00.000Z"
			maintenance := Maintenance{
				UUID:      "mw_new",
				Name:      "Planned Downtime",
				Title:     LocalizedText{En: "System Maintenance"},
				StartDate: &startDate,
				EndDate:   &endDate,
				Monitors:  []string{"mon_123"},
			}
			json.NewEncoder(w).Encode(maintenance)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient("test_key", WithBaseURL(server.URL))
	notifyMinutes := 60
	req := CreateMaintenanceRequest{
		Name:                "Planned Downtime",
		Title:               LocalizedText{En: "System Maintenance"},
		Text:                LocalizedText{En: "We will be performing scheduled maintenance."},
		StartDate:           "2025-12-20T02:00:00.000Z",
		EndDate:             "2025-12-20T06:00:00.000Z",
		Monitors:            []string{"mon_123"},
		NotificationOption:  "scheduled",
		NotificationMinutes: &notifyMinutes,
	}
	maintenance, err := client.CreateMaintenance(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if maintenance.UUID != "mw_new" {
		t.Errorf("expected UUID 'mw_new', got %s", maintenance.UUID)
	}
	if maintenance.Name != "Planned Downtime" {
		t.Errorf("expected name 'Planned Downtime', got %s", maintenance.Name)
	}
}

func TestClient_CreateMaintenance_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "validation failed",
			"details": []map[string]string{
				{"field": "name", "message": "name is required"},
			},
		})
	}))
	defer server.Close()

	client := NewClient("test_key", WithBaseURL(server.URL))
	req := CreateMaintenanceRequest{
		Name:      "",
		StartDate: "2025-12-20T02:00:00.000Z",
		EndDate:   "2025-12-20T06:00:00.000Z",
	}
	_, err := client.CreateMaintenance(context.Background(), req)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_CreateMaintenance_ReadAfterCreateFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle POST - return UUID successfully
		if r.Method == http.MethodPost && r.URL.Path == MaintenanceBasePath {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"uuid": "mw_new"})
			return
		}

		// Handle GET - return 404 (simulating failure to read created resource)
		if r.Method == http.MethodGet && r.URL.Path == MaintenanceBasePath+"/mw_new" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient("test_key", WithBaseURL(server.URL))
	notifyMinutes := 60
	req := CreateMaintenanceRequest{
		Name:                "Test Maintenance",
		Title:               LocalizedText{En: "Test"},
		StartDate:           "2025-12-20T02:00:00.000Z",
		EndDate:             "2025-12-20T06:00:00.000Z",
		Monitors:            []string{"mon_123"},
		NotificationMinutes: &notifyMinutes,
	}
	_, err := client.CreateMaintenance(context.Background(), req)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read created maintenance") {
		t.Errorf("expected read-after-create error, got: %v", err)
	}
}

func TestClient_UpdateMaintenance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != MaintenanceBasePath+"/mw_123" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var req UpdateMaintenanceRequest
		json.NewDecoder(r.Body).Decode(&req)

		startDate := "2025-12-20T02:00:00.000Z"
		endDate := "2025-12-20T06:00:00.000Z"
		maintenance := Maintenance{
			UUID:      "mw_123",
			Name:      *req.Name,
			StartDate: &startDate,
			EndDate:   &endDate,
		}
		json.NewEncoder(w).Encode(maintenance)
	}))
	defer server.Close()

	client := NewClient("test_key", WithBaseURL(server.URL))
	name := "Updated Maintenance"
	req := UpdateMaintenanceRequest{
		Name: &name,
	}
	maintenance, err := client.UpdateMaintenance(context.Background(), "mw_123", req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if maintenance.Name != "Updated Maintenance" {
		t.Errorf("expected name 'Updated Maintenance', got %s", maintenance.Name)
	}
}

func TestClient_UpdateMaintenance_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}))
	defer server.Close()

	client := NewClient("test_key", WithBaseURL(server.URL))
	name := "Updated"
	_, err := client.UpdateMaintenance(context.Background(), "nonexistent", UpdateMaintenanceRequest{Name: &name})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got: %v", err)
	}
}

func TestClient_DeleteMaintenance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != MaintenanceBasePath+"/mw_123" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("test_key", WithBaseURL(server.URL))
	err := client.DeleteMaintenance(context.Background(), "mw_123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_DeleteMaintenance_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}))
	defer server.Close()

	client := NewClient("test_key", WithBaseURL(server.URL))
	err := client.DeleteMaintenance(context.Background(), "nonexistent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got: %v", err)
	}
}

func TestClient_DeleteMaintenance_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
	}))
	defer server.Close()

	client := NewClient("test_key", WithBaseURL(server.URL))
	err := client.DeleteMaintenance(context.Background(), "mw_123")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_ListMaintenance_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	_, err := c.ListMaintenance(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_ListMaintenance_EmptyWrappedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"maintenanceWindows": []Maintenance{},
			"hasNextPage":        false,
			"total":              0,
		})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL))

	maintenance, err := c.ListMaintenance(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(maintenance) != 0 {
		t.Errorf("expected 0 maintenance, got %d", len(maintenance))
	}
}

func TestClient_ListMaintenance_UnexpectedJSONType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HeaderContentType, ContentTypeJSON)
		w.Write([]byte(`12345`))
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	_, err := c.ListMaintenance(context.Background())
	if err == nil {
		t.Fatal("expected error for unexpected JSON type, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse maintenance response") {
		t.Errorf("expected parse error, got %v", err)
	}
}
