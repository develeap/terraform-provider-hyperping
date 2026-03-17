// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// statuspagesBasePath uses the exported constant for consistency.
var statuspagesBasePath = StatuspagesBasePath

// ListStatusPages returns a paginated list of status pages.
// Parameters:
//   - page: Optional 0-indexed page number
//   - search: Optional search filter for name, hostname, or subdomain
func (c *Client) ListStatusPages(ctx context.Context, page *int, search *string) (*StatusPagePaginatedResponse, error) {
	path := statuspagesBasePath

	// Build query parameters
	params := url.Values{}
	if page != nil {
		params.Add("page", strconv.Itoa(*page))
	}
	if search != nil && *search != "" {
		params.Add("search", *search)
	}

	if len(params) > 0 {
		path = fmt.Sprintf("%s?%s", path, params.Encode())
	}

	var response StatusPagePaginatedResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to list status pages: %w", err)
	}

	return &response, nil
}

// GetStatusPage returns a single status page by UUID.
func (c *Client) GetStatusPage(ctx context.Context, uuid string) (*StatusPage, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("GetStatusPage: %w", err)
	}

	// API returns: {"statuspage": {...}}
	var response struct {
		StatusPage StatusPage `json:"statuspage"`
	}
	path := fmt.Sprintf("%s/%s", statuspagesBasePath, uuid)
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get status page %s: %w", uuid, err)
	}

	return &response.StatusPage, nil
}

// CreateStatusPage creates a new status page.
func (c *Client) CreateStatusPage(ctx context.Context, req CreateStatusPageRequest) (*StatusPage, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("CreateStatusPage: %w", err)
	}

	// API returns: {"message": "Status page created", "statuspage": {...}}
	var response struct {
		Message    string     `json:"message"`
		StatusPage StatusPage `json:"statuspage"`
	}

	if err := c.doRequest(ctx, http.MethodPost, statuspagesBasePath, req, &response); err != nil {
		return nil, fmt.Errorf("failed to create status page: %w", err)
	}

	return &response.StatusPage, nil
}

// UpdateStatusPage updates an existing status page.
// Only provided fields will be updated.
func (c *Client) UpdateStatusPage(ctx context.Context, uuid string, req UpdateStatusPageRequest) (*StatusPage, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("UpdateStatusPage: %w", err)
	}

	// API returns: {"message": "Status page updated", "statuspage": {...}}
	var response struct {
		Message    string     `json:"message"`
		StatusPage StatusPage `json:"statuspage"`
	}

	path := fmt.Sprintf("%s/%s", statuspagesBasePath, uuid)
	if err := c.doRequest(ctx, http.MethodPut, path, req, &response); err != nil {
		return nil, fmt.Errorf("failed to update status page %s: %w", uuid, err)
	}

	return &response.StatusPage, nil
}

// DeleteStatusPage deletes a status page.
// Warning: This action is irreversible.
func (c *Client) DeleteStatusPage(ctx context.Context, uuid string) error {
	if err := ValidateResourceID(uuid); err != nil {
		return fmt.Errorf("DeleteStatusPage: %w", err)
	}

	path := fmt.Sprintf("%s/%s", statuspagesBasePath, uuid)
	if err := c.doRequest(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete status page %s: %w", uuid, err)
	}

	return nil
}

// ListSubscribers returns a paginated list of subscribers for a status page.
// Parameters:
//   - uuid: Status page UUID
//   - page: Optional 0-indexed page number
//   - subscriberType: Optional filter by type (all, email, sms, slack, teams)
func (c *Client) ListSubscribers(ctx context.Context, uuid string, page *int, subscriberType *string) (*SubscriberPaginatedResponse, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("ListSubscribers: %w", err)
	}

	path := fmt.Sprintf("%s/%s/subscribers", statuspagesBasePath, uuid)

	// Build query parameters
	params := url.Values{}
	if page != nil {
		params.Add("page", strconv.Itoa(*page))
	}
	if subscriberType != nil && *subscriberType != "" && *subscriberType != "all" {
		params.Add("type", *subscriberType)
	}

	if len(params) > 0 {
		path = fmt.Sprintf("%s?%s", path, params.Encode())
	}

	var response SubscriberPaginatedResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to list subscribers for status page %s: %w", uuid, err)
	}

	return &response, nil
}

// AddSubscriber adds a new subscriber to a status page.
// Note: Slack subscribers cannot be added via API - they must use OAuth flow.
// Subscribers added via API are automatically confirmed.
func (c *Client) AddSubscriber(ctx context.Context, uuid string, req AddSubscriberRequest) (*StatusPageSubscriber, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("AddSubscriber: %w", err)
	}

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("AddSubscriber: %w", err)
	}

	// Reject Slack subscribers with clear error message
	if req.Type == "slack" {
		return nil, fmt.Errorf("slack subscribers cannot be added via API - they must use the OAuth flow through the Hyperping dashboard")
	}

	// API returns: {"message": "Subscriber added", "subscriber": {...}}
	var response struct {
		Message    string               `json:"message"`
		Subscriber StatusPageSubscriber `json:"subscriber"`
	}

	path := fmt.Sprintf("%s/%s/subscribers", statuspagesBasePath, uuid)
	if err := c.doRequest(ctx, http.MethodPost, path, req, &response); err != nil {
		return nil, fmt.Errorf("failed to add subscriber to status page %s: %w", uuid, err)
	}

	return &response.Subscriber, nil
}

// DeleteSubscriber deletes a subscriber from a status page.
func (c *Client) DeleteSubscriber(ctx context.Context, uuid string, subscriberID int) error {
	if err := ValidateResourceID(uuid); err != nil {
		return fmt.Errorf("DeleteSubscriber: %w", err)
	}

	if subscriberID <= 0 {
		return fmt.Errorf("DeleteSubscriber: subscriber ID must be positive (got %d)", subscriberID)
	}

	path := fmt.Sprintf("%s/%s/subscribers/%d", statuspagesBasePath, uuid, subscriberID)
	if err := c.doRequest(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete subscriber %d from status page %s: %w", subscriberID, uuid, err)
	}

	return nil
}
