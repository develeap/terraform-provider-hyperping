// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package dryrun

import (
	"strings"
	"testing"
)

func TestFormatComparison(t *testing.T) {
	formatter := NewDiffFormatter(false)

	comp := ResourceComparison{
		ResourceName: "test_monitor",
		ResourceType: "hyperping_monitor",
		Transformations: []Transformation{
			{
				SourceField: "check_frequency",
				TargetField: "frequency",
				SourceValue: 30,
				TargetValue: 30,
				Action:      "preserved",
				Notes:       "Frequency preserved",
			},
			{
				SourceField: "request_timeout",
				TargetField: "timeout",
				SourceValue: 10,
				TargetValue: 10,
				Action:      "mapped",
				Notes:       "Timeout mapped",
			},
		},
		Unsupported: []string{"custom_headers"},
	}

	result := formatter.FormatComparison(comp, 80)

	if !strings.Contains(result, "test_monitor") {
		t.Error("expected resource name in output")
	}

	if !strings.Contains(result, "hyperping_monitor") {
		t.Error("expected resource type in output")
	}

	if !strings.Contains(result, "check_frequency") {
		t.Error("expected source field in output")
	}

	if !strings.Contains(result, "custom_headers") {
		t.Error("expected unsupported feature in output")
	}
}

func TestFormatComparisonList(t *testing.T) {
	formatter := NewDiffFormatter(false)

	comparisons := []ResourceComparison{
		{ResourceName: "test1", ResourceType: "monitor"},
		{ResourceName: "test2", ResourceType: "monitor"},
		{ResourceName: "test3", ResourceType: "monitor"},
		{ResourceName: "test4", ResourceType: "monitor"},
		{ResourceName: "test5", ResourceType: "monitor"},
	}

	result := formatter.FormatComparisonList(comparisons, 3)

	if !strings.Contains(result, "test1") {
		t.Error("expected first comparison in output")
	}

	if !strings.Contains(result, "test2") {
		t.Error("expected second comparison in output")
	}

	if !strings.Contains(result, "test3") {
		t.Error("expected third comparison in output")
	}

	if strings.Contains(result, "test4") {
		t.Error("should not include fourth comparison when limit is 3")
	}

	if !strings.Contains(result, "(2 more resources)") {
		t.Error("expected truncation message")
	}
}

func TestGetActionIndicator(t *testing.T) {
	formatter := NewDiffFormatter(false)

	tests := []struct {
		action   string
		expected string
	}{
		{"preserved", "→"},
		{"mapped", "→"},
		{"rounded", "~"},
		{"defaulted", "+"},
		{"dropped", "✗"},
		{"unknown", "→"},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			result := formatter.getActionIndicator(tt.action)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestWrapText(t *testing.T) {
	formatter := NewDiffFormatter(false)

	tests := []struct {
		name     string
		text     string
		width    int
		expected int
	}{
		{
			name:     "short text",
			text:     "hello",
			width:    10,
			expected: 1,
		},
		{
			name:     "text at width",
			text:     "hello world",
			width:    11,
			expected: 1,
		},
		{
			name:     "text exceeds width",
			text:     "hello world this is a long text",
			width:    15,
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := formatter.wrapText(tt.text, tt.width)
			if len(lines) != tt.expected {
				t.Errorf("expected %d lines, got %d", tt.expected, len(lines))
			}
		})
	}
}
