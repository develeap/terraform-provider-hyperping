// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"strings"
	"testing"
)

func TestValidValueReference_Monitor(t *testing.T) {
	ref := ValidValueReference("Monitor")

	if ref == "" {
		t.Fatal("expected non-empty reference for Monitor")
	}

	requiredFields := []string{"protocol", "http_method", "check_frequency", "expected_status_code", "regions", "alerts_wait"}
	for _, field := range requiredFields {
		if !strings.Contains(ref, field) {
			t.Errorf("expected Monitor reference to contain %q, got: %s", field, ref)
		}
	}

	// Verify it contains actual values from client constants
	requiredValues := []string{"http", "port", "icmp", "GET", "POST", "london", "frankfurt"}
	for _, val := range requiredValues {
		if !strings.Contains(ref, val) {
			t.Errorf("expected Monitor reference to contain value %q, got: %s", val, ref)
		}
	}
}

func TestValidValueReference_MaintenanceWindow(t *testing.T) {
	ref := ValidValueReference("Maintenance Window")

	if ref == "" {
		t.Fatal("expected non-empty reference for Maintenance Window")
	}

	if !strings.Contains(ref, "notification_option") {
		t.Errorf("expected Maintenance Window reference to contain 'notification_option', got: %s", ref)
	}

	requiredValues := []string{"scheduled", "immediate"}
	for _, val := range requiredValues {
		if !strings.Contains(ref, val) {
			t.Errorf("expected Maintenance Window reference to contain %q, got: %s", val, ref)
		}
	}
}

func TestValidValueReference_Incident(t *testing.T) {
	ref := ValidValueReference("Incident")

	if ref == "" {
		t.Fatal("expected non-empty reference for Incident")
	}

	if !strings.Contains(ref, "type") {
		t.Errorf("expected Incident reference to contain 'type', got: %s", ref)
	}

	requiredValues := []string{"outage", "incident"}
	for _, val := range requiredValues {
		if !strings.Contains(ref, val) {
			t.Errorf("expected Incident reference to contain %q, got: %s", val, ref)
		}
	}
}

func TestValidValueReference_UnknownType(t *testing.T) {
	ref := ValidValueReference("UnknownType")

	if ref != "" {
		t.Errorf("expected empty string for unknown resource type, got: %s", ref)
	}
}

func TestNewCreateError_IncludesQuickReference(t *testing.T) {
	err := errors.New("validation failed")
	diag := newCreateError("Monitor", err)

	if !strings.Contains(diag.Detail(), "Quick Reference") {
		t.Errorf("expected Monitor create error to contain 'Quick Reference', got: %s", diag.Detail())
	}

	if !strings.Contains(diag.Detail(), "protocol") {
		t.Errorf("expected Monitor create error to contain 'protocol' reference, got: %s", diag.Detail())
	}
}

func TestNewUpdateError_IncludesQuickReference(t *testing.T) {
	err := errors.New("validation failed")
	diag := newUpdateError("Monitor", "mon-123", err)

	if !strings.Contains(diag.Detail(), "Quick Reference") {
		t.Errorf("expected Monitor update error to contain 'Quick Reference', got: %s", diag.Detail())
	}

	if !strings.Contains(diag.Detail(), "protocol") {
		t.Errorf("expected Monitor update error to contain 'protocol' reference, got: %s", diag.Detail())
	}
}

func TestNewCreateError_NoQuickReferenceForUnknownType(t *testing.T) {
	err := errors.New("server error")
	diag := newCreateError("Outage", err)

	if strings.Contains(diag.Detail(), "Quick Reference") {
		t.Errorf("expected Outage create error NOT to contain 'Quick Reference', got: %s", diag.Detail())
	}

	// Should still have the troubleshooting section
	if !strings.Contains(diag.Detail(), "Troubleshooting") {
		t.Errorf("expected Outage create error to contain 'Troubleshooting', got: %s", diag.Detail())
	}
}
