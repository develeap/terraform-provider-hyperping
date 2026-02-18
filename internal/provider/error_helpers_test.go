// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"strings"
	"testing"
)

func TestNewCreateError(t *testing.T) {
	err := errors.New("invalid request")
	diag := newCreateError("Monitor", err)

	if !strings.Contains(diag.Summary(), "Failed to Create Monitor") {
		t.Errorf("Expected summary to contain 'Failed to Create Monitor', got: %s", diag.Summary())
	}

	if !strings.Contains(diag.Detail(), "invalid request") {
		t.Errorf("Expected detail to contain error message, got: %s", diag.Detail())
	}

	if !strings.Contains(diag.Detail(), "Troubleshooting") {
		t.Errorf("Expected detail to contain troubleshooting section, got: %s", diag.Detail())
	}
}

func TestNewReadError(t *testing.T) {
	err := errors.New("not found")
	diag := newReadError("StatusPage", "sp_123", err)

	if !strings.Contains(diag.Summary(), "Failed to Read StatusPage") {
		t.Errorf("Expected summary to contain 'Failed to Read StatusPage', got: %s", diag.Summary())
	}

	if !strings.Contains(diag.Detail(), "sp_123") {
		t.Errorf("Expected detail to contain resource ID, got: %s", diag.Detail())
	}

	if !strings.Contains(diag.Detail(), "not found") {
		t.Errorf("Expected detail to contain error message, got: %s", diag.Detail())
	}
}

func TestNewUpdateError(t *testing.T) {
	err := errors.New("validation failed")
	diag := newUpdateError("Incident", "inc_456", err)

	if !strings.Contains(diag.Summary(), "Failed to Update Incident") {
		t.Errorf("Expected summary to contain 'Failed to Update Incident', got: %s", diag.Summary())
	}

	if !strings.Contains(diag.Detail(), "inc_456") {
		t.Errorf("Expected detail to contain resource ID, got: %s", diag.Detail())
	}
}

func TestNewDeleteError(t *testing.T) {
	err := errors.New("resource has dependencies")
	diag := newDeleteError("Maintenance", "mw_789", err)

	if !strings.Contains(diag.Summary(), "Failed to Delete Maintenance") {
		t.Errorf("Expected summary to contain 'Failed to Delete Maintenance', got: %s", diag.Summary())
	}

	if !strings.Contains(diag.Detail(), "dependencies") {
		t.Errorf("Expected detail to contain troubleshooting hint about dependencies, got: %s", diag.Detail())
	}
}

func TestNewListError(t *testing.T) {
	err := errors.New("permission denied")
	diag := newListError("Monitors", err)

	if !strings.Contains(diag.Summary(), "Failed to List Monitors") {
		t.Errorf("Expected summary to contain 'Failed to List Monitors', got: %s", diag.Summary())
	}

	if !strings.Contains(diag.Detail(), "permission denied") {
		t.Errorf("Expected detail to contain error message, got: %s", diag.Detail())
	}
}

func TestNewConfigError(t *testing.T) {
	diag := newConfigError("Invalid frequency value")

	if !strings.Contains(diag.Summary(), "Configuration Error") {
		t.Errorf("Expected summary to contain 'Configuration Error', got: %s", diag.Summary())
	}

	if !strings.Contains(diag.Detail(), "Invalid frequency value") {
		t.Errorf("Expected detail to contain error message, got: %s", diag.Detail())
	}
}

func TestNewValidationError(t *testing.T) {
	diag := newValidationError("URL", "URL must be a valid HTTP/HTTPS endpoint")

	if !strings.Contains(diag.Summary(), "Invalid URL") {
		t.Errorf("Expected summary to contain 'Invalid URL', got: %s", diag.Summary())
	}

	if !strings.Contains(diag.Detail(), "valid HTTP/HTTPS endpoint") {
		t.Errorf("Expected detail to contain validation message, got: %s", diag.Detail())
	}
}

func TestNewImportError(t *testing.T) {
	err := errors.New("invalid format")
	diag := newImportError("Monitor", err)

	if !strings.Contains(diag.Summary(), "Import Failed") {
		t.Errorf("Expected summary to contain 'Import Failed', got: %s", diag.Summary())
	}

	if !strings.Contains(diag.Detail(), "invalid format") {
		t.Errorf("Expected detail to contain error message, got: %s", diag.Detail())
	}
}

func TestNewDeleteWarning(t *testing.T) {
	diag := newDeleteWarning("Monitor", "Resource not found in Hyperping")

	if !strings.Contains(diag.Summary(), "Monitor Not Found") {
		t.Errorf("Expected summary to contain 'Monitor Not Found', got: %s", diag.Summary())
	}

	if !strings.Contains(diag.Detail(), "already been deleted") {
		t.Errorf("Expected detail to contain standard message, got: %s", diag.Detail())
	}
}

func TestNewReadAfterCreateError(t *testing.T) {
	err := errors.New("resource not ready")
	diag := newReadAfterCreateError("Monitor", "mon_123", err)

	if !strings.Contains(diag.Summary(), "Created But Read Failed") {
		t.Errorf("Expected summary to contain 'Created But Read Failed', got: %s", diag.Summary())
	}

	if !strings.Contains(diag.Detail(), "mon_123") {
		t.Errorf("Expected detail to contain resource ID, got: %s", diag.Detail())
	}

	if !strings.Contains(diag.Detail(), "terraform import") {
		t.Errorf("Expected detail to contain import instructions, got: %s", diag.Detail())
	}
}

func TestNewUnexpectedConfigTypeError(t *testing.T) {
	diag := newUnexpectedConfigTypeError("*provider.hyperpingClient", "string")

	if !strings.Contains(diag.Summary(), "Unexpected Resource Configure Type") {
		t.Errorf("Expected summary to contain 'Unexpected Resource Configure Type', got: %s", diag.Summary())
	}

	if !strings.Contains(diag.Detail(), "*provider.hyperpingClient") {
		t.Errorf("Expected detail to contain expected type, got: %s", diag.Detail())
	}

	if !strings.Contains(diag.Detail(), "provider bug") {
		t.Errorf("Expected detail to mention provider bug, got: %s", diag.Detail())
	}
}
