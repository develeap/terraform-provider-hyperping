// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"strings"
	"testing"
)

// TestValidateResourceID_Integration verifies that each client method
// returns a validation error for malicious IDs, covering the early-return paths.

func TestValidateResourceID_Monitors(t *testing.T) {
	c := NewClient("test_key", WithMaxRetries(0))
	ctx := context.Background()
	badID := "../../evil"

	t.Run("GetMonitor", func(t *testing.T) {
		_, err := c.GetMonitor(ctx, badID)
		if err == nil || !strings.Contains(err.Error(), "path traversal") {
			t.Errorf("expected validation error, got: %v", err)
		}
	})
	t.Run("UpdateMonitor", func(t *testing.T) {
		_, err := c.UpdateMonitor(ctx, badID, UpdateMonitorRequest{})
		if err == nil || !strings.Contains(err.Error(), "path traversal") {
			t.Errorf("expected validation error, got: %v", err)
		}
	})
	t.Run("DeleteMonitor", func(t *testing.T) {
		err := c.DeleteMonitor(ctx, badID)
		if err == nil || !strings.Contains(err.Error(), "path traversal") {
			t.Errorf("expected validation error, got: %v", err)
		}
	})
}

func TestValidateResourceID_Outages(t *testing.T) {
	c := NewClient("test_key", WithMaxRetries(0))
	ctx := context.Background()
	badID := "id?admin=true"

	t.Run("GetOutage", func(t *testing.T) {
		_, err := c.GetOutage(ctx, badID)
		if err == nil || !strings.Contains(err.Error(), "URL metacharacters") {
			t.Errorf("expected validation error, got: %v", err)
		}
	})
	t.Run("AcknowledgeOutage", func(t *testing.T) {
		_, err := c.AcknowledgeOutage(ctx, badID)
		if err == nil || !strings.Contains(err.Error(), "URL metacharacters") {
			t.Errorf("expected validation error, got: %v", err)
		}
	})
	t.Run("UnacknowledgeOutage", func(t *testing.T) {
		_, err := c.UnacknowledgeOutage(ctx, badID)
		if err == nil || !strings.Contains(err.Error(), "URL metacharacters") {
			t.Errorf("expected validation error, got: %v", err)
		}
	})
	t.Run("ResolveOutage", func(t *testing.T) {
		_, err := c.ResolveOutage(ctx, badID)
		if err == nil || !strings.Contains(err.Error(), "URL metacharacters") {
			t.Errorf("expected validation error, got: %v", err)
		}
	})
	t.Run("EscalateOutage", func(t *testing.T) {
		_, err := c.EscalateOutage(ctx, badID)
		if err == nil || !strings.Contains(err.Error(), "URL metacharacters") {
			t.Errorf("expected validation error, got: %v", err)
		}
	})
	t.Run("DeleteOutage", func(t *testing.T) {
		err := c.DeleteOutage(ctx, badID)
		if err == nil || !strings.Contains(err.Error(), "URL metacharacters") {
			t.Errorf("expected validation error, got: %v", err)
		}
	})
}

func TestValidateResourceID_Incidents(t *testing.T) {
	c := NewClient("test_key", WithMaxRetries(0))
	ctx := context.Background()
	badID := ""

	t.Run("GetIncident", func(t *testing.T) {
		_, err := c.GetIncident(ctx, badID)
		if err == nil || !strings.Contains(err.Error(), "must not be empty") {
			t.Errorf("expected validation error, got: %v", err)
		}
	})
	t.Run("UpdateIncident", func(t *testing.T) {
		_, err := c.UpdateIncident(ctx, badID, UpdateIncidentRequest{})
		if err == nil || !strings.Contains(err.Error(), "must not be empty") {
			t.Errorf("expected validation error, got: %v", err)
		}
	})
	t.Run("AddIncidentUpdate", func(t *testing.T) {
		_, err := c.AddIncidentUpdate(ctx, badID, AddIncidentUpdateRequest{})
		if err == nil || !strings.Contains(err.Error(), "must not be empty") {
			t.Errorf("expected validation error, got: %v", err)
		}
	})
	t.Run("DeleteIncident", func(t *testing.T) {
		err := c.DeleteIncident(ctx, badID)
		if err == nil || !strings.Contains(err.Error(), "must not be empty") {
			t.Errorf("expected validation error, got: %v", err)
		}
	})
}

func TestValidateResourceID_Maintenance(t *testing.T) {
	c := NewClient("test_key", WithMaxRetries(0))
	ctx := context.Background()
	badID := "../../secrets"

	t.Run("GetMaintenance", func(t *testing.T) {
		_, err := c.GetMaintenance(ctx, badID)
		if err == nil || !strings.Contains(err.Error(), "path traversal") {
			t.Errorf("expected validation error, got: %v", err)
		}
	})
	t.Run("UpdateMaintenance", func(t *testing.T) {
		_, err := c.UpdateMaintenance(ctx, badID, UpdateMaintenanceRequest{})
		if err == nil || !strings.Contains(err.Error(), "path traversal") {
			t.Errorf("expected validation error, got: %v", err)
		}
	})
	t.Run("DeleteMaintenance", func(t *testing.T) {
		err := c.DeleteMaintenance(ctx, badID)
		if err == nil || !strings.Contains(err.Error(), "path traversal") {
			t.Errorf("expected validation error, got: %v", err)
		}
	})
}

func TestValidateResourceID_Healthchecks(t *testing.T) {
	c := NewClient("test_key", WithMaxRetries(0))
	ctx := context.Background()
	badID := "id@evil.com"

	t.Run("GetHealthcheck", func(t *testing.T) {
		_, err := c.GetHealthcheck(ctx, badID)
		if err == nil || !strings.Contains(err.Error(), "URL metacharacters") {
			t.Errorf("expected validation error, got: %v", err)
		}
	})
	t.Run("UpdateHealthcheck", func(t *testing.T) {
		_, err := c.UpdateHealthcheck(ctx, badID, UpdateHealthcheckRequest{})
		if err == nil || !strings.Contains(err.Error(), "URL metacharacters") {
			t.Errorf("expected validation error, got: %v", err)
		}
	})
	t.Run("DeleteHealthcheck", func(t *testing.T) {
		err := c.DeleteHealthcheck(ctx, badID)
		if err == nil || !strings.Contains(err.Error(), "URL metacharacters") {
			t.Errorf("expected validation error, got: %v", err)
		}
	})
	t.Run("PauseHealthcheck", func(t *testing.T) {
		_, err := c.PauseHealthcheck(ctx, badID)
		if err == nil || !strings.Contains(err.Error(), "URL metacharacters") {
			t.Errorf("expected validation error, got: %v", err)
		}
	})
	t.Run("ResumeHealthcheck", func(t *testing.T) {
		_, err := c.ResumeHealthcheck(ctx, badID)
		if err == nil || !strings.Contains(err.Error(), "URL metacharacters") {
			t.Errorf("expected validation error, got: %v", err)
		}
	})
}

func TestValidateResourceID_Reports(t *testing.T) {
	c := NewClient("test_key", WithMaxRetries(0))
	ctx := context.Background()
	badID := "../../admin"

	t.Run("GetMonitorReport", func(t *testing.T) {
		_, err := c.GetMonitorReport(ctx, badID, "", "")
		if err == nil || !strings.Contains(err.Error(), "path traversal") {
			t.Errorf("expected validation error, got: %v", err)
		}
	})
}

func TestValidateCreateRequests(t *testing.T) {
	t.Run("CreateMonitor with invalid input", func(t *testing.T) {
		c := NewClient("test_key", WithMaxRetries(0))
		req := CreateMonitorRequest{Name: strings.Repeat("x", 256), URL: "https://example.com"}
		_, err := c.CreateMonitor(context.Background(), req)
		if err == nil || !strings.Contains(err.Error(), "exceeds maximum length") {
			t.Errorf("expected length validation error, got: %v", err)
		}
	})
	t.Run("CreateIncident with invalid input", func(t *testing.T) {
		c := NewClient("test_key", WithMaxRetries(0))
		req := CreateIncidentRequest{
			Title: LocalizedText{En: strings.Repeat("x", 256)},
			Text:  LocalizedText{En: "ok"},
		}
		_, err := c.CreateIncident(context.Background(), req)
		if err == nil || !strings.Contains(err.Error(), "exceeds maximum length") {
			t.Errorf("expected length validation error, got: %v", err)
		}
	})
	t.Run("CreateMaintenance with invalid input", func(t *testing.T) {
		c := NewClient("test_key", WithMaxRetries(0))
		req := CreateMaintenanceRequest{Name: strings.Repeat("x", 256)}
		_, err := c.CreateMaintenance(context.Background(), req)
		if err == nil || !strings.Contains(err.Error(), "exceeds maximum length") {
			t.Errorf("expected length validation error, got: %v", err)
		}
	})
	t.Run("CreateOutage with invalid input", func(t *testing.T) {
		c := NewClient("test_key", WithMaxRetries(0))
		req := CreateOutageRequest{Description: strings.Repeat("x", 10001)}
		_, err := c.CreateOutage(context.Background(), req)
		if err == nil || !strings.Contains(err.Error(), "exceeds maximum length") {
			t.Errorf("expected length validation error, got: %v", err)
		}
	})
	t.Run("CreateHealthcheck with invalid input", func(t *testing.T) {
		c := NewClient("test_key", WithMaxRetries(0))
		req := CreateHealthcheckRequest{Name: strings.Repeat("x", 256)}
		_, err := c.CreateHealthcheck(context.Background(), req)
		if err == nil || !strings.Contains(err.Error(), "exceeds maximum length") {
			t.Errorf("expected length validation error, got: %v", err)
		}
	})
}
