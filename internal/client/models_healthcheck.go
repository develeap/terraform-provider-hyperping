// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

// =============================================================================
// Healthcheck Models
// =============================================================================

// Healthcheck represents a Hyperping healthcheck (cron job monitoring).
// API: GET /v2/healthchecks, GET /v2/healthchecks/{uuid}
//
// Healthchecks monitor cron jobs via ping URLs. When the job runs, it pings
// the URL. If no ping is received within the expected period + grace period,
// an alert is triggered.
type Healthcheck struct {
	UUID             string                     `json:"uuid"` // tok_abc123def456
	Name             string                     `json:"name"`
	PingURL          string                     `json:"pingUrl"`                    // Auto-generated ping URL
	Cron             string                     `json:"cron,omitempty"`             // Cron expression (e.g., "0 0 * * *")
	Timezone         string                     `json:"timezone,omitempty"`         // Timezone (e.g., "America/New_York")
	PeriodValue      *int                       `json:"periodValue,omitempty"`      // Numeric value for period
	PeriodType       string                     `json:"periodType,omitempty"`       // seconds, minutes, hours, days
	Period           int                        `json:"period"`                     // Calculated period in seconds
	GracePeriod      int                        `json:"gracePeriod"`                // Grace period in seconds
	GracePeriodValue int                        `json:"gracePeriodValue"`           // Numeric value for grace period
	GracePeriodType  string                     `json:"gracePeriodType"`            // seconds, minutes, hours, days
	IsDown           bool                       `json:"isDown"`                     // Current status (read-only)
	IsPaused         bool                       `json:"isPaused"`                   // Whether paused (read-only)
	LastPing         string                     `json:"lastPing,omitempty"`         // ISO 8601 timestamp (read-only)
	DueDate          string                     `json:"dueDate,omitempty"`          // ISO 8601 timestamp (read-only)
	CreatedAt        string                     `json:"createdAt,omitempty"`        // ISO 8601 timestamp (read-only)
	LastLogStartDate string                     `json:"lastLogStartDate,omitempty"` // ISO 8601 timestamp (read-only)
	LastLogEndDate   string                     `json:"lastLogEndDate,omitempty"`   // ISO 8601 timestamp (read-only)
	EscalationPolicy *EscalationPolicyReference `json:"escalationPolicy,omitempty"` // Linked escalation policy
}

// CreateHealthcheckRequest represents a request to create a healthcheck.
// API: POST /v2/healthchecks
// Note: API accepts snake_case in requests but returns camelCase in responses
type CreateHealthcheckRequest struct {
	Name             string  `json:"name"`
	Cron             *string `json:"cron,omitempty"`         // Required if PeriodValue/PeriodType not set
	Timezone         *string `json:"timezone,omitempty"`     // Required if Cron is set
	PeriodValue      *int    `json:"period_value,omitempty"` // Required if Cron not set
	PeriodType       *string `json:"period_type,omitempty"`  // Required if Cron not set
	GracePeriodValue int     `json:"grace_period_value"`
	GracePeriodType  string  `json:"grace_period_type"`           // seconds, minutes, hours, days
	EscalationPolicy *string `json:"escalation_policy,omitempty"` // UUID of escalation policy
}

// Validate checks input lengths on CreateHealthcheckRequest fields.
func (r CreateHealthcheckRequest) Validate() error {
	if err := validateStringLength("name", r.Name, maxNameLength); err != nil {
		return err
	}
	return nil
}

// UpdateHealthcheckRequest represents a request to update a healthcheck.
// API: PUT /v2/healthchecks/{uuid}
// Note: API accepts snake_case in requests but returns camelCase in responses
type UpdateHealthcheckRequest struct {
	Name             *string `json:"name,omitempty"`
	Cron             *string `json:"cron,omitempty"`
	Timezone         *string `json:"timezone,omitempty"`
	PeriodValue      *int    `json:"period_value,omitempty"`
	PeriodType       *string `json:"period_type,omitempty"`
	GracePeriodValue *int    `json:"grace_period_value,omitempty"`
	GracePeriodType  *string `json:"grace_period_type,omitempty"`
	EscalationPolicy *string `json:"escalation_policy,omitempty"`
}

// HealthcheckAction represents an action performed on a healthcheck.
// Used for pause, resume responses.
type HealthcheckAction struct {
	Message string `json:"message"`
	UUID    string `json:"uuid"`
}
