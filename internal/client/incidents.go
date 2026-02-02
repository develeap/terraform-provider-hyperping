// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
)

const incidentsBasePath = "/v3/incidents"

// ListIncidents returns all incidents.
func (c *Client) ListIncidents(ctx context.Context) ([]Incident, error) {
	var rawResponse json.RawMessage
	if err := c.doRequest(ctx, "GET", incidentsBasePath, nil, &rawResponse); err != nil {
		return nil, fmt.Errorf("failed to list incidents: %w", err)
	}

	// Handle different response formats
	incidents, err := parseIncidentListResponse(rawResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse incidents response: %w", err)
	}

	return incidents, nil
}

// parseIncidentListResponse handles the various response formats the API might return.
func parseIncidentListResponse(raw json.RawMessage) ([]Incident, error) {
	// Try direct array first
	var incidents []Incident
	if err := json.Unmarshal(raw, &incidents); err == nil {
		return incidents, nil
	}

	// Try wrapped formats
	var wrapped struct {
		Incidents []Incident `json:"incidents"`
		Data      []Incident `json:"data"`
	}
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, err
	}

	if len(wrapped.Incidents) > 0 {
		return wrapped.Incidents, nil
	}
	if len(wrapped.Data) > 0 {
		return wrapped.Data, nil
	}

	// Empty response
	return []Incident{}, nil
}

// GetIncident returns a single incident by UUID.
func (c *Client) GetIncident(ctx context.Context, uuid string) (*Incident, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("GetIncident: %w", err)
	}
	var incident Incident
	path := fmt.Sprintf("%s/%s", incidentsBasePath, uuid)
	if err := c.doRequest(ctx, "GET", path, nil, &incident); err != nil {
		return nil, fmt.Errorf("failed to get incident %s: %w", uuid, err)
	}
	return &incident, nil
}

// CreateIncident creates a new incident.
func (c *Client) CreateIncident(ctx context.Context, req CreateIncidentRequest) (*Incident, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("CreateIncident: %w", err)
	}
	var incident Incident
	if err := c.doRequest(ctx, "POST", incidentsBasePath, req, &incident); err != nil {
		return nil, fmt.Errorf("failed to create incident: %w", err)
	}
	return &incident, nil
}

// UpdateIncident updates an existing incident.
func (c *Client) UpdateIncident(ctx context.Context, uuid string, req UpdateIncidentRequest) (*Incident, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("UpdateIncident: %w", err)
	}
	var incident Incident
	path := fmt.Sprintf("%s/%s", incidentsBasePath, uuid)
	if err := c.doRequest(ctx, "PUT", path, req, &incident); err != nil {
		return nil, fmt.Errorf("failed to update incident %s: %w", uuid, err)
	}
	return &incident, nil
}

// AddIncidentUpdate adds an update to an existing incident.
func (c *Client) AddIncidentUpdate(ctx context.Context, uuid string, req AddIncidentUpdateRequest) (*Incident, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("AddIncidentUpdate: %w", err)
	}
	var incident Incident
	path := fmt.Sprintf("%s/%s/updates", incidentsBasePath, uuid)
	if err := c.doRequest(ctx, "POST", path, req, &incident); err != nil {
		return nil, fmt.Errorf("failed to add update to incident %s: %w", uuid, err)
	}
	return &incident, nil
}

// ResolveIncident resolves an incident by adding a "resolved" update.
// In the Hyperping v3 API, incidents are resolved by adding an update with type "resolved".
func (c *Client) ResolveIncident(ctx context.Context, uuid string, message string) (*Incident, error) {
	return c.AddIncidentUpdate(ctx, uuid, AddIncidentUpdateRequest{
		Text: LocalizedText{En: message},
		Type: "resolved",
		Date: "", // API will use current time if empty
	})
}

// DeleteIncident deletes an incident.
func (c *Client) DeleteIncident(ctx context.Context, uuid string) error {
	if err := ValidateResourceID(uuid); err != nil {
		return fmt.Errorf("DeleteIncident: %w", err)
	}
	path := fmt.Sprintf("%s/%s", incidentsBasePath, uuid)
	if err := c.doRequest(ctx, "DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete incident %s: %w", uuid, err)
	}
	return nil
}
