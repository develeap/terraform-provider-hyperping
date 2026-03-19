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

// maxMaintenancePaginationPages is a safety limit to prevent infinite pagination loops.
const maxMaintenancePaginationPages = 100

// ListMaintenance returns all maintenance windows, paginating through all pages.
func (c *Client) ListMaintenance(ctx context.Context) ([]Maintenance, error) {
	var allMaintenance []Maintenance

	for page := 0; page < maxMaintenancePaginationPages; page++ {
		path := MaintenanceBasePath + "?page=" + strconv.Itoa(page)

		var rawResponse json.RawMessage
		if err := c.doRequest(ctx, http.MethodGet, path, nil, &rawResponse); err != nil {
			return nil, fmt.Errorf("failed to list maintenance windows (page %d): %w", page, err)
		}

		result, err := parseMaintenanceListResponse(rawResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to parse maintenance response (page %d): %w", page, err)
		}

		allMaintenance = append(allMaintenance, result.Maintenance...)

		if !result.HasNextPage {
			break
		}
	}

	if allMaintenance == nil {
		allMaintenance = []Maintenance{}
	}

	return allMaintenance, nil
}

// maintenanceListResponse holds the parsed maintenance list along with pagination metadata.
type maintenanceListResponse struct {
	Maintenance []Maintenance
	HasNextPage bool
}

// parseMaintenanceListResponse handles the various response formats the API might return.
// The API typically returns: {"maintenanceWindows": [...], "hasNextPage": bool, "total": int}
func parseMaintenanceListResponse(raw json.RawMessage) (maintenanceListResponse, error) {
	// Try direct array first (no pagination metadata available)
	var maintenance []Maintenance
	if err := json.Unmarshal(raw, &maintenance); err == nil {
		return maintenanceListResponse{Maintenance: maintenance, HasNextPage: false}, nil
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
		return maintenanceListResponse{}, err
	}

	// API uses "maintenanceWindows" as the standard key
	if len(wrapped.MaintenanceWindows) > 0 {
		return maintenanceListResponse{Maintenance: wrapped.MaintenanceWindows, HasNextPage: wrapped.HasNextPage}, nil
	}
	if len(wrapped.Data) > 0 {
		return maintenanceListResponse{Maintenance: wrapped.Data, HasNextPage: wrapped.HasNextPage}, nil
	}
	if len(wrapped.Maintenance) > 0 {
		return maintenanceListResponse{Maintenance: wrapped.Maintenance, HasNextPage: wrapped.HasNextPage}, nil
	}

	// Empty response - check if it's wrapped but empty
	// This handles {"maintenanceWindows": [], "hasNextPage": false, "total": 0}
	return maintenanceListResponse{Maintenance: []Maintenance{}, HasNextPage: false}, nil
}

// GetMaintenance returns a single maintenance window by UUID.
func (c *Client) GetMaintenance(ctx context.Context, uuid string) (*Maintenance, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("GetMaintenance: %w", err)
	}
	var maintenance Maintenance
	path := fmt.Sprintf("%s/%s", MaintenanceBasePath, uuid)
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &maintenance); err != nil {
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

	// API POST now returns the complete maintenance object (server-side fix, Bug #18)
	var maintenance Maintenance
	if err := c.doRequest(ctx, http.MethodPost, MaintenanceBasePath, req, &maintenance); err != nil {
		return nil, fmt.Errorf("failed to create maintenance: %w", err)
	}

	return &maintenance, nil
}

// UpdateMaintenance updates an existing maintenance window.
func (c *Client) UpdateMaintenance(ctx context.Context, uuid string, req UpdateMaintenanceRequest) (*Maintenance, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("UpdateMaintenance: %w", err)
	}
	var maintenance Maintenance
	path := fmt.Sprintf("%s/%s", MaintenanceBasePath, uuid)
	if err := c.doRequest(ctx, http.MethodPut, path, req, &maintenance); err != nil {
		return nil, fmt.Errorf("failed to update maintenance %s: %w", uuid, err)
	}
	return &maintenance, nil
}

// DeleteMaintenance deletes a maintenance window.
func (c *Client) DeleteMaintenance(ctx context.Context, uuid string) error {
	if err := ValidateResourceID(uuid); err != nil {
		return fmt.Errorf("DeleteMaintenance: %w", err)
	}
	path := fmt.Sprintf("%s/%s", MaintenanceBasePath, uuid)
	if err := c.doRequest(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete maintenance %s: %w", uuid, err)
	}
	return nil
}
