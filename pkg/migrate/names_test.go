// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package migrate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeResourceName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple name", "API Monitor", "api_monitor"},
		{"with brackets", "[PROD]-API-Gateway", "prod_api_gateway"},
		{"with numbers", "Monitor 123", "monitor_123"},
		{"leading number", "123 Monitor", "monitor_123_monitor"},
		{"only numbers", "123", "monitor_123"},
		{"special characters", "Test@Monitor#123!", "test_monitor_123"},
		{"multiple spaces", "Test   Monitor   Name", "test_monitor_name"},
		{"empty string", "", "unnamed_monitor"},
		{"only special chars", "@#$%", "unnamed_monitor"},
		{"trailing underscores", "__test__", "test"},
		{"parentheses", "Monitor (v2)", "monitor_v2"},
		{"mixed case", "MyProduction-API", "myproduction_api"},
		{"dots", "api.example.com", "api_example_com"},
		{"already valid", "valid_name_123", "valid_name_123"},
		{"unicode", "Mönitor Ünit", "m_nitor_nit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeResourceName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeResourceNameWith(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		opts     SanitizeOpts
		expected string
	}{
		{
			"leading number with custom prefix",
			"123check",
			SanitizeOpts{DigitPrefix: "check", EmptyFallback: "unnamed_check"},
			"check_123check",
		},
		{
			"empty with custom fallback",
			"",
			SanitizeOpts{DigitPrefix: "check", EmptyFallback: "unnamed_healthcheck"},
			"unnamed_healthcheck",
		},
		{
			"normal name ignores prefix",
			"my_monitor",
			SanitizeOpts{DigitPrefix: "check", EmptyFallback: "unnamed_check"},
			"my_monitor",
		},
		{
			"UptimeRobot convention: r_ prefix",
			"123-Monitor",
			SanitizeOpts{DigitPrefix: "r", EmptyFallback: "monitor"},
			"r_123_monitor",
		},
		{
			"UptimeRobot convention: empty fallback",
			"",
			SanitizeOpts{DigitPrefix: "r", EmptyFallback: "monitor"},
			"monitor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeResourceNameWith(tt.input, tt.opts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDeduplicateResourceName(t *testing.T) {
	tests := []struct {
		name     string
		calls    []string
		expected []string
	}{
		{
			name:     "unique names",
			calls:    []string{"monitor_a", "monitor_b", "monitor_c"},
			expected: []string{"monitor_a", "monitor_b", "monitor_c"},
		},
		{
			name:     "duplicate names",
			calls:    []string{"api", "api", "api"},
			expected: []string{"api", "api_2", "api_3"},
		},
		{
			name:     "mixed duplicates",
			calls:    []string{"web", "api", "web", "db", "api"},
			expected: []string{"web", "api", "web_2", "db", "api_2"},
		},
		{
			name:     "single name",
			calls:    []string{"solo"},
			expected: []string{"solo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seen := make(map[string]int)
			for i, name := range tt.calls {
				result := DeduplicateResourceName(name, seen)
				assert.Equal(t, tt.expected[i], result, "call %d", i)
			}
		})
	}
}
