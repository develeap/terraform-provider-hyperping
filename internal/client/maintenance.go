// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// maintenanceBasePath uses the exported constant for consistency.
var maintenanceBasePath = MaintenanceBasePath

// ListMaintenance returns all maintenance windows.
func (c *Client) ListMaintenance(ctx context.Context) ([]Maintenance, error) {
	var rawResponse json.RawMessage
	if err := c.doRequest(ctx, "GET", maintenanceBasePath, nil, &rawResponse); err != nil {
		return nil, fmt.Errorf("failed to list maintenance windows: %w", err)
	}

	// Handle different response formats
	maintenance, err := parseMaintenanceListResponse(rawResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse maintenance response: %w", err)
	}

	return maintenance, nil
}

// parseMaintenanceListResponse handles the various response formats the API might return.
// The API typically returns: {"maintenanceWindows": [...], "hasNextPage": bool, "total": int}
func parseMaintenanceListResponse(raw json.RawMessage) ([]Maintenance, error) {
	// Try direct array first
	var maintenance []Maintenance
	if err := json.Unmarshal(raw, &maintenance); err == nil {
		return maintenance, nil
	}

	// Try wrapped formats with pagination metadata
	var wrapped struct {
		MaintenanceWindows []Maintenance `json:"maintenanceWindows"`
		Data               []Maintenance `json:"data"`
		Maintenance        []Maintenance `json:"maintenance"`
		HasNextPage        bool          `json:"hasNextPage,omitempty"`
		Total              int           `json:"total,omitempty"`
	}
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, err
	}

	// API uses "maintenanceWindows" as the standard key
	if len(wrapped.MaintenanceWindows) > 0 {
		return wrapped.MaintenanceWindows, nil
	}
	if len(wrapped.Data) > 0 {
		return wrapped.Data, nil
	}
	if len(wrapped.Maintenance) > 0 {
		return wrapped.Maintenance, nil
	}

	// Empty response - check if it's wrapped but empty
	// This handles {"maintenanceWindows": [], "hasNextPage": false, "total": 0}
	return []Maintenance{}, nil
}

// GetMaintenance returns a single maintenance window by UUID.
func (c *Client) GetMaintenance(ctx context.Context, uuid string) (*Maintenance, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("GetMaintenance: %w", err)
	}
	var maintenance Maintenance
	path := fmt.Sprintf("%s/%s", maintenanceBasePath, uuid)
	if err := c.doRequest(ctx, "GET", path, nil, &maintenance); err != nil {
		return nil, fmt.Errorf("failed to get maintenance %s: %w", uuid, err)
	}
	return &maintenance, nil
}

// CreateMaintenance creates a new maintenance window.
// API POST returns minimal response: {"uuid": "mw_xxx"}
// We follow up with a GET to retrieve the full maintenance details.
func (c *Client) CreateMaintenance(ctx context.Context, req CreateMaintenanceRequest) (*Maintenance, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("CreateMaintenance: %w", err)
	}

	// API POST returns only UUID, not full object
	var createResp struct {
		UUID string `json:"uuid"`
	}
	if err := c.doRequest(ctx, "POST", maintenanceBasePath, req, &createResp); err != nil {
		return nil, fmt.Errorf("failed to create maintenance: %w", err)
	}

	// Read full maintenance details after creation
	maintenance, err := c.GetMaintenance(ctx, createResp.UUID)
	if err != nil {
		return nil, fmt.Errorf("failed to read created maintenance %s: %w", createResp.UUID, err)
	}

	return maintenance, nil
}

// UpdateMaintenance updates an existing maintenance window.
func (c *Client) UpdateMaintenance(ctx context.Context, uuid string, req UpdateMaintenanceRequest) (*Maintenance, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("UpdateMaintenance: %w", err)
	}
	var maintenance Maintenance
	path := fmt.Sprintf("%s/%s", maintenanceBasePath, uuid)
	if err := c.doRequest(ctx, "PUT", path, req, &maintenance); err != nil {
		return nil, fmt.Errorf("failed to update maintenance %s: %w", uuid, err)
	}
	return &maintenance, nil
}

// DeleteMaintenance deletes a maintenance window.
func (c *Client) DeleteMaintenance(ctx context.Context, uuid string) error {
	if err := ValidateResourceID(uuid); err != nil {
		return fmt.Errorf("DeleteMaintenance: %w", err)
	}
	path := fmt.Sprintf("%s/%s", maintenanceBasePath, uuid)
	if err := c.doRequest(ctx, "DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete maintenance %s: %w", uuid, err)
	}
	return nil
}
