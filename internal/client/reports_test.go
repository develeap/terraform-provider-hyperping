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

func TestClient_ListMonitorReports_FieldTypes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"period": {"from": "2026-01-01T00:00:00Z", "to": "2026-01-31T23:59:59Z"},
			"monitors": [{
				"uuid": "mon_field001",
				"name": "Field Type Monitor",
				"protocol": "https",
				"sla": 99.95,
				"mttr": 300,
				"mttrFormatted": "5min",
				"outages": {
					"count": 2,
					"totalDowntime": 600,
					"totalDowntimeFormatted": "10min",
					"longestOutage": 300,
					"longestOutageFormatted": "5min",
					"details": []
				}
			}]
		}`))
	}))
	defer server.Close()

	c := NewClient("test-key",
		WithHTTPClient(server.Client()),
		WithBaseURL(server.URL),
		WithMaxRetries(0),
	)

	monitors, err := c.ListMonitorReports(context.Background(), "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(monitors) != 1 {
		t.Fatalf("expected 1 monitor, got %d", len(monitors))
	}

	m := monitors[0]
	if m.UUID != "mon_field001" {
		t.Errorf("expected UUID 'mon_field001', got %q", m.UUID)
	}
	if m.Name != "Field Type Monitor" {
		t.Errorf("expected Name 'Field Type Monitor', got %q", m.Name)
	}
	if m.Protocol != "https" {
		t.Errorf("expected Protocol 'https', got %q", m.Protocol)
	}
	if m.SLA != 99.95 {
		t.Errorf("expected SLA 99.95, got %f", m.SLA)
	}
	if m.MTTR != 300 {
		t.Errorf("expected MTTR 300, got %d", m.MTTR)
	}
	if m.MTTRFormatted != "5min" {
		t.Errorf("expected MTTRFormatted '5min', got %q", m.MTTRFormatted)
	}
	if m.Outages.Count != 2 {
		t.Errorf("expected Outages.Count 2, got %d", m.Outages.Count)
	}
	if m.Outages.TotalDowntime != 600 {
		t.Errorf("expected Outages.TotalDowntime 600, got %d", m.Outages.TotalDowntime)
	}
	if m.Outages.TotalDowntimeFormatted != "10min" {
		t.Errorf("expected Outages.TotalDowntimeFormatted '10min', got %q", m.Outages.TotalDowntimeFormatted)
	}
}

func TestClient_ListMonitorReports_ZeroMTTR(t *testing.T) {
	// When "mttr" is absent from the JSON response (e.g. no outages), the
	// field must decode to zero — verifying the omitempty round-trip contract.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"period": {"from": "", "to": ""},
			"monitors": [{
				"uuid": "mon_nomttr",
				"name": "No MTTR Monitor",
				"protocol": "http",
				"sla": 100.0,
				"outages": {
					"count": 0,
					"totalDowntime": 0,
					"totalDowntimeFormatted": "",
					"longestOutage": 0,
					"longestOutageFormatted": "",
					"details": []
				}
			}]
		}`))
	}))
	defer server.Close()

	c := NewClient("test-key",
		WithHTTPClient(server.Client()),
		WithBaseURL(server.URL),
		WithMaxRetries(0),
	)

	monitors, err := c.ListMonitorReports(context.Background(), "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(monitors) != 1 {
		t.Fatalf("expected 1 monitor, got %d", len(monitors))
	}

	m := monitors[0]
	if m.MTTR != 0 {
		t.Errorf("expected MTTR 0 when absent from JSON, got %d", m.MTTR)
	}
	if m.MTTRFormatted != "" {
		t.Errorf("expected MTTRFormatted '' when absent from JSON, got %q", m.MTTRFormatted)
	}
	if m.Outages.Count != 0 {
		t.Errorf("expected Outages.Count 0, got %d", m.Outages.Count)
	}
}

func TestClient_ListMonitorReports_EmptyMonitors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"period": {"from": "", "to": ""},
			"monitors": []
		}`))
	}))
	defer server.Close()

	c := NewClient("test-key",
		WithHTTPClient(server.Client()),
		WithBaseURL(server.URL),
		WithMaxRetries(0),
	)

	monitors, err := c.ListMonitorReports(context.Background(), "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(monitors) != 0 {
		t.Errorf("expected 0 monitors for empty list response, got %d", len(monitors))
	}
}

func TestClient_GetMonitorReport(t *testing.T) {
	tests := []struct {
		name           string
		uuid           string
		from           string
		to             string
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		wantReport     *MonitorReport
	}{
		{
			name:           "success - without date range",
			uuid:           "mon_abc123",
			from:           "",
			to:             "",
			responseStatus: http.StatusOK,
			responseBody: MonitorReport{
				UUID:          "mon_abc123",
				Name:          "Production API",
				Protocol:      "http",
				Period:        ReportPeriod{From: "2026-01-01T00:00:00Z", To: "2026-01-31T23:59:59Z"},
				SLA:           99.184,
				MTTR:          4283,
				MTTRFormatted: "1hr 11min 23s",
				Outages: OutageStats{
					Count:                  3,
					TotalDowntime:          12850,
					TotalDowntimeFormatted: "3hr 34min 10s",
					LongestOutage:          4200,
					LongestOutageFormatted: "1hr 10min",
					Details: []OutageDetail{
						{
							StartDate:         "2026-01-15T14:30:00Z",
							EndDate:           "2026-01-15T15:40:00Z",
							Duration:          4200,
							DurationFormatted: "1hr 10min",
						},
					},
				},
			},
			wantErr: false,
			wantReport: &MonitorReport{
				UUID:          "mon_abc123",
				Name:          "Production API",
				Protocol:      "http",
				SLA:           99.184,
				MTTR:          4283,
				MTTRFormatted: "1hr 11min 23s",
			},
		},
		{
			name:           "success - with date range",
			uuid:           "mon_xyz789",
			from:           "2026-01-01T00:00:00Z",
			to:             "2026-01-31T23:59:59Z",
			responseStatus: http.StatusOK,
			responseBody: MonitorReport{
				UUID:     "mon_xyz789",
				Name:     "Staging API",
				Protocol: "http",
				Period:   ReportPeriod{From: "2026-01-01T00:00:00Z", To: "2026-01-31T23:59:59Z"},
				SLA:      99.9,
				Outages: OutageStats{
					Count: 0,
				},
			},
			wantErr: false,
			wantReport: &MonitorReport{
				UUID:     "mon_xyz789",
				Name:     "Staging API",
				Protocol: "http",
				SLA:      99.9,
			},
		},
		{
			name:           "not found",
			uuid:           "mon_NOTFOUND",
			from:           "",
			to:             "",
			responseStatus: http.StatusNotFound,
			responseBody: map[string]string{
				"error": "Monitor not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := ReportsBasePath + "/" + tt.uuid
				if r.URL.Path != expectedPath {
					t.Errorf("unexpected path: got %v, want %v", r.URL.Path, expectedPath)
				}
				if r.Method != http.MethodGet {
					t.Errorf("unexpected method: got %v", r.Method)
				}

				// Check query parameters if provided
				if tt.from != "" {
					if r.URL.Query().Get("from") != tt.from {
						t.Errorf("from parameter: got %v, want %v", r.URL.Query().Get("from"), tt.from)
					}
				}
				if tt.to != "" {
					if r.URL.Query().Get("to") != tt.to {
						t.Errorf("to parameter: got %v, want %v", r.URL.Query().Get("to"), tt.to)
					}
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

			got, err := client.GetMonitorReport(context.Background(), tt.uuid, tt.from, tt.to)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetMonitorReport() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.UUID != tt.wantReport.UUID {
					t.Errorf("UUID = %v, want %v", got.UUID, tt.wantReport.UUID)
				}
				if got.Name != tt.wantReport.Name {
					t.Errorf("Name = %v, want %v", got.Name, tt.wantReport.Name)
				}
				if got.SLA != tt.wantReport.SLA {
					t.Errorf("SLA = %v, want %v", got.SLA, tt.wantReport.SLA)
				}
			}
		})
	}
}

func TestClient_ListMonitorReports(t *testing.T) {
	tests := []struct {
		name           string
		from           string
		to             string
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		wantCount      int
	}{
		{
			name:           "success - without date range",
			from:           "",
			to:             "",
			responseStatus: http.StatusOK,
			responseBody: ListMonitorReportsResponse{
				Period: ReportPeriod{
					From: "2026-01-01T00:00:00Z",
					To:   "2026-01-31T23:59:59Z",
				},
				Monitors: []MonitorReport{
					{
						UUID:     "mon_001",
						Name:     "API Monitor 1",
						Protocol: "http",
						SLA:      99.5,
					},
					{
						UUID:     "mon_002",
						Name:     "API Monitor 2",
						Protocol: "http",
						SLA:      98.2,
					},
				},
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:           "success - with date range",
			from:           "2026-01-01T00:00:00Z",
			to:             "2026-01-31T23:59:59Z",
			responseStatus: http.StatusOK,
			responseBody: ListMonitorReportsResponse{
				Period: ReportPeriod{
					From: "2026-01-01T00:00:00Z",
					To:   "2026-01-31T23:59:59Z",
				},
				Monitors: []MonitorReport{
					{
						UUID:          "mon_003",
						Name:          "API Monitor 3",
						Protocol:      "http",
						SLA:           99.9,
						MTTR:          120,
						MTTRFormatted: "2min",
					},
				},
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:           "success - empty list",
			from:           "",
			to:             "",
			responseStatus: http.StatusOK,
			responseBody: ListMonitorReportsResponse{
				Period: ReportPeriod{
					From: "2026-01-01T00:00:00Z",
					To:   "2026-01-31T23:59:59Z",
				},
				Monitors: []MonitorReport{},
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:           "server error",
			from:           "",
			to:             "",
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
				if r.URL.Path != ReportsBasePath {
					t.Errorf("unexpected path: got %v", r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("unexpected method: got %v", r.Method)
				}

				// Check query parameters if provided
				if tt.from != "" {
					if r.URL.Query().Get("from") != tt.from {
						t.Errorf("from parameter: got %v, want %v", r.URL.Query().Get("from"), tt.from)
					}
				}
				if tt.to != "" {
					if r.URL.Query().Get("to") != tt.to {
						t.Errorf("to parameter: got %v, want %v", r.URL.Query().Get("to"), tt.to)
					}
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

			got, err := client.ListMonitorReports(context.Background(), tt.from, tt.to)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListMonitorReports() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(got) != tt.wantCount {
				t.Errorf("ListMonitorReports() count = %v, want %v", len(got), tt.wantCount)
			}
		})
	}
}

// TestClient_GetMonitorReport_Unauthorized verifies a 401 response produces an unauthorized error.
func TestClient_GetMonitorReport_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid API key"})
	}))
	defer server.Close()

	c := NewClient("bad-key",
		WithHTTPClient(server.Client()),
		WithBaseURL(server.URL),
		WithMaxRetries(0),
	)

	_, err := c.GetMonitorReport(context.Background(), "mon_abc", "", "")
	if err == nil {
		t.Fatal("expected error for 401 response, got nil")
	}
	if !IsUnauthorized(err) {
		t.Errorf("expected IsUnauthorized to be true, got false; err = %v", err)
	}
}

// TestClient_GetMonitorReport_MalformedJSON verifies a 200 response with invalid JSON produces an error.
func TestClient_GetMonitorReport_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	c := NewClient("test-key",
		WithHTTPClient(server.Client()),
		WithBaseURL(server.URL),
		WithMaxRetries(0),
	)

	_, err := c.GetMonitorReport(context.Background(), "mon_abc", "", "")
	if err == nil {
		t.Fatal("expected error for malformed JSON response, got nil")
	}
}

// TestClient_GetMonitorReport_ContextCancellation verifies a pre-cancelled context produces an error.
func TestClient_GetMonitorReport_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(MonitorReport{UUID: "mon_abc"})
	}))
	defer server.Close()

	c := NewClient("test-key",
		WithHTTPClient(server.Client()),
		WithBaseURL(server.URL),
		WithMaxRetries(0),
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.GetMonitorReport(ctx, "mon_abc", "", "")
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
}

// TestClient_ListMonitorReports_Unauthorized verifies a 401 response produces an unauthorized error.
func TestClient_ListMonitorReports_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid API key"})
	}))
	defer server.Close()

	c := NewClient("bad-key",
		WithHTTPClient(server.Client()),
		WithBaseURL(server.URL),
		WithMaxRetries(0),
	)

	_, err := c.ListMonitorReports(context.Background(), "", "")
	if err == nil {
		t.Fatal("expected error for 401 response, got nil")
	}
	if !IsUnauthorized(err) {
		t.Errorf("expected IsUnauthorized to be true, got false; err = %v", err)
	}
}

// TestClient_ListMonitorReports_MalformedJSON verifies a 200 response with invalid JSON produces an error.
func TestClient_ListMonitorReports_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	c := NewClient("test-key",
		WithHTTPClient(server.Client()),
		WithBaseURL(server.URL),
		WithMaxRetries(0),
	)

	_, err := c.ListMonitorReports(context.Background(), "", "")
	if err == nil {
		t.Fatal("expected error for malformed JSON response, got nil")
	}
}

// TestClient_ListMonitorReports_ContextCancellation verifies a pre-cancelled context produces an error.
func TestClient_ListMonitorReports_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ListMonitorReportsResponse{Monitors: []MonitorReport{}})
	}))
	defer server.Close()

	c := NewClient("test-key",
		WithHTTPClient(server.Client()),
		WithBaseURL(server.URL),
		WithMaxRetries(0),
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.ListMonitorReports(ctx, "", "")
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
}
