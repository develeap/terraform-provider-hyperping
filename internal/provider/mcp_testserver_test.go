// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"net/http/httptest"
	"testing"
)

// strictMCPToolHandler is the per-tool callback the strict fixture invokes
// after it has validated the JSON-RPC envelope and the tools/call argument
// shape. The handler receives the already-validated arguments map and
// returns the typed result (will be JSON-encoded into content[0].text) or
// an error (will be surfaced as a JSON-RPC -32603 wrapped the same way the
// live MCP server wraps internal errors).
type strictMCPToolHandler func(args map[string]any) (any, error)

// strictMCPTool describes one tool the fixture is willing to serve. Tests
// supply a map of these keyed by tool name to newStrictMCPTestServer.
type strictMCPTool struct {
	// Properties is the JSON-schema "properties" key whitelist for
	// params.arguments. Any incoming argument key not listed here is
	// rejected with -32602 to match the live server's
	// additionalProperties:false behaviour.
	Properties []string
	// Required is the JSON-schema "required" list. Missing keys produce
	// -32602 with an "Input validation" message matching the live server.
	Required []string
	// Handler runs after validation passes.
	Handler strictMCPToolHandler
}

// newStrictMCPTestServer constructs an httptest.Server that mimics the
// production /v1/mcp endpoint's request-shape validation closely enough
// that the v0.6.x nil-arguments regression (fixed by hyperping-go v0.7.1)
// would be caught in CI. The intent is fidelity to the request envelope
// the live server checks; per-field type validation is intentionally out
// of scope for this fixture.
//
// The server speaks just enough JSON-RPC and MCP to satisfy
// hyperping-go's transport: initialize handshake (returns a session id),
// notifications/initialized (no-op), and tools/call (validated +
// dispatched). Anything else returns a JSON-RPC -32601.
//
// Caller closes the server with the standard httptest.Server.Close().
func newStrictMCPTestServer(t *testing.T, tools map[string]strictMCPTool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(strictMCPHandler(t, tools))
}
