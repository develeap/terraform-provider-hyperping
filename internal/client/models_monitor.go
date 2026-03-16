// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"encoding/json"
	"fmt"
)

// =============================================================================
// Monitor Models
// =============================================================================

// escalationPolicyShape is used internally to unmarshal escalation_policy from
// the read API, which returns an object {"uuid":"...","name":"..."} rather than
// a plain string.
type escalationPolicyShape struct {
	UUID string `json:"uuid"`
	Name string `json:"name,omitempty"`
}

// Monitor represents a Hyperping monitor.
// API: GET /v1/monitors, GET /v1/monitors/{uuid}
//
// The escalation_policy field is polymorphic: the read API returns an object
// {"uuid":"...","name":"..."}, while POST/PUT send a plain UUID string.
// UnmarshalJSON handles both shapes and normalises to a UUID string.
type Monitor struct {
	ID                 int             `json:"id"` // v1 numeric ID (used by status page renderer)
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
	DNSRecordType      *string         `json:"dns_record_type,omitempty"`
	DNSNameserver      *string         `json:"dns_nameserver,omitempty"`
	DNSExpectedAnswer  *string         `json:"dns_expected_answer,omitempty"`
	Status             string          `json:"status,omitempty"`
	SSLExpiration      *int            `json:"ssl_expiration,omitempty"`
}

// monitorAlias is used to prevent infinite recursion in Monitor.UnmarshalJSON.
type monitorAlias Monitor

// monitorWire is the raw JSON shape for Monitor, with escalation_policy as a
// raw message so we can handle both the string and object forms.
type monitorWire struct {
	monitorAlias
	EscalationPolicy json.RawMessage `json:"escalation_policy,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler for Monitor.
// It handles the escalation_policy field being either a plain UUID string or an
// object {"uuid":"...","name":"..."} as returned by the Hyperping read API.
func (m *Monitor) UnmarshalJSON(data []byte) error {
	var wire monitorWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}

	*m = Monitor(wire.monitorAlias)

	if len(wire.EscalationPolicy) == 0 || string(wire.EscalationPolicy) == "null" {
		m.EscalationPolicy = nil
		return nil
	}

	// Try plain string first (e.g. "policy_uuid_here")
	var uuidStr string
	if err := json.Unmarshal(wire.EscalationPolicy, &uuidStr); err == nil {
		if uuidStr == "" {
			m.EscalationPolicy = nil
		} else {
			m.EscalationPolicy = &uuidStr
		}
		return nil
	}

	// Try object form {"uuid":"...","name":"..."}
	var obj escalationPolicyShape
	if err := json.Unmarshal(wire.EscalationPolicy, &obj); err == nil {
		if obj.UUID == "" {
			m.EscalationPolicy = nil
		} else {
			m.EscalationPolicy = &obj.UUID
		}
		return nil
	}

	return fmt.Errorf("cannot unmarshal escalation_policy: %s", string(wire.EscalationPolicy))
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
	DNSRecordType      *string         `json:"dns_record_type,omitempty"`
	DNSNameserver      *string         `json:"dns_nameserver,omitempty"`
	DNSExpectedAnswer  *string         `json:"dns_expected_answer,omitempty"`
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
	DNSRecordType      *string          `json:"dns_record_type,omitempty"`
	DNSNameserver      *string          `json:"dns_nameserver,omitempty"`
	DNSExpectedAnswer  *string          `json:"dns_expected_answer,omitempty"`
}
