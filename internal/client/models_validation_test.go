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

// statusPageValidationCase describes a single test case for CreateStatusPageRequest.Validate.
type statusPageValidationCase struct {
	name        string
	req         func() CreateStatusPageRequest
	wantErr     bool
	errContains string
}

// TestCreateStatusPageRequest_Validate tests all validation paths
// to increase coverage from 66.7% to 100%
func TestCreateStatusPageRequest_Validate(t *testing.T) {
	subdomain := "mystatus"
	testSubdomain := "test"

	tests := []statusPageValidationCase{
		{
			name: "valid request without website",
			req: func() CreateStatusPageRequest {
				return CreateStatusPageRequest{Name: "My Status Page", Subdomain: &subdomain}
			},
			wantErr: false,
		},
		{
			name: "valid request with website",
			req: func() CreateStatusPageRequest {
				website := "https://example.com"
				return CreateStatusPageRequest{Name: "My Status Page", Subdomain: &subdomain, Website: &website}
			},
			wantErr: false,
		},
		{
			name: "name exceeds max length",
			req: func() CreateStatusPageRequest {
				return CreateStatusPageRequest{Name: strings.Repeat("a", 256), Subdomain: &testSubdomain}
			},
			wantErr:     true,
			errContains: "name",
		},
		{
			name: "website exceeds max length",
			req: func() CreateStatusPageRequest {
				longURL := "https://" + strings.Repeat("a", 2041) + ".com"
				return CreateStatusPageRequest{Name: "Valid Name", Subdomain: &testSubdomain, Website: &longURL}
			},
			wantErr:     true,
			errContains: "website",
		},
		{
			name: "name at max length",
			req: func() CreateStatusPageRequest {
				return CreateStatusPageRequest{Name: strings.Repeat("a", 255), Subdomain: &testSubdomain}
			},
			wantErr: false,
		},
		{
			name: "website at max length",
			req: func() CreateStatusPageRequest {
				longURL := "https://" + strings.Repeat("a", 2036) + ".com"
				return CreateStatusPageRequest{Name: "Valid", Subdomain: &testSubdomain, Website: &longURL}
			},
			wantErr: false,
		},
		{
			name: "nil website is valid",
			req: func() CreateStatusPageRequest {
				return CreateStatusPageRequest{Name: "Valid Name", Subdomain: &testSubdomain, Website: nil}
			},
			wantErr: false,
		},
		{
			name: "empty website string",
			req: func() CreateStatusPageRequest {
				emptyWebsite := ""
				return CreateStatusPageRequest{Name: "Valid Name", Subdomain: &testSubdomain, Website: &emptyWebsite}
			},
			wantErr: false,
		},
		{
			name: "both name and website exceed limits",
			req: func() CreateStatusPageRequest {
				longURL := "https://" + strings.Repeat("x", 2500) + ".com"
				return CreateStatusPageRequest{
					Name:      strings.Repeat("n", 300),
					Subdomain: &testSubdomain,
					Website:   &longURL,
				}
			},
			wantErr:     true,
			errContains: "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.req()
			err := req.Validate()
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
				return
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
				return
			}
			if tt.wantErr && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("expected error to contain %q, got: %v", tt.errContains, err)
			}
		})
	}
}
