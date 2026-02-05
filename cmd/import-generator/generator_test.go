// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"strings"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestTerraformName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		prefix   string
		expected string
	}{
		{
			name:     "simple name",
			input:    "api",
			expected: "api",
		},
		{
			name:     "name with spaces",
			input:    "My API Monitor",
			expected: "my_api_monitor",
		},
		{
			name:     "name with special chars",
			input:    "[PROD]-API-Health",
			expected: "prod_api_health",
		},
		{
			name:     "name starting with number",
			input:    "123-test",
			expected: "r_123_test",
		},
		{
			name:     "with prefix",
			input:    "api",
			prefix:   "prod_",
			expected: "prod_api",
		},
		{
			name:     "empty name",
			input:    "",
			expected: "resource",
		},
		{
			name:     "unicode characters",
			input:    "API-健康检查",
			expected: "api",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := &Generator{prefix: tc.prefix}
			result := g.terraformName(tc.input)
			if result != tc.expected {
				t.Errorf("terraformName(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestEscapeHCL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with \"quotes\"", "with \\\"quotes\\\""},
		{"with\nnewline", "with\\nnewline"},
		{"with\\backslash", "with\\\\backslash"},
	}

	for _, tc := range tests {
		result := escapeHCL(tc.input)
		if result != tc.expected {
			t.Errorf("escapeHCL(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestFormatStringList(t *testing.T) {
	tests := []struct {
		input    []string
		expected string
	}{
		{nil, "[]"},
		{[]string{}, "[]"},
		{[]string{"a"}, "[\"a\"]"},
		{[]string{"a", "b"}, "[\"a\", \"b\"]"},
		{[]string{"with \"quote\""}, "[\"with \\\"quote\\\"\"]"},
	}

	for _, tc := range tests {
		result := formatStringList(tc.input)
		if result != tc.expected {
			t.Errorf("formatStringList(%v) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestGenerateMonitorHCL(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	monitor := client.Monitor{
		UUID:           "mon_123",
		Name:           "Test Monitor",
		URL:            "https://example.com",
		Protocol:       "http",
		CheckFrequency: 60,
		Regions:        []string{"virginia"},
	}

	g.generateMonitorHCL(&sb, monitor)
	result := sb.String()

	// Check essential parts
	if !strings.Contains(result, `resource "hyperping_monitor" "test_monitor"`) {
		t.Error("Missing resource declaration")
	}
	if !strings.Contains(result, `name     = "Test Monitor"`) {
		t.Error("Missing name attribute")
	}
	if !strings.Contains(result, `url      = "https://example.com"`) {
		t.Error("Missing url attribute")
	}
	if !strings.Contains(result, `protocol = "http"`) {
		t.Error("Missing protocol attribute")
	}
}

func TestGenerateImports(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	data := &ResourceData{
		Monitors: []client.Monitor{
			{UUID: "mon_123", Name: "Test"},
		},
		Healthchecks: []client.Healthcheck{
			{UUID: "hc_456", Name: "Backup Job"},
		},
	}

	g.generateImports(&sb, data)
	result := sb.String()

	if !strings.Contains(result, `terraform import hyperping_monitor.test "mon_123"`) {
		t.Error("Missing monitor import command")
	}
	if !strings.Contains(result, `terraform import hyperping_healthcheck.backup_job "hc_456"`) {
		t.Error("Missing healthcheck import command")
	}
}
