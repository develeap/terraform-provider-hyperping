// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"strings"
	"testing"
)

// TestCreateMaintenanceRequest_Validate tests all validation paths
// to increase coverage from 71.4% to 100%
func TestCreateMaintenanceRequest_Validate(t *testing.T) {
	t.Run("valid request passes", func(t *testing.T) {
		req := CreateMaintenanceRequest{
			Name: "Valid Maintenance",
			Title: LocalizedText{
				En: "System Maintenance",
			},
			Text: LocalizedText{
				En: "We will be performing maintenance",
			},
		}
		err := req.Validate()
		if err != nil {
			t.Errorf("expected no error for valid request, got: %v", err)
		}
	})

	t.Run("name exceeds max length", func(t *testing.T) {
		req := CreateMaintenanceRequest{
			Name:  strings.Repeat("a", 256), // Exceeds maxNameLength (255)
			Title: LocalizedText{En: "Title"},
			Text:  LocalizedText{En: "Text"},
		}
		err := req.Validate()
		if err == nil {
			t.Error("expected error for name exceeding max length")
		}
		if err != nil && !strings.Contains(err.Error(), "name") {
			t.Errorf("expected error to mention 'name', got: %v", err)
		}
	})

	t.Run("title exceeds max length", func(t *testing.T) {
		req := CreateMaintenanceRequest{
			Name: "Valid Name",
			Title: LocalizedText{
				En: strings.Repeat("b", 256), // Exceeds maxNameLength
			},
			Text: LocalizedText{En: "Text"},
		}
		err := req.Validate()
		if err == nil {
			t.Error("expected error for title exceeding max length")
		}
		if err != nil && !strings.Contains(err.Error(), "title") {
			t.Errorf("expected error to mention 'title', got: %v", err)
		}
	})

	t.Run("text exceeds max length", func(t *testing.T) {
		req := CreateMaintenanceRequest{
			Name:  "Valid Name",
			Title: LocalizedText{En: "Valid Title"},
			Text: LocalizedText{
				En: strings.Repeat("c", 10001), // Exceeds maxMessageLength (10000)
			},
		}
		err := req.Validate()
		if err == nil {
			t.Error("expected error for text exceeding max length")
		}
		if err != nil && !strings.Contains(err.Error(), "text") {
			t.Errorf("expected error to mention 'text', got: %v", err)
		}
	})

	t.Run("all fields at max length", func(t *testing.T) {
		req := CreateMaintenanceRequest{
			Name: strings.Repeat("a", 255),
			Title: LocalizedText{
				En: strings.Repeat("b", 255),
			},
			Text: LocalizedText{
				En: strings.Repeat("c", 10000), // Updated to actual maxMessageLength
			},
		}
		err := req.Validate()
		if err != nil {
			t.Errorf("expected no error for fields at max length, got: %v", err)
		}
	})

	t.Run("multiple localized languages in title", func(t *testing.T) {
		req := CreateMaintenanceRequest{
			Name: "Maintenance",
			Title: LocalizedText{
				En: "Maintenance",
				Fr: "Maintenance",
				De: "Wartung",
				Es: "Mantenimiento",
			},
			Text: LocalizedText{En: "Text"},
		}
		err := req.Validate()
		if err != nil {
			t.Errorf("expected no error for multiple languages, got: %v", err)
		}
	})

	t.Run("multiple localized languages in text", func(t *testing.T) {
		req := CreateMaintenanceRequest{
			Name:  "Maintenance",
			Title: LocalizedText{En: "Title"},
			Text: LocalizedText{
				En: "System maintenance window",
				Fr: "Fenêtre de maintenance du système",
				De: "System-Wartungsfenster",
				Es: "Ventana de mantenimiento del sistema",
			},
		}
		err := req.Validate()
		if err != nil {
			t.Errorf("expected no error for multiple languages, got: %v", err)
		}
	})
}

// TestCreateStatusPageRequest_Validate tests all validation paths
// to increase coverage from 66.7% to 100%
func TestCreateStatusPageRequest_Validate(t *testing.T) {
	t.Run("valid request without website", func(t *testing.T) {
		subdomain := "mystatus"
		req := CreateStatusPageRequest{
			Name:      "My Status Page",
			Subdomain: &subdomain,
		}
		err := req.Validate()
		if err != nil {
			t.Errorf("expected no error for valid request, got: %v", err)
		}
	})

	t.Run("valid request with website", func(t *testing.T) {
		subdomain := "mystatus"
		website := "https://example.com"
		req := CreateStatusPageRequest{
			Name:      "My Status Page",
			Subdomain: &subdomain,
			Website:   &website,
		}
		err := req.Validate()
		if err != nil {
			t.Errorf("expected no error for valid request with website, got: %v", err)
		}
	})

	t.Run("name exceeds max length", func(t *testing.T) {
		subdomain := "test"
		req := CreateStatusPageRequest{
			Name:      strings.Repeat("a", 256), // Exceeds maxNameLength (255)
			Subdomain: &subdomain,
		}
		err := req.Validate()
		if err == nil {
			t.Error("expected error for name exceeding max length")
		}
		if err != nil && !strings.Contains(err.Error(), "name") {
			t.Errorf("expected error to mention 'name', got: %v", err)
		}
	})

	t.Run("website exceeds max length", func(t *testing.T) {
		// Create a very long URL (> 2048 chars)
		// https:// (8) + repeat (2041) + .com (4) = 2053 total
		subdomain := "test"
		longURL := "https://" + strings.Repeat("a", 2041) + ".com"
		req := CreateStatusPageRequest{
			Name:      "Valid Name",
			Subdomain: &subdomain,
			Website:   &longURL,
		}
		err := req.Validate()
		if err == nil {
			t.Error("expected error for website exceeding max length")
		}
		if err != nil && !strings.Contains(err.Error(), "website") {
			t.Errorf("expected error to mention 'website', got: %v", err)
		}
	})

	t.Run("name at max length", func(t *testing.T) {
		subdomain := "test"
		req := CreateStatusPageRequest{
			Name:      strings.Repeat("a", 255), // Exactly at maxNameLength
			Subdomain: &subdomain,
		}
		err := req.Validate()
		if err != nil {
			t.Errorf("expected no error for name at max length, got: %v", err)
		}
	})

	t.Run("website at max length", func(t *testing.T) {
		subdomain := "test"
		// https:// (8) + repeat (2036) + .com (4) = 2048 total (exactly maxURLLength)
		longURL := "https://" + strings.Repeat("a", 2036) + ".com"
		req := CreateStatusPageRequest{
			Name:      "Valid",
			Subdomain: &subdomain,
			Website:   &longURL,
		}
		err := req.Validate()
		if err != nil {
			t.Errorf("expected no error for website at max length, got: %v", err)
		}
	})

	t.Run("nil website is valid", func(t *testing.T) {
		subdomain := "test"
		req := CreateStatusPageRequest{
			Name:      "Valid Name",
			Subdomain: &subdomain,
			Website:   nil, // Explicitly nil
		}
		err := req.Validate()
		if err != nil {
			t.Errorf("expected no error for nil website, got: %v", err)
		}
	})

	t.Run("empty website string", func(t *testing.T) {
		subdomain := "test"
		emptyWebsite := ""
		req := CreateStatusPageRequest{
			Name:      "Valid Name",
			Subdomain: &subdomain,
			Website:   &emptyWebsite,
		}
		err := req.Validate()
		if err != nil {
			t.Errorf("expected no error for empty website string, got: %v", err)
		}
	})

	t.Run("both name and website exceed limits", func(t *testing.T) {
		subdomain := "test"
		longURL := "https://" + strings.Repeat("x", 2500) + ".com"
		req := CreateStatusPageRequest{
			Name:      strings.Repeat("n", 300), // Exceeds limit
			Subdomain: &subdomain,
			Website:   &longURL, // Also exceeds limit
		}
		err := req.Validate()
		if err == nil {
			t.Error("expected error when both fields exceed limits")
		}
		// Should fail on name first (validated before website)
		if err != nil && !strings.Contains(err.Error(), "name") {
			t.Errorf("expected error to mention 'name', got: %v", err)
		}
	})
}
