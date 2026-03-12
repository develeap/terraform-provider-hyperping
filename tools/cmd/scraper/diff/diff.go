// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

// Package diff wraps oasdiff/oasdiff to detect semantic changes between two OpenAPI specs.
// It replaces the custom differ.go (363 lines) with a thin adapter (~80 lines).
package diff

import (
	"fmt"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
)

// Result contains the outcome of comparing two OpenAPI specs.
type Result struct {
	// HasChanges is true when any diff was detected (including metadata-only).
	HasChanges bool
	// HasPathChanges is true when endpoint-level changes exist (added/deleted/modified paths).
	HasPathChanges bool
	// Breaking is true when at least one breaking change is detected.
	Breaking bool
	// Summary is a human-readable markdown description of the changes.
	Summary string
}

// Compare loads two OpenAPI YAML files and returns a diff result.
// basePath is the older spec; currPath is the newer one.
func Compare(basePath, currPath string) (*Result, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	baseInfo, err := load.NewSpecInfo(loader, load.NewSource(basePath))
	if err != nil {
		return nil, fmt.Errorf("diff: load base spec %s: %w", basePath, err)
	}

	currInfo, err := load.NewSpecInfo(loader, load.NewSource(currPath))
	if err != nil {
		return nil, fmt.Errorf("diff: load curr spec %s: %w", currPath, err)
	}

	cfg := diff.NewConfig()
	d, sources, err := diff.GetWithOperationsSourcesMap(cfg, baseInfo, currInfo)
	if err != nil {
		return nil, fmt.Errorf("diff: compare specs: %w", err)
	}

	if d.Empty() {
		return &Result{}, nil
	}

	changes := checker.CheckBackwardCompatibility(checker.NewConfig(checker.GetAllChecks()), d, sources)

	hasPathChanges := d.PathsDiff != nil &&
		(len(d.PathsDiff.Added) > 0 || len(d.PathsDiff.Deleted) > 0 || len(d.PathsDiff.Modified) > 0)

	return &Result{
		HasChanges:     true,
		HasPathChanges: hasPathChanges,
		Breaking:       changes.HasLevelOrHigher(checker.ERR),
		Summary:        FormatMarkdown(d),
	}, nil
}

// FormatMarkdown converts a diff.Diff to a human-readable markdown string.
// Includes field-level change details for modified endpoints.
func FormatMarkdown(d *diff.Diff) string {
	if d == nil || d.Empty() {
		return "No API changes detected."
	}

	var sb strings.Builder
	sb.WriteString("## API Changes Detected\n\n")

	if d.PathsDiff == nil {
		sb.WriteString("_No path-level changes._\n")
		return sb.String()
	}

	pd := d.PathsDiff

	if len(pd.Added) > 0 {
		sb.WriteString("### New Endpoints\n")
		for _, p := range pd.Added {
			sb.WriteString("- `" + p + "`\n")
		}
		sb.WriteString("\n")
	}

	if len(pd.Deleted) > 0 {
		sb.WriteString("### Removed Endpoints (BREAKING)\n")
		for _, p := range pd.Deleted {
			sb.WriteString("- `" + p + "`\n")
		}
		sb.WriteString("\n")
	}

	if len(pd.Modified) > 0 {
		sb.WriteString("### Modified Endpoints\n\n")
		paths := sortedKeys(pd.Modified)
		for _, path := range paths {
			pathDiff := pd.Modified[path]
			formatPathDiff(&sb, path, pathDiff)
		}
	}

	return sb.String()
}

// formatPathDiff writes operation-level details for a single modified path.
func formatPathDiff(sb *strings.Builder, path string, pd *diff.PathDiff) {
	if pd.OperationsDiff == nil {
		sb.WriteString("#### `" + path + "`\n_No operation changes._\n\n")
		return
	}

	for _, method := range pd.OperationsDiff.Added {
		sb.WriteString("- `" + strings.ToUpper(method) + " " + path + "` (new method)\n")
	}

	for _, method := range pd.OperationsDiff.Deleted {
		sb.WriteString("- `" + strings.ToUpper(method) + " " + path + "` (removed method, BREAKING)\n")
	}

	methods := sortedKeys(pd.OperationsDiff.Modified)
	for _, method := range methods {
		methodDiff := pd.OperationsDiff.Modified[method]
		sb.WriteString("#### `" + strings.ToUpper(method) + " " + path + "`\n")
		formatMethodDiff(sb, methodDiff)
		sb.WriteString("\n")
	}
}

// formatMethodDiff writes field-level details for a single operation change.
func formatMethodDiff(sb *strings.Builder, md *diff.MethodDiff) {
	hasDetail := false

	// Parameter changes (path, query, header, cookie)
	if md.ParametersDiff != nil {
		if formatParametersDiff(sb, md.ParametersDiff) {
			hasDetail = true
		}
	}

	// Request body property changes
	if md.RequestBodyDiff != nil && md.RequestBodyDiff.ContentDiff != nil {
		if formatContentDiff(sb, md.RequestBodyDiff.ContentDiff) {
			hasDetail = true
		}
	}

	// Response body property changes (sorted by status code)
	if md.ResponsesDiff != nil {
		for _, statusCode := range sortedKeys(md.ResponsesDiff.Modified) {
			respDiff := md.ResponsesDiff.Modified[statusCode]
			if respDiff != nil && respDiff.ContentDiff != nil {
				if formatContentDiff(sb, respDiff.ContentDiff) {
					hasDetail = true
				}
			}
		}
	}

	if !hasDetail {
		sb.WriteString("- _operation metadata changed_\n")
	}
}

// formatContentDiff writes property-level details from a content diff.
// Returns true if any detail was written.
func formatContentDiff(sb *strings.Builder, cd *diff.ContentDiff) bool {
	wrote := false
	for _, mediaType := range sortedKeys(cd.MediaTypeModified) {
		mtDiff := cd.MediaTypeModified[mediaType]
		if mtDiff.SchemaDiff != nil && mtDiff.SchemaDiff.PropertiesDiff != nil {
			if formatPropertiesDiff(sb, mtDiff.SchemaDiff.PropertiesDiff) {
				wrote = true
			}
		}
	}
	return wrote
}

// formatPropertiesDiff writes added, deleted, and modified property details.
// Returns true if any detail was written.
func formatPropertiesDiff(sb *strings.Builder, pd *diff.SchemasDiff) bool {
	wrote := false

	for _, name := range pd.Added {
		sb.WriteString("- `+ " + name + "` (new property)\n")
		wrote = true
	}

	for _, name := range pd.Deleted {
		sb.WriteString("- `- " + name + "` (removed property)\n")
		wrote = true
	}

	names := sortedKeys(pd.Modified)
	for _, name := range names {
		schemaDiff := pd.Modified[name]
		detail := describePropertyChange(schemaDiff)
		sb.WriteString("- `~ " + name + "`: " + detail + "\n")
		wrote = true
	}

	return wrote
}

// formatParametersDiff writes added, deleted, and modified parameter details.
// Returns true if any detail was written.
func formatParametersDiff(sb *strings.Builder, pd *diff.ParametersDiffByLocation) bool {
	wrote := false

	for _, location := range sortedKeys(pd.Added) {
		for _, name := range pd.Added[location] {
			sb.WriteString("- `+ " + name + "` (" + location + " parameter, new)\n")
			wrote = true
		}
	}

	for _, location := range sortedKeys(pd.Deleted) {
		for _, name := range pd.Deleted[location] {
			sb.WriteString("- `- " + name + "` (" + location + " parameter, removed)\n")
			wrote = true
		}
	}

	for _, location := range sortedKeys(pd.Modified) {
		for _, name := range sortedKeys(pd.Modified[location]) {
			sb.WriteString("- `~ " + name + "` (" + location + " parameter, changed)\n")
			wrote = true
		}
	}

	return wrote
}

// describePropertyChange returns a concise description of what changed in a property.
func describePropertyChange(sd *diff.SchemaDiff) string {
	if sd == nil {
		return "schema changed"
	}

	var parts []string

	if sd.DescriptionDiff != nil {
		parts = append(parts, "description updated")
	}
	if sd.TypeDiff != nil {
		parts = append(parts, fmt.Sprintf("type changed (added: %v, removed: %v)", sd.TypeDiff.Added, sd.TypeDiff.Deleted))
	}
	if sd.DefaultDiff != nil {
		parts = append(parts, fmt.Sprintf("default %v → %v", sd.DefaultDiff.From, sd.DefaultDiff.To))
	}
	if sd.EnumDiff != nil {
		parts = append(parts, "enum values changed")
	}

	if len(parts) == 0 {
		return "schema changed"
	}
	return strings.Join(parts, ", ")
}

// sortedKeys returns the keys of a map sorted alphabetically.
// Works with any map[string]V type.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
