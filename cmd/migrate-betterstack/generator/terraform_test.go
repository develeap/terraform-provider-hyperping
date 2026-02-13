// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package generator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack/converter"
)

func TestGenerator_GenerateTerraform(t *testing.T) {
	g := New()

	monitors := []converter.ConvertedMonitor{
		{
			ResourceName:       "api_health",
			Name:               "API Health Check",
			URL:                "https://api.example.com/health",
			Protocol:           "http",
			HTTPMethod:         "GET",
			CheckFrequency:     60,
			Regions:            []string{"london", "virginia"},
			ExpectedStatusCode: "200",
			FollowRedirects:    true,
			Paused:             false,
		},
	}

	healthchecks := []converter.ConvertedHealthcheck{
		{
			ResourceName: "daily_backup",
			Name:         "Daily Backup",
			Period:       86400,
			Grace:        300,
			Paused:       false,
		},
	}

	result := g.GenerateTerraform(monitors, healthchecks)

	// Check for required sections
	assert.Contains(t, result, "terraform {")
	assert.Contains(t, result, "required_providers")
	assert.Contains(t, result, "provider \"hyperping\"")
	assert.Contains(t, result, "# ===== MONITORS =====")
	assert.Contains(t, result, "# ===== HEALTHCHECKS =====")
	assert.Contains(t, result, "resource \"hyperping_monitor\" \"api_health\"")
	assert.Contains(t, result, "resource \"hyperping_healthcheck\" \"daily_backup\"")
}

func TestGenerator_GenerateMonitorBlock(t *testing.T) {
	g := New()

	tests := []struct {
		name        string
		monitor     converter.ConvertedMonitor
		contains    []string
		notContains []string
	}{
		{
			name: "basic HTTP monitor",
			monitor: converter.ConvertedMonitor{
				ResourceName:       "test_monitor",
				Name:               "Test Monitor",
				URL:                "https://example.com",
				Protocol:           "http",
				HTTPMethod:         "GET",
				CheckFrequency:     60,
				Regions:            []string{"london"},
				ExpectedStatusCode: "200",
				FollowRedirects:    true,
				Paused:             false,
			},
			contains: []string{
				"resource \"hyperping_monitor\" \"test_monitor\"",
				"name                 = \"Test Monitor\"",
				"url                  = \"https://example.com\"",
				"check_frequency      = 60",
				"regions = [",
				"\"london\"",
			},
			notContains: []string{
				"protocol             =",       // Default, should be omitted
				"http_method          =",       // GET is default
				"expected_status_code =",       // 200 is default
				"follow_redirects     = false", // true is default
				"paused               =",       // false is default
			},
		},
		{
			name: "TCP monitor with port",
			monitor: converter.ConvertedMonitor{
				ResourceName:   "database",
				Name:           "Database",
				URL:            "db.example.com",
				Protocol:       "port",
				CheckFrequency: 120,
				Regions:        []string{"virginia"},
				Port:           5432,
			},
			contains: []string{
				"protocol             = \"port\"",
				"port                 = 5432",
			},
			notContains: []string{
				"http_method",
			},
		},
		{
			name: "monitor with custom HTTP method",
			monitor: converter.ConvertedMonitor{
				ResourceName:       "api_post",
				Name:               "API POST",
				URL:                "https://api.example.com",
				Protocol:           "http",
				HTTPMethod:         "POST",
				CheckFrequency:     60,
				Regions:            []string{"london"},
				RequestBody:        "{\"test\": \"data\"}",
				ExpectedStatusCode: "201",
			},
			contains: []string{
				"http_method          = \"POST\"",
				"expected_status_code = \"201\"",
				"request_body = \"{\\\"test\\\": \\\"data\\\"}\"",
			},
		},
		{
			name: "monitor with headers",
			monitor: converter.ConvertedMonitor{
				ResourceName:   "with_headers",
				Name:           "With Headers",
				URL:            "https://example.com",
				Protocol:       "http",
				CheckFrequency: 60,
				Regions:        []string{"london"},
				RequestHeaders: []converter.RequestHeader{
					{Name: "Authorization", Value: "Bearer token"},
					{Name: "X-Custom", Value: "value"},
				},
			},
			contains: []string{
				"request_headers = [",
				"name  = \"Authorization\"",
				"value = \"Bearer token\"",
				"name  = \"X-Custom\"",
				"value = \"value\"",
			},
		},
		{
			name: "paused monitor",
			monitor: converter.ConvertedMonitor{
				ResourceName:   "paused",
				Name:           "Paused Monitor",
				URL:            "https://example.com",
				Protocol:       "http",
				CheckFrequency: 60,
				Regions:        []string{"london"},
				Paused:         true,
			},
			contains: []string{
				"paused               = true",
			},
		},
		{
			name: "monitor with issues",
			monitor: converter.ConvertedMonitor{
				ResourceName:   "with_issues",
				Name:           "With Issues",
				URL:            "https://example.com",
				Protocol:       "http",
				CheckFrequency: 60,
				Regions:        []string{"london"},
				Issues: []string{
					"Frequency rounded from 45s to 60s",
					"Keyword monitoring not fully supported",
				},
			},
			contains: []string{
				"# MIGRATION NOTES:",
				"# - Frequency rounded from 45s to 60s",
				"# - Keyword monitoring not fully supported",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := g.generateMonitorBlock(tt.monitor)

			for _, s := range tt.contains {
				assert.Contains(t, result, s, "expected to contain: %s", s)
			}

			for _, s := range tt.notContains {
				assert.NotContains(t, result, s, "expected NOT to contain: %s", s)
			}
		})
	}
}

func TestGenerator_GenerateHealthcheckBlock(t *testing.T) {
	g := New()

	tests := []struct {
		name        string
		healthcheck converter.ConvertedHealthcheck
		contains    []string
	}{
		{
			name: "basic healthcheck",
			healthcheck: converter.ConvertedHealthcheck{
				ResourceName: "daily_backup",
				Name:         "Daily Backup",
				Period:       86400,
				Grace:        300,
				Paused:       false,
			},
			contains: []string{
				"resource \"hyperping_healthcheck\" \"daily_backup\"",
				"name               = \"Daily Backup\"",
				"cron               = \"0 0 * * *\"",
				"timezone           = \"UTC\"",
				"grace_period_value = 5",
				"grace_period_type  = \"minutes\"",
			},
		},
		{
			name: "hourly healthcheck",
			healthcheck: converter.ConvertedHealthcheck{
				ResourceName: "hourly_sync",
				Name:         "Hourly Sync",
				Period:       3600,
				Grace:        600,
				Paused:       false,
			},
			contains: []string{
				"cron               = \"0 * * * *\"",
				"grace_period_value = 10",
			},
		},
		{
			name: "paused healthcheck",
			healthcheck: converter.ConvertedHealthcheck{
				ResourceName: "paused_check",
				Name:         "Paused",
				Period:       3600,
				Grace:        300,
				Paused:       true,
			},
			contains: []string{
				"paused             = true",
			},
		},
		{
			name: "healthcheck with issues",
			healthcheck: converter.ConvertedHealthcheck{
				ResourceName: "with_issues",
				Name:         "With Issues",
				Period:       3600,
				Grace:        30,
				Issues: []string{
					"Grace period 30s is less than minimum",
				},
			},
			contains: []string{
				"# MIGRATION NOTES:",
				"# - Grace period 30s is less than minimum",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := g.generateHealthcheckBlock(tt.healthcheck)

			for _, s := range tt.contains {
				assert.Contains(t, result, s)
			}
		})
	}
}

func TestPeriodToCron(t *testing.T) {
	tests := []struct {
		period   int
		expected string
	}{
		{60, "* * * * *"},      // Every minute
		{300, "*/5 * * * *"},   // Every 5 minutes
		{600, "*/10 * * * *"},  // Every 10 minutes
		{1800, "*/30 * * * *"}, // Every 30 minutes
		{3600, "0 * * * *"},    // Hourly
		{7200, "0 */2 * * *"},  // Every 2 hours
		{86400, "0 0 * * *"},   // Daily
		{100000, "0 0 * * *"},  // > 1 day, defaults to daily
	}

	for _, tt := range tests {
		result := periodToCron(tt.period)
		assert.Equal(t, tt.expected, result)
	}
}

func TestQuoteString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "simple",
			expected: "\"simple\"",
		},
		{
			input:    "with \"quotes\"",
			expected: "\"with \\\"quotes\\\"\"",
		},
		{
			input:    "with\\backslash",
			expected: "\"with\\\\backslash\"",
		},
		{
			input:    "with\nnewline",
			expected: "\"with\\nnewline\"",
		},
		{
			input:    "with\ttab",
			expected: "\"with\\ttab\"",
		},
		{
			input:    "with\rcarriage",
			expected: "\"with\\rcarriage\"",
		},
		{
			input:    `{"json": "value"}`,
			expected: `"{\"json\": \"value\"}"`,
		},
	}

	for _, tt := range tests {
		result := quoteString(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestGenerator_GenerateTerraform_EmptyInputs(t *testing.T) {
	g := New()

	result := g.GenerateTerraform([]converter.ConvertedMonitor{}, []converter.ConvertedHealthcheck{})

	// Should still have terraform block
	assert.Contains(t, result, "terraform {")
	assert.Contains(t, result, "provider \"hyperping\"")

	// Should not have resource sections
	assert.NotContains(t, result, "# ===== MONITORS =====")
	assert.NotContains(t, result, "# ===== HEALTHCHECKS =====")
}

func TestGenerator_GenerateTerraform_MultipleRegions(t *testing.T) {
	g := New()

	monitors := []converter.ConvertedMonitor{
		{
			ResourceName:   "multi_region",
			Name:           "Multi Region",
			URL:            "https://example.com",
			Protocol:       "http",
			CheckFrequency: 60,
			Regions:        []string{"london", "virginia", "singapore", "tokyo"},
		},
	}

	result := g.GenerateTerraform(monitors, nil)

	// All regions should be present
	assert.Contains(t, result, "\"london\"")
	assert.Contains(t, result, "\"virginia\"")
	assert.Contains(t, result, "\"singapore\"")
	assert.Contains(t, result, "\"tokyo\"")
}

func TestGenerator_GenerateTerraform_ComplexExample(t *testing.T) {
	g := New()

	monitors := []converter.ConvertedMonitor{
		{
			ResourceName:   "api",
			Name:           "API",
			URL:            "https://api.example.com",
			Protocol:       "http",
			HTTPMethod:     "POST",
			CheckFrequency: 30,
			Regions:        []string{"london", "virginia"},
			RequestHeaders: []converter.RequestHeader{
				{Name: "Authorization", Value: "Bearer secret"},
			},
			RequestBody:        `{"test": true}`,
			ExpectedStatusCode: "201",
			FollowRedirects:    false,
		},
	}

	healthchecks := []converter.ConvertedHealthcheck{
		{
			ResourceName: "backup",
			Name:         "Backup",
			Period:       86400,
			Grace:        3600,
		},
	}

	result := g.GenerateTerraform(monitors, healthchecks)

	// Verify complex monitor
	assert.Contains(t, result, "http_method          = \"POST\"")
	assert.Contains(t, result, "expected_status_code = \"201\"")
	assert.Contains(t, result, "follow_redirects     = false")
	assert.Contains(t, result, "Authorization")
	assert.Contains(t, result, "request_body")

	// Verify healthcheck
	assert.Contains(t, result, "grace_period_value = 60") // 3600 seconds / 60 = 60 minutes
}

func TestGenerator_GenerateMonitorBlock_EscapedValues(t *testing.T) {
	g := New()

	monitor := converter.ConvertedMonitor{
		ResourceName:   "escaped",
		Name:           "Monitor with \"quotes\" and \\backslashes",
		URL:            "https://example.com/path?query=\"value\"",
		Protocol:       "http",
		CheckFrequency: 60,
		Regions:        []string{"london"},
	}

	result := g.generateMonitorBlock(monitor)

	// Check escaping
	assert.Contains(t, result, "\\\"quotes\\\"")
	assert.Contains(t, result, "\\\\backslashes")
}

func TestGenerator_GenerateTerraform_ValidHCL(t *testing.T) {
	g := New()

	monitors := []converter.ConvertedMonitor{
		{
			ResourceName:   "test",
			Name:           "Test",
			URL:            "https://example.com",
			Protocol:       "http",
			CheckFrequency: 60,
			Regions:        []string{"london"},
		},
	}

	result := g.GenerateTerraform(monitors, nil)

	// Basic HCL syntax checks
	assert.Equal(t, strings.Count(result, "{"), strings.Count(result, "}"), "braces should be balanced")
	assert.NotContains(t, result, "  =  ", "should have proper spacing")
	assert.Contains(t, result, "\n}\n", "resources should end with newline")
}
