// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// =============================================================================
// Mock Server Helpers
// =============================================================================

type mockResponse struct {
	statusCode int
	body       interface{}
	headers    map[string]string
}

func setupMockServer(response mockResponse) (*httptest.Server, *Client) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Set custom headers
		for key, value := range response.headers {
			w.Header().Set(key, value)
		}

		w.WriteHeader(response.statusCode)

		if response.body != nil {
			if err := json.NewEncoder(w).Encode(response.body); err != nil {
				panic(err)
			}
		}
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	client := NewClient("test-key", WithBaseURL(server.URL), WithMaxRetries(0))
	return server, client
}

// =============================================================================
// UpdateStatusPage Mock Tests
// =============================================================================

func TestUpdateStatusPage_Success_Mock(t *testing.T) {
	mockResp := mockResponse{
		statusCode: http.StatusOK,
		body: map[string]interface{}{
			"message": "Status page updated",
			"statuspage": map[string]interface{}{
				"uuid":             "sp_test123",
				"name":             "Updated Status Page",
				"hosted_subdomain": "updated.hyperping.app",
				"url":              "https://updated.hyperping.app",
				"settings": map[string]interface{}{
					"name":         "Updated Status Page",
					"theme":        "dark",
					"font":         "Roboto",
					"accent_color": "#ff0000",
					"languages":    []string{"en", "fr"},
				},
			},
		},
	}

	server, client := setupMockServer(mockResp)
	defer server.Close()

	req := UpdateStatusPageRequest{
		Name:        stringPtr("Updated Status Page"),
		Theme:       stringPtr("dark"),
		Font:        stringPtr("Roboto"),
		AccentColor: stringPtr("#ff0000"),
	}

	result, err := client.UpdateStatusPage(context.Background(), "sp_test123", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Settings.Name != "Updated Status Page" {
		t.Errorf("expected name 'Updated Status Page', got %q", result.Settings.Name)
	}
	if result.Settings.Theme != "dark" {
		t.Errorf("expected theme 'dark', got %q", result.Settings.Theme)
	}
}

func TestUpdateStatusPage_NotFound_Mock(t *testing.T) {
	mockResp := mockResponse{
		statusCode: http.StatusNotFound,
		body: map[string]interface{}{
			"error":   "Not found",
			"message": "Status page not found",
		},
	}

	server, client := setupMockServer(mockResp)
	defer server.Close()

	req := UpdateStatusPageRequest{
		Name: stringPtr("New Name"),
	}

	_, err := client.UpdateStatusPage(context.Background(), "sp_nonexistent", req)
	if err == nil {
		t.Error("expected error for non-existent status page")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected APIError, got %T", err)
		return
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected status code 404, got %d", apiErr.StatusCode)
	}
}

func TestUpdateStatusPage_InvalidJSON_Mock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{invalid json"))
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL), WithMaxRetries(0))

	req := UpdateStatusPageRequest{
		Name: stringPtr("Test"),
	}

	_, err := client.UpdateStatusPage(context.Background(), "sp_test", req)
	if err == nil {
		t.Error("expected error for invalid JSON response")
	}
}

func TestUpdateStatusPage_EmptyResponse_Mock(t *testing.T) {
	mockResp := mockResponse{
		statusCode: http.StatusOK,
		body:       map[string]interface{}{},
	}

	server, client := setupMockServer(mockResp)
	defer server.Close()

	req := UpdateStatusPageRequest{
		Name: stringPtr("Test"),
	}

	result, err := client.UpdateStatusPage(context.Background(), "sp_test", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should handle empty response gracefully
	if result.UUID != "" {
		t.Error("expected empty UUID for empty response")
	}
}

func TestUpdateStatusPage_ServerError_Mock(t *testing.T) {
	mockResp := mockResponse{
		statusCode: http.StatusInternalServerError,
		body: map[string]interface{}{
			"error": "Internal server error",
		},
	}

	server, client := setupMockServer(mockResp)
	defer server.Close()

	req := UpdateStatusPageRequest{
		Name: stringPtr("Test"),
	}

	_, err := client.UpdateStatusPage(context.Background(), "sp_test", req)
	if err == nil {
		t.Error("expected error for server error")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected APIError, got %T", err)
		return
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status code 500, got %d", apiErr.StatusCode)
	}
}

func TestUpdateStatusPage_Unauthorized_Mock(t *testing.T) {
	mockResp := mockResponse{
		statusCode: http.StatusUnauthorized,
		body: map[string]interface{}{
			"error":   "Unauthorized",
			"message": "Invalid API key",
		},
	}

	server, client := setupMockServer(mockResp)
	defer server.Close()

	req := UpdateStatusPageRequest{
		Name: stringPtr("Test"),
	}

	_, err := client.UpdateStatusPage(context.Background(), "sp_test", req)
	if err == nil {
		t.Error("expected error for unauthorized request")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected APIError, got %T", err)
		return
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status code 401, got %d", apiErr.StatusCode)
	}
}

func TestUpdateStatusPage_BadRequest_Mock(t *testing.T) {
	mockResp := mockResponse{
		statusCode: http.StatusBadRequest,
		body: map[string]interface{}{
			"error":   "Bad request",
			"message": "Invalid input",
			"details": []map[string]string{
				{"field": "name", "message": "Name is required"},
			},
		},
	}

	server, client := setupMockServer(mockResp)
	defer server.Close()

	req := UpdateStatusPageRequest{
		Name: stringPtr(""),
	}

	_, err := client.UpdateStatusPage(context.Background(), "sp_test", req)
	if err == nil {
		t.Error("expected error for bad request")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected APIError, got %T", err)
		return
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status code 400, got %d", apiErr.StatusCode)
	}
	if len(apiErr.Details) == 0 {
		t.Error("expected validation details in error")
	}
}

func TestUpdateStatusPage_MissingRequiredFields_Mock(t *testing.T) {
	mockResp := mockResponse{
		statusCode: http.StatusOK,
		body: map[string]interface{}{
			"message": "Status page updated",
			"statuspage": map[string]interface{}{
				"uuid": "sp_test",
			},
		},
	}

	server, client := setupMockServer(mockResp)
	defer server.Close()

	req := UpdateStatusPageRequest{
		Theme: stringPtr("dark"),
	}

	result, err := client.UpdateStatusPage(context.Background(), "sp_test", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.UUID != "sp_test" {
		t.Errorf("expected UUID 'sp_test', got %q", result.UUID)
	}
}

// =============================================================================
// AddSubscriber Mock Tests
// =============================================================================

func TestAddSubscriber_Email_Success_Mock(t *testing.T) {
	mockResp := mockResponse{
		statusCode: http.StatusCreated,
		body: map[string]interface{}{
			"message": "Subscriber added",
			"subscriber": map[string]interface{}{
				"id":         123,
				"type":       "email",
				"value":      "test@example.com",
				"email":      "test@example.com",
				"created_at": "2026-02-10T10:00:00Z",
			},
		},
	}

	server, client := setupMockServer(mockResp)
	defer server.Close()

	req := AddSubscriberRequest{
		Type:  "email",
		Email: stringPtr("test@example.com"),
	}

	result, err := client.AddSubscriber(context.Background(), "sp_test", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 123 {
		t.Errorf("expected ID 123, got %d", result.ID)
	}
	if result.Type != "email" {
		t.Errorf("expected type 'email', got %q", result.Type)
	}
	if *result.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %q", *result.Email)
	}
}

func TestAddSubscriber_SMS_Success_Mock(t *testing.T) {
	mockResp := mockResponse{
		statusCode: http.StatusCreated,
		body: map[string]interface{}{
			"message": "Subscriber added",
			"subscriber": map[string]interface{}{
				"id":         456,
				"type":       "sms",
				"value":      "+1234567890",
				"phone":      "+1234567890",
				"created_at": "2026-02-10T10:00:00Z",
			},
		},
	}

	server, client := setupMockServer(mockResp)
	defer server.Close()

	req := AddSubscriberRequest{
		Type:  "sms",
		Phone: stringPtr("+1234567890"),
	}

	result, err := client.AddSubscriber(context.Background(), "sp_test", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 456 {
		t.Errorf("expected ID 456, got %d", result.ID)
	}
	if result.Type != "sms" {
		t.Errorf("expected type 'sms', got %q", result.Type)
	}
}

func TestAddSubscriber_Teams_Success_Mock(t *testing.T) {
	mockResp := mockResponse{
		statusCode: http.StatusCreated,
		body: map[string]interface{}{
			"message": "Subscriber added",
			"subscriber": map[string]interface{}{
				"id":         789,
				"type":       "teams",
				"value":      "https://outlook.office.com/webhook/test",
				"created_at": "2026-02-10T10:00:00Z",
			},
		},
	}

	server, client := setupMockServer(mockResp)
	defer server.Close()

	req := AddSubscriberRequest{
		Type:            "teams",
		TeamsWebhookURL: stringPtr("https://outlook.office.com/webhook/test"),
	}

	result, err := client.AddSubscriber(context.Background(), "sp_test", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 789 {
		t.Errorf("expected ID 789, got %d", result.ID)
	}
	if result.Type != "teams" {
		t.Errorf("expected type 'teams', got %q", result.Type)
	}
}

func TestAddSubscriber_Slack_Rejected_Mock(t *testing.T) {
	// No server needed - client should reject before making request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not make HTTP request for Slack subscriber")
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL), WithMaxRetries(0))

	req := AddSubscriberRequest{
		Type: "slack",
	}

	_, err := client.AddSubscriber(context.Background(), "sp_test", req)
	if err == nil {
		t.Fatal("expected error for Slack subscriber type")
	}

	errMsg := strings.ToLower(err.Error())
	if !strings.Contains(errMsg, "slack") {
		t.Errorf("expected error message to mention 'slack', got: %v", err)
	}
	// The actual validation error comes from the Validate() method which checks type
	// The specific OAuth message is added by AddSubscriber before calling the API
	// But if validation fails first, we won't reach that check
}

func TestAddSubscriber_InvalidJSON_Mock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL), WithMaxRetries(0))

	req := AddSubscriberRequest{
		Type:  "email",
		Email: stringPtr("test@example.com"),
	}

	_, err := client.AddSubscriber(context.Background(), "sp_test", req)
	if err == nil {
		t.Error("expected error for invalid JSON response")
	}
}

func TestAddSubscriber_NetworkError_Mock(t *testing.T) {
	// Create client pointing to non-existent server
	client := NewClient("test-key", WithBaseURL("http://localhost:1"), WithMaxRetries(0))

	req := AddSubscriberRequest{
		Type:  "email",
		Email: stringPtr("test@example.com"),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.AddSubscriber(ctx, "sp_test", req)
	if err == nil {
		t.Error("expected error for network failure")
	}
}

func TestAddSubscriber_Timeout_Mock(t *testing.T) {
	// Server that hangs
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL), WithMaxRetries(0))

	req := AddSubscriberRequest{
		Type:  "email",
		Email: stringPtr("test@example.com"),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.AddSubscriber(ctx, "sp_test", req)
	if err == nil {
		t.Error("expected error for timeout")
	}
}

func TestAddSubscriber_Conflict_Mock(t *testing.T) {
	mockResp := mockResponse{
		statusCode: http.StatusConflict,
		body: map[string]interface{}{
			"error":   "Conflict",
			"message": "Subscriber already exists",
		},
	}

	server, client := setupMockServer(mockResp)
	defer server.Close()

	req := AddSubscriberRequest{
		Type:  "email",
		Email: stringPtr("existing@example.com"),
	}

	_, err := client.AddSubscriber(context.Background(), "sp_test", req)
	if err == nil {
		t.Error("expected error for conflict")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected APIError, got %T", err)
		return
	}
	if apiErr.StatusCode != http.StatusConflict {
		t.Errorf("expected status code 409, got %d", apiErr.StatusCode)
	}
}

func TestAddSubscriber_ValidationError_Mock(t *testing.T) {
	mockResp := mockResponse{
		statusCode: http.StatusBadRequest,
		body: map[string]interface{}{
			"error":   "Validation error",
			"message": "Invalid subscriber data",
			"details": []map[string]string{
				{"field": "email", "message": "Invalid email format"},
			},
		},
	}

	server, client := setupMockServer(mockResp)
	defer server.Close()

	req := AddSubscriberRequest{
		Type:  "email",
		Email: stringPtr("not-an-email"),
	}

	_, err := client.AddSubscriber(context.Background(), "sp_test", req)
	if err == nil {
		t.Error("expected error for validation failure")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected APIError, got %T", err)
		return
	}
	if len(apiErr.Details) == 0 {
		t.Error("expected validation details in error")
	}
}

// =============================================================================
// DeleteSubscriber Mock Tests
// =============================================================================

func TestDeleteSubscriber_Success_Mock(t *testing.T) {
	mockResp := mockResponse{
		statusCode: http.StatusOK,
		body: map[string]interface{}{
			"message": "Subscriber deleted",
		},
	}

	server, client := setupMockServer(mockResp)
	defer server.Close()

	err := client.DeleteSubscriber(context.Background(), "sp_test", 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteSubscriber_NotFound_Mock(t *testing.T) {
	mockResp := mockResponse{
		statusCode: http.StatusNotFound,
		body: map[string]interface{}{
			"error":   "Not found",
			"message": "Subscriber not found",
		},
	}

	server, client := setupMockServer(mockResp)
	defer server.Close()

	err := client.DeleteSubscriber(context.Background(), "sp_test", 999)
	if err == nil {
		t.Error("expected error for non-existent subscriber")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected APIError, got %T", err)
		return
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected status code 404, got %d", apiErr.StatusCode)
	}
}

func TestDeleteSubscriber_InvalidID_Zero_Mock(t *testing.T) {
	// No server needed - client should reject before making request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not make HTTP request with ID 0")
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL), WithMaxRetries(0))

	err := client.DeleteSubscriber(context.Background(), "sp_test", 0)
	if err == nil {
		t.Error("expected error for subscriber ID 0")
	}
}

func TestDeleteSubscriber_InvalidID_Negative_Mock(t *testing.T) {
	// No server needed - client should reject before making request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not make HTTP request with negative ID")
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL), WithMaxRetries(0))

	err := client.DeleteSubscriber(context.Background(), "sp_test", -1)
	if err == nil {
		t.Error("expected error for negative subscriber ID")
	}
}

func TestDeleteSubscriber_ServerError_Mock(t *testing.T) {
	mockResp := mockResponse{
		statusCode: http.StatusInternalServerError,
		body: map[string]interface{}{
			"error": "Internal server error",
		},
	}

	server, client := setupMockServer(mockResp)
	defer server.Close()

	err := client.DeleteSubscriber(context.Background(), "sp_test", 123)
	if err == nil {
		t.Error("expected error for server error")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected APIError, got %T", err)
		return
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status code 500, got %d", apiErr.StatusCode)
	}
}

func TestDeleteSubscriber_Forbidden_Mock(t *testing.T) {
	mockResp := mockResponse{
		statusCode: http.StatusForbidden,
		body: map[string]interface{}{
			"error":   "Forbidden",
			"message": "You don't have permission to delete this subscriber",
		},
	}

	server, client := setupMockServer(mockResp)
	defer server.Close()

	err := client.DeleteSubscriber(context.Background(), "sp_test", 123)
	if err == nil {
		t.Error("expected error for forbidden request")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected APIError, got %T", err)
		return
	}
	if apiErr.StatusCode != http.StatusForbidden {
		t.Errorf("expected status code 403, got %d", apiErr.StatusCode)
	}
}

// =============================================================================
// ListSubscribers Mock Tests
// =============================================================================

func TestListSubscribers_Success_Mock(t *testing.T) {
	mockResp := mockResponse{
		statusCode: http.StatusOK,
		body: map[string]interface{}{
			"subscribers": []map[string]interface{}{
				{
					"id":         1,
					"type":       "email",
					"value":      "user1@example.com",
					"email":      "user1@example.com",
					"created_at": "2026-02-10T10:00:00Z",
				},
				{
					"id":         2,
					"type":       "sms",
					"value":      "+1234567890",
					"phone":      "+1234567890",
					"created_at": "2026-02-10T11:00:00Z",
				},
			},
			"has_next_page":    false,
			"total":            2,
			"page":             0,
			"results_per_page": 20,
		},
	}

	server, client := setupMockServer(mockResp)
	defer server.Close()

	result, err := client.ListSubscribers(context.Background(), "sp_test", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Subscribers) != 2 {
		t.Errorf("expected 2 subscribers, got %d", len(result.Subscribers))
	}
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
	if result.HasNextPage {
		t.Error("expected has_next_page to be false")
	}
}

func TestListSubscribers_Empty_Mock(t *testing.T) {
	mockResp := mockResponse{
		statusCode: http.StatusOK,
		body: map[string]interface{}{
			"subscribers":      []interface{}{},
			"has_next_page":    false,
			"total":            0,
			"page":             0,
			"results_per_page": 20,
		},
	}

	server, client := setupMockServer(mockResp)
	defer server.Close()

	result, err := client.ListSubscribers(context.Background(), "sp_test", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Subscribers) != 0 {
		t.Errorf("expected 0 subscribers, got %d", len(result.Subscribers))
	}
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
}

func TestListSubscribers_Pagination_Mock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageParam := r.URL.Query().Get("page")
		if pageParam != "1" {
			t.Errorf("expected page=1, got %q", pageParam)
		}

		response := map[string]interface{}{
			"subscribers": []map[string]interface{}{
				{
					"id":         21,
					"type":       "email",
					"value":      "user21@example.com",
					"created_at": "2026-02-10T10:00:00Z",
				},
			},
			"hasNextPage":    true,
			"total":          50,
			"page":           1,
			"resultsPerPage": 20,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL), WithMaxRetries(0))

	page := 1
	result, err := client.ListSubscribers(context.Background(), "sp_test", &page, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Page)
	}
	if !result.HasNextPage {
		t.Error("expected has_next_page to be true")
	}
	if result.Total != 50 {
		t.Errorf("expected total 50, got %d", result.Total)
	}
}

func TestListSubscribers_FilterByType_Mock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		typeParam := r.URL.Query().Get("type")
		if typeParam != "email" {
			t.Errorf("expected type=email, got %q", typeParam)
		}

		response := map[string]interface{}{
			"subscribers": []map[string]interface{}{
				{
					"id":         1,
					"type":       "email",
					"value":      "user@example.com",
					"created_at": "2026-02-10T10:00:00Z",
				},
			},
			"has_next_page":    false,
			"total":            1,
			"page":             0,
			"results_per_page": 20,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL), WithMaxRetries(0))

	subscriberType := "email"
	result, err := client.ListSubscribers(context.Background(), "sp_test", nil, &subscriberType)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Subscribers) != 1 {
		t.Errorf("expected 1 subscriber, got %d", len(result.Subscribers))
	}
	if result.Subscribers[0].Type != "email" {
		t.Errorf("expected type 'email', got %q", result.Subscribers[0].Type)
	}
}

func TestListSubscribers_LargeResponse_Mock(t *testing.T) {
	// Simulate large response with many subscribers
	subscribers := make([]map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		subscribers[i] = map[string]interface{}{
			"id":         i + 1,
			"type":       "email",
			"value":      "user" + string(rune('0'+i%10)) + "@example.com",
			"created_at": "2026-02-10T10:00:00Z",
		}
	}

	mockResp := mockResponse{
		statusCode: http.StatusOK,
		body: map[string]interface{}{
			"subscribers":      subscribers,
			"has_next_page":    true,
			"total":            1000,
			"page":             0,
			"results_per_page": 100,
		},
	}

	server, client := setupMockServer(mockResp)
	defer server.Close()

	result, err := client.ListSubscribers(context.Background(), "sp_test", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Subscribers) != 100 {
		t.Errorf("expected 100 subscribers, got %d", len(result.Subscribers))
	}
}

func TestListSubscribers_InvalidJSON_Mock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL), WithMaxRetries(0))

	_, err := client.ListSubscribers(context.Background(), "sp_test", nil, nil)
	if err == nil {
		t.Error("expected error for invalid JSON response")
	}
}

func TestListSubscribers_ServerError_Mock(t *testing.T) {
	mockResp := mockResponse{
		statusCode: http.StatusInternalServerError,
		body: map[string]interface{}{
			"error": "Internal server error",
		},
	}

	server, client := setupMockServer(mockResp)
	defer server.Close()

	_, err := client.ListSubscribers(context.Background(), "sp_test", nil, nil)
	if err == nil {
		t.Error("expected error for server error")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected APIError, got %T", err)
		return
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status code 500, got %d", apiErr.StatusCode)
	}
}

// =============================================================================
// Edge Cases and Error Handling
// =============================================================================

func TestStatusPages_MalformedResponseBody_Mock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Missing required fields
		w.Write([]byte(`{"statuspage": {}}`))
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL), WithMaxRetries(0))

	result, err := client.GetStatusPage(context.Background(), "sp_test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should handle missing fields gracefully
	if result.UUID != "" {
		t.Error("expected empty UUID for malformed response")
	}
}

func TestStatusPages_NullFields_Mock(t *testing.T) {
	mockResp := mockResponse{
		statusCode: http.StatusOK,
		body: map[string]interface{}{
			"statuspage": map[string]interface{}{
				"uuid":             "sp_test",
				"name":             nil,
				"hosted_subdomain": nil,
				"url":              nil,
			},
		},
	}

	server, client := setupMockServer(mockResp)
	defer server.Close()

	result, err := client.GetStatusPage(context.Background(), "sp_test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.UUID != "sp_test" {
		t.Errorf("expected UUID 'sp_test', got %q", result.UUID)
	}
}

func TestStatusPages_RateLimitWithRetryAfter_Mock(t *testing.T) {
	mockResp := mockResponse{
		statusCode: http.StatusTooManyRequests,
		body: map[string]interface{}{
			"error":   "Rate limit exceeded",
			"message": "Too many requests",
		},
		headers: map[string]string{
			"Retry-After": "60",
		},
	}

	server, client := setupMockServer(mockResp)
	defer server.Close()

	_, err := client.GetStatusPage(context.Background(), "sp_test")
	if err == nil {
		t.Error("expected error for rate limit")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected APIError, got %T", err)
		return
	}
	if apiErr.StatusCode != http.StatusTooManyRequests {
		t.Errorf("expected status code 429, got %d", apiErr.StatusCode)
	}
	if apiErr.RetryAfter != 60 {
		t.Errorf("expected RetryAfter 60, got %d", apiErr.RetryAfter)
	}
}

func TestStatusPages_CancelledContext_Mock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Hang for a while
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL), WithMaxRetries(0))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.GetStatusPage(ctx, "sp_test")
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}
