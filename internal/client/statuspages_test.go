// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// =============================================================================
// Test Helpers
// =============================================================================

func setupStatusPageTestServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)
	client := NewClient("test-key", WithBaseURL(server.URL), WithMaxRetries(0))
	return server, client
}

func stringPtr(s string) *string {
	return &s
}

// =============================================================================
// ListStatusPages Tests
// =============================================================================

func TestListStatusPages_Success(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/statuspages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		response := StatusPagePaginatedResponse{
			StatusPages: []StatusPage{
				{
					UUID:            "sp_abc123",
					Name:            "Production Status",
					HostedSubdomain: "mycompany.hyperping.app",
					URL:             "https://mycompany.hyperping.app",
					Settings: StatusPageSettings{
						Name:        "Production Status",
						Theme:       "light",
						Font:        "Inter",
						AccentColor: "#36b27e",
						Languages:   []string{"en"},
					},
				},
			},
			HasNextPage:    false,
			Total:          1,
			Page:           0,
			ResultsPerPage: 20,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	result, err := client.ListStatusPages(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.StatusPages) != 1 {
		t.Errorf("expected 1 status page, got %d", len(result.StatusPages))
	}
	if result.StatusPages[0].UUID != "sp_abc123" {
		t.Errorf("unexpected UUID: %s", result.StatusPages[0].UUID)
	}
	if result.Total != 1 {
		t.Errorf("expected total=1, got %d", result.Total)
	}
}

func TestListStatusPages_WithPagination(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		if page != "2" {
			t.Errorf("expected page=2, got %s", page)
		}

		response := StatusPagePaginatedResponse{
			StatusPages:    []StatusPage{},
			HasNextPage:    false,
			Total:          5,
			Page:           2,
			ResultsPerPage: 20,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	page := 2
	result, err := client.ListStatusPages(context.Background(), &page, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Page != 2 {
		t.Errorf("expected page=2, got %d", result.Page)
	}
}

func TestListStatusPages_WithSearch(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("search")
		if search != "production" {
			t.Errorf("expected search=production, got %s", search)
		}

		response := StatusPagePaginatedResponse{
			StatusPages:    []StatusPage{},
			HasNextPage:    false,
			Total:          1,
			Page:           0,
			ResultsPerPage: 20,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	search := "production"
	_, err := client.ListStatusPages(context.Background(), nil, &search)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// GetStatusPage Tests
// =============================================================================

func TestGetStatusPage_Success(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/statuspages/sp_abc123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		statusPage := StatusPage{
			UUID:            "sp_abc123",
			Name:            "Production Status",
			HostedSubdomain: "mycompany.hyperping.app",
			URL:             "https://mycompany.hyperping.app",
			Settings: StatusPageSettings{
				Name:            "Production Status",
				Theme:           "dark",
				Font:            "Inter",
				AccentColor:     "#36b27e",
				Languages:       []string{"en", "fr"},
				DefaultLanguage: "en",
				AutoRefresh:     true,
				BannerHeader:    true,
				LogoHeight:      "32px",
				Subscribe: StatusPageSubscribeSettings{
					Enabled: true,
					Email:   true,
					Slack:   false,
					Teams:   false,
					SMS:     false,
				},
				Authentication: StatusPageAuthenticationSettings{
					PasswordProtection: false,
					GoogleSSO:          false,
					SAMLSSO:            false,
					AllowedDomains:     []string{},
				},
			},
			Sections: []StatusPageSection{
				{
					Name:    map[string]string{"en": "API Services"},
					IsSplit: true,
					Services: []StatusPageService{
						{
							ID:                "svc_1",
							UUID:              "mon_xyz789",
							Name:              map[string]string{"en": "Main API"},
							IsGroup:           false,
							ShowUptime:        true,
							ShowResponseTimes: true,
						},
					},
				},
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(statusPage)
	})
	defer server.Close()

	result, err := client.GetStatusPage(context.Background(), "sp_abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.UUID != "sp_abc123" {
		t.Errorf("unexpected UUID: %s", result.UUID)
	}
	if result.Settings.Theme != "dark" {
		t.Errorf("unexpected theme: %s", result.Settings.Theme)
	}
	if len(result.Sections) != 1 {
		t.Errorf("expected 1 section, got %d", len(result.Sections))
	}
}

func TestGetStatusPage_InvalidUUID(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not make request with invalid UUID")
	})
	defer server.Close()

	_, err := client.GetStatusPage(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty UUID")
	}
}

func TestGetStatusPage_NotFound(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Status page not found",
		})
	})
	defer server.Close()

	_, err := client.GetStatusPage(context.Background(), "sp_nonexistent")
	if err == nil {
		t.Error("expected error for non-existent status page")
	}
}

// =============================================================================
// CreateStatusPage Tests
// =============================================================================

func TestCreateStatusPage_Success(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/statuspages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateStatusPageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Name != "Test Status Page" {
			t.Errorf("unexpected name: %s", req.Name)
		}

		response := struct {
			Message    string     `json:"message"`
			StatusPage StatusPage `json:"statuspage"`
		}{
			Message: "Status page created",
			StatusPage: StatusPage{
				UUID:            "sp_new123",
				Name:            req.Name,
				HostedSubdomain: "test.hyperping.app",
				URL:             "https://test.hyperping.app",
				Settings: StatusPageSettings{
					Name:        req.Name,
					Theme:       "light",
					Font:        "Inter",
					AccentColor: "#36b27e",
					Languages:   []string{"en"},
				},
			},
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	req := CreateStatusPageRequest{
		Name:      "Test Status Page",
		Subdomain: stringPtr("test"),
		Theme:     stringPtr("light"),
		Languages: []string{"en"},
	}

	result, err := client.CreateStatusPage(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.UUID != "sp_new123" {
		t.Errorf("unexpected UUID: %s", result.UUID)
	}
	if result.Name != "Test Status Page" {
		t.Errorf("unexpected name: %s", result.Name)
	}
}

func TestCreateStatusPage_WithAllFields(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		var req CreateStatusPageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		// Verify all fields
		if req.Name != "Full Featured Status Page" {
			t.Errorf("unexpected name: %s", req.Name)
		}
		if *req.Theme != "dark" {
			t.Errorf("unexpected theme: %s", *req.Theme)
		}
		if *req.Font != "Roboto" {
			t.Errorf("unexpected font: %s", *req.Font)
		}
		if *req.AccentColor != "#ff0000" {
			t.Errorf("unexpected accent color: %s", *req.AccentColor)
		}
		if *req.AutoRefresh != true {
			t.Error("expected auto_refresh to be true")
		}
		if len(req.Languages) != 2 {
			t.Errorf("expected 2 languages, got %d", len(req.Languages))
		}

		response := struct {
			Message    string     `json:"message"`
			StatusPage StatusPage `json:"statuspage"`
		}{
			Message: "Status page created",
			StatusPage: StatusPage{
				UUID: "sp_full123",
				Name: req.Name,
			},
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	req := CreateStatusPageRequest{
		Name:        "Full Featured Status Page",
		Subdomain:   stringPtr("full"),
		Theme:       stringPtr("dark"),
		Font:        stringPtr("Roboto"),
		AccentColor: stringPtr("#ff0000"),
		AutoRefresh: boolPtr(true),
		Languages:   []string{"en", "fr"},
		Subscribe: &CreateStatusPageSubscribeSettings{
			Enabled: boolPtr(true),
			Email:   boolPtr(true),
			Slack:   boolPtr(false),
		},
		Sections: []CreateStatusPageSection{
			{
				Name:    "API Services",
				IsSplit: boolPtr(true),
				Services: []CreateStatusPageService{
					{
						MonitorUUID:       "mon_123",
						NameShown:         stringPtr("Main API"),
						ShowUptime:        boolPtr(true),
						ShowResponseTimes: boolPtr(true),
					},
				},
			},
		},
	}

	_, err := client.CreateStatusPage(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateStatusPage_ValidationError(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not make request with invalid input")
	})
	defer server.Close()

	// Name too long (> 255 chars)
	longName := string(make([]byte, 256))
	for i := range longName {
		longName = string(append([]byte(longName[:i]), 'a'))
	}

	req := CreateStatusPageRequest{
		Name: longName,
	}

	_, err := client.CreateStatusPage(context.Background(), req)
	if err == nil {
		t.Error("expected validation error for long name")
	}
}

// =============================================================================
// UpdateStatusPage Tests
// =============================================================================

func TestUpdateStatusPage_Success(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/v1/statuspages/sp_abc123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req UpdateStatusPageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if *req.Name != "Updated Name" {
			t.Errorf("unexpected name: %s", *req.Name)
		}

		response := struct {
			Message    string     `json:"message"`
			StatusPage StatusPage `json:"statuspage"`
		}{
			Message: "Status page updated",
			StatusPage: StatusPage{
				UUID: "sp_abc123",
				Name: *req.Name,
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	req := UpdateStatusPageRequest{
		Name: stringPtr("Updated Name"),
	}

	result, err := client.UpdateStatusPage(context.Background(), "sp_abc123", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "Updated Name" {
		t.Errorf("unexpected name: %s", result.Name)
	}
}

func TestUpdateStatusPage_PartialUpdate(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		var req UpdateStatusPageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		// Only theme should be set
		if req.Theme == nil {
			t.Error("expected theme to be set")
		}
		if req.Name != nil {
			t.Error("expected name to be nil")
		}

		response := struct {
			Message    string     `json:"message"`
			StatusPage StatusPage `json:"statuspage"`
		}{
			Message: "Status page updated",
			StatusPage: StatusPage{
				UUID: "sp_abc123",
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	req := UpdateStatusPageRequest{
		Theme: stringPtr("dark"),
	}

	_, err := client.UpdateStatusPage(context.Background(), "sp_abc123", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// DeleteStatusPage Tests
// =============================================================================

func TestDeleteStatusPage_Success(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/v1/statuspages/sp_abc123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Status page deleted",
		})
	})
	defer server.Close()

	err := client.DeleteStatusPage(context.Background(), "sp_abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteStatusPage_InvalidUUID(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not make request with invalid UUID")
	})
	defer server.Close()

	err := client.DeleteStatusPage(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty UUID")
	}
}

// =============================================================================
// ListSubscribers Tests
// =============================================================================

func TestListSubscribers_Success(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/statuspages/sp_abc123/subscribers" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		response := SubscriberPaginatedResponse{
			Subscribers: []StatusPageSubscriber{
				{
					ID:        1,
					Type:      "email",
					Value:     "user@example.com",
					Email:     stringPtr("user@example.com"),
					CreatedAt: "2026-01-31T10:00:00Z",
				},
			},
			HasNextPage:    false,
			Total:          1,
			Page:           0,
			ResultsPerPage: 20,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	result, err := client.ListSubscribers(context.Background(), "sp_abc123", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Subscribers) != 1 {
		t.Errorf("expected 1 subscriber, got %d", len(result.Subscribers))
	}
	if result.Subscribers[0].Type != "email" {
		t.Errorf("unexpected type: %s", result.Subscribers[0].Type)
	}
}

func TestListSubscribers_WithTypeFilter(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		typeParam := r.URL.Query().Get("type")
		if typeParam != "sms" {
			t.Errorf("expected type=sms, got %s", typeParam)
		}

		response := SubscriberPaginatedResponse{
			Subscribers:    []StatusPageSubscriber{},
			HasNextPage:    false,
			Total:          0,
			Page:           0,
			ResultsPerPage: 20,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	subscriberType := "sms"
	_, err := client.ListSubscribers(context.Background(), "sp_abc123", nil, &subscriberType)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListSubscribers_TypeAll(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		// "all" should not be included in query params
		if r.URL.Query().Has("type") {
			t.Error("expected no type parameter for 'all'")
		}

		response := SubscriberPaginatedResponse{
			Subscribers:    []StatusPageSubscriber{},
			HasNextPage:    false,
			Total:          0,
			Page:           0,
			ResultsPerPage: 20,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	subscriberType := "all"
	_, err := client.ListSubscribers(context.Background(), "sp_abc123", nil, &subscriberType)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// AddSubscriber Tests
// =============================================================================

func TestAddSubscriber_Email_Success(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/statuspages/sp_abc123/subscribers" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req AddSubscriberRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Type != "email" {
			t.Errorf("unexpected type: %s", req.Type)
		}
		if *req.Email != "user@example.com" {
			t.Errorf("unexpected email: %s", *req.Email)
		}

		response := struct {
			Message    string               `json:"message"`
			Subscriber StatusPageSubscriber `json:"subscriber"`
		}{
			Message: "Subscriber added",
			Subscriber: StatusPageSubscriber{
				ID:        1,
				Type:      "email",
				Value:     *req.Email,
				Email:     req.Email,
				CreatedAt: "2026-01-31T10:00:00Z",
			},
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	req := AddSubscriberRequest{
		Type:  "email",
		Email: stringPtr("user@example.com"),
	}

	result, err := client.AddSubscriber(context.Background(), "sp_abc123", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Type != "email" {
		t.Errorf("unexpected type: %s", result.Type)
	}
	if *result.Email != "user@example.com" {
		t.Errorf("unexpected email: %s", *result.Email)
	}
}

func TestAddSubscriber_SMS_Success(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		var req AddSubscriberRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Type != "sms" {
			t.Errorf("unexpected type: %s", req.Type)
		}
		if *req.Phone != "+1234567890" {
			t.Errorf("unexpected phone: %s", *req.Phone)
		}

		response := struct {
			Message    string               `json:"message"`
			Subscriber StatusPageSubscriber `json:"subscriber"`
		}{
			Message: "Subscriber added",
			Subscriber: StatusPageSubscriber{
				ID:        2,
				Type:      "sms",
				Value:     *req.Phone,
				Phone:     req.Phone,
				CreatedAt: "2026-01-31T10:00:00Z",
			},
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	req := AddSubscriberRequest{
		Type:  "sms",
		Phone: stringPtr("+1234567890"),
	}

	result, err := client.AddSubscriber(context.Background(), "sp_abc123", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Type != "sms" {
		t.Errorf("unexpected type: %s", result.Type)
	}
}

func TestAddSubscriber_Teams_Success(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		var req AddSubscriberRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Type != "teams" {
			t.Errorf("unexpected type: %s", req.Type)
		}
		if *req.TeamsWebhookURL != "https://outlook.office.com/webhook/..." {
			t.Errorf("unexpected webhook URL: %s", *req.TeamsWebhookURL)
		}

		response := struct {
			Message    string               `json:"message"`
			Subscriber StatusPageSubscriber `json:"subscriber"`
		}{
			Message: "Subscriber added",
			Subscriber: StatusPageSubscriber{
				ID:        3,
				Type:      "teams",
				Value:     *req.TeamsWebhookURL,
				CreatedAt: "2026-01-31T10:00:00Z",
			},
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	req := AddSubscriberRequest{
		Type:            "teams",
		TeamsWebhookURL: stringPtr("https://outlook.office.com/webhook/..."),
	}

	_, err := client.AddSubscriber(context.Background(), "sp_abc123", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddSubscriber_Slack_Rejected(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not make request for Slack subscriber")
	})
	defer server.Close()

	req := AddSubscriberRequest{
		Type: "slack",
	}

	_, err := client.AddSubscriber(context.Background(), "sp_abc123", req)
	if err == nil {
		t.Error("expected error for Slack subscriber type")
	}
	if err.Error() == "" {
		t.Error("error message should not be empty")
	}
}

func TestAddSubscriber_ValidationErrors(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not make request with invalid input")
	})
	defer server.Close()

	tests := []struct {
		name string
		req  AddSubscriberRequest
	}{
		{
			name: "email type without email",
			req: AddSubscriberRequest{
				Type: "email",
			},
		},
		{
			name: "sms type without phone",
			req: AddSubscriberRequest{
				Type: "sms",
			},
		},
		{
			name: "teams type without webhook URL",
			req: AddSubscriberRequest{
				Type: "teams",
			},
		},
		{
			name: "invalid type",
			req: AddSubscriberRequest{
				Type:  "invalid",
				Email: stringPtr("user@example.com"),
			},
		},
		{
			name: "invalid language",
			req: AddSubscriberRequest{
				Type:     "email",
				Email:    stringPtr("user@example.com"),
				Language: stringPtr("invalid"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.AddSubscriber(context.Background(), "sp_abc123", tt.req)
			if err == nil {
				t.Errorf("expected validation error for %s", tt.name)
			}
		})
	}
}

// =============================================================================
// DeleteSubscriber Tests
// =============================================================================

func TestDeleteSubscriber_Success(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/v1/statuspages/sp_abc123/subscribers/1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Subscriber deleted",
		})
	})
	defer server.Close()

	err := client.DeleteSubscriber(context.Background(), "sp_abc123", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteSubscriber_InvalidID(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not make request with invalid subscriber ID")
	})
	defer server.Close()

	err := client.DeleteSubscriber(context.Background(), "sp_abc123", 0)
	if err == nil {
		t.Error("expected error for subscriber ID 0")
	}

	err = client.DeleteSubscriber(context.Background(), "sp_abc123", -1)
	if err == nil {
		t.Error("expected error for negative subscriber ID")
	}
}

// =============================================================================
// Error Handling Tests
// =============================================================================

func TestStatusPages_Unauthorized(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid API key",
		})
	})
	defer server.Close()

	_, err := client.ListStatusPages(context.Background(), nil, nil)
	if err == nil {
		t.Error("expected error for unauthorized request")
	}
}

func TestStatusPages_RateLimit(t *testing.T) {
	server, client := setupStatusPageTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Rate limit exceeded",
		})
	})
	defer server.Close()

	_, err := client.GetStatusPage(context.Background(), "sp_abc123")
	if err == nil {
		t.Error("expected error for rate limit")
	}
}
