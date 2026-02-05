// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

// Package contract provides schema extraction from VCR cassette files.
// This enables contract testing by discovering actual API response schemas.
package contract

// APIFieldSchema represents a field discovered from API request/response.
type APIFieldSchema struct {
	Name      string      `json:"name"`      // Field name from JSON
	Type      string      `json:"type"`      // Inferred type: string, number, integer, boolean, array, object, null
	Nullable  bool        `json:"nullable"`  // True if field was null in any response
	Required  bool        `json:"required"`  // True if field present in all responses
	Source    FieldSource `json:"source"`    // Where this field was seen
	Examples  []any       `json:"examples"`  // Sample values (up to 3)
	ChildType string      `json:"childType"` // For arrays: type of elements
}

// FieldSource indicates where a field was observed.
type FieldSource int

const (
	SourceRequest  FieldSource = 1 << iota // Field seen in request body
	SourceResponse                         // Field seen in response body
	SourceBoth     = SourceRequest | SourceResponse
)

func (s FieldSource) String() string {
	switch s {
	case SourceRequest:
		return "request"
	case SourceResponse:
		return "response"
	case SourceBoth:
		return "both"
	default:
		return "unknown"
	}
}

// EndpointSchema represents extracted schema for an API endpoint.
type EndpointSchema struct {
	Method         string                    `json:"method"`         // HTTP method
	Path           string                    `json:"path"`           // URL path pattern
	RequestFields  map[string]APIFieldSchema `json:"requestFields"`  // Fields in request body
	ResponseFields map[string]APIFieldSchema `json:"responseFields"` // Fields in response body
	StatusCodes    []int                     `json:"statusCodes"`    // Observed status codes
}

// CassetteSchema represents schemas extracted from all cassettes.
type CassetteSchema struct {
	Endpoints map[string]*EndpointSchema `json:"endpoints"` // Keyed by "METHOD /path"
	Source    string                     `json:"source"`    // Cassette directory path
}

// NewCassetteSchema creates a new empty cassette schema.
func NewCassetteSchema(source string) *CassetteSchema {
	return &CassetteSchema{
		Endpoints: make(map[string]*EndpointSchema),
		Source:    source,
	}
}

// GetOrCreateEndpoint gets or creates an endpoint schema.
func (cs *CassetteSchema) GetOrCreateEndpoint(method, path string) *EndpointSchema {
	key := method + " " + path
	if ep, exists := cs.Endpoints[key]; exists {
		return ep
	}
	ep := &EndpointSchema{
		Method:         method,
		Path:           path,
		RequestFields:  make(map[string]APIFieldSchema),
		ResponseFields: make(map[string]APIFieldSchema),
		StatusCodes:    []int{},
	}
	cs.Endpoints[key] = ep
	return ep
}
