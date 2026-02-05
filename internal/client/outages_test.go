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

func TestClient_GetOutage(t *testing.T) {
	tests := []struct {
		name           string
		uuid           string
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		wantOutage     *Outage
	}{
		{
			name:           "success - active outage",
			uuid:           "out_ABC123",
			responseStatus: http.StatusOK,
			responseBody: Outage{
				UUID:               "out_ABC123",
				StartDate:          "2026-01-27T10:00:00Z",
				DurationMs:         0,
				StatusCode:         503,
				Description:        "Service Unavailable",
				IsResolved:         false,
				DetectedLocation:   "london",
				ConfirmedLocations: "london,paris",
				Monitor: MonitorReference{
					UUID:     "mon_test123",
					Name:     "Test Monitor",
					URL:      "https://example.com",
					Protocol: "http",
				},
			},
			wantErr: false,
			wantOutage: &Outage{
				UUID:               "out_ABC123",
				StartDate:          "2026-01-27T10:00:00Z",
				DurationMs:         0,
				StatusCode:         503,
				Description:        "Service Unavailable",
				IsResolved:         false,
				DetectedLocation:   "london",
				ConfirmedLocations: "london,paris",
				Monitor: MonitorReference{
					UUID:     "mon_test123",
					Name:     "Test Monitor",
					URL:      "https://example.com",
					Protocol: "http",
				},
			},
		},
		{
			name:           "success - resolved outage",
			uuid:           "out_DEF456",
			responseStatus: http.StatusOK,
			responseBody: func() Outage {
				endDate := "2026-01-27T10:05:00Z"
				ackAt := "2026-01-27T10:01:00Z"
				return Outage{
					UUID:               "out_DEF456",
					StartDate:          "2026-01-27T10:00:00Z",
					EndDate:            &endDate,
					DurationMs:         300000, // 5 minutes in milliseconds
					StatusCode:         503,
					Description:        "Service recovered",
					IsResolved:         true,
					DetectedLocation:   "frankfurt",
					ConfirmedLocations: "frankfurt,amsterdam",
					AcknowledgedAt:     &ackAt,
					AcknowledgedBy: &AcknowledgedByUser{
						UUID:  "user_123",
						Email: "engineer@example.com",
						Name:  "Test Engineer",
					},
					Monitor: MonitorReference{
						UUID:     "mon_test456",
						Name:     "API Monitor",
						URL:      "https://api.example.com",
						Protocol: "http",
					},
				}
			}(),
			wantErr: false,
			wantOutage: func() *Outage {
				endDate := "2026-01-27T10:05:00Z"
				ackAt := "2026-01-27T10:01:00Z"
				return &Outage{
					UUID:               "out_DEF456",
					StartDate:          "2026-01-27T10:00:00Z",
					EndDate:            &endDate,
					DurationMs:         300000,
					StatusCode:         503,
					Description:        "Service recovered",
					IsResolved:         true,
					DetectedLocation:   "frankfurt",
					ConfirmedLocations: "frankfurt,amsterdam",
					AcknowledgedAt:     &ackAt,
					AcknowledgedBy: &AcknowledgedByUser{
						UUID:  "user_123",
						Email: "engineer@example.com",
						Name:  "Test Engineer",
					},
					Monitor: MonitorReference{
						UUID:     "mon_test456",
						Name:     "API Monitor",
						URL:      "https://api.example.com",
						Protocol: "http",
					},
				}
			}(),
		},
		{
			name:           "error - not found",
			uuid:           "out_notfound",
			responseStatus: http.StatusNotFound,
			responseBody:   map[string]string{"error": "Outage not found"},
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				if r.URL.Path != outagesBasePath+"/"+tt.uuid {
					t.Errorf("Expected path %s/%s, got %s", outagesBasePath, tt.uuid, r.URL.Path)
				}

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client := NewClient("test_key", WithBaseURL(server.URL))
			outage, err := client.GetOutage(context.Background(), tt.uuid)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetOutage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if outage.UUID != tt.wantOutage.UUID {
					t.Errorf("UUID = %v, want %v", outage.UUID, tt.wantOutage.UUID)
				}
				if outage.Monitor.UUID != tt.wantOutage.Monitor.UUID {
					t.Errorf("Monitor.UUID = %v, want %v", outage.Monitor.UUID, tt.wantOutage.Monitor.UUID)
				}
				if outage.IsResolved != tt.wantOutage.IsResolved {
					t.Errorf("IsResolved = %v, want %v", outage.IsResolved, tt.wantOutage.IsResolved)
				}
				if outage.StatusCode != tt.wantOutage.StatusCode {
					t.Errorf("StatusCode = %v, want %v", outage.StatusCode, tt.wantOutage.StatusCode)
				}
			}
		})
	}
}

func TestClient_ListOutages(t *testing.T) {
	tests := []struct {
		name           string
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		wantCount      int
	}{
		{
			name:           "success - direct array",
			responseStatus: http.StatusOK,
			responseBody: []Outage{
				{
					UUID:             "out_1",
					IsResolved:       false,
					StatusCode:       503,
					Description:      "Service down",
					DetectedLocation: "london",
					Monitor: MonitorReference{
						UUID: "mon_1",
						Name: "Monitor 1",
					},
				},
				{
					UUID:             "out_2",
					IsResolved:       true,
					StatusCode:       200,
					Description:      "Recovered",
					DetectedLocation: "paris",
					Monitor: MonitorReference{
						UUID: "mon_2",
						Name: "Monitor 2",
					},
				},
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:           "success - wrapped in outages key",
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"outages": []Outage{
					{
						UUID:             "out_1",
						IsResolved:       false,
						StatusCode:       503,
						DetectedLocation: "london",
						Monitor: MonitorReference{
							UUID: "mon_1",
							Name: "Monitor 1",
						},
					},
					{
						UUID:             "out_2",
						IsResolved:       true,
						StatusCode:       200,
						DetectedLocation: "paris",
						Monitor: MonitorReference{
							UUID: "mon_2",
							Name: "Monitor 2",
						},
					},
				},
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:           "success - empty list",
			responseStatus: http.StatusOK,
			responseBody:   []Outage{},
			wantErr:        false,
			wantCount:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				if r.URL.Path != outagesBasePath {
					t.Errorf("Expected path %s, got %s", outagesBasePath, r.URL.Path)
				}

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client := NewClient("test_key", WithBaseURL(server.URL))
			outages, err := client.ListOutages(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListOutages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(outages) != tt.wantCount {
				t.Errorf("ListOutages() count = %d, want %d", len(outages), tt.wantCount)
			}
		})
	}
}

func TestClient_CreateOutage(t *testing.T) {
	tests := []struct {
		name           string
		request        CreateOutageRequest
		responseStatus int
		responseBody   interface{}
		wantErr        bool
	}{
		{
			name: "success - manual outage",
			request: CreateOutageRequest{
				MonitorUUID: "mon_abc123",
				StartDate:   "2026-01-27T10:00:00Z",
				StatusCode:  503,
				Description: "Planned maintenance outage",
				OutageType:  "manual",
			},
			responseStatus: http.StatusCreated,
			responseBody: Outage{
				UUID:               "out_NEW001",
				StartDate:          "2026-01-27T10:00:00Z",
				DurationMs:         0,
				StatusCode:         503,
				Description:        "Planned maintenance outage",
				IsResolved:         false,
				DetectedLocation:   "manual",
				ConfirmedLocations: "",
				Monitor: MonitorReference{
					UUID:     "mon_abc123",
					Name:     "Test Monitor",
					URL:      "https://example.com",
					Protocol: "http",
				},
			},
			wantErr: false,
		},
		{
			name: "validation error",
			request: CreateOutageRequest{
				MonitorUUID: "",
				StartDate:   "2026-01-27T10:00:00Z",
				StatusCode:  503,
				Description: "Invalid",
				OutageType:  "manual",
			},
			responseStatus: http.StatusBadRequest,
			responseBody: map[string]string{
				"error": "Monitor UUID is required",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != outagesBasePath {
					t.Errorf("unexpected path: got %v", r.URL.Path)
				}
				if r.Method != http.MethodPost {
					t.Errorf("unexpected method: got %v", r.Method)
				}

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client := NewClient("test-key",
				WithHTTPClient(server.Client()),
				WithBaseURL(server.URL),
				WithMaxRetries(0),
			)

			got, err := client.CreateOutage(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateOutage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got.UUID == "" {
				t.Error("CreateOutage() returned empty UUID")
			}
		})
	}
}

func TestClient_AcknowledgeOutage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != outagesBasePath+"/out_test123/acknowledge" {
			t.Errorf("Expected path %s/out_test123/acknowledge, got %s", outagesBasePath, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(OutageAction{
			Message: "Outage acknowledged successfully",
			UUID:    "out_test123",
		})
	}))
	defer server.Close()

	client := NewClient("test_key", WithBaseURL(server.URL))
	response, err := client.AcknowledgeOutage(context.Background(), "out_test123")

	if err != nil {
		t.Fatalf("AcknowledgeOutage() error = %v", err)
	}

	if response.UUID != "out_test123" {
		t.Errorf("UUID = %v, want out_test123", response.UUID)
	}
	if response.Message != "Outage acknowledged successfully" {
		t.Errorf("Message = %v, want 'Outage acknowledged successfully'", response.Message)
	}
}

func TestClient_UnacknowledgeOutage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != outagesBasePath+"/out_test123/unacknowledge" {
			t.Errorf("Expected path %s/out_test123/unacknowledge, got %s", outagesBasePath, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(OutageAction{
			Message: "Outage unacknowledged",
			UUID:    "out_test123",
		})
	}))
	defer server.Close()

	client := NewClient("test_key", WithBaseURL(server.URL))
	response, err := client.UnacknowledgeOutage(context.Background(), "out_test123")

	if err != nil {
		t.Fatalf("UnacknowledgeOutage() error = %v", err)
	}

	if response.UUID != "out_test123" {
		t.Errorf("UUID = %v, want out_test123", response.UUID)
	}
}

func TestClient_ResolveOutage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != outagesBasePath+"/out_test123/resolve" {
			t.Errorf("Expected path %s/out_test123/resolve, got %s", outagesBasePath, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(OutageAction{
			Message: "Outage resolved",
			UUID:    "out_test123",
		})
	}))
	defer server.Close()

	client := NewClient("test_key", WithBaseURL(server.URL))
	response, err := client.ResolveOutage(context.Background(), "out_test123")

	if err != nil {
		t.Fatalf("ResolveOutage() error = %v", err)
	}

	if response.UUID != "out_test123" {
		t.Errorf("UUID = %v, want out_test123", response.UUID)
	}
	if response.Message != "Outage resolved" {
		t.Errorf("Message = %v, want 'Outage resolved'", response.Message)
	}
}

func TestClient_EscalateOutage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != outagesBasePath+"/out_test123/escalate" {
			t.Errorf("Expected path %s/out_test123/escalate, got %s", outagesBasePath, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(OutageAction{
			Message: "Outage escalated",
			UUID:    "out_test123",
		})
	}))
	defer server.Close()

	client := NewClient("test_key", WithBaseURL(server.URL))
	response, err := client.EscalateOutage(context.Background(), "out_test123")

	if err != nil {
		t.Fatalf("EscalateOutage() error = %v", err)
	}

	if response.UUID != "out_test123" {
		t.Errorf("UUID = %v, want out_test123", response.UUID)
	}
	if response.Message != "Outage escalated" {
		t.Errorf("Message = %v, want 'Outage escalated'", response.Message)
	}
}

func TestClient_DeleteOutage(t *testing.T) {
	tests := []struct {
		name           string
		uuid           string
		responseStatus int
		responseBody   interface{}
		wantErr        bool
	}{
		{
			name:           "success",
			uuid:           "out_test123",
			responseStatus: http.StatusNoContent,
			responseBody:   nil,
			wantErr:        false,
		},
		{
			name:           "error - not found",
			uuid:           "out_notfound",
			responseStatus: http.StatusNotFound,
			responseBody:   map[string]string{"error": "Outage not found"},
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("Expected DELETE request, got %s", r.Method)
				}
				if r.URL.Path != outagesBasePath+"/"+tt.uuid {
					t.Errorf("Expected path %s/%s, got %s", outagesBasePath, tt.uuid, r.URL.Path)
				}

				w.WriteHeader(tt.responseStatus)
				if tt.responseBody != nil {
					json.NewEncoder(w).Encode(tt.responseBody)
				}
			}))
			defer server.Close()

			client := NewClient("test_key", WithBaseURL(server.URL))
			err := client.DeleteOutage(context.Background(), tt.uuid)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteOutage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestClient_OutageActions_ErrorHandling tests error handling for all action methods
func TestClient_OutageActions_ErrorHandling(t *testing.T) {
	tests := []struct {
		name   string
		action func(ctx context.Context, client *Client, uuid string) error
	}{
		{
			name: "acknowledge",
			action: func(ctx context.Context, client *Client, uuid string) error {
				_, err := client.AcknowledgeOutage(ctx, uuid)
				return err
			},
		},
		{
			name: "unacknowledge",
			action: func(ctx context.Context, client *Client, uuid string) error {
				_, err := client.UnacknowledgeOutage(ctx, uuid)
				return err
			},
		},
		{
			name: "resolve",
			action: func(ctx context.Context, client *Client, uuid string) error {
				_, err := client.ResolveOutage(ctx, uuid)
				return err
			},
		},
		{
			name: "escalate",
			action: func(ctx context.Context, client *Client, uuid string) error {
				_, err := client.EscalateOutage(ctx, uuid)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+" - server error", func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Internal server error",
				})
			}))
			defer server.Close()

			client := NewClient("test_key", WithBaseURL(server.URL))
			err := tt.action(context.Background(), client, "out_test")

			if err == nil {
				t.Error("Expected error for server error, got nil")
			}
		})
	}
}

func TestClient_ListOutages_ErrorPath(t *testing.T) {
	t.Run("API request error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
		}))
		defer server.Close()

		client := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))
		_, err := client.ListOutages(context.Background())

		if err == nil {
			t.Error("expected error from API request failure, got nil")
		}
	})

	t.Run("parse response error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			// Return completely invalid JSON that will fail parsing in wrapped format
			w.Write([]byte(`{"outages": "not an array"}`))
		}))
		defer server.Close()

		client := NewClient("test_key", WithBaseURL(server.URL))
		_, err := client.ListOutages(context.Background())

		if err == nil {
			t.Error("expected error from parse failure, got nil")
		}
	})
}

func TestParseOutageListResponse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
		wantErr   bool
	}{
		{
			name: "direct array",
			input: `[
				{"uuid":"out_1","isResolved":false,"statusCode":503,"monitor":{"uuid":"mon_1"}},
				{"uuid":"out_2","isResolved":true,"statusCode":200,"monitor":{"uuid":"mon_2"}}
			]`,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "wrapped in outages key",
			input:     `{"outages":[{"uuid":"out_1","isResolved":false,"statusCode":503,"monitor":{"uuid":"mon_1"}}]}`,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "wrapped in data key",
			input:     `{"data":[{"uuid":"out_1","isResolved":false,"statusCode":503,"monitor":{"uuid":"mon_1"}}]}`,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "empty array",
			input:     `[]`,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "empty object",
			input:     `{}`,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "empty outages array",
			input:     `{"outages":[]}`,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "empty data array",
			input:     `{"data":[]}`,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:    "invalid JSON",
			input:   `{invalid json syntax`,
			wantErr: true,
		},
		{
			name:      "both outages and data (outages takes priority)",
			input:     `{"outages":[{"uuid":"out_1","monitor":{"uuid":"mon_1"}}],"data":[{"uuid":"out_2","monitor":{"uuid":"mon_2"}},{"uuid":"out_3","monitor":{"uuid":"mon_3"}}]}`,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "only data when outages is empty",
			input:     `{"outages":[],"data":[{"uuid":"out_1","monitor":{"uuid":"mon_1"}}]}`,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "single outage in array",
			input:     `[{"uuid":"out_single","isResolved":false,"statusCode":500,"monitor":{"uuid":"mon_test"}}]`,
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseOutageListResponse(json.RawMessage(tt.input))

			if (err != nil) != tt.wantErr {
				t.Errorf("parseOutageListResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(result) != tt.wantCount {
				t.Errorf("parseOutageListResponse() count = %d, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestClient_ListOutages_DataWrapper(t *testing.T) {
	t.Run("success - wrapped in data key", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET request, got %s", r.Method)
			}
			if r.URL.Path != outagesBasePath {
				t.Errorf("Expected path %s, got %s", outagesBasePath, r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []Outage{
					{
						UUID:             "out_1",
						IsResolved:       false,
						StatusCode:       503,
						DetectedLocation: "london",
						Monitor: MonitorReference{
							UUID: "mon_1",
							Name: "Monitor 1",
						},
					},
					{
						UUID:             "out_2",
						IsResolved:       true,
						StatusCode:       200,
						DetectedLocation: "paris",
						Monitor: MonitorReference{
							UUID: "mon_2",
							Name: "Monitor 2",
						},
					},
				},
			})
		}))
		defer server.Close()

		client := NewClient("test_key", WithBaseURL(server.URL))
		outages, err := client.ListOutages(context.Background())

		if err != nil {
			t.Errorf("ListOutages() error = %v", err)
			return
		}

		if len(outages) != 2 {
			t.Errorf("ListOutages() count = %d, want 2", len(outages))
		}
	})
}
