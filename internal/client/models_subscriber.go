// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import "fmt"

// =============================================================================
// Status Page Subscriber Models
// =============================================================================

// StatusPageSubscriber represents a subscriber to a status page.
// API: ... /v2/statuspages/{uuid}/subscribers
type StatusPageSubscriber struct {
	ID           int     `json:"id"`
	Type         string  `json:"type"` // email, sms, slack, teams
	Value        string  `json:"value"`
	Language     string  `json:"language"` // ISO 639-1 language code
	Email        *string `json:"email"`
	Phone        *string `json:"phone"`
	SlackChannel *string `json:"slack_channel"`
	CreatedAt    string  `json:"created_at"` // ISO 8601 timestamp
}

// SubscriberPaginatedResponse represents a paginated list of subscribers.
// API: ... /v2/statuspages/{uuid}/subscribers with pagination
type SubscriberPaginatedResponse struct {
	Subscribers    []StatusPageSubscriber `json:"subscribers"`
	HasNextPage    bool                   `json:"hasNextPage"`
	Total          int                    `json:"total"`
	Page           int                    `json:"page"`
	ResultsPerPage int                    `json:"resultsPerPage"`
}

// AddSubscriberRequest represents a request to add a subscriber.
// API: ... /v2/statuspages/{uuid}/subscribers
type AddSubscriberRequest struct {
	Type            string  `json:"type"`                        // email, sms, teams (NOT slack - must use OAuth)
	Email           *string `json:"email,omitempty"`             // required if type=email
	Phone           *string `json:"phone,omitempty"`             // required if type=sms
	TeamsWebhookURL *string `json:"teams_webhook_url,omitempty"` // required if type=teams
	Language        *string `json:"language,omitempty"`          // optional, default: en
}

// Validate checks that the required fields are present based on the subscriber type.
func (r AddSubscriberRequest) Validate() error {
	// Validate type enum
	validType := false
	for _, t := range AllowedSubscriberTypes {
		if r.Type == t {
			validType = true
			break
		}
	}
	if !validType {
		return fmt.Errorf("invalid subscriber type %q, must be one of: %v", r.Type, AllowedSubscriberTypes)
	}

	// Validate conditional requirements
	switch r.Type {
	case "email":
		if r.Email == nil || *r.Email == "" {
			return fmt.Errorf("email is required when type is 'email'")
		}
	case "sms":
		if r.Phone == nil || *r.Phone == "" {
			return fmt.Errorf("phone is required when type is 'sms'")
		}
	case "teams":
		if r.TeamsWebhookURL == nil || *r.TeamsWebhookURL == "" {
			return fmt.Errorf("teams_webhook_url is required when type is 'teams'")
		}
	}

	// Validate language if provided
	if r.Language != nil {
		validLang := false
		for _, lang := range AllowedLanguages {
			if *r.Language == lang {
				validLang = true
				break
			}
		}
		if !validLang {
			return fmt.Errorf("invalid language %q, must be one of: %v", *r.Language, AllowedLanguages)
		}
	}

	return nil
}
