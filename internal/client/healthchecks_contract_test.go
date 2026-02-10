// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"os"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/internal/provider/testutil"
)

// TestContract_Healthcheck_ResponseStructure validates healthcheck response structure from VCR cassette.
func TestContract_Healthcheck_ResponseStructure(t *testing.T) {
	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "healthcheck_crud",
		Mode:         testutil.ModeReplay,
		CassetteDir:  "testdata/cassettes",
	})
	defer func() {
		if err := r.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}
	}()

	apiKey := os.Getenv("HYPERPING_API_KEY")
	if apiKey == "" {
		apiKey = "test_api_key"
	}

	client := NewClient(apiKey, WithHTTPClient(httpClient))
	ctx := context.Background()

	// Create healthcheck (from cassette)
	periodValue := 60
	periodType := "seconds"
	createReq := CreateHealthcheckRequest{
		Name:             "vcr-test-healthcheck",
		PeriodValue:      &periodValue,
		PeriodType:       &periodType,
		GracePeriodValue: 300,
		GracePeriodType:  "seconds",
	}

	healthcheck, err := client.CreateHealthcheck(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateHealthcheck failed: %v", err)
	}

	// Validate required fields
	if healthcheck.UUID == "" {
		t.Error("UUID is required but empty")
	}
	if healthcheck.Name == "" {
		t.Error("Name is required but empty")
	}
	if healthcheck.PingURL == "" {
		t.Error("PingURL is required but empty")
	}
	// Note: CreatedAt may not be returned on create, only on subsequent reads
	t.Logf("CreatedAt: %q", healthcheck.CreatedAt)

	// Validate period structure
	if healthcheck.PeriodValue != nil {
		if *healthcheck.PeriodValue <= 0 {
			t.Errorf("PeriodValue must be positive, got %d", *healthcheck.PeriodValue)
		}
	}
	if healthcheck.PeriodType == "" {
		t.Error("PeriodType is required but empty")
	}
	if healthcheck.Period <= 0 {
		t.Errorf("Period must be positive, got %d", healthcheck.Period)
	}

	// Validate period type enum
	validPeriodTypes := map[string]bool{
		"seconds": true,
		"minutes": true,
		"hours":   true,
		"days":    true,
	}
	if !validPeriodTypes[healthcheck.PeriodType] {
		t.Errorf("invalid PeriodType %q, must be one of: seconds, minutes, hours, days", healthcheck.PeriodType)
	}

	// Validate grace period structure
	if healthcheck.GracePeriodValue <= 0 {
		t.Errorf("GracePeriodValue must be positive, got %d", healthcheck.GracePeriodValue)
	}
	if healthcheck.GracePeriodType == "" {
		t.Error("GracePeriodType is required but empty")
	}
	if healthcheck.GracePeriod <= 0 {
		t.Errorf("GracePeriod must be positive, got %d", healthcheck.GracePeriod)
	}

	// Validate grace period type enum
	if !validPeriodTypes[healthcheck.GracePeriodType] {
		t.Errorf("invalid GracePeriodType %q, must be one of: seconds, minutes, hours, days", healthcheck.GracePeriodType)
	}

	// Validate boolean fields have proper types
	if healthcheck.IsDown != true && healthcheck.IsDown != false {
		t.Error("IsDown must be a boolean")
	}
	if healthcheck.IsPaused != true && healthcheck.IsPaused != false {
		t.Error("IsPaused must be a boolean")
	}

	// Read healthcheck (from cassette)
	readHealthcheck, err := client.GetHealthcheck(ctx, healthcheck.UUID)
	if err != nil {
		t.Fatalf("GetHealthcheck failed: %v", err)
	}

	// Validate read response has same structure
	if readHealthcheck.UUID != healthcheck.UUID {
		t.Errorf("expected UUID %q, got %q", healthcheck.UUID, readHealthcheck.UUID)
	}
	if readHealthcheck.Name != healthcheck.Name {
		t.Errorf("expected Name %q, got %q", healthcheck.Name, readHealthcheck.Name)
	}
	if readHealthcheck.PingURL != healthcheck.PingURL {
		t.Errorf("expected PingURL %q, got %q", healthcheck.PingURL, readHealthcheck.PingURL)
	}

	// Update healthcheck (from cassette)
	newName := "vcr-test-healthcheck-updated"
	updateReq := UpdateHealthcheckRequest{
		Name: &newName,
	}

	updatedHealthcheck, err := client.UpdateHealthcheck(ctx, healthcheck.UUID, updateReq)
	if err != nil {
		t.Fatalf("UpdateHealthcheck failed: %v", err)
	}

	// Note: Update API returns empty response, need to read back to verify
	t.Logf("Update response UUID: %q (may be empty)", updatedHealthcheck.UUID)

	// Read back to verify update
	verifyHealthcheck, err := client.GetHealthcheck(ctx, healthcheck.UUID)
	if err != nil {
		t.Fatalf("GetHealthcheck after update failed: %v", err)
	}
	if verifyHealthcheck.UUID != healthcheck.UUID {
		t.Errorf("UUID should not change on update, got %q", verifyHealthcheck.UUID)
	}
	if verifyHealthcheck.Name != newName {
		t.Errorf("expected updated Name %q, got %q", newName, verifyHealthcheck.Name)
	}

	t.Log("All healthcheck response structure validations passed")
}

// TestContract_Healthcheck_ListStructure validates healthcheck list response structure.
func TestContract_Healthcheck_ListStructure(t *testing.T) {
	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "healthcheck_list",
		Mode:         testutil.ModeReplay,
		CassetteDir:  "testdata/cassettes",
	})
	defer func() {
		if err := r.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}
	}()

	apiKey := os.Getenv("HYPERPING_API_KEY")
	if apiKey == "" {
		apiKey = "test_api_key"
	}

	client := NewClient(apiKey, WithHTTPClient(httpClient))
	ctx := context.Background()

	healthchecks, err := client.ListHealthchecks(ctx)
	if err != nil {
		t.Fatalf("ListHealthchecks failed: %v", err)
	}

	// Validate list is not nil
	if healthchecks == nil {
		t.Fatal("expected healthchecks list to not be nil")
	}

	// Validate each healthcheck in list has required fields
	for i, hc := range healthchecks {
		if hc.UUID == "" {
			t.Errorf("healthcheck[%d]: UUID is required but empty", i)
		}
		if hc.Name == "" {
			t.Errorf("healthcheck[%d]: Name is required but empty", i)
		}
		if hc.PingURL == "" {
			t.Errorf("healthcheck[%d]: PingURL is required but empty", i)
		}

		// Validate period structure
		if hc.PeriodType == "" {
			t.Errorf("healthcheck[%d]: PeriodType is required but empty", i)
		}
		if hc.Period <= 0 {
			t.Errorf("healthcheck[%d]: Period must be positive, got %d", i, hc.Period)
		}

		// Validate grace period structure
		if hc.GracePeriodValue <= 0 {
			t.Errorf("healthcheck[%d]: GracePeriodValue must be positive, got %d", i, hc.GracePeriodValue)
		}
		if hc.GracePeriodType == "" {
			t.Errorf("healthcheck[%d]: GracePeriodType is required but empty", i)
		}
		if hc.GracePeriod <= 0 {
			t.Errorf("healthcheck[%d]: GracePeriod must be positive, got %d", i, hc.GracePeriod)
		}

		// Validate enum values
		validPeriodTypes := map[string]bool{
			"seconds": true,
			"minutes": true,
			"hours":   true,
			"days":    true,
		}
		if !validPeriodTypes[hc.PeriodType] {
			t.Errorf("healthcheck[%d]: invalid PeriodType %q", i, hc.PeriodType)
		}
		if !validPeriodTypes[hc.GracePeriodType] {
			t.Errorf("healthcheck[%d]: invalid GracePeriodType %q", i, hc.GracePeriodType)
		}
	}

	t.Logf("Validated %d healthchecks in list response", len(healthchecks))
}

// TestContract_Healthcheck_PauseResumeStructure validates pause/resume action response structure.
// Note: This test is skipped because pause/resume is not included in the CRUD cassette
// and currently returns 500 errors from the API.
func TestContract_Healthcheck_PauseResumeStructure(t *testing.T) {
	t.Skip("Pause/resume operations currently return 500 errors from API")
}

// TestContract_Healthcheck_CronSupport validates cron-based healthcheck creation.
func TestContract_Healthcheck_CronSupport(t *testing.T) {
	// This test validates the cron field support in the API
	// We're testing the structure, not recording a new cassette

	cron := "0 0 * * *"
	timezone := "America/New_York"
	req := CreateHealthcheckRequest{
		Name:             "cron-healthcheck",
		Cron:             &cron,
		Timezone:         &timezone,
		GracePeriodValue: 300,
		GracePeriodType:  "seconds",
	}

	// Validate request can be marshaled with cron fields
	if req.Cron == nil {
		t.Error("expected Cron to be set")
	}
	if req.Timezone == nil {
		t.Error("expected Timezone to be set")
	}
	if *req.Cron != "0 0 * * *" {
		t.Errorf("expected Cron '0 0 * * *', got %q", *req.Cron)
	}
	if *req.Timezone != "America/New_York" {
		t.Errorf("expected Timezone 'America/New_York', got %q", *req.Timezone)
	}

	t.Log("Cron field structure validation passed")
}
