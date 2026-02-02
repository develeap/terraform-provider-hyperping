// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

// =============================================================================
// Incident Models
// =============================================================================

// Incident represents a Hyperping incident.
// API: GET /v3/incidents, GET /v3/incidents/{uuid}
type Incident struct {
	UUID               string           `json:"uuid"`
	Date               string           `json:"date,omitempty"`
	Title              LocalizedText    `json:"title"`
	Text               LocalizedText    `json:"text"`
	Type               string           `json:"type"`
	AffectedComponents []string         `json:"affectedComponents,omitempty"`
	StatusPages        []string         `json:"statuspages"`
	Updates            []IncidentUpdate `json:"updates,omitempty"`
}

// IncidentUpdate represents an update to an incident.
type IncidentUpdate struct {
	UUID string        `json:"uuid"`
	Date string        `json:"date"`
	Text LocalizedText `json:"text"`
	Type string        `json:"type"`
}

// CreateIncidentRequest represents a request to create an incident.
// API: POST /v3/incidents
type CreateIncidentRequest struct {
	Title              LocalizedText `json:"title"`
	Text               LocalizedText `json:"text"`
	Type               string        `json:"type"`
	AffectedComponents []string      `json:"affectedComponents,omitempty"`
	StatusPages        []string      `json:"statuspages"`
	Date               string        `json:"date,omitempty"`
}

// Validate checks input lengths on CreateIncidentRequest fields.
func (r CreateIncidentRequest) Validate() error {
	if err := validateLocalizedText("title", r.Title, maxNameLength); err != nil {
		return err
	}
	if err := validateLocalizedText("text", r.Text, maxMessageLength); err != nil {
		return err
	}
	return nil
}

// UpdateIncidentRequest represents a request to update an incident.
// API: PUT /v3/incidents/{uuid}
type UpdateIncidentRequest struct {
	Title              *LocalizedText `json:"title,omitempty"`
	Type               *string        `json:"type,omitempty"`
	AffectedComponents *[]string      `json:"affectedComponents,omitempty"`
	StatusPages        *[]string      `json:"statuspages,omitempty"`
}

// AddIncidentUpdateRequest represents a request to add an update to an incident.
// API: POST /v3/incidents/{uuid}/updates
type AddIncidentUpdateRequest struct {
	Text LocalizedText `json:"text"`
	Type string        `json:"type"`
	Date string        `json:"date"`
}
