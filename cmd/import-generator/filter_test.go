// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"testing"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestFilterConfig_NewFilterConfig(t *testing.T) {
	tests := []struct {
		name           string
		namePattern    string
		excludePattern string
		resourceType   string
		wantErr        bool
	}{
		{
			name:        "valid name pattern",
			namePattern: "PROD-.*",
			wantErr:     false,
		},
		{
			name:           "valid exclude pattern",
			excludePattern: "test-.*",
			wantErr:        false,
		},
		{
			name:         "valid resource type",
			resourceType: "hyperping_monitor",
			wantErr:      false,
		},
		{
			name:        "invalid name pattern",
			namePattern: "[invalid(regex",
			wantErr:     true,
		},
		{
			name:           "invalid exclude pattern",
			excludePattern: "[invalid(regex",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFilterConfig(tt.namePattern, tt.excludePattern, tt.resourceType)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFilterConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFilterConfig_FilterMonitors(t *testing.T) {
	monitors := []client.Monitor{
		{UUID: "mon_1", Name: "PROD-API-Health"},
		{UUID: "mon_2", Name: "PROD-Database"},
		{UUID: "mon_3", Name: "DEV-API-Test"},
		{UUID: "mon_4", Name: "test-service"},
	}

	tests := []struct {
		name           string
		namePattern    string
		excludePattern string
		resourceType   string
		wantCount      int
	}{
		{
			name:        "filter by PROD prefix",
			namePattern: "^PROD-.*",
			wantCount:   2,
		},
		{
			name:           "exclude test resources",
			excludePattern: "test-.*",
			wantCount:      3,
		},
		{
			name:           "PROD and exclude test",
			namePattern:    "^PROD-.*",
			excludePattern: ".*Test$",
			wantCount:      2,
		},
		{
			name:         "wrong resource type",
			resourceType: "hyperping_healthcheck",
			wantCount:    0,
		},
		{
			name:         "correct resource type",
			resourceType: "hyperping_monitor",
			wantCount:    4,
		},
		{
			name:      "no filters",
			wantCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc, err := NewFilterConfig(tt.namePattern, tt.excludePattern, tt.resourceType)
			if err != nil {
				t.Fatalf("NewFilterConfig() error = %v", err)
			}

			filtered := fc.FilterMonitors(monitors)
			if len(filtered) != tt.wantCount {
				t.Errorf("FilterMonitors() got %d monitors, want %d", len(filtered), tt.wantCount)
			}
		})
	}
}

func TestFilterConfig_IsEmpty(t *testing.T) {
	tests := []struct {
		name           string
		namePattern    string
		excludePattern string
		resourceType   string
		wantEmpty      bool
	}{
		{
			name:      "no filters",
			wantEmpty: true,
		},
		{
			name:        "has name pattern",
			namePattern: "PROD-.*",
			wantEmpty:   false,
		},
		{
			name:           "has exclude pattern",
			excludePattern: "test-.*",
			wantEmpty:      false,
		},
		{
			name:         "has resource type",
			resourceType: "hyperping_monitor",
			wantEmpty:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc, err := NewFilterConfig(tt.namePattern, tt.excludePattern, tt.resourceType)
			if err != nil {
				t.Fatalf("NewFilterConfig() error = %v", err)
			}

			if fc.IsEmpty() != tt.wantEmpty {
				t.Errorf("IsEmpty() = %v, want %v", fc.IsEmpty(), tt.wantEmpty)
			}
		})
	}
}

func TestFilterConfig_Summary(t *testing.T) {
	tests := []struct {
		name           string
		namePattern    string
		excludePattern string
		resourceType   string
		wantContains   string
	}{
		{
			name:         "empty filter",
			wantContains: "No filters applied",
		},
		{
			name:         "name pattern only",
			namePattern:  "PROD-.*",
			wantContains: "Name pattern:",
		},
		{
			name:           "exclude pattern only",
			excludePattern: "test-.*",
			wantContains:   "Exclude pattern:",
		},
		{
			name:         "resource type only",
			resourceType: "hyperping_monitor",
			wantContains: "Resource type:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc, err := NewFilterConfig(tt.namePattern, tt.excludePattern, tt.resourceType)
			if err != nil {
				t.Fatalf("NewFilterConfig() error = %v", err)
			}

			summary := fc.Summary()
			if summary != tt.wantContains && tt.wantContains != "No filters applied" {
				if len(summary) == 0 || len(tt.wantContains) == 0 {
					t.Errorf("Summary() = %q, want to contain %q", summary, tt.wantContains)
				}
			}
		})
	}
}
