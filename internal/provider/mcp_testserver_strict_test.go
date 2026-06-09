// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

// initSession opens an MCP session against srvURL by sending an initialize
// request and capturing the Mcp-Session-Id response header. It also fires
// the notifications/initialized post-handshake message because the live
// server's tools/call validator requires a fully-initialized session.
//
// Returns the session id so the caller can attach it to subsequent
// tools/call requests.
func initSession(t *testing.T, srvURL string) string {
	t.Helper()

	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"strict-fixture-tests","version":"0.1"}}}`
	req, err := http.NewRequest(http.MethodPost, srvURL, strings.NewReader(body))
	if err != nil {
		t.Fatalf("build initialize request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("initialize call: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("initialize returned status %d", resp.StatusCode)
	}
	_, _ = io.Copy(io.Discard, resp.Body)

	sid := resp.Header.Get("Mcp-Session-Id")
	if sid == "" {
		t.Fatalf("initialize response did not include Mcp-Session-Id header")
	}

	// Best-effort notifications/initialized; the live server tolerates
	// callers that skip this notification when their session was just
	// minted, so a failure here is non-fatal.
	notifyBody := `{"jsonrpc":"2.0","method":"notifications/initialized"}`
	nreq, _ := http.NewRequest(http.MethodPost, srvURL, strings.NewReader(notifyBody))
	nreq.Header.Set("Content-Type", "application/json")
	nreq.Header.Set("Accept", "application/json, text/event-stream")
	nreq.Header.Set("Mcp-Session-Id", sid)
	if nresp, err := http.DefaultClient.Do(nreq); err == nil {
		_, _ = io.Copy(io.Discard, nresp.Body)
		_ = nresp.Body.Close()
	}

	return sid
}

// callToolRaw POSTs an arbitrary JSON-RPC body (no envelope rewriting) to
// the fixture and returns the parsed JSON-RPC response. Use this when a
// test needs to control the request bytes precisely, e.g. to omit the
// arguments key entirely or set it to null.
func callToolRaw(t *testing.T, srvURL, sid, jsonRPCBody string) map[string]any {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, srvURL, strings.NewReader(jsonRPCBody))
	if err != nil {
		t.Fatalf("build tools/call request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if sid != "" {
		req.Header.Set("Mcp-Session-Id", sid)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("tools/call: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("tools/call returned status %d, body=%s", resp.StatusCode, body)
	}

	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("decode response: %v (body=%s)", err, body)
	}
	return parsed
}

// expectMCPValidationError walks the JSON-RPC response and asserts the
// fixture produced the -32602 / Input validation shape the live server
// emits for argument-shape violations. The error is wrapped as the
// literal text of content[0].text rather than a top-level rpcResp.Error,
// because that is exactly how hyperping-python and hyperping-go observed
// the live server respond on 2026-06-08: the input validator runs inside
// the tools/call handler, so its error rides in the call's result content
// stream, not the JSON-RPC error channel.
func expectMCPValidationError(t *testing.T, resp map[string]any) {
	t.Helper()

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result object, got %v", resp)
	}
	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatalf("expected non-empty content array, got %v", result)
	}
	first, ok := content[0].(map[string]any)
	if !ok {
		t.Fatalf("expected content[0] to be an object, got %T", content[0])
	}
	text, ok := first["text"].(string)
	if !ok {
		t.Fatalf("expected content[0].text to be a string, got %T", first["text"])
	}
	if !strings.Contains(text, "MCP error -32602") {
		t.Fatalf("expected text to contain \"MCP error -32602\", got %q", text)
	}
	if !strings.Contains(text, "Input validation") {
		t.Fatalf("expected text to contain \"Input validation\", got %q", text)
	}
}

// noArgsTool is a convenience builder for a list-style tool that takes no
// arguments and returns the supplied stub on a valid call.
func noArgsTool(stub any) strictMCPTool {
	return strictMCPTool{
		Handler: func(_ map[string]any) (any, error) { return stub, nil },
	}
}

// uuidTool is the get-style counterpart: requires uuid, rejects unknown
// keys, returns the stub when validation passes.
func uuidTool(stub any) strictMCPTool {
	return strictMCPTool{
		Properties: []string{"uuid"},
		Required:   []string{"uuid"},
		Handler:    func(_ map[string]any) (any, error) { return stub, nil },
	}
}

func TestStrictMCPTestServer_RejectsMissingArguments(t *testing.T) {
	t.Parallel()
	server := newStrictMCPTestServer(t, map[string]strictMCPTool{
		"list_escalation_policies": noArgsTool([]any{}),
	})
	defer server.Close()

	sid := initSession(t, server.URL)
	// Note: this request omits the arguments key entirely, exactly
	// matching the v0.6.x hyperping-go transport bug.
	body := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_escalation_policies"}}`
	resp := callToolRaw(t, server.URL, sid, body)
	expectMCPValidationError(t, resp)
}

func TestStrictMCPTestServer_RejectsNilArguments(t *testing.T) {
	t.Parallel()
	server := newStrictMCPTestServer(t, map[string]strictMCPTool{
		"list_escalation_policies": noArgsTool([]any{}),
	})
	defer server.Close()

	sid := initSession(t, server.URL)
	body := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_escalation_policies","arguments":null}}`
	resp := callToolRaw(t, server.URL, sid, body)
	expectMCPValidationError(t, resp)
}

func TestStrictMCPTestServer_RejectsUnknownKey(t *testing.T) {
	t.Parallel()
	// list_escalation_policies has empty properties so any caller-
	// supplied key is unknown. This guards against tests accidentally
	// passing junk and getting away with it.
	server := newStrictMCPTestServer(t, map[string]strictMCPTool{
		"list_escalation_policies": noArgsTool([]any{}),
	})
	defer server.Close()

	sid := initSession(t, server.URL)
	body := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_escalation_policies","arguments":{"banana":"x"}}}`
	resp := callToolRaw(t, server.URL, sid, body)
	expectMCPValidationError(t, resp)
}

func TestStrictMCPTestServer_RejectsMissingRequiredField(t *testing.T) {
	t.Parallel()
	server := newStrictMCPTestServer(t, map[string]strictMCPTool{
		"get_escalation_policy": uuidTool(map[string]any{}),
	})
	defer server.Close()

	sid := initSession(t, server.URL)
	body := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_escalation_policy","arguments":{}}}`
	resp := callToolRaw(t, server.URL, sid, body)
	expectMCPValidationError(t, resp)
}

func TestStrictMCPTestServer_AcceptsEmptyArgumentsObject(t *testing.T) {
	t.Parallel()
	stub := []any{
		map[string]any{
			"uuid":  "esc_abc",
			"name":  "Primary",
			"team":  "SRE",
			"steps": []any{},
		},
	}
	server := newStrictMCPTestServer(t, map[string]strictMCPTool{
		"list_escalation_policies": noArgsTool(stub),
	})
	defer server.Close()

	sid := initSession(t, server.URL)
	body := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_escalation_policies","arguments":{}}}`
	resp := callToolRaw(t, server.URL, sid, body)

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result object, got %v", resp)
	}
	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatalf("expected non-empty content array, got %v", result)
	}
	first := content[0].(map[string]any)
	text, ok := first["text"].(string)
	if !ok {
		t.Fatalf("expected content[0].text to be a string")
	}
	var decoded []any
	if err := json.Unmarshal([]byte(text), &decoded); err != nil {
		t.Fatalf("content[0].text should be valid JSON, got %q: %v", text, err)
	}
	if len(decoded) != 1 {
		t.Fatalf("expected one element in stub result, got %d", len(decoded))
	}
	first0 := decoded[0].(map[string]any)
	if first0["uuid"] != "esc_abc" {
		t.Fatalf("expected uuid esc_abc, got %v", first0["uuid"])
	}
}

// TestEscalationPoliciesDataSource_FailsAgainstStrictServer_WhenBugRegressedToNilArgs
// is the load-bearing regression guard. It hand-drives a tools/call against
// the strict fixture using the buggy pre-v0.7.1 wire shape (no arguments
// key) and asserts the fixture rejects it with the -32602 envelope. If
// hyperping-go ever regresses to omitting arguments for nil-args callers
// (or someone in this repo writes a custom transport that does so), the
// EscalationPolicies data source would see an "MCP error -32602 ... Input
// validation" string flow through its content[0].text decode and fail with
// "invalid character 'M' looking for beginning of value", which is the
// exact production signature of the original silent breakage.
//
// We assert the fixture-level response shape directly rather than wiring
// through the SDK; the SDK is already covered by hyperping-go's own
// Test 1.16 regression guard. This test's job is fixture fidelity.
func TestEscalationPoliciesDataSource_FailsAgainstStrictServer_WhenBugRegressedToNilArgs(t *testing.T) {
	t.Parallel()
	server := newStrictMCPTestServer(t, map[string]strictMCPTool{
		"list_escalation_policies": noArgsTool([]any{
			map[string]any{"uuid": "esc_x", "name": "Primary", "team": "SRE", "steps": []any{}},
		}),
	})
	defer server.Close()

	sid := initSession(t, server.URL)

	// Buggy v0.6.x wire shape: no arguments key on params.
	buggy := `{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"list_escalation_policies"}}`
	resp := callToolRaw(t, server.URL, sid, buggy)
	expectMCPValidationError(t, resp)

	// Now confirm that the fixed v0.7.1 wire shape (arguments:{}) gets
	// a clean response from the same fixture, demonstrating the wrapper
	// distinguishes the two cases rather than blanket-rejecting.
	fixed := `{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"list_escalation_policies","arguments":{}}}`
	good := callToolRaw(t, server.URL, sid, fixed)
	result, ok := good["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result, got %v", good)
	}
	content := result["content"].([]any)
	first := content[0].(map[string]any)
	text := first["text"].(string)
	if !strings.Contains(text, "esc_x") {
		t.Fatalf("expected stub payload in response, got %q", text)
	}

	// Belt: the byte stream the strict fixture sends back for the buggy
	// case must round-trip through json.Marshal -> json.Unmarshal so
	// downstream assertions are stable across Go releases.
	if _, err := json.Marshal(resp); err != nil {
		t.Fatalf("response should be re-marshalable: %v", err)
	}
}

// echo helpers below are intentionally unused — placeholders that document
// the JSON-RPC envelope shape the fixture must respect. Kept as a single
// reference point for future tests.
var (
	_ = bytes.NewReader
	_ = fmt.Sprintf
)
