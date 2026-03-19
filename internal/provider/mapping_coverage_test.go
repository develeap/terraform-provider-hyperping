// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
	"github.com/develeap/terraform-provider-hyperping/internal/provider/testutil"
)

// ---------------------------------------------------------------------------
// MapMonitorCommonFields
// ---------------------------------------------------------------------------

func TestMapMonitorCommonFields_FullMonitor(t *testing.T) {
	port := 443
	sslExp := 30
	keyword := "healthy"
	ep := "ep_abc"
	dnsRecType := "A"
	dnsNS := "8.8.8.8"
	dnsAnswer := "1.2.3.4"

	monitor := &client.Monitor{
		UUID:               "mon_full",
		Name:               "Full Monitor",
		URL:                "https://example.com",
		Protocol:           "http",
		HTTPMethod:         "GET",
		CheckFrequency:     60,
		ExpectedStatusCode: "200",
		FollowRedirects:    true,
		Paused:             false,
		Regions:            []string{"london", "frankfurt"},
		RequestHeaders: []client.RequestHeader{
			{Name: "Authorization", Value: "Bearer token123"},
		},
		RequestBody:       `{"ping":true}`,
		Port:              &port,
		AlertsWait:        5,
		EscalationPolicy:  &ep,
		DNSRecordType:     &dnsRecType,
		DNSNameserver:     &dnsNS,
		DNSExpectedAnswer: &dnsAnswer,
		RequiredKeyword:   &keyword,
		Status:            "up",
		SSLExpiration:     &sslExp,
		ProjectUUID:       "proj_001",
	}

	var diags diag.Diagnostics
	result := MapMonitorCommonFields(monitor, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	assertStringField(t, "ID", result.ID, "mon_full")
	assertStringField(t, "Name", result.Name, "Full Monitor")
	assertStringField(t, "URL", result.URL, "https://example.com")
	assertStringField(t, "Protocol", result.Protocol, "http")
	assertStringField(t, "HTTPMethod", result.HTTPMethod, "GET")
	assertInt64Field(t, "CheckFrequency", result.CheckFrequency, 60)
	assertStringField(t, "ExpectedStatusCode", result.ExpectedStatusCode, "200")
	assertBoolField(t, "FollowRedirects", result.FollowRedirects, true)
	assertBoolField(t, "Paused", result.Paused, false)
	assertStringField(t, "RequestBody", result.RequestBody, `{"ping":true}`)
	assertInt64Field(t, "Port", result.Port, 443)
	assertInt64Field(t, "AlertsWait", result.AlertsWait, 5)
	assertStringField(t, "EscalationPolicy", result.EscalationPolicy, "ep_abc")
	assertStringField(t, "DNSRecordType", result.DNSRecordType, "A")
	assertStringField(t, "DNSNameserver", result.DNSNameserver, "8.8.8.8")
	assertStringField(t, "DNSExpectedAnswer", result.DNSExpectedAnswer, "1.2.3.4")
	assertStringField(t, "RequiredKeyword", result.RequiredKeyword, "healthy")
	assertStringField(t, "Status", result.Status, "up")
	assertInt64Field(t, "SSLExpiration", result.SSLExpiration, 30)
	assertStringField(t, "ProjectUUID", result.ProjectUUID, "proj_001")

	// Regions list
	if result.Regions.IsNull() {
		t.Fatal("expected Regions to be non-null")
	}
	regionElems := result.Regions.Elements()
	if len(regionElems) != 2 {
		t.Fatalf("expected 2 regions, got %d", len(regionElems))
	}
	if regionElems[0].(types.String).ValueString() != "london" {
		t.Errorf("expected first region 'london', got %s", regionElems[0].(types.String).ValueString())
	}
	if regionElems[1].(types.String).ValueString() != "frankfurt" {
		t.Errorf("expected second region 'frankfurt', got %s", regionElems[1].(types.String).ValueString())
	}

	// Request headers list
	if result.RequestHeaders.IsNull() {
		t.Fatal("expected RequestHeaders to be non-null")
	}
	headerElems := result.RequestHeaders.Elements()
	if len(headerElems) != 1 {
		t.Fatalf("expected 1 header, got %d", len(headerElems))
	}
	headerObj := headerElems[0].(types.Object)
	headerAttrs := headerObj.Attributes()
	if headerAttrs["name"].(types.String).ValueString() != "Authorization" {
		t.Errorf("expected header name 'Authorization', got %s", headerAttrs["name"].(types.String).ValueString())
	}
	if headerAttrs["value"].(types.String).ValueString() != "Bearer token123" {
		t.Errorf("expected header value 'Bearer token123', got %s", headerAttrs["value"].(types.String).ValueString())
	}
}

func TestMapMonitorCommonFields_MinimalMonitor(t *testing.T) {
	monitor := &client.Monitor{
		UUID:               "mon_min",
		Name:               "Minimal",
		URL:                "https://example.com",
		Protocol:           "http",
		HTTPMethod:         "GET",
		CheckFrequency:     30,
		ExpectedStatusCode: "200",
		FollowRedirects:    false,
		Paused:             true,
	}

	var diags diag.Diagnostics
	result := MapMonitorCommonFields(monitor, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	assertStringField(t, "ID", result.ID, "mon_min")
	assertBoolField(t, "FollowRedirects", result.FollowRedirects, false)
	assertBoolField(t, "Paused", result.Paused, true)
	assertStringField(t, "Status", result.Status, "")

	assertNullString(t, "RequestBody", result.RequestBody)
	assertNullInt64(t, "Port", result.Port)
	assertNullInt64(t, "AlertsWait", result.AlertsWait)
	assertNullString(t, "EscalationPolicy", result.EscalationPolicy)
	assertNullString(t, "DNSRecordType", result.DNSRecordType)
	assertNullString(t, "DNSNameserver", result.DNSNameserver)
	assertNullString(t, "DNSExpectedAnswer", result.DNSExpectedAnswer)
	assertNullString(t, "RequiredKeyword", result.RequiredKeyword)
	assertNullInt64(t, "SSLExpiration", result.SSLExpiration)
	assertNullString(t, "ProjectUUID", result.ProjectUUID)

	if !result.Regions.IsNull() {
		t.Error("expected Regions to be null when empty")
	}
	if !result.RequestHeaders.IsNull() {
		t.Error("expected RequestHeaders to be null when empty")
	}
}

func TestMapMonitorCommonFields_AlertsWait(t *testing.T) {
	tests := []struct {
		name       string
		alertsWait int
		wantNull   bool
		wantValue  int64
	}{
		{
			name:       "zero maps to null",
			alertsWait: 0,
			wantNull:   true,
		},
		{
			name:       "negative one (disabled) maps to value",
			alertsWait: -1,
			wantNull:   false,
			wantValue:  -1,
		},
		{
			name:       "positive value maps to value",
			alertsWait: 10,
			wantNull:   false,
			wantValue:  10,
		},
		{
			name:       "large positive value maps to value",
			alertsWait: 3600,
			wantNull:   false,
			wantValue:  3600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := &client.Monitor{
				UUID:               "mon_aw",
				Name:               "AW Test",
				URL:                "https://example.com",
				Protocol:           "http",
				HTTPMethod:         "GET",
				CheckFrequency:     60,
				ExpectedStatusCode: "200",
				AlertsWait:         tt.alertsWait,
			}

			var diags diag.Diagnostics
			result := MapMonitorCommonFields(monitor, &diags)

			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags.Errors())
			}

			if tt.wantNull {
				assertNullInt64(t, "AlertsWait", result.AlertsWait)
			} else {
				assertInt64Field(t, "AlertsWait", result.AlertsWait, tt.wantValue)
			}
		})
	}
}

func TestMapMonitorCommonFields_PortSetVsNil(t *testing.T) {
	tests := []struct {
		name      string
		port      *int
		wantNull  bool
		wantValue int64
	}{
		{
			name:     "nil port maps to null",
			port:     nil,
			wantNull: true,
		},
		{
			name:      "port 80",
			port:      testutil.Ptr(80),
			wantNull:  false,
			wantValue: 80,
		},
		{
			name:      "port 443",
			port:      testutil.Ptr(443),
			wantNull:  false,
			wantValue: 443,
		},
		{
			name:      "port 0",
			port:      testutil.Ptr(0),
			wantNull:  false,
			wantValue: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := &client.Monitor{
				UUID:               "mon_port",
				Name:               "Port Test",
				URL:                "https://example.com",
				Protocol:           "http",
				HTTPMethod:         "GET",
				CheckFrequency:     60,
				ExpectedStatusCode: "200",
				Port:               tt.port,
			}

			var diags diag.Diagnostics
			result := MapMonitorCommonFields(monitor, &diags)

			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags.Errors())
			}

			if tt.wantNull {
				assertNullInt64(t, "Port", result.Port)
			} else {
				assertInt64Field(t, "Port", result.Port, tt.wantValue)
			}
		})
	}
}

func TestMapMonitorCommonFields_EscalationPolicyVariants(t *testing.T) {
	empty := ""
	populated := "ep_456"

	tests := []struct {
		name      string
		ep        *string
		wantNull  bool
		wantValue string
	}{
		{
			name:     "nil escalation policy maps to null",
			ep:       nil,
			wantNull: true,
		},
		{
			name:     "empty string escalation policy maps to null",
			ep:       &empty,
			wantNull: true,
		},
		{
			name:      "populated escalation policy maps to value",
			ep:        &populated,
			wantNull:  false,
			wantValue: "ep_456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := &client.Monitor{
				UUID:               "mon_ep",
				Name:               "EP Test",
				URL:                "https://example.com",
				Protocol:           "http",
				HTTPMethod:         "GET",
				CheckFrequency:     60,
				ExpectedStatusCode: "200",
				EscalationPolicy:   tt.ep,
			}

			var diags diag.Diagnostics
			result := MapMonitorCommonFields(monitor, &diags)

			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags.Errors())
			}

			if tt.wantNull {
				assertNullString(t, "EscalationPolicy", result.EscalationPolicy)
			} else {
				assertStringField(t, "EscalationPolicy", result.EscalationPolicy, tt.wantValue)
			}
		})
	}
}

func TestMapMonitorCommonFields_DNSFields(t *testing.T) {
	tests := []struct {
		name              string
		dnsRecordType     *string
		dnsNameserver     *string
		dnsExpectedAnswer *string
		wantRecTypeNull   bool
		wantNSNull        bool
		wantAnswerNull    bool
	}{
		{
			name:            "all nil",
			wantRecTypeNull: true,
			wantNSNull:      true,
			wantAnswerNull:  true,
		},
		{
			name:              "all empty strings",
			dnsRecordType:     testutil.Ptr(""),
			dnsNameserver:     testutil.Ptr(""),
			dnsExpectedAnswer: testutil.Ptr(""),
			wantRecTypeNull:   true,
			wantNSNull:        true,
			wantAnswerNull:    true,
		},
		{
			name:              "all populated",
			dnsRecordType:     testutil.Ptr("AAAA"),
			dnsNameserver:     testutil.Ptr("1.1.1.1"),
			dnsExpectedAnswer: testutil.Ptr("2001:db8::1"),
			wantRecTypeNull:   false,
			wantNSNull:        false,
			wantAnswerNull:    false,
		},
		{
			name:            "partially populated (only record type)",
			dnsRecordType:   testutil.Ptr("MX"),
			wantRecTypeNull: false,
			wantNSNull:      true,
			wantAnswerNull:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := &client.Monitor{
				UUID:               "mon_dns",
				Name:               "DNS Test",
				URL:                "dns://example.com",
				Protocol:           "dns",
				HTTPMethod:         "GET",
				CheckFrequency:     60,
				ExpectedStatusCode: "200",
				DNSRecordType:      tt.dnsRecordType,
				DNSNameserver:      tt.dnsNameserver,
				DNSExpectedAnswer:  tt.dnsExpectedAnswer,
			}

			var diags diag.Diagnostics
			result := MapMonitorCommonFields(monitor, &diags)

			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags.Errors())
			}

			if tt.wantRecTypeNull {
				assertNullString(t, "DNSRecordType", result.DNSRecordType)
			} else {
				assertStringField(t, "DNSRecordType", result.DNSRecordType, *tt.dnsRecordType)
			}
			if tt.wantNSNull {
				assertNullString(t, "DNSNameserver", result.DNSNameserver)
			} else {
				assertStringField(t, "DNSNameserver", result.DNSNameserver, *tt.dnsNameserver)
			}
			if tt.wantAnswerNull {
				assertNullString(t, "DNSExpectedAnswer", result.DNSExpectedAnswer)
			} else {
				assertStringField(t, "DNSExpectedAnswer", result.DNSExpectedAnswer, *tt.dnsExpectedAnswer)
			}
		})
	}
}

func TestMapMonitorCommonFields_SSLExpiration(t *testing.T) {
	tests := []struct {
		name      string
		sslExp    *int
		wantNull  bool
		wantValue int64
	}{
		{
			name:     "nil maps to null",
			sslExp:   nil,
			wantNull: true,
		},
		{
			name:      "positive value",
			sslExp:    testutil.Ptr(90),
			wantNull:  false,
			wantValue: 90,
		},
		{
			name:      "zero value",
			sslExp:    testutil.Ptr(0),
			wantNull:  false,
			wantValue: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := &client.Monitor{
				UUID:               "mon_ssl",
				Name:               "SSL Test",
				URL:                "https://example.com",
				Protocol:           "http",
				HTTPMethod:         "GET",
				CheckFrequency:     60,
				ExpectedStatusCode: "200",
				SSLExpiration:      tt.sslExp,
			}

			var diags diag.Diagnostics
			result := MapMonitorCommonFields(monitor, &diags)

			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags.Errors())
			}

			if tt.wantNull {
				assertNullInt64(t, "SSLExpiration", result.SSLExpiration)
			} else {
				assertInt64Field(t, "SSLExpiration", result.SSLExpiration, tt.wantValue)
			}
		})
	}
}

func TestMapMonitorCommonFields_RegionsVariants(t *testing.T) {
	tests := []struct {
		name     string
		regions  []string
		wantNull bool
		wantLen  int
	}{
		{
			name:     "nil regions maps to null",
			regions:  nil,
			wantNull: true,
		},
		{
			name:     "empty regions maps to null",
			regions:  []string{},
			wantNull: true,
		},
		{
			name:     "single region",
			regions:  []string{"virginia"},
			wantNull: false,
			wantLen:  1,
		},
		{
			name:     "all regions",
			regions:  []string{"london", "frankfurt", "singapore", "sydney", "tokyo", "virginia", "saopaulo", "bahrain"},
			wantNull: false,
			wantLen:  8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := &client.Monitor{
				UUID:               "mon_reg",
				Name:               "Regions Test",
				URL:                "https://example.com",
				Protocol:           "http",
				HTTPMethod:         "GET",
				CheckFrequency:     60,
				ExpectedStatusCode: "200",
				Regions:            tt.regions,
			}

			var diags diag.Diagnostics
			result := MapMonitorCommonFields(monitor, &diags)

			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags.Errors())
			}

			if tt.wantNull {
				if !result.Regions.IsNull() {
					t.Error("expected Regions to be null")
				}
			} else {
				if result.Regions.IsNull() {
					t.Fatal("expected Regions to be non-null")
				}
				elems := result.Regions.Elements()
				if len(elems) != tt.wantLen {
					t.Errorf("expected %d regions, got %d", tt.wantLen, len(elems))
				}
				for i, region := range tt.regions {
					got := elems[i].(types.String).ValueString()
					if got != region {
						t.Errorf("region[%d] = %s, want %s", i, got, region)
					}
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// MapHealthcheckCommonFields
// ---------------------------------------------------------------------------

func TestMapHealthcheckCommonFields_NilReturnsAllNulls(t *testing.T) {
	f := MapHealthcheckCommonFields(nil)

	nullStrings := map[string]types.String{
		"ID":               f.ID,
		"Name":             f.Name,
		"PingURL":          f.PingURL,
		"Cron":             f.Cron,
		"Timezone":         f.Timezone,
		"PeriodType":       f.PeriodType,
		"GracePeriodType":  f.GracePeriodType,
		"EscalationPolicy": f.EscalationPolicy,
		"LastPing":         f.LastPing,
		"CreatedAt":        f.CreatedAt,
	}
	for name, field := range nullStrings {
		if !field.IsNull() {
			t.Errorf("expected %s to be null for nil healthcheck", name)
		}
	}

	nullInt64s := map[string]types.Int64{
		"PeriodValue":      f.PeriodValue,
		"GracePeriodValue": f.GracePeriodValue,
		"Period":           f.Period,
		"GracePeriod":      f.GracePeriod,
	}
	for name, field := range nullInt64s {
		if !field.IsNull() {
			t.Errorf("expected %s to be null for nil healthcheck", name)
		}
	}

	nullBools := map[string]types.Bool{
		"IsPaused": f.IsPaused,
		"IsDown":   f.IsDown,
	}
	for name, field := range nullBools {
		if !field.IsNull() {
			t.Errorf("expected %s to be null for nil healthcheck", name)
		}
	}
}

func TestMapHealthcheckCommonFields_AllFieldsPopulated(t *testing.T) {
	pv := 5
	hc := &client.Healthcheck{
		UUID:             "hc_full",
		Name:             "Full HC",
		PingURL:          "https://hb.tinyping.io/hc_full",
		Cron:             "*/5 * * * *",
		Timezone:         "America/New_York",
		PeriodValue:      &pv,
		PeriodType:       "minutes",
		GracePeriodValue: 120,
		GracePeriodType:  "seconds",
		IsPaused:         false,
		IsDown:           true,
		Period:           300,
		GracePeriod:      120,
		LastPing:         "2026-03-01T12:00:00Z",
		CreatedAt:        "2026-01-15T08:00:00Z",
		EscalationPolicy: &client.EscalationPolicyReference{UUID: "ep_hc_full"},
	}

	f := MapHealthcheckCommonFields(hc)

	assertStringField(t, "ID", f.ID, "hc_full")
	assertStringField(t, "Name", f.Name, "Full HC")
	assertStringField(t, "PingURL", f.PingURL, "https://hb.tinyping.io/hc_full")
	assertStringField(t, "Cron", f.Cron, "*/5 * * * *")
	assertStringField(t, "Timezone", f.Timezone, "America/New_York")
	assertInt64Field(t, "PeriodValue", f.PeriodValue, 5)
	assertStringField(t, "PeriodType", f.PeriodType, "minutes")
	assertInt64Field(t, "GracePeriodValue", f.GracePeriodValue, 120)
	assertStringField(t, "GracePeriodType", f.GracePeriodType, "seconds")
	assertBoolField(t, "IsPaused", f.IsPaused, false)
	assertBoolField(t, "IsDown", f.IsDown, true)
	assertInt64Field(t, "Period", f.Period, 300)
	assertInt64Field(t, "GracePeriod", f.GracePeriod, 120)
	assertStringField(t, "LastPing", f.LastPing, "2026-03-01T12:00:00Z")
	assertStringField(t, "CreatedAt", f.CreatedAt, "2026-01-15T08:00:00Z")
	assertStringField(t, "EscalationPolicy", f.EscalationPolicy, "ep_hc_full")
}

func TestMapHealthcheckCommonFields_OptionalFieldsEmpty(t *testing.T) {
	hc := &client.Healthcheck{
		UUID:             "hc_sparse",
		Name:             "Sparse",
		PingURL:          "https://hb.tinyping.io/hc_sparse",
		GracePeriodValue: 60,
		GracePeriodType:  "seconds",
		Period:           0,
		GracePeriod:      60,
	}

	f := MapHealthcheckCommonFields(hc)

	assertStringField(t, "ID", f.ID, "hc_sparse")
	assertStringField(t, "Name", f.Name, "Sparse")
	assertNullString(t, "Cron", f.Cron)
	assertNullString(t, "Timezone", f.Timezone)
	assertNullInt64(t, "PeriodValue", f.PeriodValue)
	assertNullString(t, "PeriodType", f.PeriodType)
	assertNullString(t, "EscalationPolicy", f.EscalationPolicy)
	assertNullString(t, "LastPing", f.LastPing)
	assertNullString(t, "CreatedAt", f.CreatedAt)
	assertBoolField(t, "IsPaused", f.IsPaused, false)
	assertBoolField(t, "IsDown", f.IsDown, false)
}

// ---------------------------------------------------------------------------
// MapOutageNestedObjects
// ---------------------------------------------------------------------------

func TestMapOutageNestedObjects_NilReturnsNulls(t *testing.T) {
	var diags diag.Diagnostics
	monObj, ackObj := MapOutageNestedObjects(nil, &diags)

	if !monObj.IsNull() {
		t.Error("expected monitor object to be null for nil outage")
	}
	if !ackObj.IsNull() {
		t.Error("expected acknowledged_by object to be null for nil outage")
	}
	if diags.HasError() {
		t.Errorf("unexpected diagnostics: %v", diags.Errors())
	}
}

func TestMapOutageNestedObjects_ZeroValueMonitorReference(t *testing.T) {
	outage := &client.Outage{
		UUID:    "out_zero",
		Monitor: client.MonitorReference{},
	}

	var diags diag.Diagnostics
	monObj, ackObj := MapOutageNestedObjects(outage, &diags)

	if !monObj.IsNull() {
		t.Error("expected monitor object to be null for zero-value MonitorReference")
	}
	if !ackObj.IsNull() {
		t.Error("expected acknowledged_by to be null when AcknowledgedBy is nil")
	}
	if diags.HasError() {
		t.Errorf("unexpected diagnostics: %v", diags.Errors())
	}
}

func TestMapOutageNestedObjects_PopulatedMonitorNoAck(t *testing.T) {
	outage := &client.Outage{
		UUID: "out_mon",
		Monitor: client.MonitorReference{
			UUID:     "mon_ref_001",
			Name:     "Web Monitor",
			URL:      "https://web.example.com",
			Protocol: "http",
		},
	}

	var diags diag.Diagnostics
	monObj, ackObj := MapOutageNestedObjects(outage, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	if monObj.IsNull() {
		t.Fatal("expected monitor object to be non-null")
	}
	attrs := monObj.Attributes()
	assertObjectAttrString(t, attrs, "uuid", "mon_ref_001")
	assertObjectAttrString(t, attrs, "name", "Web Monitor")
	assertObjectAttrString(t, attrs, "url", "https://web.example.com")
	assertObjectAttrString(t, attrs, "protocol", "http")

	if !ackObj.IsNull() {
		t.Error("expected acknowledged_by to be null")
	}
}

func TestMapOutageNestedObjects_WithAcknowledgedBy(t *testing.T) {
	outage := &client.Outage{
		UUID: "out_ack",
		Monitor: client.MonitorReference{
			UUID:     "mon_ref_002",
			Name:     "API",
			URL:      "https://api.example.com",
			Protocol: "http",
		},
		AcknowledgedBy: &client.AcknowledgedByUser{
			UUID:  "user_42",
			Email: "admin@example.com",
			Name:  "Admin User",
		},
	}

	var diags diag.Diagnostics
	monObj, ackObj := MapOutageNestedObjects(outage, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	if monObj.IsNull() {
		t.Fatal("expected monitor object to be non-null")
	}
	if ackObj.IsNull() {
		t.Fatal("expected acknowledged_by object to be non-null")
	}

	ackAttrs := ackObj.Attributes()
	assertObjectAttrString(t, ackAttrs, "uuid", "user_42")
	assertObjectAttrString(t, ackAttrs, "email", "admin@example.com")
	assertObjectAttrString(t, ackAttrs, "name", "Admin User")
}

func TestMapOutageNestedObjects_PartialMonitorReference(t *testing.T) {
	// Only UUID set -- should still be non-null because UUID is non-empty.
	outage := &client.Outage{
		UUID: "out_partial",
		Monitor: client.MonitorReference{
			UUID: "mon_ref_003",
		},
	}

	var diags diag.Diagnostics
	monObj, _ := MapOutageNestedObjects(outage, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	if monObj.IsNull() {
		t.Fatal("expected monitor object to be non-null when UUID is set")
	}

	attrs := monObj.Attributes()
	assertObjectAttrString(t, attrs, "uuid", "mon_ref_003")
	// Empty strings for the rest
	assertObjectAttrString(t, attrs, "name", "")
	assertObjectAttrString(t, attrs, "url", "")
	assertObjectAttrString(t, attrs, "protocol", "")
}

// ---------------------------------------------------------------------------
// mapStringSliceToList
// ---------------------------------------------------------------------------

func TestMapStringSliceToList_NilInput(t *testing.T) {
	var diags diag.Diagnostics
	result := mapStringSliceToList(nil, &diags)

	if !result.IsNull() {
		t.Error("expected null list for nil input")
	}
	if diags.HasError() {
		t.Errorf("unexpected diagnostics: %v", diags.Errors())
	}
}

func TestMapStringSliceToList_EmptySlice(t *testing.T) {
	var diags diag.Diagnostics
	result := mapStringSliceToList([]string{}, &diags)

	if !result.IsNull() {
		t.Error("expected null list for empty slice")
	}
	if diags.HasError() {
		t.Errorf("unexpected diagnostics: %v", diags.Errors())
	}
}

func TestMapStringSliceToList_PopulatedSlice(t *testing.T) {
	var diags diag.Diagnostics
	input := []string{"alpha", "beta", "gamma"}
	result := mapStringSliceToList(input, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	if result.IsNull() {
		t.Fatal("expected non-null list")
	}

	elems := result.Elements()
	if len(elems) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(elems))
	}
	for i, want := range input {
		got := elems[i].(types.String).ValueString()
		if got != want {
			t.Errorf("element[%d] = %s, want %s", i, got, want)
		}
	}
}

func TestMapStringSliceToList_SingleElement(t *testing.T) {
	var diags diag.Diagnostics
	result := mapStringSliceToList([]string{"only"}, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	if result.IsNull() {
		t.Fatal("expected non-null list for single element")
	}
	elems := result.Elements()
	if len(elems) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elems))
	}
	if elems[0].(types.String).ValueString() != "only" {
		t.Errorf("expected 'only', got %s", elems[0].(types.String).ValueString())
	}
}

// ---------------------------------------------------------------------------
// mapRequestHeadersToTFList
// ---------------------------------------------------------------------------

func TestMapRequestHeadersToTFList_NilInput(t *testing.T) {
	var diags diag.Diagnostics
	result := mapRequestHeadersToTFList(nil, &diags)

	if !result.IsNull() {
		t.Error("expected null list for nil input")
	}
	if diags.HasError() {
		t.Errorf("unexpected diagnostics: %v", diags.Errors())
	}
}

func TestMapRequestHeadersToTFList_EmptySlice(t *testing.T) {
	var diags diag.Diagnostics
	result := mapRequestHeadersToTFList([]client.RequestHeader{}, &diags)

	if !result.IsNull() {
		t.Error("expected null list for empty slice")
	}
	if diags.HasError() {
		t.Errorf("unexpected diagnostics: %v", diags.Errors())
	}
}

func TestMapRequestHeadersToTFList_PopulatedHeaders(t *testing.T) {
	var diags diag.Diagnostics
	headers := []client.RequestHeader{
		{Name: "Content-Type", Value: "application/json"},
		{Name: "X-Request-ID", Value: "req-12345"},
	}
	result := mapRequestHeadersToTFList(headers, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	if result.IsNull() {
		t.Fatal("expected non-null list")
	}

	elems := result.Elements()
	if len(elems) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(elems))
	}

	for i, want := range headers {
		obj := elems[i].(types.Object)
		attrs := obj.Attributes()
		gotName := attrs["name"].(types.String).ValueString()
		gotValue := attrs["value"].(types.String).ValueString()
		if gotName != want.Name {
			t.Errorf("header[%d].name = %s, want %s", i, gotName, want.Name)
		}
		if gotValue != want.Value {
			t.Errorf("header[%d].value = %s, want %s", i, gotValue, want.Value)
		}
	}
}

// ---------------------------------------------------------------------------
// mapTFListToRequestHeaders
// ---------------------------------------------------------------------------

func TestMapTFListToRequestHeaders_NullListReturnsNil(t *testing.T) {
	var diags diag.Diagnostics
	result := mapTFListToRequestHeaders(
		types.ListNull(types.ObjectType{AttrTypes: RequestHeaderAttrTypes()}),
		&diags,
	)

	if result != nil {
		t.Error("expected nil for null list")
	}
	if diags.HasError() {
		t.Errorf("unexpected diagnostics: %v", diags.Errors())
	}
}

func TestMapTFListToRequestHeaders_PopulatedList(t *testing.T) {
	var diags diag.Diagnostics

	h1, _ := types.ObjectValue(RequestHeaderAttrTypes(), map[string]attr.Value{
		"name":  types.StringValue("Accept"),
		"value": types.StringValue("text/html"),
	})
	h2, _ := types.ObjectValue(RequestHeaderAttrTypes(), map[string]attr.Value{
		"name":  types.StringValue("Cache-Control"),
		"value": types.StringValue("no-cache"),
	})

	list, _ := types.ListValue(
		types.ObjectType{AttrTypes: RequestHeaderAttrTypes()},
		[]attr.Value{h1, h2},
	)
	result := mapTFListToRequestHeaders(list, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 headers, got %d", len(result))
	}
	if result[0].Name != "Accept" || result[0].Value != "text/html" {
		t.Errorf("header[0] = {%s, %s}, want {Accept, text/html}", result[0].Name, result[0].Value)
	}
	if result[1].Name != "Cache-Control" || result[1].Value != "no-cache" {
		t.Errorf("header[1] = {%s, %s}, want {Cache-Control, no-cache}", result[1].Name, result[1].Value)
	}
}

func TestMapTFListToRequestHeaders_SkipsNullNameValue(t *testing.T) {
	var diags diag.Diagnostics

	nullHeader, _ := types.ObjectValue(RequestHeaderAttrTypes(), map[string]attr.Value{
		"name":  types.StringNull(),
		"value": types.StringNull(),
	})
	validHeader, _ := types.ObjectValue(RequestHeaderAttrTypes(), map[string]attr.Value{
		"name":  types.StringValue("X-Keep"),
		"value": types.StringValue("yes"),
	})

	list, _ := types.ListValue(
		types.ObjectType{AttrTypes: RequestHeaderAttrTypes()},
		[]attr.Value{nullHeader, validHeader},
	)
	result := mapTFListToRequestHeaders(list, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 header (null pair skipped), got %d", len(result))
	}
	if result[0].Name != "X-Keep" {
		t.Errorf("expected header name 'X-Keep', got %s", result[0].Name)
	}
}

func TestMapTFListToRequestHeaders_SkipsNullNameOnly(t *testing.T) {
	var diags diag.Diagnostics

	// If name is null but value is not, the pair is still skipped
	header, _ := types.ObjectValue(RequestHeaderAttrTypes(), map[string]attr.Value{
		"name":  types.StringNull(),
		"value": types.StringValue("orphan-value"),
	})

	list, _ := types.ListValue(
		types.ObjectType{AttrTypes: RequestHeaderAttrTypes()},
		[]attr.Value{header},
	)
	result := mapTFListToRequestHeaders(list, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	if len(result) != 0 {
		t.Errorf("expected 0 headers (null name skips pair), got %d", len(result))
	}
}

// ---------------------------------------------------------------------------
// Roundtrip: mapRequestHeadersToTFList -> mapTFListToRequestHeaders
// ---------------------------------------------------------------------------

func TestRequestHeaders_Roundtrip(t *testing.T) {
	original := []client.RequestHeader{
		{Name: "Authorization", Value: "Bearer abc"},
		{Name: "X-Custom", Value: "value-with-special chars !@#"},
	}

	var diags diag.Diagnostics
	tfList := mapRequestHeadersToTFList(original, &diags)
	if diags.HasError() {
		t.Fatalf("forward mapping error: %v", diags.Errors())
	}

	roundtripped := mapTFListToRequestHeaders(tfList, &diags)
	if diags.HasError() {
		t.Fatalf("reverse mapping error: %v", diags.Errors())
	}

	if len(roundtripped) != len(original) {
		t.Fatalf("expected %d headers, got %d", len(original), len(roundtripped))
	}
	for i := range original {
		if roundtripped[i].Name != original[i].Name {
			t.Errorf("header[%d].Name = %s, want %s", i, roundtripped[i].Name, original[i].Name)
		}
		if roundtripped[i].Value != original[i].Value {
			t.Errorf("header[%d].Value = %s, want %s", i, roundtripped[i].Value, original[i].Value)
		}
	}
}

// ---------------------------------------------------------------------------
// Test helpers (unique to this file)
// ---------------------------------------------------------------------------

func assertStringField(t *testing.T, name string, field types.String, want string) {
	t.Helper()
	if field.IsNull() {
		t.Errorf("%s: expected value %q, got null", name, want)
		return
	}
	if field.ValueString() != want {
		t.Errorf("%s = %q, want %q", name, field.ValueString(), want)
	}
}

func assertNullString(t *testing.T, name string, field types.String) {
	t.Helper()
	if !field.IsNull() {
		t.Errorf("%s: expected null, got %q", name, field.ValueString())
	}
}

func assertInt64Field(t *testing.T, name string, field types.Int64, want int64) {
	t.Helper()
	if field.IsNull() {
		t.Errorf("%s: expected value %d, got null", name, want)
		return
	}
	if field.ValueInt64() != want {
		t.Errorf("%s = %d, want %d", name, field.ValueInt64(), want)
	}
}

func assertNullInt64(t *testing.T, name string, field types.Int64) {
	t.Helper()
	if !field.IsNull() {
		t.Errorf("%s: expected null, got %d", name, field.ValueInt64())
	}
}

func assertBoolField(t *testing.T, name string, field types.Bool, want bool) {
	t.Helper()
	if field.IsNull() {
		t.Errorf("%s: expected value %v, got null", name, want)
		return
	}
	if field.ValueBool() != want {
		t.Errorf("%s = %v, want %v", name, field.ValueBool(), want)
	}
}

func assertObjectAttrString(t *testing.T, attrs map[string]attr.Value, key string, want string) {
	t.Helper()
	val, ok := attrs[key]
	if !ok {
		t.Errorf("attribute %q not found in object", key)
		return
	}
	strVal, ok := val.(types.String)
	if !ok {
		t.Errorf("attribute %q is not types.String", key)
		return
	}
	if strVal.ValueString() != want {
		t.Errorf("attribute %q = %q, want %q", key, strVal.ValueString(), want)
	}
}
