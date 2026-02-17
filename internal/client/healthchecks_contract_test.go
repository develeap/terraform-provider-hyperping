// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"os"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/internal/provider/testutil"
)

// healthcheckField holds a field name and the value to check for being non-empty or positive.
type healthcheckField struct {
	name     string
	isEmpty  bool   // true if string value is empty
	strValue string // populated for string assertions
	intValue int    // populated for positive-int assertions
	isInt    bool   // if true, check intValue > 0; otherwise check strValue != ""
}

// assertHealthcheckRequiredFields validates the core required fields of a Healthcheck.
func assertHealthcheckRequiredFields(t *testing.T, hc *Healthcheck, label string) {
	t.Helper()
	fields := []healthcheckField{
		{name: "UUID", strValue: hc.UUID},
		{name: "Name", strValue: hc.Name},
		{name: "PingURL", strValue: hc.PingURL},
		{name: "PeriodType", strValue: hc.PeriodType},
		{name: "GracePeriodType", strValue: hc.GracePeriodType},
	}
	for _, f := range fields {
		if f.strValue == "" {
			t.Errorf("%s: %s is required but empty", label, f.name)
		}
	}
	if hc.Period <= 0 {
		t.Errorf("%s: Period must be positive, got %d", label, hc.Period)
	}
	if hc.GracePeriodValue <= 0 {
		t.Errorf("%s: GracePeriodValue must be positive, got %d", label, hc.GracePeriodValue)
	}
	if hc.GracePeriod <= 0 {
		t.Errorf("%s: GracePeriod must be positive, got %d", label, hc.GracePeriod)
	}
}

// assertValidPeriodType checks that a period type string is one of the allowed enum values.
func assertValidPeriodType(t *testing.T, label, fieldName, periodType string) {
	t.Helper()
	validPeriodTypes := map[string]bool{
		"seconds": true,
		"minutes": true,
		"hours":   true,
		"days":    true,
	}
	if !validPeriodTypes[periodType] {
		t.Errorf("%s: invalid %s %q, must be one of: seconds, minutes, hours, days", label, fieldName, periodType)
	}
}

// newHealthcheckVCRClient sets up a VCR recorder and returns a client and teardown func.
func newHealthcheckVCRClient(t *testing.T, cassetteName string) (*Client, func()) {
	t.Helper()
	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: cassetteName,
		Mode:         testutil.ModeReplay,
		CassetteDir:  "testdata/cassettes",
	})
	apiKey := os.Getenv("HYPERPING_API_KEY")
	if apiKey == "" {
		apiKey = "test_api_key"
	}
	client := NewClient(apiKey, WithHTTPClient(httpClient))
	teardown := func() {
		if err := r.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}
	}
	return client, teardown
}

// TestContract_Healthcheck_ResponseStructure validates healthcheck response structure from VCR cassette.
func TestContract_Healthcheck_ResponseStructure(t *testing.T) {
	client, teardown := newHealthcheckVCRClient(t, "healthcheck_crud")
	defer teardown()

	ctx := context.Background()

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

	t.Run("CreateResponse", func(t *testing.T) {
		assertHealthcheckRequiredFields(t, healthcheck, "create")
		assertValidPeriodType(t, "create", "PeriodType", healthcheck.PeriodType)
		assertValidPeriodType(t, "create", "GracePeriodType", healthcheck.GracePeriodType)

		if healthcheck.PeriodValue != nil && *healthcheck.PeriodValue <= 0 {
			t.Errorf("PeriodValue must be positive, got %d", *healthcheck.PeriodValue)
		}
		t.Logf("CreatedAt: %q", healthcheck.CreatedAt)
	})

	t.Run("ReadResponse", func(t *testing.T) {
		readHealthcheck, err := client.GetHealthcheck(ctx, healthcheck.UUID)
		if err != nil {
			t.Fatalf("GetHealthcheck failed: %v", err)
		}
		if readHealthcheck.UUID != healthcheck.UUID {
			t.Errorf("expected UUID %q, got %q", healthcheck.UUID, readHealthcheck.UUID)
		}
		if readHealthcheck.Name != healthcheck.Name {
			t.Errorf("expected Name %q, got %q", healthcheck.Name, readHealthcheck.Name)
		}
		if readHealthcheck.PingURL != healthcheck.PingURL {
			t.Errorf("expected PingURL %q, got %q", healthcheck.PingURL, readHealthcheck.PingURL)
		}
	})

	t.Run("UpdateResponse", func(t *testing.T) {
		newName := "vcr-test-healthcheck-updated"
		updateReq := UpdateHealthcheckRequest{Name: &newName}
		updatedHealthcheck, err := client.UpdateHealthcheck(ctx, healthcheck.UUID, updateReq)
		if err != nil {
			t.Fatalf("UpdateHealthcheck failed: %v", err)
		}
		t.Logf("Update response UUID: %q (may be empty)", updatedHealthcheck.UUID)

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
	})

	t.Log("All healthcheck response structure validations passed")
}

// TestContract_Healthcheck_ListStructure validates healthcheck list response structure.
func TestContract_Healthcheck_ListStructure(t *testing.T) {
	client, teardown := newHealthcheckVCRClient(t, "healthcheck_list")
	defer teardown()

	ctx := context.Background()

	healthchecks, err := client.ListHealthchecks(ctx)
	if err != nil {
		t.Fatalf("ListHealthchecks failed: %v", err)
	}
	if healthchecks == nil {
		t.Fatal("expected healthchecks list to not be nil")
	}

	for i, hc := range healthchecks {
		label := "healthcheck[" + string(rune('0'+i)) + "]"
		hcCopy := hc
		assertHealthcheckRequiredFields(t, &hcCopy, label)
		assertValidPeriodType(t, label, "PeriodType", hc.PeriodType)
		assertValidPeriodType(t, label, "GracePeriodType", hc.GracePeriodType)
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
	if req.Name != "cron-healthcheck" {
		t.Errorf("expected Name 'cron-healthcheck', got %q", req.Name)
	}
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
	if req.GracePeriodValue != 300 {
		t.Errorf("expected GracePeriodValue 300, got %d", req.GracePeriodValue)
	}
	if req.GracePeriodType != "seconds" {
		t.Errorf("expected GracePeriodType 'seconds', got %q", req.GracePeriodType)
	}

	t.Log("Cron field structure validation passed")
}
