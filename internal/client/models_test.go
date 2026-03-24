// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFlexibleString_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected FlexibleString
		wantErr  bool
	}{
		{
			name:     "string value",
			input:    `"200"`,
			expected: FlexibleString("200"),
			wantErr:  false,
		},
		{
			name:     "integer number",
			input:    `200`,
			expected: FlexibleString("200"),
			wantErr:  false,
		},
		{
			name:     "float number",
			input:    `200.5`,
			expected: FlexibleString("200.5"),
			wantErr:  false,
		},
		{
			name:     "empty string",
			input:    `""`,
			expected: FlexibleString(""),
			wantErr:  false,
		},
		{
			name:     "zero number",
			input:    `0`,
			expected: FlexibleString("0"),
			wantErr:  false,
		},
		{
			name:     "negative number",
			input:    `-1`,
			expected: FlexibleString("-1"),
			wantErr:  false,
		},
		{
			name:     "boolean - should error",
			input:    `true`,
			expected: FlexibleString(""),
			wantErr:  true,
		},
		{
			name:     "null becomes empty string",
			input:    `null`,
			expected: FlexibleString(""),
			wantErr:  false,
		},
		{
			name:     "array - should error",
			input:    `[1,2,3]`,
			expected: FlexibleString(""),
			wantErr:  true,
		},
		{
			name:     "object - should error",
			input:    `{"key":"value"}`,
			expected: FlexibleString(""),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result FlexibleString
			err := json.Unmarshal([]byte(tt.input), &result)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for input %s, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %s: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("UnmarshalJSON() = %q, want %q", result, tt.expected)
				}
			}
		})
	}
}

func TestFlexibleString_String(t *testing.T) {
	tests := []struct {
		name     string
		input    FlexibleString
		expected string
	}{
		{
			name:     "normal string",
			input:    FlexibleString("200"),
			expected: "200",
		},
		{
			name:     "empty string",
			input:    FlexibleString(""),
			expected: "",
		},
		{
			name:     "numeric string",
			input:    FlexibleString("404"),
			expected: "404",
		},
		{
			name:     "string with special characters",
			input:    FlexibleString("status-code: 200"),
			expected: "status-code: 200",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.String()
			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFlexibleString_Integration(t *testing.T) {
	// Test that FlexibleString works in a real struct
	type testStruct struct {
		StatusCode FlexibleString `json:"status_code"`
	}

	t.Run("unmarshal string value", func(t *testing.T) {
		data := []byte(`{"status_code": "200"}`)
		var result testStruct
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if result.StatusCode != "200" {
			t.Errorf("expected '200', got %q", result.StatusCode)
		}
	})

	t.Run("unmarshal integer value", func(t *testing.T) {
		data := []byte(`{"status_code": 404}`)
		var result testStruct
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if result.StatusCode != "404" {
			t.Errorf("expected '404', got %q", result.StatusCode)
		}
	})

	t.Run("unmarshal float value", func(t *testing.T) {
		data := []byte(`{"status_code": 200.5}`)
		var result testStruct
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if result.StatusCode != "200.5" {
			t.Errorf("expected '200.5', got %q", result.StatusCode)
		}
	})

	t.Run("oversized input rejected", func(t *testing.T) {
		oversized := `"` + strings.Repeat("a", 200) + `"`
		var fs FlexibleString
		err := json.Unmarshal([]byte(oversized), &fs)
		if err == nil {
			t.Error("expected error for oversized FlexibleString input, got nil")
		}
		if !strings.Contains(err.Error(), "exceeds maximum size") {
			t.Errorf("expected size error, got: %v", err)
		}
	})
}

// TestMonitor_UnmarshalJSON_escalationPolicy verifies the custom unmarshaler
// for Monitor.escalation_policy, which the real API returns as an object
// {"uuid":"...","name":"..."} on GET even though POST/PUT send a plain string.
func TestMonitor_UnmarshalJSON_escalationPolicy(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		wantUUID string // empty string means nil
		wantErr  bool
	}{
		{
			name:     "object shape (real API GET response)",
			json:     `{"uuid":"mon_1","name":"test","url":"https://example.com","protocol":"HTTPS","http_method":"GET","check_frequency":60,"follow_redirects":true,"expected_status_code":"200","escalation_policy":{"uuid":"policy_abc123","name":"Core-Escalation"}}`,
			wantUUID: "policy_abc123",
		},
		{
			name:     "plain string (legacy / write path)",
			json:     `{"uuid":"mon_1","name":"test","url":"https://example.com","protocol":"HTTPS","http_method":"GET","check_frequency":60,"follow_redirects":true,"expected_status_code":"200","escalation_policy":"policy_abc123"}`,
			wantUUID: "policy_abc123",
		},
		{
			name:     "null value",
			json:     `{"uuid":"mon_1","name":"test","url":"https://example.com","protocol":"HTTPS","http_method":"GET","check_frequency":60,"follow_redirects":true,"expected_status_code":"200","escalation_policy":null}`,
			wantUUID: "",
		},
		{
			name:     "field absent",
			json:     `{"uuid":"mon_1","name":"test","url":"https://example.com","protocol":"HTTPS","http_method":"GET","check_frequency":60,"follow_redirects":true,"expected_status_code":"200"}`,
			wantUUID: "",
		},
		{
			name:     "object with empty UUID",
			json:     `{"uuid":"mon_1","name":"test","url":"https://example.com","protocol":"HTTPS","http_method":"GET","check_frequency":60,"follow_redirects":true,"expected_status_code":"200","escalation_policy":{"uuid":"","name":"Unnamed"}}`,
			wantUUID: "",
		},
		{
			name:     "empty string value",
			json:     `{"uuid":"mon_1","name":"test","url":"https://example.com","protocol":"HTTPS","http_method":"GET","check_frequency":60,"follow_redirects":true,"expected_status_code":"200","escalation_policy":""}`,
			wantUUID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m Monitor
			err := json.Unmarshal([]byte(tt.json), &m)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantUUID == "" {
				if m.EscalationPolicy != nil {
					t.Errorf("expected nil EscalationPolicy, got %q", *m.EscalationPolicy)
				}
			} else {
				if m.EscalationPolicy == nil {
					t.Errorf("expected EscalationPolicy = %q, got nil", tt.wantUUID)
				} else if *m.EscalationPolicy != tt.wantUUID {
					t.Errorf("EscalationPolicy = %q, want %q", *m.EscalationPolicy, tt.wantUUID)
				}
			}
		})
	}
}

func TestCreateMonitorRequest_Validate(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		req := CreateMonitorRequest{Name: "test", URL: "https://example.com"}
		if err := req.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("name too long", func(t *testing.T) {
		req := CreateMonitorRequest{Name: strings.Repeat("a", 256), URL: "https://example.com"}
		if err := req.Validate(); err == nil {
			t.Error("expected validation error for oversized name")
		}
	})

	t.Run("url too long", func(t *testing.T) {
		req := CreateMonitorRequest{Name: "test", URL: strings.Repeat("a", 2049)}
		if err := req.Validate(); err == nil {
			t.Error("expected validation error for oversized URL")
		}
	})
}

func TestCreateIncidentRequest_Validate(t *testing.T) {
	t.Run("title too long", func(t *testing.T) {
		req := CreateIncidentRequest{
			Title: LocalizedText{En: strings.Repeat("x", 256)},
			Text:  LocalizedText{En: "short"},
		}
		if err := req.Validate(); err == nil {
			t.Error("expected validation error for oversized title")
		}
	})

	t.Run("text too long", func(t *testing.T) {
		req := CreateIncidentRequest{
			Title: LocalizedText{En: "short"},
			Text:  LocalizedText{En: strings.Repeat("x", 10001)},
		}
		if err := req.Validate(); err == nil {
			t.Error("expected validation error for oversized text")
		}
	})
}

// TestUpdateMonitorRequest_DNSRecordTypeOmittedWhenNil verifies that an empty
// UpdateMonitorRequest omits dns_record_type from the JSON body. The workaround
// for the Hyperping API validation bug (defaulting to "A") is applied in
// UpdateMonitor(), not at the struct serialisation level.
func TestUpdateMonitorRequest_DNSRecordTypeOmittedWhenNil(t *testing.T) {
	req := UpdateMonitorRequest{}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}
	body := string(b)

	if strings.Contains(body, "dns_record_type") {
		t.Errorf("expected dns_record_type to be omitted when nil, got: %s", body)
	}
}

// =============================================================================
// Gap 1: Incident updates array
// =============================================================================

func TestCreateIncidentRequest_WithUpdates_JSON(t *testing.T) {
	req := CreateIncidentRequest{
		Title:       LocalizedText{En: "Incident"},
		Text:        LocalizedText{En: "Details"},
		Type:        "investigating",
		StatusPages: []string{"sp_1"},
		Updates: []AddIncidentUpdateRequest{
			{
				Text: LocalizedText{En: "Investigating..."},
				Type: "investigating",
				Date: "2026-03-20T10:00:00Z",
			},
		},
	}

	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}
	body := string(b)

	if !strings.Contains(body, `"updates"`) {
		t.Errorf("expected JSON to contain 'updates' key, got: %s", body)
	}
	if !strings.Contains(body, `"investigating"`) {
		t.Errorf("expected JSON to contain update type, got: %s", body)
	}
}

func TestUpdateIncidentRequest_WithUpdates_JSON(t *testing.T) {
	req := UpdateIncidentRequest{
		Updates: []AddIncidentUpdateRequest{
			{
				Text: LocalizedText{En: "Resolved"},
				Type: "resolved",
				Date: "2026-03-20T12:00:00Z",
			},
		},
	}

	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}
	body := string(b)

	if !strings.Contains(body, `"updates"`) {
		t.Errorf("expected JSON to contain 'updates' key, got: %s", body)
	}
}

func TestCreateIncidentRequest_WithoutUpdates_OmitsField(t *testing.T) {
	t.Run("nil updates omitted", func(t *testing.T) {
		req := CreateIncidentRequest{
			Title:       LocalizedText{En: "Incident"},
			Text:        LocalizedText{En: "Details"},
			Type:        "investigating",
			StatusPages: []string{"sp_1"},
		}

		b, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("unexpected marshal error: %v", err)
		}
		if strings.Contains(string(b), `"updates"`) {
			t.Errorf("expected 'updates' to be omitted when nil, got: %s", string(b))
		}
	})

	t.Run("empty updates omitted", func(t *testing.T) {
		req := CreateIncidentRequest{
			Title:       LocalizedText{En: "Incident"},
			Text:        LocalizedText{En: "Details"},
			Type:        "investigating",
			StatusPages: []string{"sp_1"},
			Updates:     []AddIncidentUpdateRequest{},
		}

		b, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("unexpected marshal error: %v", err)
		}
		if strings.Contains(string(b), `"updates"`) {
			t.Errorf("expected 'updates' to be omitted when empty, got: %s", string(b))
		}
	})
}

// =============================================================================
// Gap 2: StatusPage sso_connection_uuid
// =============================================================================

func TestStatusPageAuthenticationSettings_SSOConnectionUUID_JSON(t *testing.T) {
	t.Run("null sso_connection_uuid", func(t *testing.T) {
		data := []byte(`{"password_protection":false,"google_sso":false,"saml_sso":false,"google_allowed_domains":[],"sso_connection_uuid":null}`)
		var auth StatusPageAuthenticationSettings
		if err := json.Unmarshal(data, &auth); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if auth.SSOConnectionUUID != nil {
			t.Errorf("expected nil SSOConnectionUUID, got %q", *auth.SSOConnectionUUID)
		}
	})

	t.Run("string sso_connection_uuid", func(t *testing.T) {
		data := []byte(`{"password_protection":false,"google_sso":false,"saml_sso":false,"google_allowed_domains":[],"sso_connection_uuid":"uuid-abc"}`)
		var auth StatusPageAuthenticationSettings
		if err := json.Unmarshal(data, &auth); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if auth.SSOConnectionUUID == nil {
			t.Fatal("expected non-nil SSOConnectionUUID")
		}
		if *auth.SSOConnectionUUID != "uuid-abc" {
			t.Errorf("expected 'uuid-abc', got %q", *auth.SSOConnectionUUID)
		}
	})

	t.Run("absent sso_connection_uuid", func(t *testing.T) {
		data := []byte(`{"password_protection":false,"google_sso":false,"saml_sso":false,"google_allowed_domains":[]}`)
		var auth StatusPageAuthenticationSettings
		if err := json.Unmarshal(data, &auth); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if auth.SSOConnectionUUID != nil {
			t.Errorf("expected nil SSOConnectionUUID for absent field, got %q", *auth.SSOConnectionUUID)
		}
	})
}

func TestCreateStatusPageAuthenticationSettings_SSOConnectionUUID_JSON(t *testing.T) {
	t.Run("nil omits field", func(t *testing.T) {
		auth := CreateStatusPageAuthenticationSettings{}
		b, err := json.Marshal(auth)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}
		if strings.Contains(string(b), "sso_connection_uuid") {
			t.Errorf("expected sso_connection_uuid to be omitted when nil, got: %s", string(b))
		}
	})

	t.Run("non-nil includes field", func(t *testing.T) {
		val := "uuid-abc"
		auth := CreateStatusPageAuthenticationSettings{SSOConnectionUUID: &val}
		b, err := json.Marshal(auth)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}
		if !strings.Contains(string(b), `"sso_connection_uuid":"uuid-abc"`) {
			t.Errorf("expected sso_connection_uuid in output, got: %s", string(b))
		}
	})
}

// =============================================================================
// Gap 3: StatusPage service description
// =============================================================================

func TestStatusPageService_Description_JSON(t *testing.T) {
	t.Run("localized description map", func(t *testing.T) {
		data := []byte(`{"uuid":"svc_1","name":{"en":"API"},"is_group":false,"show_uptime":false,"show_response_times":false,"description":{"en":"My service"}}`)
		var svc StatusPageService
		if err := json.Unmarshal(data, &svc); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if svc.Description["en"] != "My service" {
			t.Errorf("expected description[en]='My service', got %q", svc.Description["en"])
		}
	})

	t.Run("multi-language description", func(t *testing.T) {
		data := []byte(`{"uuid":"svc_1","name":{"en":"API"},"is_group":false,"show_uptime":false,"show_response_times":false,"description":{"en":"English","fr":"French"}}`)
		var svc StatusPageService
		if err := json.Unmarshal(data, &svc); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if len(svc.Description) != 2 {
			t.Errorf("expected 2 description entries, got %d", len(svc.Description))
		}
	})

	t.Run("absent description", func(t *testing.T) {
		data := []byte(`{"uuid":"svc_1","name":{"en":"API"},"is_group":false,"show_uptime":false,"show_response_times":false}`)
		var svc StatusPageService
		if err := json.Unmarshal(data, &svc); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if svc.Description != nil {
			t.Errorf("expected nil description for absent field, got %v", svc.Description)
		}
	})
}

// =============================================================================
// Outage severity and summary fields
// =============================================================================

func TestOutage_SeverityAndSummary_Unmarshal(t *testing.T) {
	t.Run("both fields present", func(t *testing.T) {
		data := []byte(`{"uuid":"out_1","startDate":"2026-03-20T10:00:00Z","statusCode":500,"description":"down","outageType":"manual","isResolved":false,"detectedLocation":"us-east","confirmedLocations":"us-east","monitor":{"uuid":"mon_1","name":"test","url":"https://example.com","protocol":"HTTPS"},"severity":"critical","summary":"Server outage"}`)
		var outage Outage
		if err := json.Unmarshal(data, &outage); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if outage.Severity != "critical" {
			t.Errorf("expected severity 'critical', got %q", outage.Severity)
		}
		if outage.Summary != "Server outage" {
			t.Errorf("expected summary 'Server outage', got %q", outage.Summary)
		}
	})

	t.Run("fields absent", func(t *testing.T) {
		data := []byte(`{"uuid":"out_1","startDate":"2026-03-20T10:00:00Z","statusCode":500,"description":"down","outageType":"manual","isResolved":false,"detectedLocation":"us-east","confirmedLocations":"us-east","monitor":{"uuid":"mon_1","name":"test","url":"https://example.com","protocol":"HTTPS"}}`)
		var outage Outage
		if err := json.Unmarshal(data, &outage); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if outage.Severity != "" {
			t.Errorf("expected empty severity, got %q", outage.Severity)
		}
		if outage.Summary != "" {
			t.Errorf("expected empty summary, got %q", outage.Summary)
		}
	})
}

func TestCreateOutageRequest_SeverityAndSummary_Marshal(t *testing.T) {
	t.Run("nil omits fields", func(t *testing.T) {
		req := CreateOutageRequest{
			MonitorUUID: "mon_1",
			StartDate:   "2026-03-20T10:00:00Z",
			StatusCode:  500,
			Description: "down",
			OutageType:  "manual",
		}
		b, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}
		body := string(b)
		if strings.Contains(body, "severity") {
			t.Errorf("expected severity to be omitted when nil, got: %s", body)
		}
		if strings.Contains(body, "summary") {
			t.Errorf("expected summary to be omitted when nil, got: %s", body)
		}
	})

	t.Run("non-nil includes fields", func(t *testing.T) {
		sev := "critical"
		sum := "Server outage"
		req := CreateOutageRequest{
			MonitorUUID: "mon_1",
			StartDate:   "2026-03-20T10:00:00Z",
			StatusCode:  500,
			Description: "down",
			OutageType:  "manual",
			Severity:    &sev,
			Summary:     &sum,
		}
		b, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}
		body := string(b)
		if !strings.Contains(body, `"severity":"critical"`) {
			t.Errorf("expected severity in output, got: %s", body)
		}
		if !strings.Contains(body, `"summary":"Server outage"`) {
			t.Errorf("expected summary in output, got: %s", body)
		}
	})
}

func TestCreateStatusPageService_Description_JSON(t *testing.T) {
	t.Run("nil omits field", func(t *testing.T) {
		svc := CreateStatusPageService{}
		b, err := json.Marshal(svc)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}
		if strings.Contains(string(b), "description") {
			t.Errorf("expected description to be omitted when nil, got: %s", string(b))
		}
	})

	t.Run("non-nil includes field as string", func(t *testing.T) {
		val := "Service desc"
		svc := CreateStatusPageService{Description: &val}
		b, err := json.Marshal(svc)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}
		if !strings.Contains(string(b), `"description":"Service desc"`) {
			t.Errorf("expected description in output, got: %s", string(b))
		}
	})
}
