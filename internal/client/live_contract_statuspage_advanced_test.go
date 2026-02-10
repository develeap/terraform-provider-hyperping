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

// TestLiveContract_StatusPage_Update tests updating a status page.
func TestLiveContract_StatusPage_Update(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "statuspage_update",
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

	subdomain := "vcr-test-update-" + time.Now().Format("20060102150405")
	createReq := CreateStatusPageRequest{
		Name:      "VCR Test Update",
		Subdomain: &subdomain,
		Languages: []string{"en"},
	}

	statusPage, err := client.CreateStatusPage(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateStatusPage failed: %v", err)
	}
	defer client.DeleteStatusPage(ctx, statusPage.UUID)

	// Update status page
	newName := "VCR Test Updated Name"
	updateReq := UpdateStatusPageRequest{
		Name: &newName,
	}

	updated, err := client.UpdateStatusPage(ctx, statusPage.UUID, updateReq)
	if err != nil {
		t.Fatalf("UpdateStatusPage failed: %v", err)
	}

	responseJSON, _ := json.MarshalIndent(updated, "", "  ")
	t.Logf("Updated status page:\n%s", string(responseJSON))

	if updated.Name != newName {
		t.Errorf("expected name %q, got %q", newName, updated.Name)
	}

	t.Log("Successfully updated status page")
}

// TestLiveContract_StatusPage_Subscribers tests subscriber management.
func TestLiveContract_StatusPage_Subscribers(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "statuspage_subscribers",
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

	// Get existing status page
	statusPages, err := client.ListStatusPages(ctx, nil, nil)
	if err != nil || len(statusPages.StatusPages) == 0 {
		t.Skip("Skipping: status pages not accessible")
	}

	statusPageUUID := statusPages.StatusPages[0].UUID

	// Add subscriber
	email := "test-subscriber@example.com"
	addReq := AddSubscriberRequest{
		Type:  "email",
		Email: &email,
	}

	subscriber, err := client.AddSubscriber(ctx, statusPageUUID, addReq)
	if err != nil {
		t.Fatalf("AddSubscriber failed: %v", err)
	}

	t.Logf("Added subscriber: ID %d", subscriber.ID)
	responseJSON, _ := json.MarshalIndent(subscriber, "", "  ")
	t.Logf("Subscriber:\n%s", string(responseJSON))

	// List subscribers
	subscriberList, err := client.ListSubscribers(ctx, statusPageUUID, nil, nil)
	if err != nil {
		t.Fatalf("ListSubscribers failed: %v", err)
	}

	t.Logf("Found %d subscribers", len(subscriberList.Subscribers))

	// Delete subscriber
	err = client.DeleteSubscriber(ctx, statusPageUUID, subscriber.ID)
	if err != nil {
		t.Fatalf("DeleteSubscriber failed: %v", err)
	}

	t.Log("Successfully managed status page subscribers")
}

// TestLiveContract_StatusPage_ListWithPagination tests listing with pagination.
func TestLiveContract_StatusPage_ListWithPagination(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "statuspage_list_pagination",
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

	// List with page 1
	page := 1
	response, err := client.ListStatusPages(ctx, &page, nil)
	if err != nil {
		t.Fatalf("ListStatusPages failed: %v", err)
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")
	t.Logf("Status pages (page %d):\n%s", page, string(responseJSON))

	t.Logf("Found %d status pages on page %d", len(response.StatusPages), page)
	t.Logf("Total: %d, Has next page: %v", response.Total, response.HasNextPage)

	t.Log("Successfully listed status pages with pagination")
}

// TestLiveContract_StatusPage_ListWithSearch tests listing with search.
func TestLiveContract_StatusPage_ListWithSearch(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "statuspage_list_search",
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

	// List with search term
	search := "vcr"
	response, err := client.ListStatusPages(ctx, nil, &search)
	if err != nil {
		t.Fatalf("ListStatusPages with search failed: %v", err)
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")
	t.Logf("Status pages (search=%q):\n%s", search, string(responseJSON))

	t.Logf("Found %d status pages matching %q", len(response.StatusPages), search)

	t.Log("Successfully searched status pages")
}

// TestLiveContract_StatusPage_NotFound tests getting a non-existent status page.
func TestLiveContract_StatusPage_NotFound(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "statuspage_not_found",
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

	// Try to get non-existent status page
	_, err := client.GetStatusPage(ctx, "sp_nonexistent123456")
	if err == nil {
		t.Fatal("expected error for non-existent status page")
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
