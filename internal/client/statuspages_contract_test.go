// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/internal/provider/testutil"
)

// =============================================================================
// UpdateStatusPage Tests
// =============================================================================

// TestContract_UpdateStatusPage_NameChange tests updating a status page name.
func TestContract_UpdateStatusPage_NameChange(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "contract_updatestatuspage_name",
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

	// Create status page first
	subdomain := "vcr-test-contract-update"
	createReq := CreateStatusPageRequest{
		Name:      "VCR Contract Test Original",
		Subdomain: &subdomain,
		Languages: []string{"en"},
	}

	statusPage, err := client.CreateStatusPage(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateStatusPage failed: %v", err)
	}
	defer client.DeleteStatusPage(ctx, statusPage.UUID)

	// Test: Update name
	newName := "VCR Contract Test Updated"
	updateReq := UpdateStatusPageRequest{
		Name: &newName,
	}

	updated, err := client.UpdateStatusPage(ctx, statusPage.UUID, updateReq)
	if err != nil {
		t.Fatalf("UpdateStatusPage failed: %v", err)
	}

	responseJSON, _ := json.MarshalIndent(updated, "", "  ")
	t.Logf("Updated status page:\n%s", string(responseJSON))

	if updated.Settings.Name != newName {
		t.Errorf("expected name %q, got %q", newName, updated.Settings.Name)
	}

	t.Log("Successfully updated status page name")
}

// TestContract_UpdateStatusPage_MultipleFields tests updating multiple fields at once.
func TestContract_UpdateStatusPage_MultipleFields(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "contract_updatestatuspage_multiple",
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

	// Create status page first
	subdomain := "vcr-test-multi-update"
	createReq := CreateStatusPageRequest{
		Name:      "VCR Multi Update Test",
		Subdomain: &subdomain,
		Languages: []string{"en"},
	}

	statusPage, err := client.CreateStatusPage(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateStatusPage failed: %v", err)
	}
	defer client.DeleteStatusPage(ctx, statusPage.UUID)

	// Test: Update multiple fields
	newName := "VCR Multi Updated"
	newWebsite := "https://example.com"
	newTheme := "dark"
	newAccentColor := "#ff5733"
	autoRefresh := true

	updateReq := UpdateStatusPageRequest{
		Name:        &newName,
		Website:     &newWebsite,
		Theme:       &newTheme,
		AccentColor: &newAccentColor,
		AutoRefresh: &autoRefresh,
	}

	updated, err := client.UpdateStatusPage(ctx, statusPage.UUID, updateReq)
	if err != nil {
		t.Fatalf("UpdateStatusPage failed: %v", err)
	}

	responseJSON, _ := json.MarshalIndent(updated, "", "  ")
	t.Logf("Updated status page:\n%s", string(responseJSON))

	if updated.Settings.Name != newName {
		t.Errorf("expected name %q, got %q", newName, updated.Settings.Name)
	}
	if updated.Settings.Website != newWebsite {
		t.Errorf("expected website %q, got %q", newWebsite, updated.Settings.Website)
	}
	if updated.Settings.Theme != newTheme {
		t.Errorf("expected theme %q, got %q", newTheme, updated.Settings.Theme)
	}
	if updated.Settings.AccentColor != newAccentColor {
		t.Errorf("expected accent color %q, got %q", newAccentColor, updated.Settings.AccentColor)
	}
	if updated.Settings.AutoRefresh != autoRefresh {
		t.Errorf("expected auto refresh %v, got %v", autoRefresh, updated.Settings.AutoRefresh)
	}

	t.Log("Successfully updated multiple status page fields")
}

// TestContract_UpdateStatusPage_InvalidID tests updating with invalid UUID.
func TestContract_UpdateStatusPage_InvalidID(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "contract_updatestatuspage_invalid",
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

	// Test: Update with empty UUID (validation error)
	newName := "Should Not Work"
	updateReq := UpdateStatusPageRequest{
		Name: &newName,
	}

	_, err := client.UpdateStatusPage(ctx, "", updateReq)
	if err == nil {
		t.Fatal("expected error for empty UUID")
	}

	t.Logf("Got expected validation error: %v", err)

	// Test: Update with non-existent UUID (API error)
	_, err = client.UpdateStatusPage(ctx, "sp_nonexistent123456", updateReq)
	if err == nil {
		t.Fatal("expected error for non-existent status page")
	}

	t.Logf("Got expected API error: %v", err)
}

// =============================================================================
// AddSubscriber Tests
// =============================================================================

// TestContract_AddSubscriber_Email tests adding an email subscriber.
func TestContract_AddSubscriber_Email(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "contract_addsubscriber_email",
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

	// Get or create status page
	statusPages, err := client.ListStatusPages(ctx, nil, nil)
	if err != nil || len(statusPages.StatusPages) == 0 {
		t.Skip("Skipping: status pages not accessible")
	}

	statusPageUUID := statusPages.StatusPages[0].UUID

	// Test: Add email subscriber
	email := "test-contract@example.com"
	addReq := AddSubscriberRequest{
		Type:  "email",
		Email: &email,
	}

	subscriber, err := client.AddSubscriber(ctx, statusPageUUID, addReq)
	if err != nil {
		t.Fatalf("AddSubscriber failed: %v", err)
	}
	defer client.DeleteSubscriber(ctx, statusPageUUID, subscriber.ID)

	t.Logf("Added subscriber: ID %d", subscriber.ID)
	responseJSON, _ := json.MarshalIndent(subscriber, "", "  ")
	t.Logf("Subscriber:\n%s", string(responseJSON))

	if subscriber.Type != "email" {
		t.Errorf("expected type 'email', got %q", subscriber.Type)
	}
	if subscriber.Email == nil || *subscriber.Email != email {
		t.Errorf("expected email %q, got %v", email, subscriber.Email)
	}

	t.Log("Successfully added email subscriber")
}

// TestContract_AddSubscriber_MultipleTypes tests adding different subscriber types.
func TestContract_AddSubscriber_MultipleTypes(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "contract_addsubscriber_types",
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

	// Get status page
	statusPages, err := client.ListStatusPages(ctx, nil, nil)
	if err != nil || len(statusPages.StatusPages) == 0 {
		t.Skip("Skipping: status pages not accessible")
	}

	statusPageUUID := statusPages.StatusPages[0].UUID

	// Test: Add email subscriber
	email := "multi-test@example.com"
	emailReq := AddSubscriberRequest{
		Type:  "email",
		Email: &email,
	}

	emailSub, err := client.AddSubscriber(ctx, statusPageUUID, emailReq)
	if err != nil {
		t.Fatalf("AddSubscriber (email) failed: %v", err)
	}
	defer client.DeleteSubscriber(ctx, statusPageUUID, emailSub.ID)

	if emailSub.Type != "email" {
		t.Errorf("expected type 'email', got %q", emailSub.Type)
	}

	// Test: Add SMS subscriber
	phone := "+12125551234"
	smsReq := AddSubscriberRequest{
		Type:  "sms",
		Phone: &phone,
	}

	smsSub, err := client.AddSubscriber(ctx, statusPageUUID, smsReq)
	if err != nil {
		t.Fatalf("AddSubscriber (sms) failed: %v", err)
	}
	defer client.DeleteSubscriber(ctx, statusPageUUID, smsSub.ID)

	if smsSub.Type != "sms" {
		t.Errorf("expected type 'sms', got %q", smsSub.Type)
	}

	t.Log("Successfully added multiple subscriber types")
}

// TestContract_AddSubscriber_ValidationErrors tests validation errors.
func TestContract_AddSubscriber_ValidationErrors(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "contract_addsubscriber_validation",
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

	// Get status page
	statusPages, err := client.ListStatusPages(ctx, nil, nil)
	if err != nil || len(statusPages.StatusPages) == 0 {
		t.Skip("Skipping: status pages not accessible")
	}

	statusPageUUID := statusPages.StatusPages[0].UUID

	// Test: Empty UUID
	email := "test@example.com"
	addReq := AddSubscriberRequest{
		Type:  "email",
		Email: &email,
	}

	_, err = client.AddSubscriber(ctx, "", addReq)
	if err == nil {
		t.Error("expected error for empty UUID")
	} else {
		t.Logf("Got expected validation error for empty UUID: %v", err)
	}

	// Test: Email type without email field
	invalidReq := AddSubscriberRequest{
		Type: "email",
	}

	_, err = client.AddSubscriber(ctx, statusPageUUID, invalidReq)
	if err == nil {
		t.Error("expected validation error for missing email")
	} else {
		t.Logf("Got expected validation error for missing email: %v", err)
	}

	// Test: Slack subscriber (should be rejected)
	slackReq := AddSubscriberRequest{
		Type: "slack",
	}

	_, err = client.AddSubscriber(ctx, statusPageUUID, slackReq)
	if err == nil {
		t.Error("expected error for slack subscriber")
	} else {
		t.Logf("Got expected error for slack subscriber: %v", err)
	}

	t.Log("Successfully validated subscriber errors")
}

// TestContract_AddSubscriber_WithLanguage tests adding subscriber with custom language.
func TestContract_AddSubscriber_WithLanguage(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "contract_addsubscriber_language",
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

	// Get status page
	statusPages, err := client.ListStatusPages(ctx, nil, nil)
	if err != nil || len(statusPages.StatusPages) == 0 {
		t.Skip("Skipping: status pages not accessible")
	}

	statusPageUUID := statusPages.StatusPages[0].UUID

	// Test: Add subscriber with Spanish language
	email := "spanish-test@example.com"
	language := "es"
	addReq := AddSubscriberRequest{
		Type:     "email",
		Email:    &email,
		Language: &language,
	}

	subscriber, err := client.AddSubscriber(ctx, statusPageUUID, addReq)
	if err != nil {
		t.Fatalf("AddSubscriber with language failed: %v", err)
	}
	defer client.DeleteSubscriber(ctx, statusPageUUID, subscriber.ID)

	if subscriber.Language != language {
		t.Errorf("expected language %q, got %q", language, subscriber.Language)
	}

	t.Log("Successfully added subscriber with custom language")
}

// =============================================================================
// DeleteSubscriber Tests
// =============================================================================

// TestContract_DeleteSubscriber_Success tests successful deletion.
func TestContract_DeleteSubscriber_Success(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "contract_deletesubscriber_success",
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

	// Get status page
	statusPages, err := client.ListStatusPages(ctx, nil, nil)
	if err != nil || len(statusPages.StatusPages) == 0 {
		t.Skip("Skipping: status pages not accessible")
	}

	statusPageUUID := statusPages.StatusPages[0].UUID

	// Create subscriber to delete
	email := "delete-test@example.com"
	addReq := AddSubscriberRequest{
		Type:  "email",
		Email: &email,
	}

	subscriber, err := client.AddSubscriber(ctx, statusPageUUID, addReq)
	if err != nil {
		t.Fatalf("AddSubscriber failed: %v", err)
	}

	// Test: Delete subscriber
	err = client.DeleteSubscriber(ctx, statusPageUUID, subscriber.ID)
	if err != nil {
		t.Fatalf("DeleteSubscriber failed: %v", err)
	}

	t.Logf("Successfully deleted subscriber ID %d", subscriber.ID)

	// Verify deletion by listing subscribers
	subscribers, err := client.ListSubscribers(ctx, statusPageUUID, nil, nil)
	if err != nil {
		t.Fatalf("ListSubscribers failed: %v", err)
	}

	for _, sub := range subscribers.Subscribers {
		if sub.ID == subscriber.ID {
			t.Errorf("subscriber ID %d still exists after deletion", subscriber.ID)
		}
	}

	t.Log("Successfully verified subscriber deletion")
}

// TestContract_DeleteSubscriber_InvalidID tests deletion with invalid IDs.
func TestContract_DeleteSubscriber_InvalidID(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "contract_deletesubscriber_invalid",
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

	// Get status page
	statusPages, err := client.ListStatusPages(ctx, nil, nil)
	if err != nil || len(statusPages.StatusPages) == 0 {
		t.Skip("Skipping: status pages not accessible")
	}

	statusPageUUID := statusPages.StatusPages[0].UUID

	// Test: Delete with zero ID (validation error)
	err = client.DeleteSubscriber(ctx, statusPageUUID, 0)
	if err == nil {
		t.Error("expected error for zero subscriber ID")
	} else {
		t.Logf("Got expected validation error for zero ID: %v", err)
	}

	// Test: Delete with negative ID (validation error)
	err = client.DeleteSubscriber(ctx, statusPageUUID, -1)
	if err == nil {
		t.Error("expected error for negative subscriber ID")
	} else {
		t.Logf("Got expected validation error for negative ID: %v", err)
	}

	// Test: Delete with empty UUID (validation error)
	err = client.DeleteSubscriber(ctx, "", 123)
	if err == nil {
		t.Error("expected error for empty UUID")
	} else {
		t.Logf("Got expected validation error for empty UUID: %v", err)
	}

	t.Log("Successfully validated deletion errors")
}

// TestContract_DeleteSubscriber_NonExistent tests deleting non-existent subscriber.
func TestContract_DeleteSubscriber_NonExistent(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "contract_deletesubscriber_nonexistent",
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

	// Get status page
	statusPages, err := client.ListStatusPages(ctx, nil, nil)
	if err != nil || len(statusPages.StatusPages) == 0 {
		t.Skip("Skipping: status pages not accessible")
	}

	statusPageUUID := statusPages.StatusPages[0].UUID

	// Test: Delete non-existent subscriber
	err = client.DeleteSubscriber(ctx, statusPageUUID, 999999)
	if err == nil {
		t.Fatal("expected error for non-existent subscriber")
	}

	t.Logf("Got expected API error: %v", err)
}

// =============================================================================
// ListSubscribers Tests
// =============================================================================

// TestContract_ListSubscribers_Success tests listing subscribers.
func TestContract_ListSubscribers_Success(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "contract_listsubscribers_success",
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

	// Get status page
	statusPages, err := client.ListStatusPages(ctx, nil, nil)
	if err != nil || len(statusPages.StatusPages) == 0 {
		t.Skip("Skipping: status pages not accessible")
	}

	statusPageUUID := statusPages.StatusPages[0].UUID

	// Test: List all subscribers
	response, err := client.ListSubscribers(ctx, statusPageUUID, nil, nil)
	if err != nil {
		t.Fatalf("ListSubscribers failed: %v", err)
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")
	t.Logf("Subscribers:\n%s", string(responseJSON))

	t.Logf("Found %d subscribers", len(response.Subscribers))
	t.Logf("Total: %d, Has next page: %v", response.Total, response.HasNextPage)

	t.Log("Successfully listed subscribers")
}

// TestContract_ListSubscribers_WithPagination tests pagination.
func TestContract_ListSubscribers_WithPagination(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "contract_listsubscribers_pagination",
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

	// Get status page
	statusPages, err := client.ListStatusPages(ctx, nil, nil)
	if err != nil || len(statusPages.StatusPages) == 0 {
		t.Skip("Skipping: status pages not accessible")
	}

	statusPageUUID := statusPages.StatusPages[0].UUID

	// Test: List with page parameter
	page := 0
	response, err := client.ListSubscribers(ctx, statusPageUUID, &page, nil)
	if err != nil {
		t.Fatalf("ListSubscribers with pagination failed: %v", err)
	}

	if response.Page != page {
		t.Errorf("expected page %d, got %d", page, response.Page)
	}

	t.Logf("Found %d subscribers on page %d", len(response.Subscribers), page)
	t.Log("Successfully tested pagination")
}

// TestContract_ListSubscribers_WithTypeFilter tests filtering by type.
func TestContract_ListSubscribers_WithTypeFilter(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "contract_listsubscribers_filter",
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

	// Get status page
	statusPages, err := client.ListStatusPages(ctx, nil, nil)
	if err != nil || len(statusPages.StatusPages) == 0 {
		t.Skip("Skipping: status pages not accessible")
	}

	statusPageUUID := statusPages.StatusPages[0].UUID

	// Add an email subscriber first
	email := "filter-test@example.com"
	addReq := AddSubscriberRequest{
		Type:  "email",
		Email: &email,
	}

	subscriber, err := client.AddSubscriber(ctx, statusPageUUID, addReq)
	if err != nil {
		t.Fatalf("AddSubscriber failed: %v", err)
	}
	defer client.DeleteSubscriber(ctx, statusPageUUID, subscriber.ID)

	// Test: Filter by email type
	filterType := "email"
	response, err := client.ListSubscribers(ctx, statusPageUUID, nil, &filterType)
	if err != nil {
		t.Fatalf("ListSubscribers with filter failed: %v", err)
	}

	// Verify all returned subscribers are email type
	for _, sub := range response.Subscribers {
		if sub.Type != "email" {
			t.Errorf("expected type 'email', got %q", sub.Type)
		}
	}

	t.Logf("Found %d email subscribers", len(response.Subscribers))

	// Test: Filter with "all" type (should not add filter parameter)
	allType := "all"
	responseAll, err := client.ListSubscribers(ctx, statusPageUUID, nil, &allType)
	if err != nil {
		t.Fatalf("ListSubscribers with 'all' filter failed: %v", err)
	}

	t.Logf("Found %d subscribers with 'all' filter", len(responseAll.Subscribers))
	t.Log("Successfully tested type filtering")
}

// TestContract_ListSubscribers_InvalidUUID tests listing with invalid UUID.
func TestContract_ListSubscribers_InvalidUUID(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "contract_listsubscribers_invalid",
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

	// Test: List with empty UUID (validation error)
	_, err := client.ListSubscribers(ctx, "", nil, nil)
	if err == nil {
		t.Error("expected error for empty UUID")
	} else {
		t.Logf("Got expected validation error for empty UUID: %v", err)
	}

	// Test: List with non-existent UUID (API error)
	_, err = client.ListSubscribers(ctx, "sp_nonexistent123456", nil, nil)
	if err == nil {
		t.Fatal("expected error for non-existent status page")
	}

	t.Logf("Got expected API error: %v", err)
}
