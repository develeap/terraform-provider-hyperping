// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// healthchecksBasePath uses the exported constant for consistency.
var healthchecksBasePath = HealthchecksBasePath

// parseHealthcheckListResponse handles the various response formats the API might return.
func parseHealthcheckListResponse(raw json.RawMessage) ([]Healthcheck, error) {
	// Try direct array first
	var healthchecks []Healthcheck
	if err := json.Unmarshal(raw, &healthchecks); err == nil {
		return healthchecks, nil
	}

	// Try wrapped formats
	var wrapped struct {
		Healthchecks []Healthcheck `json:"healthchecks"`
		Data         []Healthcheck `json:"data"`
	}
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, err
	}

	if len(wrapped.Healthchecks) > 0 {
		return wrapped.Healthchecks, nil
	}
	if len(wrapped.Data) > 0 {
		return wrapped.Data, nil
	}

	// Empty response
	return []Healthcheck{}, nil
}

// GetHealthcheck retrieves a single healthcheck by UUID.
// API: ... /v2/healthchecks/{uuid}
func (c *Client) GetHealthcheck(ctx context.Context, uuid string) (*Healthcheck, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("GetHealthcheck: %w", err)
	}
	path := fmt.Sprintf("%s/%s", healthchecksBasePath, uuid)

	// API GET returns wrapped response: {"healthcheck":{...}}
	var getResp struct {
		Healthcheck Healthcheck `json:"healthcheck"`
	}
	if err := c.doRequest(ctx, "GET", path, nil, &getResp); err != nil {
		return nil, fmt.Errorf("failed to get healthcheck: %w", err)
	}

	return &getResp.Healthcheck, nil
}

// ListHealthchecks retrieves all healthchecks.
// API: ... /v2/healthchecks
//
// The response can vary in format:
//   - Direct array: [{...}, {...}]
//   - Wrapped in "healthchecks": {"healthchecks": [{...}]}
//   - Wrapped in "data": {"data": [{...}]}
func (c *Client) ListHealthchecks(ctx context.Context) ([]Healthcheck, error) {
	path := healthchecksBasePath

	var rawResponse json.RawMessage
	if err := c.doRequest(ctx, "GET", path, nil, &rawResponse); err != nil {
		return nil, fmt.Errorf("failed to list healthchecks: %w", err)
	}

	// Handle different response formats
	healthchecks, err := parseHealthcheckListResponse(rawResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse healthchecks response: %w", err)
	}

	return healthchecks, nil
}

// CreateHealthcheck creates a new healthcheck.
// API: ... /v2/healthchecks
func (c *Client) CreateHealthcheck(ctx context.Context, req CreateHealthcheckRequest) (*Healthcheck, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("CreateHealthcheck: %w", err)
	}

	// API POST returns wrapped response: {"message":"...","healthcheck":{...}}
	var createResp struct {
		Message     string      `json:"message"`
		Healthcheck Healthcheck `json:"healthcheck"`
	}
	if err := c.doRequest(ctx, "POST", healthchecksBasePath, req, &createResp); err != nil {
		return nil, fmt.Errorf("failed to create healthcheck: %w", err)
	}

	return &createResp.Healthcheck, nil
}

// UpdateHealthcheck updates an existing healthcheck.
// API: ... /v2/healthchecks/{uuid}
func (c *Client) UpdateHealthcheck(ctx context.Context, uuid string, req UpdateHealthcheckRequest) (*Healthcheck, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("UpdateHealthcheck: %w", err)
	}
	path := fmt.Sprintf("%s/%s", healthchecksBasePath, uuid)

	var healthcheck Healthcheck
	if err := c.doRequest(ctx, "PUT", path, req, &healthcheck); err != nil {
		return nil, fmt.Errorf("failed to update healthcheck: %w", err)
	}
	return &healthcheck, nil
}

// DeleteHealthcheck deletes a healthcheck.
// API: ... /v2/healthchecks/{uuid}
func (c *Client) DeleteHealthcheck(ctx context.Context, uuid string) error {
	if err := ValidateResourceID(uuid); err != nil {
		return fmt.Errorf("DeleteHealthcheck: %w", err)
	}
	path := fmt.Sprintf("%s/%s", healthchecksBasePath, uuid)

	if err := c.doRequest(ctx, "DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete healthcheck: %w", err)
	}
	return nil
}

// PauseHealthcheck pauses a healthcheck.
// API: ... /v2/healthchecks/{uuid}/pause
func (c *Client) PauseHealthcheck(ctx context.Context, uuid string) (*HealthcheckAction, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("PauseHealthcheck: %w", err)
	}
	path := fmt.Sprintf("%s/%s/pause", healthchecksBasePath, uuid)

	var action HealthcheckAction
	if err := c.doRequest(ctx, "POST", path, nil, &action); err != nil {
		return nil, fmt.Errorf("failed to pause healthcheck: %w", err)
	}
	return &action, nil
}

// ResumeHealthcheck resumes a paused healthcheck.
// API: ... /v2/healthchecks/{uuid}/resume
func (c *Client) ResumeHealthcheck(ctx context.Context, uuid string) (*HealthcheckAction, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("ResumeHealthcheck: %w", err)
	}
	path := fmt.Sprintf("%s/%s/resume", healthchecksBasePath, uuid)

	var action HealthcheckAction
	if err := c.doRequest(ctx, "POST", path, nil, &action); err != nil {
		return nil, fmt.Errorf("failed to resume healthcheck: %w", err)
	}
	return &action, nil
}
