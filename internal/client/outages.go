// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// outageListResponse holds the parsed outage list along with pagination metadata.
type outageListResponse struct {
	Outages     []Outage
	HasNextPage bool
}

// parseOutageListResponse handles the various response formats the API might return.
// It also extracts pagination metadata when available.
func parseOutageListResponse(raw json.RawMessage) (outageListResponse, error) {
	// Try direct array first (no pagination metadata available)
	var outages []Outage
	if err := json.Unmarshal(raw, &outages); err == nil {
		return outageListResponse{Outages: outages, HasNextPage: false}, nil
	}

	// Try wrapped formats with pagination metadata
	var wrapped struct {
		Outages     []Outage `json:"outages"`
		Data        []Outage `json:"data"`
		HasNextPage bool     `json:"hasNextPage"`
	}
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return outageListResponse{}, err
	}

	if len(wrapped.Outages) > 0 {
		return outageListResponse{Outages: wrapped.Outages, HasNextPage: wrapped.HasNextPage}, nil
	}
	if len(wrapped.Data) > 0 {
		return outageListResponse{Outages: wrapped.Data, HasNextPage: wrapped.HasNextPage}, nil
	}

	// Empty response
	return outageListResponse{Outages: []Outage{}, HasNextPage: false}, nil
}

// GetOutage retrieves a single outage by UUID.
// API: ... /v2/outages/{uuid}
func (c *Client) GetOutage(ctx context.Context, uuid string) (*Outage, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("GetOutage: %w", err)
	}
	path := fmt.Sprintf("%s/%s", OutagesBasePath, uuid)

	// API GET returns wrapped response: {"outage":{...}}
	var getResp struct {
		Outage Outage `json:"outage"`
	}
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &getResp); err != nil {
		return nil, fmt.Errorf("failed to get outage: %w", err)
	}

	return &getResp.Outage, nil
}

// maxPaginationPages is a safety limit to prevent infinite pagination loops.
const maxOutagePaginationPages = 100

// ListOutages retrieves all outages, paginating through all pages.
// API: ... /v2/outages
//
// The response can vary in format:
//   - Direct array: [{...}, {...}]
//   - Wrapped in "outages": {"outages": [{...}], "hasNextPage": bool}
//   - Wrapped in "data": {"data": [{...}]}
func (c *Client) ListOutages(ctx context.Context) ([]Outage, error) {
	var allOutages []Outage

	for page := 0; page < maxOutagePaginationPages; page++ {
		path := OutagesBasePath + "?page=" + strconv.Itoa(page)

		var rawResponse json.RawMessage
		if err := c.doRequest(ctx, http.MethodGet, path, nil, &rawResponse); err != nil {
			return nil, fmt.Errorf("failed to list outages (page %d): %w", page, err)
		}

		result, err := parseOutageListResponse(rawResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to parse outages response (page %d): %w", page, err)
		}

		allOutages = append(allOutages, result.Outages...)

		if !result.HasNextPage {
			break
		}
	}

	if allOutages == nil {
		allOutages = []Outage{}
	}

	return allOutages, nil
}

// CreateOutage creates a manual outage.
// API: ... /v2/outages
//
// Manual outages are used to track incidents that don't originate from monitor failures.
// They must be manually resolved via the ResolveOutage endpoint.
func (c *Client) CreateOutage(ctx context.Context, req CreateOutageRequest) (*Outage, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("CreateOutage: %w", err)
	}

	// API POST returns wrapped response: {"message":"Incident created","outage":{...}}
	var createResp struct {
		Message string `json:"message"`
		Outage  Outage `json:"outage"`
	}
	if err := c.doRequest(ctx, http.MethodPost, OutagesBasePath, req, &createResp); err != nil {
		return nil, fmt.Errorf("failed to create outage: %w", err)
	}

	return &createResp.Outage, nil
}

// AcknowledgeOutage acknowledges an outage.
// API: ... /v2/outages/{uuid}/acknowledge
func (c *Client) AcknowledgeOutage(ctx context.Context, uuid string) (*OutageAction, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("AcknowledgeOutage: %w", err)
	}
	path := fmt.Sprintf("%s/%s/acknowledge", OutagesBasePath, uuid)

	var response OutageAction
	if err := c.doRequest(ctx, http.MethodPost, path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to acknowledge outage: %w", err)
	}

	return &response, nil
}

// UnacknowledgeOutage removes acknowledgement from an outage.
// API: ... /v2/outages/{uuid}/unacknowledge
func (c *Client) UnacknowledgeOutage(ctx context.Context, uuid string) (*OutageAction, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("UnacknowledgeOutage: %w", err)
	}
	path := fmt.Sprintf("%s/%s/unacknowledge", OutagesBasePath, uuid)

	var response OutageAction
	if err := c.doRequest(ctx, http.MethodPost, path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to unacknowledge outage: %w", err)
	}

	return &response, nil
}

// ResolveOutage manually resolves an outage.
// API: ... /v2/outages/{uuid}/resolve
func (c *Client) ResolveOutage(ctx context.Context, uuid string) (*OutageAction, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("ResolveOutage: %w", err)
	}
	path := fmt.Sprintf("%s/%s/resolve", OutagesBasePath, uuid)

	var response OutageAction
	if err := c.doRequest(ctx, http.MethodPost, path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to resolve outage: %w", err)
	}

	return &response, nil
}

// EscalateOutage escalates an outage (triggers additional notifications).
// API: ... /v2/outages/{uuid}/escalate
func (c *Client) EscalateOutage(ctx context.Context, uuid string) (*OutageAction, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("EscalateOutage: %w", err)
	}
	path := fmt.Sprintf("%s/%s/escalate", OutagesBasePath, uuid)

	var response OutageAction
	if err := c.doRequest(ctx, http.MethodPost, path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to escalate outage: %w", err)
	}

	return &response, nil
}

// DeleteOutage deletes an outage.
// API: ... /v2/outages/{uuid}
func (c *Client) DeleteOutage(ctx context.Context, uuid string) error {
	if err := ValidateResourceID(uuid); err != nil {
		return fmt.Errorf("DeleteOutage: %w", err)
	}
	path := fmt.Sprintf("%s/%s", OutagesBasePath, uuid)

	if err := c.doRequest(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete outage: %w", err)
	}

	return nil
}
