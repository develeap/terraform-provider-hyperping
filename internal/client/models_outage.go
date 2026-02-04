// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

// =============================================================================
// Outage Models
// =============================================================================

// Outage represents a Hyperping outage (system-generated monitor failure).
// API: GET /v1/outages, GET /v1/outages/:uuid
//
// Outages are different from Incidents:
// - Outages are automatically detected monitor failures (system-generated)
// - Incidents are manually created status page notifications (user-created)
//
// Outage Types:
// - monitor: Automatically resolved when monitor recovers
// - manual: Must be manually resolved via API or dashboard
type Outage struct {
	UUID               string                     `json:"uuid"`
	StartDate          string                     `json:"startDate"`
	EndDate            *string                    `json:"endDate"`            // null if ongoing
	DurationMs         int                        `json:"durationMs"`         // Duration in milliseconds
	StatusCode         int                        `json:"statusCode"`         // HTTP status code
	Description        string                     `json:"description"`        // Error message
	OutageType         string                     `json:"outageType"`         // "manual" or "automatic"
	IsResolved         bool                       `json:"isResolved"`         // Whether resolved
	DetectedLocation   string                     `json:"detectedLocation"`   // Location that detected outage
	ConfirmedLocations string                     `json:"confirmedLocations"` // Comma-separated locations
	AcknowledgedAt     *string                    `json:"acknowledgedAt"`     // ISO 8601 timestamp
	AcknowledgedBy     *AcknowledgedByUser        `json:"acknowledgedBy"`     // User who acknowledged
	Monitor            MonitorReference           `json:"monitor"`            // Associated monitor
	EscalationPolicy   *EscalationPolicyReference `json:"escalationPolicy"`   // Linked escalation policy
}

// AcknowledgedByUser represents the user who acknowledged an outage.
type AcknowledgedByUser struct {
	UUID  string `json:"uuid"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// MonitorReference is a simplified monitor object included in outage responses.
type MonitorReference struct {
	UUID     string `json:"uuid"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Protocol string `json:"protocol"`
}

// EscalationPolicyReference is a simplified escalation policy object.
type EscalationPolicyReference struct {
	UUID         string `json:"uuid"`
	Name         string `json:"name"`
	AlertedSteps int    `json:"alertedSteps"`
	TotalSteps   int    `json:"totalSteps"`
}

// OutageAction represents an action performed on an outage.
// Used for acknowledge, resolve, escalate responses.
type OutageAction struct {
	Message string `json:"message"`
	UUID    string `json:"uuid"`
}

// CreateOutageRequest represents a request to create a manual outage.
// API: POST /v1/outages
type CreateOutageRequest struct {
	MonitorUUID          string  `json:"monitorUuid"`
	StartDate            string  `json:"startDate"`                      // ISO 8601 format
	EndDate              *string `json:"endDate,omitempty"`              // ISO 8601 format, optional
	StatusCode           int     `json:"statusCode"`
	Description          string  `json:"description"`
	OutageType           string  `json:"outageType"`                     // "manual" for manually created outages
	EscalationPolicyUuid *string `json:"escalationPolicyUuid,omitempty"` // UUID of escalation policy to trigger
}

// Validate checks input lengths on CreateOutageRequest fields.
func (r CreateOutageRequest) Validate() error {
	if err := validateStringLength("description", r.Description, maxMessageLength); err != nil {
		return err
	}
	return nil
}
