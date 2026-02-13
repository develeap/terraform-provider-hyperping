// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack/betterstack"
)

func TestConverter_MapProtocol(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"http status", "status", "http"},
		{"tcp", "tcp", "port"},
		{"ping", "ping", "icmp"},
		{"keyword", "keyword", "http"},
		{"heartbeat", "heartbeat", "healthcheck"},
		{"unknown", "unknown", ""},
	}

	c := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.mapProtocol(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConverter_MapFrequency(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"exact match 60", 60, 60},
		{"exact match 120", 120, 120},
		{"round 45 to 60", 45, 60},
		{"round 90 to 60", 90, 60},
		{"round 240 to 300", 240, 300},
		{"round 7200 to 3600", 7200, 3600},
		{"very small to 10", 5, 10},
		{"very large to 86400", 100000, 86400},
	}

	c := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.mapFrequency(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConverter_MapRegions(t *testing.T) {
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
			name:     "specific AWS regions",
			input:    []string{"us-east-1", "eu-west-1", "ap-southeast-1"},
			expected: []string{"virginia", "london", "singapore"},
		},
		{
			name:     "duplicates removed",
			input:    []string{"us", "us-east", "us-east-1"},
			expected: []string{"virginia"},
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "mixed case",
			input:    []string{"US", "EU", "Asia"},
			expected: []string{"virginia", "london", "singapore"},
		},
		{
			name:     "unknown regions ignored",
			input:    []string{"unknown", "us"},
			expected: []string{"virginia"},
		},
	}

	c := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.mapRegions(tt.input)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

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
			name:     "with brackets",
			input:    "[PROD]-API-Gateway",
			expected: "prod_api_gateway",
		},
		{
			name:     "with numbers",
			input:    "Monitor 123",
			expected: "monitor_123",
		},
		{
			name:     "leading number",
			input:    "123 Monitor",
			expected: "monitor_123_monitor",
		},
		{
			name:     "only numbers",
			input:    "123",
			expected: "monitor_123",
		},
		{
			name:     "special characters",
			input:    "Test@Monitor#123!",
			expected: "test_monitor_123",
		},
		{
			name:     "multiple spaces",
			input:    "Test   Monitor   Name",
			expected: "test_monitor_name",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "unnamed_monitor",
		},
		{
			name:     "only special chars",
			input:    "@#$%",
			expected: "unnamed_monitor",
		},
		{
			name:     "trailing underscores",
			input:    "__test__",
			expected: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeResourceName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConverter_ConvertMonitor_HTTPMonitor(t *testing.T) {
	c := New()
	monitor := betterstack.Monitor{
		ID:   "mon-123",
		Type: "monitor",
		Attributes: betterstack.MonitorAttributes{
			PronouncableName: "API Health Check",
			URL:              "https://api.example.com/health",
			MonitorType:      "status",
			CheckFrequency:   60,
			RequestMethod:    "GET",
			RequestHeaders: []betterstack.RequestHeader{
				{Name: "Authorization", Value: "Bearer token"},
				{Name: "X-Custom", Value: "value"},
			},
			ExpectedStatusCodes: []int{200},
			FollowRedirects:     true,
			Paused:              false,
			Regions:             []string{"us-east-1", "eu-west-1"},
		},
	}

	converted, issues := c.convertMonitor(monitor)

	assert.Equal(t, "api_health_check", converted.ResourceName)
	assert.Equal(t, "API Health Check", converted.Name)
	assert.Equal(t, "https://api.example.com/health", converted.URL)
	assert.Equal(t, "http", converted.Protocol)
	assert.Equal(t, "GET", converted.HTTPMethod)
	assert.Equal(t, 60, converted.CheckFrequency)
	assert.Equal(t, "200", converted.ExpectedStatusCode)
	assert.True(t, converted.FollowRedirects)
	assert.False(t, converted.Paused)
	assert.ElementsMatch(t, []string{"virginia", "london"}, converted.Regions)
	assert.Len(t, converted.RequestHeaders, 2)
	assert.Empty(t, issues)
}

func TestConverter_ConvertMonitor_TCPMonitor(t *testing.T) {
	c := New()
	monitor := betterstack.Monitor{
		ID:   "mon-456",
		Type: "monitor",
		Attributes: betterstack.MonitorAttributes{
			PronouncableName: "Database",
			URL:              "db.example.com",
			MonitorType:      "tcp",
			CheckFrequency:   120,
			Port:             5432,
			Regions:          []string{"us-east-1"},
		},
	}

	converted, issues := c.convertMonitor(monitor)

	assert.Equal(t, "database", converted.ResourceName)
	assert.Equal(t, "port", converted.Protocol)
	assert.Equal(t, 5432, converted.Port)
	assert.Equal(t, 120, converted.CheckFrequency)
	assert.Empty(t, issues)
}

func TestConverter_ConvertMonitor_KeywordMonitor(t *testing.T) {
	c := New()
	monitor := betterstack.Monitor{
		ID:   "mon-789",
		Type: "monitor",
		Attributes: betterstack.MonitorAttributes{
			PronouncableName:    "Search Page",
			URL:                 "https://example.com/search",
			MonitorType:         "keyword",
			CheckFrequency:      60,
			RequestMethod:       "GET",
			ExpectedStatusCodes: []int{200},
		},
	}

	converted, issues := c.convertMonitor(monitor)

	assert.Equal(t, "http", converted.Protocol)
	assert.NotEmpty(t, issues)

	// Should have warning about keyword monitoring
	found := false
	for _, issue := range issues {
		if issue.Severity == "warning" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected warning about keyword monitoring")
}

func TestConverter_ConvertMonitor_FrequencyRounding(t *testing.T) {
	c := New()
	monitor := betterstack.Monitor{
		ID:   "mon-freq",
		Type: "monitor",
		Attributes: betterstack.MonitorAttributes{
			PronouncableName:    "Test",
			URL:                 "https://example.com",
			MonitorType:         "status",
			CheckFrequency:      45, // Should round to 60
			RequestMethod:       "GET",
			ExpectedStatusCodes: []int{200},
		},
	}

	converted, issues := c.convertMonitor(monitor)

	assert.Equal(t, 60, converted.CheckFrequency)
	assert.NotEmpty(t, issues)

	// Should have warning about frequency rounding
	found := false
	for _, issue := range issues {
		if issue.Message != "" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected warning about frequency rounding")
}

func TestConverter_ConvertMonitor_MultipleStatusCodes(t *testing.T) {
	c := New()
	monitor := betterstack.Monitor{
		ID:   "mon-status",
		Type: "monitor",
		Attributes: betterstack.MonitorAttributes{
			PronouncableName:    "Test",
			URL:                 "https://example.com",
			MonitorType:         "status",
			CheckFrequency:      60,
			RequestMethod:       "GET",
			ExpectedStatusCodes: []int{200, 201, 204},
		},
	}

	converted, issues := c.convertMonitor(monitor)

	assert.Equal(t, "200", converted.ExpectedStatusCode)
	assert.NotEmpty(t, issues)

	// Should have warning about multiple status codes
	found := false
	for _, issue := range issues {
		if issue.Message != "" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected warning about multiple status codes")
}

func TestConverter_ConvertHeartbeat(t *testing.T) {
	c := New()
	heartbeat := betterstack.Heartbeat{
		ID:   "hb-123",
		Type: "heartbeat",
		Attributes: betterstack.HeartbeatAttributes{
			Name:   "Daily Backup",
			Period: 86400,
			Grace:  300,
			Paused: false,
		},
	}

	converted, issues := c.convertHeartbeat(heartbeat)

	assert.Equal(t, "daily_backup", converted.ResourceName)
	assert.Equal(t, "Daily Backup", converted.Name)
	assert.Equal(t, 86400, converted.Period)
	assert.Equal(t, 300, converted.Grace)
	assert.False(t, converted.Paused)
	assert.Empty(t, issues)
}

func TestConverter_ConvertHeartbeat_LowGrace(t *testing.T) {
	c := New()
	heartbeat := betterstack.Heartbeat{
		ID:   "hb-456",
		Type: "heartbeat",
		Attributes: betterstack.HeartbeatAttributes{
			Name:   "Fast Check",
			Period: 300,
			Grace:  30, // Less than minimum
			Paused: false,
		},
	}

	converted, issues := c.convertHeartbeat(heartbeat)

	assert.Equal(t, 30, converted.Grace)
	assert.NotEmpty(t, issues)

	// Should have warning about low grace period
	found := false
	for _, issue := range issues {
		if issue.Severity == "warning" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected warning about low grace period")
}

func TestConverter_ConvertMonitors_Batch(t *testing.T) {
	c := New()
	monitors := []betterstack.Monitor{
		{
			ID:   "mon-1",
			Type: "monitor",
			Attributes: betterstack.MonitorAttributes{
				PronouncableName:    "Monitor 1",
				URL:                 "https://example.com/1",
				MonitorType:         "status",
				CheckFrequency:      60,
				RequestMethod:       "GET",
				ExpectedStatusCodes: []int{200},
			},
		},
		{
			ID:   "mon-2",
			Type: "monitor",
			Attributes: betterstack.MonitorAttributes{
				PronouncableName: "Monitor 2",
				URL:              "https://example.com/2",
				MonitorType:      "tcp",
				CheckFrequency:   120,
				Port:             8080,
			},
		},
	}

	converted, issues := c.ConvertMonitors(monitors)

	assert.Len(t, converted, 2)
	assert.Equal(t, "monitor_1", converted[0].ResourceName)
	assert.Equal(t, "monitor_2", converted[1].ResourceName)

	// Second monitor should have default regions warning
	assert.NotEmpty(t, issues)
}

func TestConverter_ConvertHeartbeats_Batch(t *testing.T) {
	c := New()
	heartbeats := []betterstack.Heartbeat{
		{
			ID:   "hb-1",
			Type: "heartbeat",
			Attributes: betterstack.HeartbeatAttributes{
				Name:   "Heartbeat 1",
				Period: 3600,
				Grace:  300,
			},
		},
		{
			ID:   "hb-2",
			Type: "heartbeat",
			Attributes: betterstack.HeartbeatAttributes{
				Name:   "Heartbeat 2",
				Period: 86400,
				Grace:  600,
			},
		},
	}

	converted, issues := c.ConvertHeartbeats(heartbeats)

	require.Len(t, converted, 2)
	assert.Equal(t, "heartbeat_1", converted[0].ResourceName)
	assert.Equal(t, "heartbeat_2", converted[1].ResourceName)
	assert.Empty(t, issues)
}

func TestAbs(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{5, 5},
		{-5, 5},
		{0, 0},
		{100, 100},
		{-100, 100},
	}

	for _, tt := range tests {
		result := abs(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}
