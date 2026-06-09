// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
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

// sessionCounter mints unique session ids for each initialize call. Atomic
// so concurrent fixtures across parallel subtests do not collide.
var sessionCounter atomic.Uint64

// strictMCPHandler returns the http.Handler that backs the strict MCP
// test server. It is separated from newStrictMCPTestServer so the handler
// can be wrapped in middleware by callers that need to assert on raw
// request bytes without re-implementing the JSON-RPC dispatch.
func strictMCPHandler(t *testing.T, tools map[string]strictMCPTool) http.Handler {
	t.Helper()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read body", http.StatusBadRequest)
			return
		}
		_ = r.Body.Close()

		// Probe the method first; we need to special-case initialize
		// (mints a session) and the notifications/initialized post-
		// handshake message before we touch params.
		var envelope struct {
			JSONRPC string         `json:"jsonrpc"`
			ID      any            `json:"id"`
			Method  string         `json:"method"`
			Params  map[string]any `json:"params"`
		}
		if err := json.Unmarshal(body, &envelope); err != nil {
			writeRPCError(w, nil, -32700, "Parse error: "+err.Error())
			return
		}

		switch envelope.Method {
		case "initialize":
			sid := fmt.Sprintf("strict-%d", sessionCounter.Add(1))
			w.Header().Set("Mcp-Session-Id", sid)
			w.Header().Set("Content-Type", "application/json")
			writeRPC(w, envelope.ID, map[string]any{
				"protocolVersion": "2025-03-26",
				"capabilities":    map[string]any{},
				"serverInfo": map[string]any{
					"name":    "strict-mcp-fixture",
					"version": "1.0",
				},
			})
			return
		case "notifications/initialized":
			// Spec: no response body for a notification. Send 202 with
			// empty body so the transport's response decoder, which is
			// permissive about notifications, does not block.
			w.WriteHeader(http.StatusAccepted)
			return
		case "tools/call":
			handleToolsCall(t, w, envelope.ID, envelope.Params, tools, body)
			return
		default:
			writeRPCError(w, envelope.ID, -32601, "method not found: "+envelope.Method)
			return
		}
	})
}

// handleToolsCall does the request-shape validation that production
// /v1/mcp performs. Validation failures are returned as content[0].text
// starting with "MCP error -32602: ... Input validation ..." because the
// live server wraps its input-validator failures inside the tool result
// stream rather than the JSON-RPC error channel. Mirroring this exactly
// is the entire point of the fixture; the original bug surfaced as a
// json.Unmarshal failure on that literal text in hyperping-go's
// content[0].text decoder.
func handleToolsCall(t *testing.T, w http.ResponseWriter, id any, params map[string]any, tools map[string]strictMCPTool, rawBody []byte) {
	t.Helper()

	// Validation rule 1: params present and non-null. JSON-RPC spec
	// allows omitted params on some methods, but tools/call always
	// requires a params object.
	if params == nil {
		writeWrappedValidationError(w, id, "tools/call: params is required")
		return
	}

	nameAny, hasName := params["name"]
	if !hasName {
		writeWrappedValidationError(w, id, "tools/call: params.name is required")
		return
	}
	name, ok := nameAny.(string)
	if !ok {
		writeWrappedValidationError(w, id, "tools/call: params.name must be a string")
		return
	}

	tool, known := tools[name]
	if !known {
		writeRPCError(w, id, -32601, "unknown tool: "+name)
		return
	}

	// Validation rule 2: arguments key MUST be present AND non-null.
	// The arguments-key-omitted variant is the load-bearing case (the
	// hyperping-go v0.6.x bug); the explicit-null variant is the belt
	// for callers who "fix" it by sending arguments:null. To detect
	// either case reliably we re-parse the raw bytes; a map[string]any
	// drops the distinction between "key absent" and "key set to null"
	// after Unmarshal completes, both surfacing as a missing entry.
	var rawProbe struct {
		Params *json.RawMessage `json:"params"`
	}
	if err := json.Unmarshal(rawBody, &rawProbe); err != nil || rawProbe.Params == nil {
		writeWrappedValidationError(w, id, "tools/call: params.arguments is required")
		return
	}
	var paramsProbe struct {
		Arguments json.RawMessage `json:"arguments"`
	}
	// json.Decoder with DisallowUnknownFields would defeat the purpose
	// here; we only care that the bytes contain the arguments key.
	_ = json.Unmarshal(*rawProbe.Params, &paramsProbe)
	if len(paramsProbe.Arguments) == 0 || string(paramsProbe.Arguments) == "null" {
		writeWrappedValidationError(w, id,
			"tools/call: params.arguments is required, got missing or null")
		return
	}

	argsAny := params["arguments"]
	args, ok := argsAny.(map[string]any)
	if !ok {
		writeWrappedValidationError(w, id, "tools/call: params.arguments must be an object")
		return
	}

	// Validation rule 3: unknown keys are rejected when the tool
	// declares a properties whitelist. A tool with empty Properties
	// declares "no allowed keys", matching the empty-properties +
	// implicit-additionalProperties:false posture of the three list
	// tools on the live server.
	allowed := make(map[string]struct{}, len(tool.Properties))
	for _, p := range tool.Properties {
		allowed[p] = struct{}{}
	}
	for k := range args {
		if _, ok := allowed[k]; !ok {
			writeWrappedValidationError(w, id,
				fmt.Sprintf("tools/call: additionalProperties not allowed: %q", k))
			return
		}
	}

	// Validation rule 4: required fields must be present.
	for _, req := range tool.Required {
		if _, ok := args[req]; !ok {
			writeWrappedValidationError(w, id,
				fmt.Sprintf("tools/call: required field %q is missing", req))
			return
		}
	}

	// Dispatch to the per-tool handler.
	result, err := tool.Handler(args)
	if err != nil {
		// Internal handler errors map to -32603, wrapped in
		// content[0].text the same way the live server wraps
		// post-validator runtime failures.
		writeWrappedError(w, id, -32603, "MCP error -32603: "+err.Error())
		return
	}

	// The hyperping-go transport extracts content[0].text and json-
	// decodes it. So result must be JSON-encoded as a string and
	// returned in content[0].text.
	resultBytes, err := json.Marshal(result)
	if err != nil {
		writeWrappedError(w, id, -32603, "MCP error -32603: marshal: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeRPC(w, id, map[string]any{
		"content": []any{
			map[string]any{
				"type": "text",
				"text": string(resultBytes),
			},
		},
	})
}

// writeRPC writes a JSON-RPC success response.
func writeRPC(w http.ResponseWriter, id any, result any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	})
}

// writeRPCError writes a JSON-RPC error response (top-level error
// channel). Use this for envelope-level failures (method not found,
// parse error). Tool-level validation errors go through
// writeWrappedValidationError to match the live server's wrapping.
func writeRPCError(w http.ResponseWriter, id any, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	})
}

// writeWrappedValidationError emits the -32602 / Input validation shape
// the live server uses when params.arguments is malformed. The error
// rides inside result.content[0].text as a literal "MCP error -32602:
// Input validation: ..." string, NOT in the JSON-RPC error channel. This
// is what makes the original bug invisible to clients that only check
// rpcResp.Error and then try to json.Unmarshal the content[0].text body.
func writeWrappedValidationError(w http.ResponseWriter, id any, detail string) {
	writeWrappedError(w, id, -32602, "MCP error -32602: Input validation: "+detail)
}

// writeWrappedError is the generic content[0].text-wrapped error path.
func writeWrappedError(w http.ResponseWriter, id any, code int, text string) {
	_ = code // documented in text per the live server's shape
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]any{
			"content": []any{
				map[string]any{
					"type": "text",
					"text": text,
				},
			},
		},
	})
}
