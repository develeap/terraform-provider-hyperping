// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// ---------------------------------------------------------------------------
// applySimpleFieldChanges
// ---------------------------------------------------------------------------

func TestApplySimpleFieldChanges_noChanges(t *testing.T) {
	t.Parallel()

	plan := &MonitorResourceModel{
		Name:               types.StringValue("test"),
		URL:                types.StringValue("https://example.com"),
		Protocol:           types.StringValue("http"),
		HTTPMethod:         types.StringValue("GET"),
		CheckFrequency:     types.Int64Value(60),
		ExpectedStatusCode: types.StringValue("2xx"),
		FollowRedirects:    types.BoolValue(true),
		Paused:             types.BoolValue(false),
		ProjectUUID:        types.StringNull(),
	}

	state := &MonitorResourceModel{
		Name:               types.StringValue("test"),
		URL:                types.StringValue("https://example.com"),
		Protocol:           types.StringValue("http"),
		HTTPMethod:         types.StringValue("GET"),
		CheckFrequency:     types.Int64Value(60),
		ExpectedStatusCode: types.StringValue("2xx"),
		FollowRedirects:    types.BoolValue(true),
		Paused:             types.BoolValue(false),
		ProjectUUID:        types.StringNull(),
	}

	var req client.UpdateMonitorRequest
	r := &MonitorResource{}
	r.applySimpleFieldChanges(plan, state, &req)

	if req.Name != nil {
		t.Errorf("expected Name nil, got %v", *req.Name)
	}
	if req.URL != nil {
		t.Errorf("expected URL nil, got %v", *req.URL)
	}
	if req.Protocol != nil {
		t.Errorf("expected Protocol nil, got %v", *req.Protocol)
	}
	if req.HTTPMethod != nil {
		t.Errorf("expected HTTPMethod nil, got %v", *req.HTTPMethod)
	}
	if req.CheckFrequency != nil {
		t.Errorf("expected CheckFrequency nil, got %v", *req.CheckFrequency)
	}
	if req.ExpectedStatusCode != nil {
		t.Errorf("expected ExpectedStatusCode nil, got %v", *req.ExpectedStatusCode)
	}
	if req.FollowRedirects != nil {
		t.Errorf("expected FollowRedirects nil, got %v", *req.FollowRedirects)
	}
	if req.Paused != nil {
		t.Errorf("expected Paused nil, got %v", *req.Paused)
	}
	if req.ProjectUUID != nil {
		t.Errorf("expected ProjectUUID nil, got %v", *req.ProjectUUID)
	}
}

func TestApplySimpleFieldChanges_nameChanged(t *testing.T) {
	t.Parallel()

	plan := &MonitorResourceModel{
		Name:               types.StringValue("new-name"),
		URL:                types.StringValue("https://example.com"),
		Protocol:           types.StringValue("http"),
		HTTPMethod:         types.StringValue("GET"),
		CheckFrequency:     types.Int64Value(60),
		ExpectedStatusCode: types.StringValue("2xx"),
		FollowRedirects:    types.BoolValue(true),
		Paused:             types.BoolValue(false),
		ProjectUUID:        types.StringNull(),
	}

	state := &MonitorResourceModel{
		Name:               types.StringValue("old-name"),
		URL:                types.StringValue("https://example.com"),
		Protocol:           types.StringValue("http"),
		HTTPMethod:         types.StringValue("GET"),
		CheckFrequency:     types.Int64Value(60),
		ExpectedStatusCode: types.StringValue("2xx"),
		FollowRedirects:    types.BoolValue(true),
		Paused:             types.BoolValue(false),
		ProjectUUID:        types.StringNull(),
	}

	var req client.UpdateMonitorRequest
	r := &MonitorResource{}
	r.applySimpleFieldChanges(plan, state, &req)

	if req.Name == nil || *req.Name != "new-name" {
		t.Fatalf("expected Name=%q, got %v", "new-name", req.Name)
	}
	// Other fields must remain nil.
	if req.URL != nil {
		t.Errorf("expected URL nil, got %v", *req.URL)
	}
	if req.Protocol != nil {
		t.Errorf("expected Protocol nil, got %v", *req.Protocol)
	}
}

func TestApplySimpleFieldChanges_multipleFieldsChanged(t *testing.T) {
	t.Parallel()

	plan := &MonitorResourceModel{
		Name:               types.StringValue("new-name"),
		URL:                types.StringValue("https://new.example.com"),
		Protocol:           types.StringValue("icmp"),
		HTTPMethod:         types.StringValue("POST"),
		CheckFrequency:     types.Int64Value(300),
		ExpectedStatusCode: types.StringValue("200"),
		FollowRedirects:    types.BoolValue(false),
		Paused:             types.BoolValue(true),
		ProjectUUID:        types.StringValue("proj-123"),
	}

	state := &MonitorResourceModel{
		Name:               types.StringValue("old-name"),
		URL:                types.StringValue("https://old.example.com"),
		Protocol:           types.StringValue("http"),
		HTTPMethod:         types.StringValue("GET"),
		CheckFrequency:     types.Int64Value(60),
		ExpectedStatusCode: types.StringValue("2xx"),
		FollowRedirects:    types.BoolValue(true),
		Paused:             types.BoolValue(false),
		ProjectUUID:        types.StringNull(),
	}

	var req client.UpdateMonitorRequest
	r := &MonitorResource{}
	r.applySimpleFieldChanges(plan, state, &req)

	if req.Name == nil || *req.Name != "new-name" {
		t.Errorf("expected Name=%q, got %v", "new-name", req.Name)
	}
	if req.URL == nil || *req.URL != "https://new.example.com" {
		t.Errorf("expected URL=%q, got %v", "https://new.example.com", req.URL)
	}
	if req.Protocol == nil || *req.Protocol != "icmp" {
		t.Errorf("expected Protocol=%q, got %v", "icmp", req.Protocol)
	}
	if req.HTTPMethod == nil || *req.HTTPMethod != "POST" {
		t.Errorf("expected HTTPMethod=%q, got %v", "POST", req.HTTPMethod)
	}
	if req.CheckFrequency == nil || *req.CheckFrequency != 300 {
		t.Errorf("expected CheckFrequency=300, got %v", req.CheckFrequency)
	}
	if req.ExpectedStatusCode == nil || *req.ExpectedStatusCode != "200" {
		t.Errorf("expected ExpectedStatusCode=%q, got %v", "200", req.ExpectedStatusCode)
	}
	if req.FollowRedirects == nil || *req.FollowRedirects != false {
		t.Errorf("expected FollowRedirects=false, got %v", req.FollowRedirects)
	}
	if req.Paused == nil || *req.Paused != true {
		t.Errorf("expected Paused=true, got %v", req.Paused)
	}
	if req.ProjectUUID == nil || *req.ProjectUUID != "proj-123" {
		t.Errorf("expected ProjectUUID=%q, got %v", "proj-123", req.ProjectUUID)
	}
}

// ---------------------------------------------------------------------------
// applyHTTPFieldChanges
// ---------------------------------------------------------------------------

// buildHeaderList is a test helper that builds a types.List of request header objects.
func buildHeaderList(t *testing.T, headers []client.RequestHeader) types.List {
	t.Helper()
	if len(headers) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: RequestHeaderAttrTypes()})
	}

	values := make([]attr.Value, len(headers))
	for i, h := range headers {
		obj, diags := types.ObjectValue(RequestHeaderAttrTypes(), map[string]attr.Value{
			"name":  types.StringValue(h.Name),
			"value": types.StringValue(h.Value),
		})
		if diags.HasError() {
			t.Fatalf("failed to build header object: %v", diags)
		}
		values[i] = obj
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: RequestHeaderAttrTypes()}, values)
	if diags.HasError() {
		t.Fatalf("failed to build header list: %v", diags)
	}
	return list
}

func TestApplyHTTPFieldChanges_noChanges(t *testing.T) {
	t.Parallel()

	headers := buildHeaderList(t, []client.RequestHeader{{Name: "X-Test", Value: "val"}})
	plan := &MonitorResourceModel{
		RequestHeaders: headers,
		RequestBody:    types.StringValue("body"),
	}
	state := &MonitorResourceModel{
		RequestHeaders: headers,
		RequestBody:    types.StringValue("body"),
	}

	var req client.UpdateMonitorRequest
	var diags diag.Diagnostics
	applyHTTPFieldChanges(plan, state, &req, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.RequestHeaders != nil {
		t.Error("expected RequestHeaders nil")
	}
	if req.RequestBody != nil {
		t.Error("expected RequestBody nil")
	}
}

func TestApplyHTTPFieldChanges_headersChanged(t *testing.T) {
	t.Parallel()

	oldHeaders := buildHeaderList(t, []client.RequestHeader{{Name: "X-Old", Value: "old"}})
	newHeaders := buildHeaderList(t, []client.RequestHeader{{Name: "X-New", Value: "new"}})

	plan := &MonitorResourceModel{
		RequestHeaders: newHeaders,
		RequestBody:    types.StringNull(),
	}
	state := &MonitorResourceModel{
		RequestHeaders: oldHeaders,
		RequestBody:    types.StringNull(),
	}

	var req client.UpdateMonitorRequest
	var diags diag.Diagnostics
	applyHTTPFieldChanges(plan, state, &req, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.RequestHeaders == nil {
		t.Fatal("expected RequestHeaders to be set")
	}
	if len(*req.RequestHeaders) != 1 {
		t.Fatalf("expected 1 header, got %d", len(*req.RequestHeaders))
	}
	if (*req.RequestHeaders)[0].Name != "X-New" {
		t.Errorf("expected header name=%q, got %q", "X-New", (*req.RequestHeaders)[0].Name)
	}
}

func TestApplyHTTPFieldChanges_headersSetToNull(t *testing.T) {
	t.Parallel()

	oldHeaders := buildHeaderList(t, []client.RequestHeader{{Name: "X-Old", Value: "old"}})

	plan := &MonitorResourceModel{
		RequestHeaders: types.ListNull(types.ObjectType{AttrTypes: RequestHeaderAttrTypes()}),
		RequestBody:    types.StringNull(),
	}
	state := &MonitorResourceModel{
		RequestHeaders: oldHeaders,
		RequestBody:    types.StringNull(),
	}

	var req client.UpdateMonitorRequest
	var diags diag.Diagnostics
	applyHTTPFieldChanges(plan, state, &req, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.RequestHeaders == nil {
		t.Fatal("expected RequestHeaders to be set (empty slice)")
	}
	if len(*req.RequestHeaders) != 0 {
		t.Errorf("expected empty headers slice, got %d", len(*req.RequestHeaders))
	}
}

func TestApplyHTTPFieldChanges_requestBodyChanged(t *testing.T) {
	t.Parallel()

	nullHeaders := types.ListNull(types.ObjectType{AttrTypes: RequestHeaderAttrTypes()})
	plan := &MonitorResourceModel{
		RequestHeaders: nullHeaders,
		RequestBody:    types.StringValue("new-body"),
	}
	state := &MonitorResourceModel{
		RequestHeaders: nullHeaders,
		RequestBody:    types.StringValue("old-body"),
	}

	var req client.UpdateMonitorRequest
	var diags diag.Diagnostics
	applyHTTPFieldChanges(plan, state, &req, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.RequestBody == nil || *req.RequestBody != "new-body" {
		t.Errorf("expected RequestBody=%q, got %v", "new-body", req.RequestBody)
	}
}

func TestApplyHTTPFieldChanges_requestBodyRemoved(t *testing.T) {
	t.Parallel()

	nullHeaders := types.ListNull(types.ObjectType{AttrTypes: RequestHeaderAttrTypes()})
	plan := &MonitorResourceModel{
		RequestHeaders: nullHeaders,
		RequestBody:    types.StringNull(),
	}
	state := &MonitorResourceModel{
		RequestHeaders: nullHeaders,
		RequestBody:    types.StringValue("had-a-body"),
	}

	var req client.UpdateMonitorRequest
	var diags diag.Diagnostics
	applyHTTPFieldChanges(plan, state, &req, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.RequestBody == nil {
		t.Fatal("expected RequestBody to be set (empty string)")
	}
	if *req.RequestBody != "" {
		t.Errorf("expected empty string, got %q", *req.RequestBody)
	}
}

func TestApplyHTTPFieldChanges_unknownHeadersSkipped(t *testing.T) {
	t.Parallel()

	unknownHeaders := types.ListUnknown(types.ObjectType{AttrTypes: RequestHeaderAttrTypes()})
	oldHeaders := buildHeaderList(t, []client.RequestHeader{{Name: "X-Old", Value: "old"}})

	plan := &MonitorResourceModel{
		RequestHeaders: unknownHeaders,
		RequestBody:    types.StringNull(),
	}
	state := &MonitorResourceModel{
		RequestHeaders: oldHeaders,
		RequestBody:    types.StringNull(),
	}

	var req client.UpdateMonitorRequest
	var diags diag.Diagnostics
	applyHTTPFieldChanges(plan, state, &req, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	// Unknown headers are skipped entirely.
	if req.RequestHeaders != nil {
		t.Error("expected RequestHeaders nil when plan headers are unknown")
	}
}

// ---------------------------------------------------------------------------
// applyMonitoringFieldChanges
// ---------------------------------------------------------------------------

// buildStringList is a test helper that builds a types.List of strings.
func buildStringList(t *testing.T, values []string) types.List {
	t.Helper()
	if values == nil {
		return types.ListNull(types.StringType)
	}

	elems := make([]attr.Value, len(values))
	for i, v := range values {
		elems[i] = types.StringValue(v)
	}

	list, diags := types.ListValue(types.StringType, elems)
	if diags.HasError() {
		t.Fatalf("failed to build string list: %v", diags)
	}
	return list
}

func TestApplyMonitoringFieldChanges_regionsChanged(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	plan := &MonitorResourceModel{
		Regions:           buildStringList(t, []string{"london", "frankfurt"}),
		AlertsWait:        types.Int64Null(),
		EscalationPolicy:  types.StringNull(),
		RequiredKeyword:   types.StringNull(),
		Port:              types.Int64Null(),
		DNSRecordType:     types.StringNull(),
		DNSNameserver:     types.StringNull(),
		DNSExpectedAnswer: types.StringNull(),
	}
	state := &MonitorResourceModel{
		Regions:           buildStringList(t, []string{"virginia"}),
		AlertsWait:        types.Int64Null(),
		EscalationPolicy:  types.StringNull(),
		RequiredKeyword:   types.StringNull(),
		Port:              types.Int64Null(),
		DNSRecordType:     types.StringNull(),
		DNSNameserver:     types.StringNull(),
		DNSExpectedAnswer: types.StringNull(),
	}

	var req client.UpdateMonitorRequest
	var diags diag.Diagnostics
	applyMonitoringFieldChanges(ctx, plan, state, &req, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.Regions == nil {
		t.Fatal("expected Regions to be set")
	}
	if len(*req.Regions) != 2 {
		t.Fatalf("expected 2 regions, got %d", len(*req.Regions))
	}
	if (*req.Regions)[0] != "london" || (*req.Regions)[1] != "frankfurt" {
		t.Errorf("expected [london frankfurt], got %v", *req.Regions)
	}
}

func TestApplyMonitoringFieldChanges_regionsRemoved(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	plan := &MonitorResourceModel{
		Regions:           types.ListNull(types.StringType),
		AlertsWait:        types.Int64Null(),
		EscalationPolicy:  types.StringNull(),
		RequiredKeyword:   types.StringNull(),
		Port:              types.Int64Null(),
		DNSRecordType:     types.StringNull(),
		DNSNameserver:     types.StringNull(),
		DNSExpectedAnswer: types.StringNull(),
	}
	state := &MonitorResourceModel{
		Regions:           buildStringList(t, []string{"london"}),
		AlertsWait:        types.Int64Null(),
		EscalationPolicy:  types.StringNull(),
		RequiredKeyword:   types.StringNull(),
		Port:              types.Int64Null(),
		DNSRecordType:     types.StringNull(),
		DNSNameserver:     types.StringNull(),
		DNSExpectedAnswer: types.StringNull(),
	}

	var req client.UpdateMonitorRequest
	var diags diag.Diagnostics
	applyMonitoringFieldChanges(ctx, plan, state, &req, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.Regions == nil {
		t.Fatal("expected Regions to be set (empty slice)")
	}
	if len(*req.Regions) != 0 {
		t.Errorf("expected empty regions slice, got %d", len(*req.Regions))
	}
}

func TestApplyMonitoringFieldChanges_alertsWaitRemoved(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	plan := &MonitorResourceModel{
		Regions:           types.ListNull(types.StringType),
		AlertsWait:        types.Int64Null(),
		EscalationPolicy:  types.StringNull(),
		RequiredKeyword:   types.StringNull(),
		Port:              types.Int64Null(),
		DNSRecordType:     types.StringNull(),
		DNSNameserver:     types.StringNull(),
		DNSExpectedAnswer: types.StringNull(),
	}
	state := &MonitorResourceModel{
		Regions:           types.ListNull(types.StringType),
		AlertsWait:        types.Int64Value(5),
		EscalationPolicy:  types.StringNull(),
		RequiredKeyword:   types.StringNull(),
		Port:              types.Int64Null(),
		DNSRecordType:     types.StringNull(),
		DNSNameserver:     types.StringNull(),
		DNSExpectedAnswer: types.StringNull(),
	}

	var req client.UpdateMonitorRequest
	var diags diag.Diagnostics
	applyMonitoringFieldChanges(ctx, plan, state, &req, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.AlertsWait == nil {
		t.Fatal("expected AlertsWait to be set")
	}
	if *req.AlertsWait != 0 {
		t.Errorf("expected AlertsWait=0, got %d", *req.AlertsWait)
	}
}

func TestApplyMonitoringFieldChanges_escalationPolicyRemoved(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	plan := &MonitorResourceModel{
		Regions:           types.ListNull(types.StringType),
		AlertsWait:        types.Int64Null(),
		EscalationPolicy:  types.StringNull(),
		RequiredKeyword:   types.StringNull(),
		Port:              types.Int64Null(),
		DNSRecordType:     types.StringNull(),
		DNSNameserver:     types.StringNull(),
		DNSExpectedAnswer: types.StringNull(),
	}
	state := &MonitorResourceModel{
		Regions:           types.ListNull(types.StringType),
		AlertsWait:        types.Int64Null(),
		EscalationPolicy:  types.StringValue("policy-uuid-123"),
		RequiredKeyword:   types.StringNull(),
		Port:              types.Int64Null(),
		DNSRecordType:     types.StringNull(),
		DNSNameserver:     types.StringNull(),
		DNSExpectedAnswer: types.StringNull(),
	}

	var req client.UpdateMonitorRequest
	var diags diag.Diagnostics
	applyMonitoringFieldChanges(ctx, plan, state, &req, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.EscalationPolicy == nil {
		t.Fatal("expected EscalationPolicy to be set")
	}
	if *req.EscalationPolicy != "none" {
		t.Errorf("expected EscalationPolicy=%q, got %q", "none", *req.EscalationPolicy)
	}
}

func TestApplyMonitoringFieldChanges_requiredKeywordRemoved(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	plan := &MonitorResourceModel{
		Regions:           types.ListNull(types.StringType),
		AlertsWait:        types.Int64Null(),
		EscalationPolicy:  types.StringNull(),
		RequiredKeyword:   types.StringNull(),
		Port:              types.Int64Null(),
		DNSRecordType:     types.StringNull(),
		DNSNameserver:     types.StringNull(),
		DNSExpectedAnswer: types.StringNull(),
	}
	state := &MonitorResourceModel{
		Regions:           types.ListNull(types.StringType),
		AlertsWait:        types.Int64Null(),
		EscalationPolicy:  types.StringNull(),
		RequiredKeyword:   types.StringValue("healthy"),
		Port:              types.Int64Null(),
		DNSRecordType:     types.StringNull(),
		DNSNameserver:     types.StringNull(),
		DNSExpectedAnswer: types.StringNull(),
	}

	var req client.UpdateMonitorRequest
	var diags diag.Diagnostics
	applyMonitoringFieldChanges(ctx, plan, state, &req, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.RequiredKeyword == nil {
		t.Fatal("expected RequiredKeyword to be set")
	}
	if *req.RequiredKeyword != "" {
		t.Errorf("expected empty string, got %q", *req.RequiredKeyword)
	}
}

func TestApplyMonitoringFieldChanges_dnsFieldsChanged(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	plan := &MonitorResourceModel{
		Regions:           types.ListNull(types.StringType),
		AlertsWait:        types.Int64Null(),
		EscalationPolicy:  types.StringNull(),
		RequiredKeyword:   types.StringNull(),
		Port:              types.Int64Null(),
		DNSRecordType:     types.StringValue("AAAA"),
		DNSNameserver:     types.StringValue("1.1.1.1"),
		DNSExpectedAnswer: types.StringValue("::1"),
	}
	state := &MonitorResourceModel{
		Regions:           types.ListNull(types.StringType),
		AlertsWait:        types.Int64Null(),
		EscalationPolicy:  types.StringNull(),
		RequiredKeyword:   types.StringNull(),
		Port:              types.Int64Null(),
		DNSRecordType:     types.StringValue("A"),
		DNSNameserver:     types.StringValue("8.8.8.8"),
		DNSExpectedAnswer: types.StringValue("1.2.3.4"),
	}

	var req client.UpdateMonitorRequest
	var diags diag.Diagnostics
	applyMonitoringFieldChanges(ctx, plan, state, &req, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.DNSRecordType == nil || *req.DNSRecordType != "AAAA" {
		t.Errorf("expected DNSRecordType=%q, got %v", "AAAA", req.DNSRecordType)
	}
	if req.DNSNameserver == nil || *req.DNSNameserver != "1.1.1.1" {
		t.Errorf("expected DNSNameserver=%q, got %v", "1.1.1.1", req.DNSNameserver)
	}
	if req.DNSExpectedAnswer == nil || *req.DNSExpectedAnswer != "::1" {
		t.Errorf("expected DNSExpectedAnswer=%q, got %v", "::1", req.DNSExpectedAnswer)
	}
}

func TestApplyMonitoringFieldChanges_dnsFieldsRemoved(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	plan := &MonitorResourceModel{
		Regions:           types.ListNull(types.StringType),
		AlertsWait:        types.Int64Null(),
		EscalationPolicy:  types.StringNull(),
		RequiredKeyword:   types.StringNull(),
		Port:              types.Int64Null(),
		DNSRecordType:     types.StringNull(),
		DNSNameserver:     types.StringNull(),
		DNSExpectedAnswer: types.StringNull(),
	}
	state := &MonitorResourceModel{
		Regions:           types.ListNull(types.StringType),
		AlertsWait:        types.Int64Null(),
		EscalationPolicy:  types.StringNull(),
		RequiredKeyword:   types.StringNull(),
		Port:              types.Int64Null(),
		DNSRecordType:     types.StringValue("A"),
		DNSNameserver:     types.StringValue("8.8.8.8"),
		DNSExpectedAnswer: types.StringValue("1.2.3.4"),
	}

	var req client.UpdateMonitorRequest
	var diags diag.Diagnostics
	applyMonitoringFieldChanges(ctx, plan, state, &req, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	// DNSRecordType: null plan uses tfStringToPtr which returns nil for null
	if req.DNSRecordType != nil {
		t.Errorf("expected DNSRecordType nil, got %v", *req.DNSRecordType)
	}
	// DNSNameserver: null plan sends empty string
	if req.DNSNameserver == nil || *req.DNSNameserver != "" {
		t.Errorf("expected DNSNameserver empty string, got %v", req.DNSNameserver)
	}
	// DNSExpectedAnswer: null plan sends empty string
	if req.DNSExpectedAnswer == nil || *req.DNSExpectedAnswer != "" {
		t.Errorf("expected DNSExpectedAnswer empty string, got %v", req.DNSExpectedAnswer)
	}
}

func TestApplyMonitoringFieldChanges_portChanged(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	plan := &MonitorResourceModel{
		Regions:           types.ListNull(types.StringType),
		AlertsWait:        types.Int64Null(),
		EscalationPolicy:  types.StringNull(),
		RequiredKeyword:   types.StringNull(),
		Port:              types.Int64Value(8080),
		DNSRecordType:     types.StringNull(),
		DNSNameserver:     types.StringNull(),
		DNSExpectedAnswer: types.StringNull(),
	}
	state := &MonitorResourceModel{
		Regions:           types.ListNull(types.StringType),
		AlertsWait:        types.Int64Null(),
		EscalationPolicy:  types.StringNull(),
		RequiredKeyword:   types.StringNull(),
		Port:              types.Int64Value(443),
		DNSRecordType:     types.StringNull(),
		DNSNameserver:     types.StringNull(),
		DNSExpectedAnswer: types.StringNull(),
	}

	var req client.UpdateMonitorRequest
	var diags diag.Diagnostics
	applyMonitoringFieldChanges(ctx, plan, state, &req, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.Port == nil || *req.Port != 8080 {
		t.Errorf("expected Port=8080, got %v", req.Port)
	}
}

// ---------------------------------------------------------------------------
// Request spy tests
// ---------------------------------------------------------------------------

func TestMockServer_RequestSpy_RecordsCreateAndGet(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	// Initially no requests recorded
	if reqs := server.getRequests(); len(reqs) != 0 {
		t.Fatalf("expected 0 requests, got %d", len(reqs))
	}
	if server.lastRequest() != nil {
		t.Fatal("expected nil lastRequest initially")
	}

	// Create a monitor via the mock
	body := `{"name":"spy-test","url":"https://example.com"}`
	resp, err := http.Post(server.URL+client.MonitorsBasePath, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	resp.Body.Close()

	reqs := server.getRequests()
	if len(reqs) != 1 {
		t.Fatalf("expected 1 request, got %d", len(reqs))
	}
	if reqs[0].Method != "POST" {
		t.Errorf("expected POST, got %s", reqs[0].Method)
	}
	if reqs[0].Path != client.MonitorsBasePath {
		t.Errorf("expected path %s, got %s", client.MonitorsBasePath, reqs[0].Path)
	}
	if reqs[0].Body == nil {
		t.Fatal("expected non-nil body")
	}
	if reqs[0].Body["name"] != "spy-test" {
		t.Errorf("expected body.name=spy-test, got %v", reqs[0].Body["name"])
	}

	last := server.lastRequest()
	if last == nil || last.Method != "POST" {
		t.Errorf("lastRequest should be POST, got %v", last)
	}
}

func TestMockServer_RequestSpy_GETHasNilBody(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + client.MonitorsBasePath)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	resp.Body.Close()

	reqs := server.getRequests()
	if len(reqs) != 1 {
		t.Fatalf("expected 1 request, got %d", len(reqs))
	}
	if reqs[0].Body != nil {
		t.Errorf("GET body should be nil, got %v", reqs[0].Body)
	}
}
