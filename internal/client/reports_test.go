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
				expectedPath := "/v2/reporting/monitor-reports/" + tt.uuid
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
			responseBody: []MonitorReport{
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
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:           "success - with date range",
			from:           "2026-01-01T00:00:00Z",
			to:             "2026-01-31T23:59:59Z",
			responseStatus: http.StatusOK,
			responseBody: []MonitorReport{
				{
					UUID:          "mon_003",
					Name:          "API Monitor 3",
					Protocol:      "http",
					SLA:           99.9,
					MTTR:          120,
					MTTRFormatted: "2min",
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
			responseBody:   []MonitorReport{},
			wantErr:        false,
			wantCount:      0,
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
				if r.URL.Path != "/v2/reporting/monitor-reports" {
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
