// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

// =============================================================================
// Monitor Models
// =============================================================================

// Monitor represents a Hyperping monitor.
// API: GET /v1/monitors, GET /v1/monitors/{uuid}
type Monitor struct {
	UUID               string          `json:"uuid"`
	Name               string          `json:"name"`
	URL                string          `json:"url"`
	Protocol           string          `json:"protocol"`
	ProjectUUID        string          `json:"projectUuid,omitempty"`
	HTTPMethod         string          `json:"http_method"`
	Regions            []string        `json:"regions"`
	CheckFrequency     int             `json:"check_frequency"`
	RequestHeaders     []RequestHeader `json:"request_headers"`
	RequestBody        string          `json:"request_body,omitempty"`
	FollowRedirects    bool            `json:"follow_redirects"`
	ExpectedStatusCode FlexibleString  `json:"expected_status_code"`
	RequiredKeyword    *string         `json:"required_keyword,omitempty"`
	Paused             bool            `json:"paused"`
	Port               *int            `json:"port,omitempty"`
	AlertsWait         int             `json:"alerts_wait,omitempty"`
	EscalationPolicy   *string         `json:"escalation_policy,omitempty"`
	Status             string          `json:"status,omitempty"`         // up, down (read-only)
	SSLExpiration      *int            `json:"ssl_expiration,omitempty"` // Days until SSL cert expiration (read-only)
}

// CreateMonitorRequest represents a request to create a monitor.
// API: POST /v1/monitors
type CreateMonitorRequest struct {
	Name               string          `json:"name"`
	URL                string          `json:"url"`
	Protocol           string          `json:"protocol"`
	ProjectUUID        string          `json:"projectUuid,omitempty"`
	HTTPMethod         string          `json:"http_method,omitempty"`
	CheckFrequency     int             `json:"check_frequency,omitempty"`
	Regions            []string        `json:"regions,omitempty"`
	RequestHeaders     []RequestHeader `json:"request_headers,omitempty"`
	RequestBody        *string         `json:"request_body,omitempty"`
	FollowRedirects    *bool           `json:"follow_redirects,omitempty"`
	ExpectedStatusCode string          `json:"expected_status_code,omitempty"`
	RequiredKeyword    *string         `json:"required_keyword,omitempty"`
	Paused             bool            `json:"paused,omitempty"`
	Port               *int            `json:"port,omitempty"`
	AlertsWait         *int            `json:"alerts_wait,omitempty"`
	EscalationPolicy   *string         `json:"escalation_policy,omitempty"`
}

// Validate checks input lengths on CreateMonitorRequest fields.
func (r CreateMonitorRequest) Validate() error {
	if err := validateStringLength("name", r.Name, maxNameLength); err != nil {
		return err
	}
	if err := validateStringLength("url", r.URL, maxURLLength); err != nil {
		return err
	}
	return nil
}

// UpdateMonitorRequest represents a request to update a monitor.
// API: PUT /v1/monitors/{uuid}
type UpdateMonitorRequest struct {
	Name               *string          `json:"name,omitempty"`
	URL                *string          `json:"url,omitempty"`
	Protocol           *string          `json:"protocol,omitempty"`
	ProjectUUID        *string          `json:"projectUuid,omitempty"`
	HTTPMethod         *string          `json:"http_method,omitempty"`
	CheckFrequency     *int             `json:"check_frequency,omitempty"`
	Regions            *[]string        `json:"regions,omitempty"`
	RequestHeaders     *[]RequestHeader `json:"request_headers,omitempty"`
	RequestBody        *string          `json:"request_body,omitempty"`
	FollowRedirects    *bool            `json:"follow_redirects,omitempty"`
	ExpectedStatusCode *string          `json:"expected_status_code,omitempty"`
	RequiredKeyword    *string          `json:"required_keyword,omitempty"`
	Paused             *bool            `json:"paused,omitempty"`
	Port               *int             `json:"port,omitempty"`
	AlertsWait         *int             `json:"alerts_wait,omitempty"`
	EscalationPolicy   *string          `json:"escalation_policy,omitempty"`
}
