// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
)

const outagesBasePath = "/v1/outages"

// parseOutageListResponse handles the various response formats the API might return.
func parseOutageListResponse(raw json.RawMessage) ([]Outage, error) {
	// Try direct array first
	var outages []Outage
	if err := json.Unmarshal(raw, &outages); err == nil {
		return outages, nil
	}

	// Try wrapped formats
	var wrapped struct {
		Outages []Outage `json:"outages"`
		Data    []Outage `json:"data"`
	}
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, err
	}

	if len(wrapped.Outages) > 0 {
		return wrapped.Outages, nil
	}
	if len(wrapped.Data) > 0 {
		return wrapped.Data, nil
	}

	// Empty response
	return []Outage{}, nil
}

// GetOutage retrieves a single outage by UUID.
// API: GET /v1/outages/{uuid}
func (c *Client) GetOutage(ctx context.Context, uuid string) (*Outage, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("GetOutage: %w", err)
	}
	path := fmt.Sprintf("%s/%s", outagesBasePath, uuid)

	var outage Outage
	if err := c.doRequest(ctx, "GET", path, nil, &outage); err != nil {
		return nil, fmt.Errorf("failed to get outage: %w", err)
	}

	return &outage, nil
}

// ListOutages retrieves all outages.
// API: GET /v1/outages
//
// The response can vary in format:
//   - Direct array: [{...}, {...}]
//   - Wrapped in "outages": {"outages": [{...}]}
//   - Wrapped in "data": {"data": [{...}]}
func (c *Client) ListOutages(ctx context.Context) ([]Outage, error) {
	path := outagesBasePath

	var rawResponse json.RawMessage
	if err := c.doRequest(ctx, "GET", path, nil, &rawResponse); err != nil {
		return nil, fmt.Errorf("failed to list outages: %w", err)
	}

	// Handle different response formats
	outages, err := parseOutageListResponse(rawResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse outages response: %w", err)
	}

	return outages, nil
}

// CreateOutage creates a manual outage.
// API: POST /v1/outages
//
// Manual outages are used to track incidents that don't originate from monitor failures.
// They must be manually resolved via the ResolveOutage endpoint.
func (c *Client) CreateOutage(ctx context.Context, req CreateOutageRequest) (*Outage, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("CreateOutage: %w", err)
	}
	var outage Outage
	if err := c.doRequest(ctx, "POST", outagesBasePath, req, &outage); err != nil {
		return nil, fmt.Errorf("failed to create outage: %w", err)
	}
	return &outage, nil
}

// AcknowledgeOutage acknowledges an outage.
// API: POST /v1/outages/{uuid}/acknowledge
func (c *Client) AcknowledgeOutage(ctx context.Context, uuid string) (*OutageAction, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("AcknowledgeOutage: %w", err)
	}
	path := fmt.Sprintf("%s/%s/acknowledge", outagesBasePath, uuid)

	var response OutageAction
	if err := c.doRequest(ctx, "POST", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to acknowledge outage: %w", err)
	}

	return &response, nil
}

// UnacknowledgeOutage removes acknowledgement from an outage.
// API: POST /v1/outages/{uuid}/unacknowledge
func (c *Client) UnacknowledgeOutage(ctx context.Context, uuid string) (*OutageAction, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("UnacknowledgeOutage: %w", err)
	}
	path := fmt.Sprintf("%s/%s/unacknowledge", outagesBasePath, uuid)

	var response OutageAction
	if err := c.doRequest(ctx, "POST", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to unacknowledge outage: %w", err)
	}

	return &response, nil
}

// ResolveOutage manually resolves an outage.
// API: POST /v1/outages/{uuid}/resolve
func (c *Client) ResolveOutage(ctx context.Context, uuid string) (*OutageAction, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("ResolveOutage: %w", err)
	}
	path := fmt.Sprintf("%s/%s/resolve", outagesBasePath, uuid)

	var response OutageAction
	if err := c.doRequest(ctx, "POST", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to resolve outage: %w", err)
	}

	return &response, nil
}

// EscalateOutage escalates an outage (triggers additional notifications).
// API: POST /v1/outages/{uuid}/escalate
func (c *Client) EscalateOutage(ctx context.Context, uuid string) (*OutageAction, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("EscalateOutage: %w", err)
	}
	path := fmt.Sprintf("%s/%s/escalate", outagesBasePath, uuid)

	var response OutageAction
	if err := c.doRequest(ctx, "POST", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to escalate outage: %w", err)
	}

	return &response, nil
}

// DeleteOutage deletes an outage.
// API: DELETE /v1/outages/{uuid}
func (c *Client) DeleteOutage(ctx context.Context, uuid string) error {
	if err := ValidateResourceID(uuid); err != nil {
		return fmt.Errorf("DeleteOutage: %w", err)
	}
	path := fmt.Sprintf("%s/%s", outagesBasePath, uuid)

	if err := c.doRequest(ctx, "DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete outage: %w", err)
	}

	return nil
}
