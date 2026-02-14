// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package dryrun

import (
	"strings"
	"testing"
)

func TestExtractResources(t *testing.T) {
	generator := NewPreviewGenerator(false)

	tfContent := `
resource "hyperping_monitor" "test1" {
  name = "Test 1"
  url  = "https://example.com"
}

resource "hyperping_monitor" "test2" {
  name = "Test 2"
  url  = "https://example2.com"
}

resource "hyperping_healthcheck" "test3" {
  name = "Test 3"
}
`

	resources := generator.extractResources(tfContent)

	if len(resources) != 3 {
		t.Errorf("expected 3 resources, got %d", len(resources))
	}

	if !strings.Contains(resources[0], "test1") {
		t.Error("first resource should contain test1")
	}

	if !strings.Contains(resources[1], "test2") {
		t.Error("second resource should contain test2")
	}

	if !strings.Contains(resources[2], "test3") {
		t.Error("third resource should contain test3")
	}
}

func TestAnalyzeResources(t *testing.T) {
	generator := NewPreviewGenerator(false)

	resources := []string{
		`resource "hyperping_monitor" "test1" {}`,
		`resource "hyperping_monitor" "test2" {}`,
		`resource "hyperping_healthcheck" "test3" {}`,
	}

	breakdown := generator.analyzeResources(resources)

	if breakdown["hyperping_monitor"] != 2 {
		t.Errorf("expected 2 monitors, got %d", breakdown["hyperping_monitor"])
	}

	if breakdown["hyperping_healthcheck"] != 1 {
		t.Errorf("expected 1 healthcheck, got %d", breakdown["hyperping_healthcheck"])
	}
}

func TestExtractResourceType(t *testing.T) {
	generator := NewPreviewGenerator(false)

	tests := []struct {
		name     string
		resource string
		expected string
	}{
		{
			name:     "monitor",
			resource: `resource "hyperping_monitor" "test" {}`,
			expected: "hyperping_monitor",
		},
		{
			name:     "healthcheck",
			resource: `resource "hyperping_healthcheck" "test" {}`,
			expected: "hyperping_healthcheck",
		},
		{
			name:     "multiline",
			resource: "resource \"hyperping_monitor\" \"test\" {\n  name = \"test\"\n}",
			expected: "hyperping_monitor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.extractResourceType(tt.resource)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	generator := NewPreviewGenerator(false)

	tests := []struct {
		bytes    int64
		expected string
	}{
		{100, "100 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := generator.formatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGeneratePreview(t *testing.T) {
	generator := NewPreviewGenerator(false)

	tfContent := `
resource "hyperping_monitor" "test1" {
  name = "Test 1"
}

resource "hyperping_monitor" "test2" {
  name = "Test 2"
}
`

	result := generator.GeneratePreview(tfContent, 2, 3)

	if !strings.Contains(result, "Terraform Preview") {
		t.Error("expected preview header")
	}

	if !strings.Contains(result, "Resource Breakdown") {
		t.Error("expected resource breakdown")
	}

	if !strings.Contains(result, "hyperping_monitor") {
		t.Error("expected resource type in breakdown")
	}
}

func TestGenerateResourceSummary(t *testing.T) {
	generator := NewPreviewGenerator(false)

	freqDist := map[int]int{
		30:  5,
		60:  10,
		120: 2,
	}

	regionDist := map[string]int{
		"london":   12,
		"virginia": 10,
		"tokyo":    5,
	}

	result := generator.GenerateResourceSummary(17, 3, freqDist, regionDist)

	if !strings.Contains(result, "Resource Distribution") {
		t.Error("expected distribution header")
	}

	if !strings.Contains(result, "Monitors:     17") {
		t.Error("expected monitor count")
	}

	if !strings.Contains(result, "Healthchecks: 3") {
		t.Error("expected healthcheck count")
	}

	if !strings.Contains(result, "Check Frequency Distribution") {
		t.Error("expected frequency distribution")
	}

	if !strings.Contains(result, "Region Distribution") {
		t.Error("expected region distribution")
	}
}
