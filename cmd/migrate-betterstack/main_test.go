// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack/betterstack"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack/converter"
)

func TestSanitizeResourceName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "API Monitor",
			expected: "api_monitor",
		},
		{
			name:     "special characters",
			input:    "[PROD]-API-Gateway",
			expected: "prod_api_gateway",
		},
		{
			name:     "leading digit",
			input:    "123-monitor",
			expected: "monitor_123_monitor",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "unnamed_monitor",
		},
		{
			name:     "unicode characters",
			input:    "Mönitor-Test",
			expected: "m_nitor_test",
		},
		{
			name:     "multiple underscores",
			input:    "test___monitor___name",
			expected: "test_monitor_name",
		},
	}

	conv := converter.New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We test through conversion since sanitizeResourceName is not exported
			monitor := betterstack.Monitor{
				ID:   "test-id",
				Type: "monitor",
				Attributes: betterstack.MonitorAttributes{
					PronouncableName:    tt.input,
					URL:                 "https://example.com",
					MonitorType:         "status",
					CheckFrequency:      60,
					RequestMethod:       "GET",
					ExpectedStatusCodes: []int{200},
					FollowRedirects:     true,
				},
			}

			converted, _ := conv.ConvertMonitors([]betterstack.Monitor{monitor})
			require.Len(t, converted, 1)
			assert.Equal(t, tt.expected, converted[0].ResourceName)
		})
	}
}

func TestMonitorConversion(t *testing.T) {
	tests := []struct {
		name          string
		monitor       betterstack.Monitor
		expectedName  string
		expectedProto string
		expectIssues  bool
	}{
		{
			name: "basic HTTP monitor",
			monitor: betterstack.Monitor{
				ID:   "mon-1",
				Type: "monitor",
				Attributes: betterstack.MonitorAttributes{
					PronouncableName:    "API Health",
					URL:                 "https://api.example.com/health",
					MonitorType:         "status",
					CheckFrequency:      60,
					RequestMethod:       "GET",
					ExpectedStatusCodes: []int{200},
					FollowRedirects:     true,
					Regions:             []string{"us", "eu"},
				},
			},
			expectedName:  "api_health",
			expectedProto: "http",
			expectIssues:  false,
		},
		{
			name: "TCP port monitor",
			monitor: betterstack.Monitor{
				ID:   "mon-2",
				Type: "monitor",
				Attributes: betterstack.MonitorAttributes{
					PronouncableName: "Database",
					URL:              "db.example.com",
					MonitorType:      "tcp",
					CheckFrequency:   60,
					Port:             5432,
					Regions:          []string{"us-east-1"},
				},
			},
			expectedName:  "database",
			expectedProto: "port",
			expectIssues:  false,
		},
		{
			name: "keyword monitor with warning",
			monitor: betterstack.Monitor{
				ID:   "mon-3",
				Type: "monitor",
				Attributes: betterstack.MonitorAttributes{
					PronouncableName:    "Search Page",
					URL:                 "https://example.com/search",
					MonitorType:         "keyword",
					CheckFrequency:      60,
					RequestMethod:       "GET",
					ExpectedStatusCodes: []int{200},
					FollowRedirects:     true,
				},
			},
			expectedName:  "search_page",
			expectedProto: "http",
			expectIssues:  true,
		},
	}

	conv := converter.New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converted, issues := conv.ConvertMonitors([]betterstack.Monitor{tt.monitor})

			require.Len(t, converted, 1)
			assert.Equal(t, tt.expectedName, converted[0].ResourceName)
			assert.Equal(t, tt.expectedProto, converted[0].Protocol)

			if tt.expectIssues {
				assert.NotEmpty(t, issues, "expected conversion issues")
			}
		})
	}
}

func TestFrequencyMapping(t *testing.T) {
	tests := []struct {
		name        string
		input       int
		expected    int
		expectIssue bool
	}{
		{
			name:        "exact match",
			input:       60,
			expected:    60,
			expectIssue: false,
		},
		{
			name:        "round up 45s to 60s",
			input:       45,
			expected:    60,
			expectIssue: true,
		},
		{
			name:        "round down 90s to 60s",
			input:       90,
			expected:    60,
			expectIssue: true,
		},
		{
			name:        "round 240s to 300s",
			input:       240,
			expected:    300,
			expectIssue: true,
		},
	}

	conv := converter.New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := betterstack.Monitor{
				ID:   "test",
				Type: "monitor",
				Attributes: betterstack.MonitorAttributes{
					PronouncableName:    "Test",
					URL:                 "https://example.com",
					MonitorType:         "status",
					CheckFrequency:      tt.input,
					RequestMethod:       "GET",
					ExpectedStatusCodes: []int{200},
				},
			}

			converted, issues := conv.ConvertMonitors([]betterstack.Monitor{monitor})
			require.Len(t, converted, 1)
			assert.Equal(t, tt.expected, converted[0].CheckFrequency)

			if tt.expectIssue {
				assert.NotEmpty(t, issues)
				hasFrequencyIssue := false
				for _, issue := range issues {
					if issue.Message != "" {
						hasFrequencyIssue = true
						break
					}
				}
				assert.True(t, hasFrequencyIssue, "expected frequency conversion issue")
			}
		})
	}
}

func TestRegionMapping(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "simple regions",
			input:    []string{"us", "eu"},
			expected: []string{"virginia", "london"},
		},
		{
			name:     "specific regions",
			input:    []string{"us-east-1", "eu-west-1", "ap-southeast-1"},
			expected: []string{"virginia", "london", "singapore"},
		},
		{
			name:     "duplicate regions",
			input:    []string{"us", "us-east-1"},
			expected: []string{"virginia"},
		},
		{
			name:     "empty regions",
			input:    []string{},
			expected: []string{"london", "virginia", "singapore"},
		},
	}

	conv := converter.New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := betterstack.Monitor{
				ID:   "test",
				Type: "monitor",
				Attributes: betterstack.MonitorAttributes{
					PronouncableName:    "Test",
					URL:                 "https://example.com",
					MonitorType:         "status",
					CheckFrequency:      60,
					RequestMethod:       "GET",
					ExpectedStatusCodes: []int{200},
					Regions:             tt.input,
				},
			}

			converted, _ := conv.ConvertMonitors([]betterstack.Monitor{monitor})
			require.Len(t, converted, 1)
			assert.ElementsMatch(t, tt.expected, converted[0].Regions)
		})
	}
}

func TestHeartbeatConversion(t *testing.T) {
	tests := []struct {
		name         string
		heartbeat    betterstack.Heartbeat
		expectedName string
		expectIssues bool
	}{
		{
			name: "basic heartbeat",
			heartbeat: betterstack.Heartbeat{
				ID:   "hb-1",
				Type: "heartbeat",
				Attributes: betterstack.HeartbeatAttributes{
					Name:   "Daily Backup",
					Period: 86400,
					Grace:  300,
					Paused: false,
				},
			},
			expectedName: "daily_backup",
			expectIssues: false,
		},
		{
			name: "heartbeat with low grace period",
			heartbeat: betterstack.Heartbeat{
				ID:   "hb-2",
				Type: "heartbeat",
				Attributes: betterstack.HeartbeatAttributes{
					Name:   "Fast Check",
					Period: 300,
					Grace:  30,
					Paused: false,
				},
			},
			expectedName: "fast_check",
			expectIssues: true,
		},
	}

	conv := converter.New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converted, issues := conv.ConvertHeartbeats([]betterstack.Heartbeat{tt.heartbeat})

			require.Len(t, converted, 1)
			assert.Equal(t, tt.expectedName, converted[0].ResourceName)

			if tt.expectIssues {
				assert.NotEmpty(t, issues, "expected conversion issues")
			}
		})
	}
}
