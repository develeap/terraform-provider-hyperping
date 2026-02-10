// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/develeap/terraform-provider-hyperping/internal/provider/testutil"
)

// =============================================================================
// Healthcheck Contract Tests
// NOTE: Requires healthcheck API access (different API key permissions)
// =============================================================================

func TestLiveContract_Healthcheck_CRUD(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "healthcheck_crud",
		Mode:         testutil.GetRecordMode(),
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

	// Test Create
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

	t.Logf("Created healthcheck: %s", healthcheck.UUID)
	responseJSON, _ := json.MarshalIndent(healthcheck, "", "  ")
	t.Logf("Healthcheck response:\n%s", string(responseJSON))

	if healthcheck.UUID == "" {
		t.Error("expected UUID to be set")
	}
	if healthcheck.Name != "vcr-test-healthcheck" {
		t.Errorf("expected name 'vcr-test-healthcheck', got %q", healthcheck.Name)
	}
	if healthcheck.PingURL == "" {
		t.Error("expected PingURL to be set")
	}

	// Test Read
	readHealthcheck, err := client.GetHealthcheck(ctx, healthcheck.UUID)
	if err != nil {
		t.Fatalf("GetHealthcheck failed: %v", err)
	}
	if readHealthcheck.UUID != healthcheck.UUID {
		t.Errorf("expected UUID %q, got %q", healthcheck.UUID, readHealthcheck.UUID)
	}
	if readHealthcheck.Name != "vcr-test-healthcheck" {
		t.Errorf("expected name 'vcr-test-healthcheck', got %q", readHealthcheck.Name)
	}

	// Test Update - change name
	newName := "vcr-test-healthcheck-updated"
	updateReq := UpdateHealthcheckRequest{
		Name: &newName,
	}

	updatedHealthcheck, err := client.UpdateHealthcheck(ctx, healthcheck.UUID, updateReq)
	if err != nil {
		t.Fatalf("UpdateHealthcheck failed: %v", err)
	}

	t.Logf("Update response:\n%s", string(func() []byte { b, _ := json.MarshalIndent(updatedHealthcheck, "", "  "); return b }()))

	// Read back to verify update
	verifyHealthcheck, err := client.GetHealthcheck(ctx, healthcheck.UUID)
	if err != nil {
		t.Fatalf("GetHealthcheck after update failed: %v", err)
	}
	if verifyHealthcheck.Name != newName {
		t.Errorf("expected updated name %q, got %q", newName, verifyHealthcheck.Name)
	}

	t.Logf("Updated healthcheck name to: %s", verifyHealthcheck.Name)

	// Test Delete
	if err = client.DeleteHealthcheck(ctx, healthcheck.UUID); err != nil {
		t.Fatalf("DeleteHealthcheck failed: %v", err)
	}

	t.Log("Successfully completed healthcheck CRUD cycle")
}

func TestLiveContract_Healthcheck_List(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "healthcheck_list",
		Mode:         testutil.GetRecordMode(),
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

	// Create a test healthcheck first to ensure we have at least one
	periodValue := 60
	periodType := "seconds"
	createReq := CreateHealthcheckRequest{
		Name:             "vcr-test-healthcheck-list",
		PeriodValue:      &periodValue,
		PeriodType:       &periodType,
		GracePeriodValue: 300,
		GracePeriodType:  "seconds",
	}

	healthcheck, err := client.CreateHealthcheck(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateHealthcheck failed: %v", err)
	}
	defer client.DeleteHealthcheck(ctx, healthcheck.UUID)

	// Test List
	healthchecks, err := client.ListHealthchecks(ctx)
	if err != nil {
		t.Fatalf("ListHealthchecks failed: %v", err)
	}

	t.Logf("Found %d healthchecks", len(healthchecks))
	for i, hc := range healthchecks {
		data, _ := json.MarshalIndent(hc, "", "  ")
		t.Logf("Healthcheck %d:\n%s", i, string(data))
	}

	// Verify our created healthcheck is in the list
	found := false
	for _, hc := range healthchecks {
		if hc.UUID == healthcheck.UUID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find created healthcheck %s in list", healthcheck.UUID)
	}
}

func TestLiveContract_Healthcheck_NotFound(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "healthcheck_not_found",
		Mode:         testutil.GetRecordMode(),
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

	// Try to get non-existent healthcheck
	_, err := client.GetHealthcheck(ctx, "tok_nonexistent123456")
	if err == nil {
		t.Fatal("expected error for non-existent healthcheck")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.StatusCode != 404 {
		t.Errorf("expected status code 404, got %d", apiErr.StatusCode)
	}

	t.Logf("Got expected error: %v", apiErr)
}

// =============================================================================
// Outage Contract Tests
// NOTE: Requires outage API access (different API key permissions)
// =============================================================================

func TestLiveContract_Outage_CRUD(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "outage_crud",
		Mode:         testutil.GetRecordMode(),
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

	// Need a monitor for outage
	monitors, err := client.ListMonitors(ctx)
	if err != nil {
		t.Fatalf("ListMonitors failed: %v", err)
	}
	if len(monitors) == 0 {
		t.Skip("No monitors available for outage test")
	}

	monitorUUID := monitors[0].UUID
	endDate := time.Now().Add(-23 * time.Hour).UTC().Format(time.RFC3339)
	createReq := CreateOutageRequest{
		MonitorUUID: monitorUUID,
		StartDate:   time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339),
		EndDate:     &endDate,
		StatusCode:  500,
		Description: "Manual outage for contract testing",
		OutageType:  "manual",
	}

	outage, err := client.CreateOutage(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateOutage failed: %v", err)
	}

	t.Logf("Created outage: %s", outage.UUID)
	if outage.UUID == "" {
		t.Error("expected UUID to be set")
	}

	// Test Read
	readOutage, err := client.GetOutage(ctx, outage.UUID)
	if err != nil {
		t.Fatalf("GetOutage failed: %v", err)
	}
	if readOutage.UUID != outage.UUID {
		t.Errorf("expected UUID %q, got %q", outage.UUID, readOutage.UUID)
	}

	// Test Delete
	if err = client.DeleteOutage(ctx, outage.UUID); err != nil {
		t.Fatalf("DeleteOutage failed: %v", err)
	}

	t.Log("Successfully completed outage CRUD cycle")
}

func TestLiveContract_Outage_List(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "outage_list",
		Mode:         testutil.GetRecordMode(),
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

	outages, err := client.ListOutages(ctx)
	if err != nil {
		t.Fatalf("ListOutages failed: %v", err)
	}

	t.Logf("Found %d outages", len(outages))
	for i, o := range outages {
		data, _ := json.MarshalIndent(o, "", "  ")
		t.Logf("Outage %d:\n%s", i, string(data))
	}
}

// =============================================================================
// Status Page Contract Tests
// NOTE: Requires status page API access (different API key permissions)
// =============================================================================

func TestLiveContract_StatusPage_CRUD(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "statuspage_crud",
		Mode:         testutil.GetRecordMode(),
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

	subdomain := "vcr-test-" + time.Now().Format("20060102150405")
	createReq := CreateStatusPageRequest{
		Name:      "VCR Test Status Page",
		Subdomain: &subdomain,
		Languages: []string{"en"},
	}

	statusPage, err := client.CreateStatusPage(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateStatusPage failed: %v", err)
	}

	t.Logf("Created status page: %s", statusPage.UUID)
	if statusPage.UUID == "" {
		t.Error("expected UUID to be set")
	}

	// Test Read
	readStatusPage, err := client.GetStatusPage(ctx, statusPage.UUID)
	if err != nil {
		t.Fatalf("GetStatusPage failed: %v", err)
	}
	if readStatusPage.UUID != statusPage.UUID {
		t.Errorf("expected UUID %q, got %q", statusPage.UUID, readStatusPage.UUID)
	}

	// Test Delete
	if err = client.DeleteStatusPage(ctx, statusPage.UUID); err != nil {
		t.Fatalf("DeleteStatusPage failed: %v", err)
	}

	t.Log("Successfully completed status page CRUD cycle")
}

func TestLiveContract_StatusPage_List(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "statuspage_list",
		Mode:         testutil.GetRecordMode(),
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

	response, err := client.ListStatusPages(ctx, nil, nil)
	if err != nil {
		t.Fatalf("ListStatusPages failed: %v", err)
	}

	t.Logf("Found %d status pages", len(response.StatusPages))
	for i, sp := range response.StatusPages {
		data, _ := json.MarshalIndent(sp, "", "  ")
		t.Logf("StatusPage %d:\n%s", i, string(data))
	}
}

// =============================================================================
// Reports Contract Tests
// NOTE: Requires reports API access (different API key permissions)
// =============================================================================

func TestLiveContract_Report_Get(t *testing.T) {
	// Skip this test in replay mode - cassette needs re-recording
	if os.Getenv("VCR_MODE") != "record" && os.Getenv("HYPERPING_API_KEY") == "" {
		t.Skip("Skipping report test in replay mode (cassette incomplete)")
	}

	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "report_get",
		Mode:         testutil.GetRecordMode(),
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

	monitors, err := client.ListMonitors(ctx)
	if err != nil {
		t.Fatalf("ListMonitors failed: %v", err)
	}
	if len(monitors) == 0 {
		t.Skip("No monitors available for report test")
	}

	monitorUUID := monitors[0].UUID
	// Use fixed timestamps for VCR cassette compatibility
	// When recording, these should cover the period when cassette was created
	from := "2026-02-03T00:00:00Z"
	to := "2026-02-10T00:00:00Z"

	report, err := client.GetMonitorReport(ctx, monitorUUID, from, to)
	if err != nil {
		t.Fatalf("GetMonitorReport failed: %v", err)
	}

	responseJSON, _ := json.MarshalIndent(report, "", "  ")
	t.Logf("Report response:\n%s", string(responseJSON))
}

func TestLiveContract_Report_List(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "report_list",
		Mode:         testutil.GetRecordMode(),
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

	// Use static timestamps to match VCR cassette
	from := "2026-02-03T05:36:12Z"
	to := "2026-02-10T05:36:12Z"

	reports, err := client.ListMonitorReports(ctx, from, to)
	if err != nil {
		t.Fatalf("ListMonitorReports failed: %v", err)
	}

	t.Logf("Found %d monitor reports", len(reports))
	for i, r := range reports {
		data, _ := json.MarshalIndent(r, "", "  ")
		t.Logf("Report %d:\n%s", i, string(data))
	}
}
