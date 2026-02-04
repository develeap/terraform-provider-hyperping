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

func TestClient_GetHealthcheck(t *testing.T) {
	tests := []struct {
		name            string
		uuid            string
		responseStatus  int
		responseBody    interface{}
		wantErr         bool
		wantHealthcheck *Healthcheck
	}{
		{
			name:           "success - cron-based healthcheck",
			uuid:           "tok_ABC123",
			responseStatus: http.StatusOK,
			responseBody: Healthcheck{
				UUID:             "tok_ABC123",
				Name:             "Daily Backup Job",
				PingURL:          "https://ping.hyperping.io/tok_ABC123",
				Cron:             "0 2 * * *",
				Timezone:               "America/New_York",
				Period:           86400,
				GracePeriod:      3600,
				GracePeriodValue: 1,
				GracePeriodType:  "hours",
				IsDown:           false,
				IsPaused:         false,
				CreatedAt:        "2026-01-27T10:00:00Z",
			},
			wantErr: false,
			wantHealthcheck: &Healthcheck{
				UUID:             "tok_ABC123",
				Name:             "Daily Backup Job",
				PingURL:          "https://ping.hyperping.io/tok_ABC123",
				Cron:             "0 2 * * *",
				Timezone:               "America/New_York",
				Period:           86400,
				GracePeriod:      3600,
				GracePeriodValue: 1,
				GracePeriodType:  "hours",
				IsDown:           false,
				IsPaused:         false,
				CreatedAt:        "2026-01-27T10:00:00Z",
			},
		},
		{
			name:           "success - period-based healthcheck",
			uuid:           "tok_DEF456",
			responseStatus: http.StatusOK,
			responseBody: Healthcheck{
				UUID:             "tok_DEF456",
				Name:             "Hourly Task",
				PingURL:          "https://ping.hyperping.io/tok_DEF456",
				PeriodValue:      intPtr(1),
				PeriodType:       "hours",
				Period:           3600,
				GracePeriod:      300,
				GracePeriodValue: 5,
				GracePeriodType:  "minutes",
				IsDown:           false,
				IsPaused:         false,
				CreatedAt:        "2026-01-27T11:00:00Z",
			},
			wantErr: false,
			wantHealthcheck: &Healthcheck{
				UUID:             "tok_DEF456",
				Name:             "Hourly Task",
				PingURL:          "https://ping.hyperping.io/tok_DEF456",
				PeriodValue:      intPtr(1),
				PeriodType:       "hours",
				Period:           3600,
				GracePeriod:      300,
				GracePeriodValue: 5,
				GracePeriodType:  "minutes",
				IsDown:           false,
				IsPaused:         false,
				CreatedAt:        "2026-01-27T11:00:00Z",
			},
		},
		{
			name:           "not found",
			uuid:           "tok_NOTFOUND",
			responseStatus: http.StatusNotFound,
			responseBody: map[string]string{
				"error": "Healthcheck not found",
			},
			wantErr: true,
		},
		{
			name:           "server error",
			uuid:           "tok_ERROR",
			responseStatus: http.StatusInternalServerError,
			responseBody: map[string]string{
				"error": "Internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/healthchecks/"+tt.uuid {
					t.Errorf("unexpected path: got %v", r.URL.Path)
				}
				if r.Method != http.MethodGet {
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

			got, err := client.GetHealthcheck(context.Background(), tt.uuid)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetHealthcheck() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.UUID != tt.wantHealthcheck.UUID {
					t.Errorf("UUID = %v, want %v", got.UUID, tt.wantHealthcheck.UUID)
				}
				if got.Name != tt.wantHealthcheck.Name {
					t.Errorf("Name = %v, want %v", got.Name, tt.wantHealthcheck.Name)
				}
				if got.IsDown != tt.wantHealthcheck.IsDown {
					t.Errorf("IsDown = %v, want %v", got.IsDown, tt.wantHealthcheck.IsDown)
				}
			}
		})
	}
}

func TestClient_ListHealthchecks(t *testing.T) {
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
			responseBody: []Healthcheck{
				{
					UUID:    "tok_001",
					Name:    "Backup Job",
					PingURL: "https://ping.hyperping.io/tok_001",
					IsDown:  false,
				},
				{
					UUID:    "tok_002",
					Name:    "Sync Task",
					PingURL: "https://ping.hyperping.io/tok_002",
					IsDown:  true,
				},
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:           "success - wrapped in healthchecks key",
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"healthchecks": []Healthcheck{
					{
						UUID:    "tok_003",
						Name:    "Report Job",
						PingURL: "https://ping.hyperping.io/tok_003",
						IsDown:  false,
					},
				},
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:           "success - wrapped in data key",
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"data": []Healthcheck{
					{
						UUID:    "tok_004",
						Name:    "Cleanup Job",
						PingURL: "https://ping.hyperping.io/tok_004",
						IsDown:  false,
					},
				},
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:           "success - empty list",
			responseStatus: http.StatusOK,
			responseBody:   []Healthcheck{},
			wantErr:        false,
			wantCount:      0,
		},
		{
			name:           "server error",
			responseStatus: http.StatusInternalServerError,
			responseBody: map[string]string{
				"error": "Internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/healthchecks" {
					t.Errorf("unexpected path: got %v", r.URL.Path)
				}
				if r.Method != http.MethodGet {
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

			got, err := client.ListHealthchecks(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListHealthchecks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(got) != tt.wantCount {
				t.Errorf("ListHealthchecks() count = %v, want %v", len(got), tt.wantCount)
			}
		})
	}
}

func TestClient_CreateHealthcheck(t *testing.T) {
	cronExpr := "0 2 * * *"
	tz := "America/New_York"
	escalationPolicy := "ep_test123"

	tests := []struct {
		name           string
		request        CreateHealthcheckRequest
		responseStatus int
		responseBody   interface{}
		wantErr        bool
	}{
		{
			name: "success - cron-based",
			request: CreateHealthcheckRequest{
				Name:             "Daily Backup",
				Cron:             &cronExpr,
				Timezone:               &tz,
				GracePeriodValue: 1,
				GracePeriodType:  "hours",
				EscalationPolicy: &escalationPolicy,
			},
			responseStatus: http.StatusCreated,
			responseBody: Healthcheck{
				UUID:             "tok_NEW001",
				Name:             "Daily Backup",
				PingURL:          "https://ping.hyperping.io/tok_NEW001",
				Cron:             "0 2 * * *",
				Timezone:               "America/New_York",
				Period:           86400,
				GracePeriod:      3600,
				GracePeriodValue: 1,
				GracePeriodType:  "hours",
				IsDown:           false,
				CreatedAt:        "2026-01-27T12:00:00Z",
			},
			wantErr: false,
		},
		{
			name: "success - period-based",
			request: CreateHealthcheckRequest{
				Name:             "Hourly Task",
				PeriodValue:      intPtr(1),
				PeriodType:       strPtr("hours"),
				GracePeriodValue: 5,
				GracePeriodType:  "minutes",
			},
			responseStatus: http.StatusOK,
			responseBody: Healthcheck{
				UUID:             "tok_NEW002",
				Name:             "Hourly Task",
				PingURL:          "https://ping.hyperping.io/tok_NEW002",
				PeriodValue:      intPtr(1),
				PeriodType:       "hours",
				Period:           3600,
				GracePeriod:      300,
				GracePeriodValue: 5,
				GracePeriodType:  "minutes",
				IsDown:           false,
				CreatedAt:        "2026-01-27T12:00:00Z",
			},
			wantErr: false,
		},
		{
			name: "validation error",
			request: CreateHealthcheckRequest{
				Name:             "Invalid",
				GracePeriodValue: 1,
				GracePeriodType:  "hours",
			},
			responseStatus: http.StatusBadRequest,
			responseBody: map[string]interface{}{
				"error":   "Validation error",
				"message": "Either cron or period must be specified",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/healthchecks" {
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

			got, err := client.CreateHealthcheck(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateHealthcheck() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got.UUID == "" {
				t.Error("CreateHealthcheck() returned empty UUID")
			}
		})
	}
}

func TestClient_UpdateHealthcheck(t *testing.T) {
	newName := "Updated Name"
	newGracePeriod := 10

	tests := []struct {
		name           string
		uuid           string
		request        UpdateHealthcheckRequest
		responseStatus int
		responseBody   interface{}
		wantErr        bool
	}{
		{
			name: "success",
			uuid: "tok_UPDATE001",
			request: UpdateHealthcheckRequest{
				Name:             &newName,
				GracePeriodValue: &newGracePeriod,
			},
			responseStatus: http.StatusOK,
			responseBody: Healthcheck{
				UUID:             "tok_UPDATE001",
				Name:             "Updated Name",
				PingURL:          "https://ping.hyperping.io/tok_UPDATE001",
				Period:           3600,
				GracePeriod:      600,
				GracePeriodValue: 10,
				GracePeriodType:  "minutes",
				IsDown:           false,
			},
			wantErr: false,
		},
		{
			name: "not found",
			uuid: "tok_NOTFOUND",
			request: UpdateHealthcheckRequest{
				Name: &newName,
			},
			responseStatus: http.StatusNotFound,
			responseBody: map[string]string{
				"error": "Healthcheck not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/healthchecks/"+tt.uuid {
					t.Errorf("unexpected path: got %v", r.URL.Path)
				}
				if r.Method != http.MethodPut {
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

			got, err := client.UpdateHealthcheck(context.Background(), tt.uuid, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateHealthcheck() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got.UUID != tt.uuid {
				t.Errorf("UpdateHealthcheck() UUID = %v, want %v", got.UUID, tt.uuid)
			}
		})
	}
}

func TestClient_DeleteHealthcheck(t *testing.T) {
	tests := []struct {
		name           string
		uuid           string
		responseStatus int
		responseBody   interface{}
		wantErr        bool
	}{
		{
			name:           "success - 200 OK",
			uuid:           "tok_DELETE001",
			responseStatus: http.StatusOK,
			responseBody: map[string]string{
				"message": "Healthcheck deleted successfully",
			},
			wantErr: false,
		},
		{
			name:           "success - 204 No Content",
			uuid:           "tok_DELETE002",
			responseStatus: http.StatusNoContent,
			responseBody:   nil,
			wantErr:        false,
		},
		{
			name:           "not found",
			uuid:           "tok_NOTFOUND",
			responseStatus: http.StatusNotFound,
			responseBody: map[string]string{
				"error": "Healthcheck not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/healthchecks/"+tt.uuid {
					t.Errorf("unexpected path: got %v", r.URL.Path)
				}
				if r.Method != http.MethodDelete {
					t.Errorf("unexpected method: got %v", r.Method)
				}

				w.WriteHeader(tt.responseStatus)
				if tt.responseBody != nil {
					json.NewEncoder(w).Encode(tt.responseBody)
				}
			}))
			defer server.Close()

			client := NewClient("test-key",
				WithHTTPClient(server.Client()),
				WithBaseURL(server.URL),
				WithMaxRetries(0),
			)

			err := client.DeleteHealthcheck(context.Background(), tt.uuid)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteHealthcheck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_PauseHealthcheck(t *testing.T) {
	tests := []struct {
		name           string
		uuid           string
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		wantAction     *HealthcheckAction
	}{
		{
			name:           "success",
			uuid:           "tok_PAUSE001",
			responseStatus: http.StatusOK,
			responseBody: HealthcheckAction{
				Message: "Healthcheck paused successfully",
				UUID:    "tok_PAUSE001",
			},
			wantErr: false,
			wantAction: &HealthcheckAction{
				Message: "Healthcheck paused successfully",
				UUID:    "tok_PAUSE001",
			},
		},
		{
			name:           "not found",
			uuid:           "tok_NOTFOUND",
			responseStatus: http.StatusNotFound,
			responseBody: map[string]string{
				"error": "Healthcheck not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/healthchecks/"+tt.uuid+"/pause" {
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

			got, err := client.PauseHealthcheck(context.Background(), tt.uuid)

			if (err != nil) != tt.wantErr {
				t.Errorf("PauseHealthcheck() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.UUID != tt.wantAction.UUID {
					t.Errorf("PauseHealthcheck() UUID = %v, want %v", got.UUID, tt.wantAction.UUID)
				}
			}
		})
	}
}

func TestClient_ResumeHealthcheck(t *testing.T) {
	tests := []struct {
		name           string
		uuid           string
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		wantAction     *HealthcheckAction
	}{
		{
			name:           "success",
			uuid:           "tok_RESUME001",
			responseStatus: http.StatusOK,
			responseBody: HealthcheckAction{
				Message: "Healthcheck resumed successfully",
				UUID:    "tok_RESUME001",
			},
			wantErr: false,
			wantAction: &HealthcheckAction{
				Message: "Healthcheck resumed successfully",
				UUID:    "tok_RESUME001",
			},
		},
		{
			name:           "not found",
			uuid:           "tok_NOTFOUND",
			responseStatus: http.StatusNotFound,
			responseBody: map[string]string{
				"error": "Healthcheck not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/healthchecks/"+tt.uuid+"/resume" {
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

			got, err := client.ResumeHealthcheck(context.Background(), tt.uuid)

			if (err != nil) != tt.wantErr {
				t.Errorf("ResumeHealthcheck() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.UUID != tt.wantAction.UUID {
					t.Errorf("ResumeHealthcheck() UUID = %v, want %v", got.UUID, tt.wantAction.UUID)
				}
			}
		})
	}
}

func TestClient_HealthcheckActions_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		actionFunc     func(*Client, context.Context, string) error
		responseStatus int
		wantErr        bool
	}{
		{
			name: "pause - unauthorized",
			actionFunc: func(c *Client, ctx context.Context, uuid string) error {
				_, err := c.PauseHealthcheck(ctx, uuid)
				return err
			},
			responseStatus: http.StatusUnauthorized,
			wantErr:        true,
		},
		{
			name: "resume - forbidden",
			actionFunc: func(c *Client, ctx context.Context, uuid string) error {
				_, err := c.ResumeHealthcheck(ctx, uuid)
				return err
			},
			responseStatus: http.StatusForbidden,
			wantErr:        true,
		},
		{
			name: "pause - rate limited",
			actionFunc: func(c *Client, ctx context.Context, uuid string) error {
				_, err := c.PauseHealthcheck(ctx, uuid)
				return err
			},
			responseStatus: http.StatusTooManyRequests,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Error occurred",
				})
			}))
			defer server.Close()

			client := NewClient("test-key",
				WithHTTPClient(server.Client()),
				WithBaseURL(server.URL),
				WithMaxRetries(0),
			)

			err := tt.actionFunc(client, context.Background(), "tok_TEST")

			if (err != nil) != tt.wantErr {
				t.Errorf("action error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_ListHealthchecks_ErrorPath(t *testing.T) {
	t.Run("API request error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
		}))
		defer server.Close()

		client := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))
		_, err := client.ListHealthchecks(context.Background())

		if err == nil {
			t.Error("expected error from API request failure, got nil")
		}
	})

	t.Run("parse response error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			// Return completely invalid JSON that will fail parsing in wrapped format
			w.Write([]byte(`{"healthchecks": "not an array"}`))
		}))
		defer server.Close()

		client := NewClient("test_key", WithBaseURL(server.URL))
		_, err := client.ListHealthchecks(context.Background())

		if err == nil {
			t.Error("expected error from parse failure, got nil")
		}
	})
}

func TestParseHealthcheckListResponse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "direct array",
			input:     `[{"uuid":"tok_1","name":"Test 1"},{"uuid":"tok_2","name":"Test 2"}]`,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "wrapped in healthchecks key",
			input:     `{"healthchecks":[{"uuid":"tok_1","name":"Test 1"}]}`,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "wrapped in data key",
			input:     `{"data":[{"uuid":"tok_1","name":"Test 1"}]}`,
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
			name:      "empty healthchecks array",
			input:     `{"healthchecks":[]}`,
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
			input:   `{invalid json`,
			wantErr: true,
		},
		{
			name:      "both healthchecks and data (healthchecks takes priority)",
			input:     `{"healthchecks":[{"uuid":"tok_1"}],"data":[{"uuid":"tok_2"},{"uuid":"tok_3"}]}`,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "only data when healthchecks is empty",
			input:     `{"healthchecks":[],"data":[{"uuid":"tok_1"}]}`,
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseHealthcheckListResponse(json.RawMessage(tt.input))

			if (err != nil) != tt.wantErr {
				t.Errorf("parseHealthcheckListResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(result) != tt.wantCount {
				t.Errorf("parseHealthcheckListResponse() count = %d, want %d", len(result), tt.wantCount)
			}
		})
	}
}
