// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"strings"
	"testing"
)

// =============================================================================
// Tests for validateEmailSubscriber
// =============================================================================

func TestValidateEmailSubscriber(t *testing.T) {
	t.Run("valid email passes", func(t *testing.T) {
		req := AddSubscriberRequest{
			Type:  "email",
			Email: strPtr("user@example.com"),
		}
		if err := validateEmailSubscriber(req); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("nil email fails", func(t *testing.T) {
		req := AddSubscriberRequest{Type: "email", Email: nil}
		err := validateEmailSubscriber(req)
		if err == nil {
			t.Fatal("expected error for nil email")
		}
		if !strings.Contains(err.Error(), "email is required") {
			t.Errorf("unexpected error message: %q", err.Error())
		}
	})

	t.Run("empty email fails", func(t *testing.T) {
		req := AddSubscriberRequest{Type: "email", Email: strPtr("")}
		err := validateEmailSubscriber(req)
		if err == nil {
			t.Fatal("expected error for empty email")
		}
		if !strings.Contains(err.Error(), "email is required") {
			t.Errorf("unexpected error message: %q", err.Error())
		}
	})

	t.Run("extra fields do not affect email validation", func(t *testing.T) {
		// Phone and TeamsWebhookURL are irrelevant for email type.
		req := AddSubscriberRequest{
			Type:            "email",
			Email:           strPtr("a@b.com"),
			Phone:           strPtr("+1234567890"),
			TeamsWebhookURL: strPtr("https://teams.example.com/webhook"),
		}
		if err := validateEmailSubscriber(req); err != nil {
			t.Errorf("extra fields should not cause failure: %v", err)
		}
	})
}

// =============================================================================
// Tests for validateSMSSubscriber
// =============================================================================

func TestValidateSMSSubscriber(t *testing.T) {
	t.Run("valid phone passes", func(t *testing.T) {
		req := AddSubscriberRequest{
			Type:  "sms",
			Phone: strPtr("+15551234567"),
		}
		if err := validateSMSSubscriber(req); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("nil phone fails", func(t *testing.T) {
		req := AddSubscriberRequest{Type: "sms", Phone: nil}
		err := validateSMSSubscriber(req)
		if err == nil {
			t.Fatal("expected error for nil phone")
		}
		if !strings.Contains(err.Error(), "phone is required") {
			t.Errorf("unexpected error message: %q", err.Error())
		}
	})

	t.Run("empty phone fails", func(t *testing.T) {
		req := AddSubscriberRequest{Type: "sms", Phone: strPtr("")}
		err := validateSMSSubscriber(req)
		if err == nil {
			t.Fatal("expected error for empty phone")
		}
		if !strings.Contains(err.Error(), "phone is required") {
			t.Errorf("unexpected error message: %q", err.Error())
		}
	})

	t.Run("extra fields do not affect sms validation", func(t *testing.T) {
		req := AddSubscriberRequest{
			Type:  "sms",
			Phone: strPtr("+15551234567"),
			Email: strPtr("extra@example.com"),
		}
		if err := validateSMSSubscriber(req); err != nil {
			t.Errorf("extra fields should not cause failure: %v", err)
		}
	})
}

// =============================================================================
// Tests for validateTeamsSubscriber
// =============================================================================

func TestValidateTeamsSubscriber(t *testing.T) {
	t.Run("valid webhook url passes", func(t *testing.T) {
		req := AddSubscriberRequest{
			Type:            "teams",
			TeamsWebhookURL: strPtr("https://teams.microsoft.com/webhook/abc"),
		}
		if err := validateTeamsSubscriber(req); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("nil teams_webhook_url fails", func(t *testing.T) {
		req := AddSubscriberRequest{Type: "teams", TeamsWebhookURL: nil}
		err := validateTeamsSubscriber(req)
		if err == nil {
			t.Fatal("expected error for nil teams_webhook_url")
		}
		if !strings.Contains(err.Error(), "teams_webhook_url is required") {
			t.Errorf("unexpected error message: %q", err.Error())
		}
	})

	t.Run("empty teams_webhook_url fails", func(t *testing.T) {
		req := AddSubscriberRequest{Type: "teams", TeamsWebhookURL: strPtr("")}
		err := validateTeamsSubscriber(req)
		if err == nil {
			t.Fatal("expected error for empty teams_webhook_url")
		}
		if !strings.Contains(err.Error(), "teams_webhook_url is required") {
			t.Errorf("unexpected error message: %q", err.Error())
		}
	})

	t.Run("extra fields do not affect teams validation", func(t *testing.T) {
		req := AddSubscriberRequest{
			Type:            "teams",
			TeamsWebhookURL: strPtr("https://teams.microsoft.com/webhook/abc"),
			Email:           strPtr("extra@example.com"),
			Phone:           strPtr("+1234567890"),
		}
		if err := validateTeamsSubscriber(req); err != nil {
			t.Errorf("extra fields should not cause failure: %v", err)
		}
	})
}

// =============================================================================
// Tests for validateSubscriberType
// =============================================================================

func TestValidateSubscriberType(t *testing.T) {
	validTypes := []string{"email", "sms", "teams"}
	for _, typ := range validTypes {
		typ := typ
		t.Run("valid type: "+typ, func(t *testing.T) {
			if err := validateSubscriberType(typ); err != nil {
				t.Errorf("unexpected error for valid type %q: %v", typ, err)
			}
		})
	}

	invalidTypes := []string{"slack", "webhook", "discord", "", "EMAIL", "SMS"}
	for _, typ := range invalidTypes {
		typ := typ
		t.Run("invalid type: "+typ, func(t *testing.T) {
			err := validateSubscriberType(typ)
			if err == nil {
				t.Errorf("expected error for invalid type %q", typ)
			}
			if !strings.Contains(err.Error(), "invalid subscriber type") {
				t.Errorf("unexpected error message: %q", err.Error())
			}
		})
	}
}

// =============================================================================
// Tests for validateSubscriberLanguage
// =============================================================================

func TestValidateSubscriberLanguage(t *testing.T) {
	validLangs := []string{"en", "fr", "de", "ru", "nl", "es", "it", "pt", "ja", "zh"}
	for _, lang := range validLangs {
		lang := lang
		t.Run("valid language: "+lang, func(t *testing.T) {
			if err := validateSubscriberLanguage(lang); err != nil {
				t.Errorf("unexpected error for valid language %q: %v", lang, err)
			}
		})
	}

	invalidLangs := []string{"EN", "english", "xx", "", "zz"}
	for _, lang := range invalidLangs {
		lang := lang
		t.Run("invalid language: "+lang, func(t *testing.T) {
			err := validateSubscriberLanguage(lang)
			if err == nil {
				t.Errorf("expected error for invalid language %q", lang)
			}
			if !strings.Contains(err.Error(), "invalid language") {
				t.Errorf("unexpected error message: %q", err.Error())
			}
		})
	}
}

// =============================================================================
// Tests for AddSubscriberRequest.Validate (top-level dispatcher)
// =============================================================================

func TestAddSubscriberRequestValidate(t *testing.T) {
	t.Run("valid email subscriber passes", func(t *testing.T) {
		req := AddSubscriberRequest{Type: "email", Email: strPtr("user@example.com")}
		if err := req.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("valid sms subscriber passes", func(t *testing.T) {
		req := AddSubscriberRequest{Type: "sms", Phone: strPtr("+15551234567")}
		if err := req.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("valid teams subscriber passes", func(t *testing.T) {
		req := AddSubscriberRequest{
			Type:            "teams",
			TeamsWebhookURL: strPtr("https://teams.microsoft.com/webhook/abc"),
		}
		if err := req.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("invalid type fails before type-specific check", func(t *testing.T) {
		req := AddSubscriberRequest{Type: "slack", Email: strPtr("user@example.com")}
		err := req.Validate()
		if err == nil {
			t.Fatal("expected error for invalid type")
		}
		if !strings.Contains(err.Error(), "invalid subscriber type") {
			t.Errorf("unexpected error: %q", err.Error())
		}
	})

	t.Run("email type with missing email fails", func(t *testing.T) {
		req := AddSubscriberRequest{Type: "email"}
		err := req.Validate()
		if err == nil {
			t.Fatal("expected error for missing email")
		}
		if !strings.Contains(err.Error(), "email is required") {
			t.Errorf("unexpected error: %q", err.Error())
		}
	})

	t.Run("valid type with invalid language fails", func(t *testing.T) {
		lang := "xx"
		req := AddSubscriberRequest{
			Type:     "email",
			Email:    strPtr("user@example.com"),
			Language: &lang,
		}
		err := req.Validate()
		if err == nil {
			t.Fatal("expected error for invalid language")
		}
		if !strings.Contains(err.Error(), "invalid language") {
			t.Errorf("unexpected error: %q", err.Error())
		}
	})

	t.Run("valid type with valid language passes", func(t *testing.T) {
		lang := "fr"
		req := AddSubscriberRequest{
			Type:     "email",
			Email:    strPtr("user@example.com"),
			Language: &lang,
		}
		if err := req.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
