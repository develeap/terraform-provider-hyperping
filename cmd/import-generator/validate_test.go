// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestValidate_AllValid(t *testing.T) {
	mock := &mockClient{
		monitors: []client.Monitor{
			{UUID: "mon_abc123"},
			{UUID: "mon_def456"},
		},
		healthchecks: []client.Healthcheck{
			{UUID: "hc_abc123"},
		},
	}

	gen := &Generator{
		client:    mock,
		resources: []string{"monitors", "healthchecks"},
	}

	result := gen.Validate(context.Background())

	if !result.IsValid() {
		t.Error("Expected validation to pass")
	}

	if result.Monitors.ValidCount != 2 {
		t.Errorf("Expected 2 valid monitors, got %d", result.Monitors.ValidCount)
	}

	if result.Healthchecks.ValidCount != 1 {
		t.Errorf("Expected 1 valid healthcheck, got %d", result.Healthchecks.ValidCount)
	}
}

func TestValidate_InvalidIDs(t *testing.T) {
	mock := &mockClient{
		monitors: []client.Monitor{
			{UUID: "mon_valid123"},
			{UUID: "invalid_id!"},
		},
	}

	gen := &Generator{
		client:    mock,
		resources: []string{"monitors"},
	}

	result := gen.Validate(context.Background())

	if result.IsValid() {
		t.Error("Expected validation to fail")
	}

	if result.Monitors.ValidCount != 1 {
		t.Errorf("Expected 1 valid monitor, got %d", result.Monitors.ValidCount)
	}

	if len(result.Monitors.InvalidIDs) != 1 {
		t.Errorf("Expected 1 invalid ID, got %d", len(result.Monitors.InvalidIDs))
	}

	if result.Monitors.InvalidIDs[0] != "invalid_id!" {
		t.Errorf("Expected invalid ID 'invalid_id!', got %s", result.Monitors.InvalidIDs[0])
	}
}

func TestValidate_FetchError(t *testing.T) {
	mock := &mockClient{
		monitorsErr: errors.New("API error"),
	}

	gen := &Generator{
		client:    mock,
		resources: []string{"monitors"},
	}

	result := gen.Validate(context.Background())

	if result.IsValid() {
		t.Error("Expected validation to fail")
	}

	if result.Monitors.FetchError == nil {
		t.Error("Expected fetch error to be recorded")
	}
}

func TestValidationResult_Print(t *testing.T) {
	result := &ValidationResult{
		Monitors: ValidationResourceResult{
			ResourceType: "Monitors",
			ValidCount:   2,
			InvalidIDs:   []string{"invalid_id"},
		},
		Healthchecks: ValidationResourceResult{
			ResourceType: "Healthchecks",
			ValidCount:   1,
		},
	}

	var buf bytes.Buffer
	result.Print(&buf)

	output := buf.String()

	if output == "" {
		t.Error("Expected non-empty output")
	}

	// Check for expected markers
	if !bytes.Contains(buf.Bytes(), []byte("✗ Monitors")) {
		t.Error("Expected monitors with error marker")
	}

	if !bytes.Contains(buf.Bytes(), []byte("✓ Healthchecks")) {
		t.Error("Expected healthchecks with success marker")
	}

	if !bytes.Contains(buf.Bytes(), []byte("invalid_id")) {
		t.Error("Expected invalid ID in output")
	}

	if !bytes.Contains(buf.Bytes(), []byte("Validation failed")) {
		t.Error("Expected validation failed message")
	}
}

func TestValidate_AllResourceTypes(t *testing.T) {
	mock := &mockClient{
		monitors:     []client.Monitor{{UUID: "mon_123"}},
		healthchecks: []client.Healthcheck{{UUID: "hc_123"}},
		statusPages:  []client.StatusPage{{UUID: "sp_123"}},
		incidents:    []client.Incident{{UUID: "inc_123"}},
		maintenance:  []client.Maintenance{{UUID: "maint_123"}},
		outages:      []client.Outage{{UUID: "outage_123"}},
	}

	gen := &Generator{
		client:    mock,
		resources: []string{"monitors", "healthchecks", "statuspages", "incidents", "maintenance", "outages"},
	}

	result := gen.Validate(context.Background())

	if !result.IsValid() {
		t.Error("Expected validation to pass")
	}

	if result.Monitors.ValidCount != 1 {
		t.Errorf("Expected 1 valid monitor, got %d", result.Monitors.ValidCount)
	}
	if result.Healthchecks.ValidCount != 1 {
		t.Errorf("Expected 1 valid healthcheck, got %d", result.Healthchecks.ValidCount)
	}
	if result.StatusPages.ValidCount != 1 {
		t.Errorf("Expected 1 valid status page, got %d", result.StatusPages.ValidCount)
	}
	if result.Incidents.ValidCount != 1 {
		t.Errorf("Expected 1 valid incident, got %d", result.Incidents.ValidCount)
	}
	if result.Maintenance.ValidCount != 1 {
		t.Errorf("Expected 1 valid maintenance, got %d", result.Maintenance.ValidCount)
	}
	if result.Outages.ValidCount != 1 {
		t.Errorf("Expected 1 valid outage, got %d", result.Outages.ValidCount)
	}
}

func TestValidate_InvalidIDsAllTypes(t *testing.T) {
	mock := &mockClient{
		statusPages: []client.StatusPage{{UUID: "invalid_sp"}},
		incidents:   []client.Incident{{UUID: "invalid_inc"}},
		maintenance: []client.Maintenance{{UUID: "invalid_maint"}},
		outages:     []client.Outage{{UUID: "invalid_outage"}},
	}

	gen := &Generator{
		client:    mock,
		resources: []string{"statuspages", "incidents", "maintenance", "outages"},
	}

	result := gen.Validate(context.Background())

	if result.IsValid() {
		t.Error("Expected validation to fail")
	}

	if len(result.StatusPages.InvalidIDs) != 1 {
		t.Errorf("Expected 1 invalid status page, got %d", len(result.StatusPages.InvalidIDs))
	}
	if len(result.Incidents.InvalidIDs) != 1 {
		t.Errorf("Expected 1 invalid incident, got %d", len(result.Incidents.InvalidIDs))
	}
	if len(result.Maintenance.InvalidIDs) != 1 {
		t.Errorf("Expected 1 invalid maintenance, got %d", len(result.Maintenance.InvalidIDs))
	}
	if len(result.Outages.InvalidIDs) != 1 {
		t.Errorf("Expected 1 invalid outage, got %d", len(result.Outages.InvalidIDs))
	}
}

func TestValidationResult_ErrorCount(t *testing.T) {
	tests := []struct {
		name     string
		result   ValidationResult
		expected int
	}{
		{
			name: "no errors",
			result: ValidationResult{
				Monitors: ValidationResourceResult{
					ResourceType: "Monitors",
					ValidCount:   2,
				},
			},
			expected: 0,
		},
		{
			name: "one invalid ID",
			result: ValidationResult{
				Monitors: ValidationResourceResult{
					ResourceType: "Monitors",
					ValidCount:   1,
					InvalidIDs:   []string{"invalid"},
				},
			},
			expected: 1,
		},
		{
			name: "one fetch error",
			result: ValidationResult{
				Monitors: ValidationResourceResult{
					ResourceType: "Monitors",
					FetchError:   errors.New("error"),
				},
			},
			expected: 1,
		},
		{
			name: "multiple resource errors",
			result: ValidationResult{
				Monitors: ValidationResourceResult{
					ResourceType: "Monitors",
					InvalidIDs:   []string{"invalid"},
				},
				Healthchecks: ValidationResourceResult{
					ResourceType: "Healthchecks",
					FetchError:   errors.New("error"),
				},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := tt.result.ErrorCount()
			if count != tt.expected {
				t.Errorf("Expected %d errors, got %d", tt.expected, count)
			}
		})
	}
}
