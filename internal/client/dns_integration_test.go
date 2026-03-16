// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

//go:build integration

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

// =============================================================================
// Comprehensive integration tests for dns_record_type PUT workaround removal
// and DNS monitor support.
//
// Run all:
//   export $(grep HYPERPING_API_KEY .env | head -1)
//   go test ./internal/client/ -tags integration -run TestAPI_ -v
//
// These tests create real monitors against the Hyperping API and clean up after
// themselves. Each test is independent and safe to run in parallel.
// =============================================================================

func mustAPIKey(t *testing.T) string {
	t.Helper()
	key := os.Getenv("HYPERPING_API_KEY")
	if key == "" {
		t.Skip("HYPERPING_API_KEY not set")
	}
	return key
}

func mustCreateMonitor(t *testing.T, c *Client, ctx context.Context, req CreateMonitorRequest) *Monitor {
	t.Helper()
	m, err := c.CreateMonitor(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	t.Logf("Created %s monitor %q (uuid: %s)", m.Protocol, m.Name, m.UUID)
	return m
}

func cleanupMonitor(t *testing.T, c *Client, uuid string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := c.DeleteMonitor(ctx, uuid); err != nil {
		t.Logf("Warning: failed to delete monitor %s: %v", uuid, err)
	} else {
		t.Logf("Cleaned up monitor %s", uuid)
	}
}

func safeDeref(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func strp(s string) *string { return &s }

func uniqueName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano()%1_000_000)
}

// ─── Workaround removal tests ───────────────────────────────────────────────

// TestAPI_PUT_HTTPMonitor_WithoutDNSRecordType verifies PUT on an HTTP monitor
// succeeds when dns_record_type is absent from the request body.
// This was the original bug: API returned 422 for missing dns_record_type.
func TestAPI_PUT_HTTPMonitor_WithoutDNSRecordType(t *testing.T) {
	c := NewClient(mustAPIKey(t))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	m := mustCreateMonitor(t, c, ctx, CreateMonitorRequest{
		Name: uniqueName("put-no-drt"), URL: "https://httpstat.us/200", Protocol: "http",
	})
	defer cleanupMonitor(t, c, m.UUID)

	updated, err := c.UpdateMonitor(ctx, m.UUID, UpdateMonitorRequest{
		Name: strp(m.Name + "-updated"),
	})
	if err != nil {
		t.Fatalf("PUT without dns_record_type FAILED: %v", err)
	}
	t.Logf("PASS: PUT without dns_record_type succeeded (name: %s)", updated.Name)
}

// TestAPI_PUT_HTTPMonitor_WithNullDNSRecordType sends dns_record_type: null explicitly.
// This was another variant of the bug (v1.3.7 regression).
func TestAPI_PUT_HTTPMonitor_WithNullDNSRecordType(t *testing.T) {
	c := NewClient(mustAPIKey(t))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	m := mustCreateMonitor(t, c, ctx, CreateMonitorRequest{
		Name: uniqueName("put-null-drt"), URL: "https://httpstat.us/200", Protocol: "http",
	})
	defer cleanupMonitor(t, c, m.UUID)

	// Build raw JSON with explicit null to bypass Go's omitempty
	body := map[string]interface{}{
		"name":            m.Name + "-null-test",
		"dns_record_type": nil,
	}
	var result Monitor
	err := c.doRequest(ctx, "PUT", fmt.Sprintf("/v1/monitors/%s", m.UUID), body, &result)
	if err != nil {
		t.Fatalf("PUT with dns_record_type:null FAILED: %v", err)
	}
	t.Logf("PASS: PUT with dns_record_type:null succeeded (name: %s)", result.Name)
}

// TestAPI_PUT_HTTPMonitor_WithEmptyStringDNSRecordType sends dns_record_type: "".
// This was the third rejection variant.
func TestAPI_PUT_HTTPMonitor_WithEmptyStringDNSRecordType(t *testing.T) {
	c := NewClient(mustAPIKey(t))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	m := mustCreateMonitor(t, c, ctx, CreateMonitorRequest{
		Name: uniqueName("put-empty-drt"), URL: "https://httpstat.us/200", Protocol: "http",
	})
	defer cleanupMonitor(t, c, m.UUID)

	body := map[string]interface{}{
		"name":            m.Name + "-empty-test",
		"dns_record_type": "",
	}
	var result Monitor
	err := c.doRequest(ctx, "PUT", fmt.Sprintf("/v1/monitors/%s", m.UUID), body, &result)
	if err != nil {
		t.Logf("PUT with dns_record_type:\"\" FAILED: %v", err)
		t.Logf("INFO: API still rejects empty string for dns_record_type. This is fine — " +
			"the provider never sends an empty string (uses omitempty).")
	} else {
		t.Logf("PASS: PUT with dns_record_type:\"\" succeeded (name: %s)", result.Name)
	}
}

// TestAPI_PUT_ICMPMonitor_WithoutDNSRecordType verifies the bug is fixed for ICMP too.
func TestAPI_PUT_ICMPMonitor_WithoutDNSRecordType(t *testing.T) {
	c := NewClient(mustAPIKey(t))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	m := mustCreateMonitor(t, c, ctx, CreateMonitorRequest{
		Name: uniqueName("icmp-no-drt"), URL: "1.1.1.1", Protocol: "icmp",
	})
	defer cleanupMonitor(t, c, m.UUID)

	updated, err := c.UpdateMonitor(ctx, m.UUID, UpdateMonitorRequest{
		Name: strp(m.Name + "-updated"),
	})
	if err != nil {
		t.Fatalf("PUT on ICMP monitor without dns_record_type FAILED: %v", err)
	}
	t.Logf("PASS: ICMP monitor PUT without dns_record_type succeeded (name: %s)", updated.Name)
}

// TestAPI_PUT_PortMonitor_WithoutDNSRecordType verifies the bug is fixed for port too.
func TestAPI_PUT_PortMonitor_WithoutDNSRecordType(t *testing.T) {
	c := NewClient(mustAPIKey(t))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	port := 443
	m := mustCreateMonitor(t, c, ctx, CreateMonitorRequest{
		Name: uniqueName("port-no-drt"), URL: "httpstat.us", Protocol: "port", Port: &port,
	})
	defer cleanupMonitor(t, c, m.UUID)

	newPort := 8443
	updated, err := c.UpdateMonitor(ctx, m.UUID, UpdateMonitorRequest{
		Port: &newPort,
	})
	if err != nil {
		t.Fatalf("PUT on port monitor without dns_record_type FAILED: %v", err)
	}
	t.Logf("PASS: Port monitor PUT without dns_record_type succeeded (port: %v)", updated.Port)
}

// TestAPI_PUT_MultipleFieldsUpdate_NoDNSRecordType does a realistic multi-field
// update (name + url + check_frequency) without dns_record_type.
func TestAPI_PUT_MultipleFieldsUpdate_NoDNSRecordType(t *testing.T) {
	c := NewClient(mustAPIKey(t))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	m := mustCreateMonitor(t, c, ctx, CreateMonitorRequest{
		Name: uniqueName("multi-update"), URL: "https://httpstat.us/200", Protocol: "http",
	})
	defer cleanupMonitor(t, c, m.UUID)

	freq := 120
	updated, err := c.UpdateMonitor(ctx, m.UUID, UpdateMonitorRequest{
		Name:           strp(m.Name + "-multi"),
		URL:            strp("https://httpstat.us/201"),
		CheckFrequency: &freq,
	})
	if err != nil {
		t.Fatalf("Multi-field PUT without dns_record_type FAILED: %v", err)
	}
	if updated.CheckFrequency != 120 {
		t.Errorf("Expected check_frequency 120, got %d", updated.CheckFrequency)
	}
	t.Logf("PASS: Multi-field update succeeded (name: %s, freq: %d)", updated.Name, updated.CheckFrequency)
}

// TestAPI_PUT_RawHTTP_OmittedField uses raw HTTP to send a PUT with absolutely
// no dns_record_type in the JSON — not null, not empty, just absent.
func TestAPI_PUT_RawHTTP_OmittedField(t *testing.T) {
	apiKey := mustAPIKey(t)
	c := NewClient(apiKey)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	m := mustCreateMonitor(t, c, ctx, CreateMonitorRequest{
		Name: uniqueName("raw-http"), URL: "https://httpstat.us/200", Protocol: "http",
	})
	defer cleanupMonitor(t, c, m.UUID)

	// Build raw JSON manually — guaranteed no dns_record_type key
	rawJSON := fmt.Sprintf(`{"name":"%s-raw"}`, m.Name)
	url := fmt.Sprintf("https://api.hyperping.io/v1/monitors/%s", m.UUID)

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBufferString(rawJSON))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Raw HTTP PUT failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 422 {
		t.Fatalf("API returned 422 — bug has REGRESSED. Body: %s", string(body))
	}
	if resp.StatusCode != 200 {
		t.Fatalf("Unexpected status %d. Body: %s", resp.StatusCode, string(body))
	}

	var result Monitor
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	t.Logf("PASS: Raw HTTP PUT (no dns_record_type) returned 200 (name: %s)", result.Name)
}

// ─── DNS monitor CRUD tests ─────────────────────────────────────────────────

// TestAPI_DNS_FullCRUD tests create, read, update, and delete of a DNS monitor.
func TestAPI_DNS_FullCRUD(t *testing.T) {
	c := NewClient(mustAPIKey(t))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// CREATE
	m := mustCreateMonitor(t, c, ctx, CreateMonitorRequest{
		Name:              uniqueName("dns-crud"),
		URL:               "example.com",
		Protocol:          "dns",
		DNSRecordType:     strp("CNAME"),
		DNSNameserver:     strp("8.8.8.8"),
		DNSExpectedAnswer: strp("www.example.com"),
	})
	defer cleanupMonitor(t, c, m.UUID)

	// Verify create response
	if m.Protocol != "dns" {
		t.Errorf("Create: expected protocol 'dns', got %q", m.Protocol)
	}
	if safeDeref(m.DNSRecordType) != "CNAME" {
		t.Errorf("Create: expected dns_record_type 'CNAME', got %q", safeDeref(m.DNSRecordType))
	}
	if safeDeref(m.DNSNameserver) != "8.8.8.8" {
		t.Errorf("Create: expected dns_nameserver '8.8.8.8', got %q", safeDeref(m.DNSNameserver))
	}
	if safeDeref(m.DNSExpectedAnswer) != "www.example.com" {
		t.Errorf("Create: expected dns_expected_answer 'www.example.com', got %q", safeDeref(m.DNSExpectedAnswer))
	}

	// READ
	got, err := c.GetMonitor(ctx, m.UUID)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if safeDeref(got.DNSRecordType) != "CNAME" {
		t.Errorf("Read: expected dns_record_type 'CNAME', got %q", safeDeref(got.DNSRecordType))
	}
	if safeDeref(got.DNSNameserver) != "8.8.8.8" {
		t.Errorf("Read: expected dns_nameserver '8.8.8.8', got %q", safeDeref(got.DNSNameserver))
	}
	if safeDeref(got.DNSExpectedAnswer) != "www.example.com" {
		t.Errorf("Read: expected dns_expected_answer 'www.example.com', got %q", safeDeref(got.DNSExpectedAnswer))
	}
	t.Logf("READ verified: all DNS fields match")

	// UPDATE
	updated, err := c.UpdateMonitor(ctx, m.UUID, UpdateMonitorRequest{
		DNSRecordType:     strp("MX"),
		DNSNameserver:     strp("1.1.1.1"),
		DNSExpectedAnswer: strp("mail.example.com"),
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if safeDeref(updated.DNSRecordType) != "MX" {
		t.Errorf("Update: expected dns_record_type 'MX', got %q", safeDeref(updated.DNSRecordType))
	}
	if safeDeref(updated.DNSNameserver) != "1.1.1.1" {
		t.Errorf("Update: expected dns_nameserver '1.1.1.1', got %q", safeDeref(updated.DNSNameserver))
	}
	if safeDeref(updated.DNSExpectedAnswer) != "mail.example.com" {
		t.Errorf("Update: expected dns_expected_answer 'mail.example.com', got %q", safeDeref(updated.DNSExpectedAnswer))
	}
	t.Logf("UPDATE verified: all DNS fields updated correctly")

	// READ again to confirm persistence
	got2, err := c.GetMonitor(ctx, m.UUID)
	if err != nil {
		t.Fatalf("Read after update failed: %v", err)
	}
	if safeDeref(got2.DNSRecordType) != "MX" {
		t.Errorf("Read after update: expected 'MX', got %q", safeDeref(got2.DNSRecordType))
	}
	t.Logf("READ after UPDATE verified: changes persisted")
}

// TestAPI_DNS_AllRecordTypes creates a DNS monitor and cycles through all record types.
func TestAPI_DNS_AllRecordTypes(t *testing.T) {
	c := NewClient(mustAPIKey(t))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	m := mustCreateMonitor(t, c, ctx, CreateMonitorRequest{
		Name: uniqueName("dns-all-rt"), URL: "example.com", Protocol: "dns",
		DNSRecordType: strp("A"),
	})
	defer cleanupMonitor(t, c, m.UUID)

	recordTypes := []string{"A", "AAAA", "CNAME", "MX", "NS", "TXT", "SOA", "SRV", "CAA", "PTR"}
	for _, rt := range recordTypes {
		updated, err := c.UpdateMonitor(ctx, m.UUID, UpdateMonitorRequest{
			DNSRecordType: strp(rt),
		})
		if err != nil {
			t.Errorf("Failed to set dns_record_type=%q: %v", rt, err)
			continue
		}
		if safeDeref(updated.DNSRecordType) != rt {
			t.Errorf("Expected dns_record_type %q, got %q", rt, safeDeref(updated.DNSRecordType))
		} else {
			t.Logf("PASS: dns_record_type=%q accepted and persisted", rt)
		}
	}
}

// TestAPI_DNS_DefaultRecordType verifies API defaults dns_record_type to "A" when omitted.
func TestAPI_DNS_DefaultRecordType(t *testing.T) {
	c := NewClient(mustAPIKey(t))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	m := mustCreateMonitor(t, c, ctx, CreateMonitorRequest{
		Name: uniqueName("dns-default-rt"), URL: "example.com", Protocol: "dns",
		// No DNSRecordType — should default to "A"
	})
	defer cleanupMonitor(t, c, m.UUID)

	got, err := c.GetMonitor(ctx, m.UUID)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	rt := safeDeref(got.DNSRecordType)
	t.Logf("dns_record_type when omitted from create: %q", rt)
	if rt != "A" {
		t.Errorf("Expected default dns_record_type 'A', got %q", rt)
	}
}

// TestAPI_DNS_HTTPMonitorHasNullDNSFields verifies HTTP monitors return null DNS fields.
func TestAPI_DNS_HTTPMonitorHasNullDNSFields(t *testing.T) {
	c := NewClient(mustAPIKey(t))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	m := mustCreateMonitor(t, c, ctx, CreateMonitorRequest{
		Name: uniqueName("http-null-dns"), URL: "https://httpstat.us/200", Protocol: "http",
	})
	defer cleanupMonitor(t, c, m.UUID)

	got, err := c.GetMonitor(ctx, m.UUID)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if got.DNSRecordType != nil {
		t.Errorf("HTTP monitor should have nil dns_record_type, got %q", *got.DNSRecordType)
	}
	if got.DNSNameserver != nil {
		t.Errorf("HTTP monitor should have nil dns_nameserver, got %q", *got.DNSNameserver)
	}
	if got.DNSExpectedAnswer != nil {
		t.Errorf("HTTP monitor should have nil dns_expected_answer, got %q", *got.DNSExpectedAnswer)
	}
	t.Logf("PASS: HTTP monitor has null DNS fields as expected")
}

// TestAPI_DNS_ListIncludesDNSFields verifies DNS fields appear in list responses.
func TestAPI_DNS_ListIncludesDNSFields(t *testing.T) {
	c := NewClient(mustAPIKey(t))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	m := mustCreateMonitor(t, c, ctx, CreateMonitorRequest{
		Name: uniqueName("dns-list"), URL: "example.com", Protocol: "dns",
		DNSRecordType: strp("TXT"), DNSNameserver: strp("8.8.4.4"),
	})
	defer cleanupMonitor(t, c, m.UUID)

	monitors, err := c.ListMonitors(ctx)
	if err != nil {
		t.Fatalf("ListMonitors failed: %v", err)
	}

	var found *Monitor
	for i := range monitors {
		if monitors[i].UUID == m.UUID {
			found = &monitors[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("DNS monitor %s not found in list response", m.UUID)
	}

	if safeDeref(found.DNSRecordType) != "TXT" {
		t.Errorf("List: expected dns_record_type 'TXT', got %q", safeDeref(found.DNSRecordType))
	}
	if safeDeref(found.DNSNameserver) != "8.8.4.4" {
		t.Errorf("List: expected dns_nameserver '8.8.4.4', got %q", safeDeref(found.DNSNameserver))
	}
	t.Logf("PASS: DNS fields present in list response")
}

// TestAPI_PUT_DNSMonitor_WithoutDNSRecordType verifies that updating a DNS monitor
// without resending dns_record_type preserves the existing value.
func TestAPI_PUT_DNSMonitor_WithoutDNSRecordType(t *testing.T) {
	c := NewClient(mustAPIKey(t))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	m := mustCreateMonitor(t, c, ctx, CreateMonitorRequest{
		Name: uniqueName("dns-preserve-rt"), URL: "example.com", Protocol: "dns",
		DNSRecordType: strp("CNAME"),
	})
	defer cleanupMonitor(t, c, m.UUID)

	// Update name only — dns_record_type not in payload
	updated, err := c.UpdateMonitor(ctx, m.UUID, UpdateMonitorRequest{
		Name: strp(m.Name + "-nameonly"),
	})
	if err != nil {
		t.Fatalf("PUT on DNS monitor without dns_record_type failed: %v", err)
	}

	// Read back to check
	got, err := c.GetMonitor(ctx, updated.UUID)
	if err != nil {
		t.Fatalf("Read after update failed: %v", err)
	}

	rt := safeDeref(got.DNSRecordType)
	t.Logf("dns_record_type after updating name only: %q", rt)
	if rt != "CNAME" {
		t.Errorf("Expected dns_record_type preserved as 'CNAME', got %q", rt)
	}
	t.Logf("PASS: dns_record_type preserved when not included in PUT")
}
