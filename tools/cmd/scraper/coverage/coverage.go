// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

// Package coverage compares an OpenAPI spec against a Terraform provider schema
// to identify fields that are documented in the API but missing from the provider,
// or provider fields that are no longer in the API.
//
// It requires a schema JSON file produced by:
//
//	terraform providers schema -json > schema.json
package coverage

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	tfjson "github.com/hashicorp/terraform-json"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/analyzer"
	"github.com/develeap/terraform-provider-hyperping/tools/scraper/openapi"
)

// Gap describes a single discrepancy between the API spec and the TF provider.
type Gap struct {
	Resource string // e.g., "monitors"
	APIField string // Field name as documented in the API
	TFField  string // Corresponding Terraform attribute name
	GapType  string // "missing" | "stale"
	Severity string // "error" | "warning" | "info"
	Details  string
}

// Report aggregates coverage statistics.
type Report struct {
	TotalAPIFields  int
	CoveredFields   int
	MissingFields   int
	StaleFields     int
	CoveragePercent float64
	Gaps            []Gap
}

// LoadProviderSchema reads a terraform-json ProviderSchemas from a file.
func LoadProviderSchema(schemaPath string) (*tfjson.ProviderSchemas, error) {
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("coverage: read schema file: %w", err)
	}

	var schemas tfjson.ProviderSchemas
	if err := json.Unmarshal(data, &schemas); err != nil {
		return nil, fmt.Errorf("coverage: parse schema JSON: %w", err)
	}

	return &schemas, nil
}

// Analyze compares the OpenAPI spec against the TF provider schema.
// It reuses analyzer.ResourceMappings for the API↔TF resource name translation,
// and analyzer.MapAPIFieldToTerraform() for field-level name normalization.
func Analyze(spec *openapi.Spec, schemas *tfjson.ProviderSchemas) *Report {
	report := &Report{}

	// Locate the hyperping provider schemas (first key that contains "hyperping").
	resourceSchemas := findHyperpingSchemas(schemas)
	if resourceSchemas == nil {
		return report
	}

	for _, mapping := range analyzer.ResourceMappings {
		apiFields := extractAPIFields(spec, mapping.APISection)
		tfAttrs := extractTFAttributes(resourceSchemas, mapping.TerraformResource)

		gaps, covered, missing, stale := compareFields(mapping, apiFields, tfAttrs)

		// Count only comparable (non-skipped) API fields in the denominator.
		// Skipped fields (dotted paths like "subscribe.email", "sections[].name")
		// are never checked against the TF schema, so including them inflates the
		// denominator and deflates the coverage percentage.
		comparable := 0
		for _, f := range apiFields {
			if !analyzer.IsSkippedField(f) {
				comparable++
			}
		}
		report.TotalAPIFields += comparable
		report.CoveredFields += covered
		report.MissingFields += missing
		report.StaleFields += stale
		report.Gaps = append(report.Gaps, gaps...)
	}

	if report.TotalAPIFields > 0 {
		report.CoveragePercent = float64(report.CoveredFields) / float64(report.TotalAPIFields) * 100
	}

	return report
}

// --- internal helpers ---

// findHyperpingSchemas locates the resource schema map for the hyperping provider.
func findHyperpingSchemas(schemas *tfjson.ProviderSchemas) map[string]*tfjson.Schema {
	for key, ps := range schemas.Schemas {
		if strings.Contains(key, "hyperping") {
			if ps.ResourceSchemas != nil {
				return ps.ResourceSchemas
			}
		}
	}
	return nil
}

// extractAPIFields returns all unique field names documented for a given section
// across all matching OAS operations (e.g., all POST /v1/monitors request body fields).
func extractAPIFields(spec *openapi.Spec, section string) []string {
	seen := make(map[string]bool)
	var fields []string

	for _, item := range spec.Paths {
		for _, op := range opsForItem(item) {
			if op == nil || !containsTag(op.Tags, section) {
				continue
			}
			if op.RequestBody == nil {
				continue
			}
			for _, media := range op.RequestBody.Content {
				for name := range media.Schema.Properties {
					if !seen[name] {
						seen[name] = true
						fields = append(fields, name)
					}
				}
			}
		}
	}

	return fields
}

func opsForItem(item openapi.PathItem) []*openapi.Operation {
	return []*openapi.Operation{item.Get, item.Post, item.Put, item.Patch, item.Delete}
}

func containsTag(tags []string, section string) bool {
	for _, t := range tags {
		if t == section {
			return true
		}
	}
	return false
}

// extractTFAttributes returns all attribute names for a given terraform resource.
func extractTFAttributes(resourceSchemas map[string]*tfjson.Schema, tfResource string) map[string]bool {
	attrs := make(map[string]bool)
	schema, ok := resourceSchemas[tfResource]
	if !ok || schema.Block == nil {
		return attrs
	}
	for name := range schema.Block.Attributes {
		attrs[name] = true
	}
	for name := range schema.Block.NestedBlocks {
		attrs[name] = true
	}
	return attrs
}

// compareFields checks each API field against TF attributes and vice versa.
func compareFields(mapping analyzer.ResourceMapping, apiFields []string, tfAttrs map[string]bool) (gaps []Gap, covered, missing, stale int) {
	apiSet := make(map[string]bool, len(apiFields))

	for _, apiName := range apiFields {
		if analyzer.IsSkippedField(apiName) {
			continue
		}

		tfName := analyzer.MapAPIFieldToTerraform(apiName, &mapping)
		apiSet[tfName] = true

		if tfAttrs[tfName] || tfAttrs[apiName] || analyzer.IsNestedField(apiName, &mapping) {
			covered++
		} else {
			missing++
			gaps = append(gaps, Gap{
				Resource: mapping.APISection,
				APIField: apiName,
				TFField:  tfName,
				GapType:  "missing",
				Severity: "warning",
				Details:  fmt.Sprintf("API field %q not implemented in %s", apiName, mapping.TerraformResource),
			})
		}
	}

	// Check for stale TF fields
	for tfName := range tfAttrs {
		if analyzer.IsSkippedField(tfName) || analyzer.IsExpectedStaleField(tfName, &mapping) {
			continue
		}
		if !apiSet[tfName] {
			stale++
			gaps = append(gaps, Gap{
				Resource: mapping.APISection,
				TFField:  tfName,
				GapType:  "stale",
				Severity: "info",
				Details:  fmt.Sprintf("TF attribute %q not documented in API for %s", tfName, mapping.APISection),
			})
		}
	}

	return gaps, covered, missing, stale
}
