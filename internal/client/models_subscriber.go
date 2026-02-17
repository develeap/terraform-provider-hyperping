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

// validateEmailSubscriber checks that the email field is present for email-type subscribers.
func validateEmailSubscriber(req AddSubscriberRequest) error {
	if req.Email == nil || *req.Email == "" {
		return fmt.Errorf("email is required when type is 'email'")
	}
	return nil
}

// validateSMSSubscriber checks that the phone field is present for sms-type subscribers.
func validateSMSSubscriber(req AddSubscriberRequest) error {
	if req.Phone == nil || *req.Phone == "" {
		return fmt.Errorf("phone is required when type is 'sms'")
	}
	return nil
}

// validateTeamsSubscriber checks that teams_webhook_url is present for teams-type subscribers.
func validateTeamsSubscriber(req AddSubscriberRequest) error {
	if req.TeamsWebhookURL == nil || *req.TeamsWebhookURL == "" {
		return fmt.Errorf("teams_webhook_url is required when type is 'teams'")
	}
	return nil
}

// validateSubscriberType checks that the subscriber type is one of the allowed values.
func validateSubscriberType(subscriberType string) error {
	for _, t := range AllowedSubscriberTypes {
		if subscriberType == t {
			return nil
		}
	}
	return fmt.Errorf("invalid subscriber type %q, must be one of: %v", subscriberType, AllowedSubscriberTypes)
}

// validateSubscriberLanguage checks that the language code is one of the allowed values.
func validateSubscriberLanguage(language string) error {
	for _, lang := range AllowedLanguages {
		if language == lang {
			return nil
		}
	}
	return fmt.Errorf("invalid language %q, must be one of: %v", language, AllowedLanguages)
}

// typeValidators maps subscriber type names to their per-type validation functions.
var typeValidators = map[string]func(AddSubscriberRequest) error{
	"email": validateEmailSubscriber,
	"sms":   validateSMSSubscriber,
	"teams": validateTeamsSubscriber,
}

// Validate checks that the required fields are present based on the subscriber type.
func (r AddSubscriberRequest) Validate() error {
	if err := validateSubscriberType(r.Type); err != nil {
		return err
	}

	if validator, ok := typeValidators[r.Type]; ok {
		if err := validator(r); err != nil {
			return err
		}
	}

	if r.Language != nil {
		return validateSubscriberLanguage(*r.Language)
	}

	return nil
}
