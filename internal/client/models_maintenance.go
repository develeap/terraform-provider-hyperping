// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

// =============================================================================
// Maintenance Models
// =============================================================================

// Maintenance represents a Hyperping maintenance window.
// API: GET /v1/maintenance-windows, GET /v1/maintenance-windows/{uuid}
type Maintenance struct {
	UUID                string              `json:"uuid"`
	Name                string              `json:"name"`
	Title               LocalizedText       `json:"title,omitempty"`
	Text                LocalizedText       `json:"text,omitempty"`
	StartDate           *string             `json:"start_date,omitempty"`
	EndDate             *string             `json:"end_date,omitempty"`
	Timezone            string              `json:"timezone,omitempty"`
	Monitors            []string            `json:"monitors"`
	StatusPages         []string            `json:"statuspages,omitempty"`
	BulkUUID            *string             `json:"bulkUuid,omitempty"`
	CreatedBy           string              `json:"createdBy,omitempty"`
	CreatedAt           string              `json:"createdAt,omitempty"`
	NotificationOption  string              `json:"notificationOption,omitempty"` // none, immediate, scheduled
	NotificationMinutes *int                `json:"notificationMinutes,omitempty"`
	Status              string              `json:"status,omitempty"`  // upcoming, ongoing, completed (read-only)
	Updates             []MaintenanceUpdate `json:"updates,omitempty"` // Array of updates (read-only)
}

// MaintenanceUpdate represents an update to a maintenance window.
type MaintenanceUpdate struct {
	Text LocalizedText `json:"text"`
	Date string        `json:"date"`
}

// CreateMaintenanceRequest represents a request to create a maintenance window.
// API: POST /v1/maintenance-windows
type CreateMaintenanceRequest struct {
	Name                string        `json:"name"`
	Title               LocalizedText `json:"title,omitempty"`
	Text                LocalizedText `json:"text,omitempty"`
	StartDate           string        `json:"start_date"`
	EndDate             string        `json:"end_date"`
	Monitors            []string      `json:"monitors"`
	StatusPages         []string      `json:"statuspages,omitempty"`
	NotificationOption  string        `json:"notificationOption,omitempty"`
	NotificationMinutes *int          `json:"notificationMinutes,omitempty"`
}

// Validate checks input lengths on CreateMaintenanceRequest fields.
func (r CreateMaintenanceRequest) Validate() error {
	if err := validateStringLength("name", r.Name, maxNameLength); err != nil {
		return err
	}
	if err := validateLocalizedText("title", r.Title, maxNameLength); err != nil {
		return err
	}
	if err := validateLocalizedText("text", r.Text, maxMessageLength); err != nil {
		return err
	}
	return nil
}

// UpdateMaintenanceRequest represents a request to update a maintenance window.
// API: PUT /v1/maintenance-windows/{uuid}
type UpdateMaintenanceRequest struct {
	Name      *string   `json:"name,omitempty"`
	StartDate *string   `json:"start_date,omitempty"`
	EndDate   *string   `json:"end_date,omitempty"`
	Monitors  *[]string `json:"monitors,omitempty"`
}
