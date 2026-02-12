// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"errors"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestFetchResources_ContinueOnError_Monitors(t *testing.T) {
	mock := &mockClient{
		monitorsErr: errors.New("API error"),
		healthchecks: []client.Healthcheck{
			{UUID: "hc_123", Name: "Healthcheck"},
		},
	}

	gen := &Generator{
		client:          mock,
		resources:       []string{"monitors", "healthchecks"},
		continueOnError: true,
	}

	data, err := gen.fetchResources(context.Background())
	if err != nil {
		t.Fatalf("fetchResources failed: %v", err)
	}

	// Monitors should be empty due to error
	if len(data.Monitors) != 0 {
		t.Errorf("Expected no monitors, got %d", len(data.Monitors))
	}

	// Healthchecks should still be fetched
	if len(data.Healthchecks) != 1 {
		t.Errorf("Expected 1 healthcheck, got %d", len(data.Healthchecks))
	}
}

func TestFetchResources_ContinueOnError_MultipleErrors(t *testing.T) {
	mock := &mockClient{
		monitorsErr:     errors.New("monitors error"),
		healthchecksErr: errors.New("healthchecks error"),
		statusPages: []client.StatusPage{
			{UUID: "sp_123", Name: "Status Page"},
		},
	}

	gen := &Generator{
		client:          mock,
		resources:       []string{"monitors", "healthchecks", "statuspages"},
		continueOnError: true,
	}

	data, err := gen.fetchResources(context.Background())
	if err != nil {
		t.Fatalf("fetchResources failed: %v", err)
	}

	// Failed resources should be empty
	if len(data.Monitors) != 0 {
		t.Errorf("Expected no monitors, got %d", len(data.Monitors))
	}

	if len(data.Healthchecks) != 0 {
		t.Errorf("Expected no healthchecks, got %d", len(data.Healthchecks))
	}

	// Successful resource should be fetched
	if len(data.StatusPages) != 1 {
		t.Errorf("Expected 1 status page, got %d", len(data.StatusPages))
	}
}

func TestFetchResources_FailOnError_Default(t *testing.T) {
	mock := &mockClient{
		monitorsErr: errors.New("API error"),
		healthchecks: []client.Healthcheck{
			{UUID: "hc_123", Name: "Healthcheck"},
		},
	}

	gen := &Generator{
		client:          mock,
		resources:       []string{"monitors", "healthchecks"},
		continueOnError: false, // default behavior
	}

	_, err := gen.fetchResources(context.Background())
	if err == nil {
		t.Error("Expected error when continueOnError is false")
	}

	if !errors.Is(err, mock.monitorsErr) {
		expectedMsg := "fetching monitors"
		if err.Error() == "" || !contains(err.Error(), expectedMsg) {
			t.Errorf("Expected error message to contain %q, got: %v", expectedMsg, err)
		}
	}
}

func TestFetchResources_ContinueOnError_AllResourceTypes(t *testing.T) {
	mock := &mockClient{
		monitorsErr:    errors.New("monitors error"),
		healthchecks:   []client.Healthcheck{{UUID: "hc_1", Name: "HC1"}},
		statusPagesErr: errors.New("status pages error"),
		incidents:      []client.Incident{{UUID: "inc_1", Title: client.LocalizedText{En: "Inc1"}}},
		maintenanceErr: errors.New("maintenance error"),
		outages:        []client.Outage{{UUID: "outage_1", Monitor: client.MonitorReference{Name: "Mon1"}}},
	}

	gen := &Generator{
		client:          mock,
		resources:       []string{"monitors", "healthchecks", "statuspages", "incidents", "maintenance", "outages"},
		continueOnError: true,
	}

	data, err := gen.fetchResources(context.Background())
	if err != nil {
		t.Fatalf("fetchResources failed: %v", err)
	}

	// Failed resources
	if len(data.Monitors) != 0 {
		t.Errorf("Expected no monitors, got %d", len(data.Monitors))
	}
	if len(data.StatusPages) != 0 {
		t.Errorf("Expected no status pages, got %d", len(data.StatusPages))
	}
	if len(data.Maintenance) != 0 {
		t.Errorf("Expected no maintenance, got %d", len(data.Maintenance))
	}

	// Successful resources
	if len(data.Healthchecks) != 1 {
		t.Errorf("Expected 1 healthcheck, got %d", len(data.Healthchecks))
	}
	if len(data.Incidents) != 1 {
		t.Errorf("Expected 1 incident, got %d", len(data.Incidents))
	}
	if len(data.Outages) != 1 {
		t.Errorf("Expected 1 outage, got %d", len(data.Outages))
	}
}

func TestGenerate_ContinueOnError_ProducesPartialOutput(t *testing.T) {
	mock := &mockClient{
		monitorsErr: errors.New("monitors error"),
		healthchecks: []client.Healthcheck{
			{UUID: "hc_123", Name: "Healthcheck"},
		},
	}

	gen := &Generator{
		client:          mock,
		resources:       []string{"monitors", "healthchecks"},
		continueOnError: true,
	}

	output, err := gen.Generate(context.Background(), "import")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Should contain healthcheck import
	if !contains(output, "hyperping_healthcheck.healthcheck") {
		t.Error("Expected output to contain healthcheck import")
	}

	if !contains(output, "hc_123") {
		t.Error("Expected output to contain healthcheck UUID")
	}

	// Should not contain monitor import
	if contains(output, "hyperping_monitor") {
		t.Error("Expected output to not contain monitor import")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
