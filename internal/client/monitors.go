// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
)

const monitorsBasePath = "/v1/monitors"

// ListMonitors returns all monitors.
func (c *Client) ListMonitors(ctx context.Context) ([]Monitor, error) {
	var rawResponse json.RawMessage
	if err := c.doRequest(ctx, "GET", monitorsBasePath, nil, &rawResponse); err != nil {
		return nil, fmt.Errorf("failed to list monitors: %w", err)
	}

	// Handle different response formats:
	// 1. Direct array: [{...}, {...}]
	// 2. Wrapped in "monitors": {"monitors": [{...}]}
	// 3. Wrapped in "data": {"data": [{...}]}
	monitors, err := parseMonitorListResponse(rawResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse monitors response: %w", err)
	}

	return monitors, nil
}

// parseMonitorListResponse handles the various response formats the API might return.
func parseMonitorListResponse(raw json.RawMessage) ([]Monitor, error) {
	// Try direct array first
	var monitors []Monitor
	if err := json.Unmarshal(raw, &monitors); err == nil {
		return monitors, nil
	}

	// Try wrapped formats
	var wrapped struct {
		Monitors []Monitor `json:"monitors"`
		Data     []Monitor `json:"data"`
	}
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, err
	}

	if len(wrapped.Monitors) > 0 {
		return wrapped.Monitors, nil
	}
	if len(wrapped.Data) > 0 {
		return wrapped.Data, nil
	}

	// Empty response
	return []Monitor{}, nil
}

// GetMonitor returns a single monitor by ID.
func (c *Client) GetMonitor(ctx context.Context, id string) (*Monitor, error) {
	if err := ValidateResourceID(id); err != nil {
		return nil, fmt.Errorf("GetMonitor: %w", err)
	}
	var monitor Monitor
	path := fmt.Sprintf("%s/%s", monitorsBasePath, id)
	if err := c.doRequest(ctx, "GET", path, nil, &monitor); err != nil {
		return nil, fmt.Errorf("failed to get monitor %s: %w", id, err)
	}
	return &monitor, nil
}

// CreateMonitor creates a new monitor.
func (c *Client) CreateMonitor(ctx context.Context, req CreateMonitorRequest) (*Monitor, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("CreateMonitor: %w", err)
	}
	var monitor Monitor
	if err := c.doRequest(ctx, "POST", monitorsBasePath, req, &monitor); err != nil {
		return nil, fmt.Errorf("failed to create monitor: %w", err)
	}
	return &monitor, nil
}

// UpdateMonitor updates an existing monitor.
func (c *Client) UpdateMonitor(ctx context.Context, id string, req UpdateMonitorRequest) (*Monitor, error) {
	if err := ValidateResourceID(id); err != nil {
		return nil, fmt.Errorf("UpdateMonitor: %w", err)
	}
	var monitor Monitor
	path := fmt.Sprintf("%s/%s", monitorsBasePath, id)
	if err := c.doRequest(ctx, "PUT", path, req, &monitor); err != nil {
		return nil, fmt.Errorf("failed to update monitor %s: %w", id, err)
	}
	return &monitor, nil
}

// DeleteMonitor deletes a monitor.
func (c *Client) DeleteMonitor(ctx context.Context, id string) error {
	if err := ValidateResourceID(id); err != nil {
		return fmt.Errorf("DeleteMonitor: %w", err)
	}
	path := fmt.Sprintf("%s/%s", monitorsBasePath, id)
	if err := c.doRequest(ctx, "DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete monitor %s: %w", id, err)
	}
	return nil
}

// PauseMonitor pauses a monitor.
func (c *Client) PauseMonitor(ctx context.Context, id string) (*Monitor, error) {
	paused := true
	return c.UpdateMonitor(ctx, id, UpdateMonitorRequest{Paused: &paused})
}

// ResumeMonitor resumes a paused monitor.
func (c *Client) ResumeMonitor(ctx context.Context, id string) (*Monitor, error) {
	paused := false
	return c.UpdateMonitor(ctx, id, UpdateMonitorRequest{Paused: &paused})
}
