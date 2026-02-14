// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package dryrun

import (
	"testing"
)

func TestAnalyzeCompatibility(t *testing.T) {
	analyzer := NewCompatibilityAnalyzer()

	tests := []struct {
		name               string
		comparisons        []ResourceComparison
		warnings           []Warning
		expectedScore      float64
		expectedComplexity string
	}{
		{
			name:               "empty comparisons",
			comparisons:        []ResourceComparison{},
			warnings:           []Warning{},
			expectedScore:      100.0,
			expectedComplexity: "Simple",
		},
		{
			name: "all clean",
			comparisons: []ResourceComparison{
				{ResourceName: "test1", ResourceType: "monitor", HasWarnings: false, HasErrors: false},
				{ResourceName: "test2", ResourceType: "monitor", HasWarnings: false, HasErrors: false},
				{ResourceName: "test3", ResourceType: "monitor", HasWarnings: false, HasErrors: false},
			},
			warnings:           []Warning{},
			expectedScore:      100.0,
			expectedComplexity: "Simple",
		},
		{
			name: "some warnings",
			comparisons: []ResourceComparison{
				{ResourceName: "test1", ResourceType: "monitor", HasWarnings: false, HasErrors: false},
				{ResourceName: "test2", ResourceType: "monitor", HasWarnings: true, HasErrors: false},
				{ResourceName: "test3", ResourceType: "monitor", HasWarnings: false, HasErrors: false},
				{ResourceName: "test4", ResourceType: "monitor", HasWarnings: false, HasErrors: false},
			},
			warnings:           []Warning{},
			expectedScore:      75.0,
			expectedComplexity: "Simple",
		},
		{
			name: "many warnings",
			comparisons: []ResourceComparison{
				{ResourceName: "test1", ResourceType: "monitor", HasWarnings: false, HasErrors: false},
				{ResourceName: "test2", ResourceType: "monitor", HasWarnings: true, HasErrors: false},
				{ResourceName: "test3", ResourceType: "monitor", HasWarnings: true, HasErrors: false},
				{ResourceName: "test4", ResourceType: "monitor", HasWarnings: true, HasErrors: false},
			},
			warnings:           []Warning{},
			expectedScore:      25.0,
			expectedComplexity: "Complex",
		},
		{
			name: "with errors",
			comparisons: []ResourceComparison{
				{ResourceName: "test1", ResourceType: "monitor", HasWarnings: false, HasErrors: false},
				{ResourceName: "test2", ResourceType: "monitor", HasWarnings: false, HasErrors: true},
				{ResourceName: "test3", ResourceType: "monitor", HasWarnings: false, HasErrors: false},
				{ResourceName: "test4", ResourceType: "monitor", HasWarnings: false, HasErrors: false},
			},
			warnings:           []Warning{},
			expectedScore:      75.0,
			expectedComplexity: "Complex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.AnalyzeCompatibility(tt.comparisons, tt.warnings)

			if result.OverallScore != tt.expectedScore {
				t.Errorf("expected score %.1f, got %.1f", tt.expectedScore, result.OverallScore)
			}

			if result.Complexity != tt.expectedComplexity {
				t.Errorf("expected complexity %s, got %s", tt.expectedComplexity, result.Complexity)
			}
		})
	}
}

func TestCategorizeWarnings(t *testing.T) {
	analyzer := NewCompatibilityAnalyzer()

	warnings := []Warning{
		{Severity: "critical", Message: "Critical issue 1"},
		{Severity: "critical", Message: "Critical issue 2"},
		{Severity: "warning", Message: "Warning 1"},
		{Severity: "warning", Message: "Warning 2"},
		{Severity: "warning", Message: "Warning 3"},
		{Severity: "info", Message: "Info 1"},
	}

	categorized := analyzer.CategorizeWarnings(warnings)

	if len(categorized["critical"]) != 2 {
		t.Errorf("expected 2 critical warnings, got %d", len(categorized["critical"]))
	}

	if len(categorized["warning"]) != 3 {
		t.Errorf("expected 3 warnings, got %d", len(categorized["warning"]))
	}

	if len(categorized["info"]) != 1 {
		t.Errorf("expected 1 info, got %d", len(categorized["info"]))
	}
}

func TestEstimateManualEffort(t *testing.T) {
	analyzer := NewCompatibilityAnalyzer()

	tests := []struct {
		name            string
		warnings        []Warning
		expectedMinutes int
	}{
		{
			name:            "no warnings",
			warnings:        []Warning{},
			expectedMinutes: 0,
		},
		{
			name: "only info",
			warnings: []Warning{
				{Severity: "info"},
				{Severity: "info"},
			},
			expectedMinutes: 2, // 2 * 2 / 2
		},
		{
			name: "mixed warnings",
			warnings: []Warning{
				{Severity: "critical"},
				{Severity: "warning"},
				{Severity: "warning"},
				{Severity: "info"},
			},
			expectedMinutes: 11, // (10 + 5 + 5 + 2) / 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minutes, _ := analyzer.EstimateManualEffort(tt.warnings)
			if minutes != tt.expectedMinutes {
				t.Errorf("expected %d minutes, got %d", tt.expectedMinutes, minutes)
			}
		})
	}
}
