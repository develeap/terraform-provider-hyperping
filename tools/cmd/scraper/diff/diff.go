// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

// Package diff wraps tufin/oasdiff to detect semantic changes between two OpenAPI specs.
// It replaces the custom differ.go (363 lines) with a thin adapter (~80 lines).
package diff

import (
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/tufin/oasdiff/checker"
	"github.com/tufin/oasdiff/diff"
	"github.com/tufin/oasdiff/load"
)

// Result contains the outcome of comparing two OpenAPI specs.
type Result struct {
	// HasChanges is true when any diff was detected.
	HasChanges bool
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

	changes := checker.CheckBackwardCompatibility(checker.GetDefaultChecks(), d, sources)

	return &Result{
		HasChanges: true,
		Breaking:   changes.HasLevelOrHigher(checker.ERR),
		Summary:    FormatMarkdown(d),
	}, nil
}

// FormatMarkdown converts a diff.Diff to a human-readable markdown string.
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
		sb.WriteString("### Modified Endpoints\n")
		for path := range pd.Modified {
			sb.WriteString("- `" + path + "`\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
