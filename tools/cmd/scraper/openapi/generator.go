// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package openapi

import (
	"fmt"
	"os"
	"strings"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/extractor"
	"github.com/develeap/terraform-provider-hyperping/tools/scraper/utils"
	"gopkg.in/yaml.v3"
)

// --- OpenAPI 3.0 data structures ---

// Spec is the root OpenAPI 3.0 document.
type Spec struct {
	OpenAPI    string              `yaml:"openapi"`
	Info       Info                `yaml:"info"`
	Servers    []Server            `yaml:"servers"`
	Paths      map[string]PathItem `yaml:"paths"`
	Components Components          `yaml:"components,omitempty"`
}

// Info holds API metadata.
type Info struct {
	Title   string `yaml:"title"`
	Version string `yaml:"version"`
}

// Server holds a base URL.
type Server struct {
	URL string `yaml:"url"`
}

// PathItem holds operations per HTTP method for a single path.
type PathItem struct {
	Get    *Operation `yaml:"get,omitempty"`
	Post   *Operation `yaml:"post,omitempty"`
	Put    *Operation `yaml:"put,omitempty"`
	Patch  *Operation `yaml:"patch,omitempty"`
	Delete *Operation `yaml:"delete,omitempty"`
}

// Operation represents a single HTTP operation.
type Operation struct {
	OperationID string              `yaml:"operationId"`
	Summary     string              `yaml:"summary,omitempty"`
	Tags        []string            `yaml:"tags,omitempty"`
	Parameters  []Parameter         `yaml:"parameters,omitempty"`
	RequestBody *RequestBody        `yaml:"requestBody,omitempty"`
	Responses   map[string]Response `yaml:"responses"`
}

// Parameter represents a path/query/header parameter.
type Parameter struct {
	Name        string `yaml:"name"`
	In          string `yaml:"in"` // "path", "query", "header"
	Required    bool   `yaml:"required"`
	Description string `yaml:"description,omitempty"`
	Schema      Schema `yaml:"schema"`
}

// RequestBody holds the request payload schema.
type RequestBody struct {
	Required bool                 `yaml:"required"`
	Content  map[string]MediaType `yaml:"content"`
}

// MediaType wraps a schema for a content type.
type MediaType struct {
	Schema Schema `yaml:"schema"`
}

// Schema represents a JSON Schema subset used by OAS.
type Schema struct {
	Type        string            `yaml:"type,omitempty"`
	Format      string            `yaml:"format,omitempty"`
	Description string            `yaml:"description,omitempty"`
	Default     interface{}       `yaml:"default,omitempty"`
	Enum        []string          `yaml:"enum,omitempty"`
	Deprecated  bool              `yaml:"deprecated,omitempty"`
	Properties  map[string]Schema `yaml:"properties,omitempty"`
	Required    []string          `yaml:"required,omitempty"`
	Items       *Schema           `yaml:"items,omitempty"`
}

// Response is a simplified OAS response object.
type Response struct {
	Description string `yaml:"description"`
}

// Components holds reusable schemas (reserved for future use).
type Components struct{}

// --- Generator ---

// Generate converts scraped API parameters to an OpenAPI 3.0 Spec.
// The input map is keyed by the Hyperping doc path (e.g., "monitors/create").
// Unmapped paths are skipped; an info line is logged to stderr.
func Generate(sections map[string][]extractor.APIParameter, apiVersion string) *Spec {
	spec := &Spec{
		OpenAPI: "3.0.3",
		Info: Info{
			Title:   "Hyperping API",
			Version: apiVersion,
		},
		Servers: []Server{
			{URL: "https://api.hyperping.io"},
		},
		Paths: make(map[string]PathItem),
	}

	for docPath, params := range sections {
		mapping := LookupByDocPath(docPath)
		if mapping == nil {
			fmt.Fprintf(os.Stderr, "openapi: unknown doc path %q, skipping\n", docPath)
			continue
		}
		op := buildOperation(mapping, params)
		addOperation(spec.Paths, mapping.OASPath, mapping.HTTPMethod, op)
	}

	return spec
}

// Save serialises spec to YAML and writes it atomically.
func Save(spec *Spec, path string) error {
	data, err := yaml.Marshal(spec)
	if err != nil {
		return fmt.Errorf("openapi: marshal: %w", err)
	}
	if err := utils.AtomicWriteFile(path, data, utils.FilePermPrivate); err != nil {
		return fmt.Errorf("openapi: write %s: %w", path, err)
	}
	return nil
}

// Load reads an OpenAPI YAML file from disk.
func Load(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("openapi: read %s: %w", path, err)
	}
	var spec Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("openapi: parse %s: %w", path, err)
	}
	return &spec, nil
}

// AllDocPaths returns every doc path from the current spec's paths map.
// Useful for iterating over scraped sections without needing the original map.
func AllDocPaths() []string {
	paths := make([]string, 0, len(Mappings))
	for _, m := range Mappings {
		paths = append(paths, m.DocPath)
	}
	return paths
}

// --- internal helpers ---

func buildOperation(m *EndpointMapping, params []extractor.APIParameter) *Operation {
	op := &Operation{
		OperationID: operationID(m.HTTPMethod, m.OASPath),
		Tags:        []string{m.Section},
		Responses: map[string]Response{
			"200": {Description: "OK"},
			"400": {Description: "Bad Request"},
			"401": {Description: "Unauthorized"},
			"404": {Description: "Not Found"},
			"429": {Description: "Too Many Requests"},
		},
	}

	bodyParams, pathParams, queryParams := classifyParams(m.OASPath, params)

	for _, p := range pathParams {
		op.Parameters = append(op.Parameters, toOASParameter(p, "path"))
	}
	for _, p := range queryParams {
		op.Parameters = append(op.Parameters, toOASParameter(p, "query"))
	}

	if len(bodyParams) > 0 && hasBody(m.HTTPMethod) {
		op.RequestBody = buildRequestBody(bodyParams)
	}

	return op
}

func classifyParams(oasPath string, params []extractor.APIParameter) (body, path, query []extractor.APIParameter) {
	for _, p := range params {
		placeholder := "{" + p.Name + "}"
		if strings.Contains(oasPath, placeholder) {
			path = append(path, p)
		} else if isQueryMethod(oasPath) {
			query = append(query, p)
		} else {
			body = append(body, p)
		}
	}
	return body, path, query
}

func isQueryMethod(method string) bool {
	// GET operations use query parameters, not request bodies
	return strings.HasPrefix(method, "GET") || strings.HasSuffix(method, "/list")
}

func hasBody(httpMethod string) bool {
	switch strings.ToUpper(httpMethod) {
	case "POST", "PUT", "PATCH":
		return true
	}
	return false
}

func toOASParameter(p extractor.APIParameter, in string) Parameter {
	return Parameter{
		Name:        p.Name,
		In:          in,
		Required:    p.Required,
		Description: p.Description,
		Schema:      toSchema(p),
	}
}

func buildRequestBody(params []extractor.APIParameter) *RequestBody {
	props := make(map[string]Schema, len(params))
	var required []string

	for _, p := range params {
		props[p.Name] = toSchema(p)
		if p.Required {
			required = append(required, p.Name)
		}
	}

	return &RequestBody{
		Required: true,
		Content: map[string]MediaType{
			"application/json": {
				Schema: Schema{
					Type:       "object",
					Properties: props,
					Required:   required,
				},
			},
		},
	}
}

func toSchema(p extractor.APIParameter) Schema {
	s := Schema{
		Type:        normalizeType(p.Type),
		Description: p.Description,
		Default:     p.Default,
		Deprecated:  p.Deprecated,
	}

	if len(p.ValidValues) > 0 {
		s.Enum = p.ValidValues
	}

	if s.Type == "array" {
		s.Items = &Schema{Type: "string"}
	}

	return s
}

func normalizeType(t string) string {
	switch strings.ToLower(t) {
	case "string", "str":
		return "string"
	case "integer", "int", "number":
		return "integer"
	case "boolean", "bool":
		return "boolean"
	case "array", "[]string", "[]":
		return "array"
	case "object", "map":
		return "object"
	default:
		return "string"
	}
}

func operationID(httpMethod, oasPath string) string {
	// Produce a readable ID like "GET_v1_monitors" or "POST_v3_incidents"
	clean := strings.NewReplacer(
		"/", "_",
		"{", "",
		"}", "",
		"-", "_",
	).Replace(oasPath)
	clean = strings.Trim(clean, "_")
	return strings.ToUpper(httpMethod) + "_" + clean
}

func addOperation(paths map[string]PathItem, oasPath, httpMethod string, op *Operation) {
	item := paths[oasPath]
	switch strings.ToUpper(httpMethod) {
	case "GET":
		item.Get = op
	case "POST":
		item.Post = op
	case "PUT":
		item.Put = op
	case "PATCH":
		item.Patch = op
	case "DELETE":
		item.Delete = op
	}
	paths[oasPath] = item
}
