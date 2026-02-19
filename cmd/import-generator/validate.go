// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"fmt"
	"io"
	"regexp"
)

// ValidationResult holds the results of resource validation.
type ValidationResult struct {
	Monitors     ValidationResourceResult
	Healthchecks ValidationResourceResult
	StatusPages  ValidationResourceResult
	Incidents    ValidationResourceResult
	Maintenance  ValidationResourceResult
	Outages      ValidationResourceResult
}

// ValidationResourceResult holds validation results for a single resource type.
type ValidationResourceResult struct {
	ResourceType string
	ValidCount   int
	InvalidIDs   []string
	FetchError   error
}

// Validate performs validation of resources without generating output.
func (g *Generator) Validate(ctx context.Context) *ValidationResult {
	result := &ValidationResult{}

	for _, r := range g.resources {
		switch r {
		case "monitors":
			result.Monitors = g.validateMonitors(ctx)
		case "healthchecks":
			result.Healthchecks = g.validateHealthchecks(ctx)
		case "statuspages":
			result.StatusPages = g.validateStatusPages(ctx)
		case "incidents":
			result.Incidents = g.validateIncidents(ctx)
		case "maintenance":
			result.Maintenance = g.validateMaintenance(ctx)
		case "outages":
			result.Outages = g.validateOutages(ctx)
		}
	}

	return result
}

func (g *Generator) validateMonitors(ctx context.Context) ValidationResourceResult {
	result := ValidationResourceResult{ResourceType: "Monitors"}

	monitors, err := g.client.ListMonitors(ctx)
	if err != nil {
		result.FetchError = err
		return result
	}

	validPattern := regexp.MustCompile(`^mon_[a-zA-Z0-9]+$`)
	for _, m := range monitors {
		if validPattern.MatchString(m.UUID) {
			result.ValidCount++
		} else {
			result.InvalidIDs = append(result.InvalidIDs, m.UUID)
		}
	}

	return result
}

func (g *Generator) validateHealthchecks(ctx context.Context) ValidationResourceResult {
	result := ValidationResourceResult{ResourceType: "Healthchecks"}

	healthchecks, err := g.client.ListHealthchecks(ctx)
	if err != nil {
		result.FetchError = err
		return result
	}

	validPattern := regexp.MustCompile(`^hc_[a-zA-Z0-9]+$`)
	for _, h := range healthchecks {
		if validPattern.MatchString(h.UUID) {
			result.ValidCount++
		} else {
			result.InvalidIDs = append(result.InvalidIDs, h.UUID)
		}
	}

	return result
}

func (g *Generator) validateStatusPages(ctx context.Context) ValidationResourceResult {
	result := ValidationResourceResult{ResourceType: "Status Pages"}

	resp, err := g.client.ListStatusPages(ctx, nil, nil)
	if err != nil {
		result.FetchError = err
		return result
	}

	validPattern := regexp.MustCompile(`^sp_[a-zA-Z0-9]+$`)
	for _, sp := range resp.StatusPages {
		if validPattern.MatchString(sp.UUID) {
			result.ValidCount++
		} else {
			result.InvalidIDs = append(result.InvalidIDs, sp.UUID)
		}
	}

	return result
}

func (g *Generator) validateIncidents(ctx context.Context) ValidationResourceResult {
	result := ValidationResourceResult{ResourceType: "Incidents"}

	incidents, err := g.client.ListIncidents(ctx)
	if err != nil {
		result.FetchError = err
		return result
	}

	validPattern := regexp.MustCompile(`^inc_[a-zA-Z0-9]+$`)
	for _, i := range incidents {
		if validPattern.MatchString(i.UUID) {
			result.ValidCount++
		} else {
			result.InvalidIDs = append(result.InvalidIDs, i.UUID)
		}
	}

	return result
}

func (g *Generator) validateMaintenance(ctx context.Context) ValidationResourceResult {
	result := ValidationResourceResult{ResourceType: "Maintenance"}

	maintenance, err := g.client.ListMaintenance(ctx)
	if err != nil {
		result.FetchError = err
		return result
	}

	validPattern := regexp.MustCompile(`^maint_[a-zA-Z0-9]+$`)
	for _, m := range maintenance {
		if validPattern.MatchString(m.UUID) {
			result.ValidCount++
		} else {
			result.InvalidIDs = append(result.InvalidIDs, m.UUID)
		}
	}

	return result
}

func (g *Generator) validateOutages(ctx context.Context) ValidationResourceResult {
	result := ValidationResourceResult{ResourceType: "Outages"}

	outages, err := g.client.ListOutages(ctx)
	if err != nil {
		result.FetchError = err
		return result
	}

	validPattern := regexp.MustCompile(`^outage_[a-zA-Z0-9]+$`)
	for _, o := range outages {
		if validPattern.MatchString(o.UUID) {
			result.ValidCount++
		} else {
			result.InvalidIDs = append(result.InvalidIDs, o.UUID)
		}
	}

	return result
}

// Print outputs the validation results in a human-readable format.
func (vr *ValidationResult) Print(w io.Writer) {
	vr.printResourceResult(w, vr.Monitors)
	vr.printResourceResult(w, vr.Healthchecks)
	vr.printResourceResult(w, vr.StatusPages)
	vr.printResourceResult(w, vr.Incidents)
	vr.printResourceResult(w, vr.Maintenance)
	vr.printResourceResult(w, vr.Outages)

	if vr.IsValid() {
		fmt.Fprintln(w, "\nValidation passed: All resources are valid")
	} else {
		errorCount := vr.ErrorCount()
		fmt.Fprintf(w, "\nValidation failed: %d error(s) found\n", errorCount) //nolint:gosec // G705: errorCount is an integer, not user-controlled HTML content
	}
}

func (vr *ValidationResult) printResourceResult(w io.Writer, result ValidationResourceResult) {
	// Skip if resource type wasn't checked
	if result.ResourceType == "" {
		return
	}

	if result.FetchError != nil {
		fmt.Fprintf(w, "✗ %s: Failed to fetch (%v)\n", result.ResourceType, result.FetchError)
		return
	}

	if len(result.InvalidIDs) > 0 {
		fmt.Fprintf(w, "✗ %s: %d valid, %d invalid ID(s)\n",
			result.ResourceType, result.ValidCount, len(result.InvalidIDs))
		for _, id := range result.InvalidIDs {
			fmt.Fprintf(w, "  - %s\n", id)
		}
	} else {
		fmt.Fprintf(w, "✓ %s: %d valid ID(s)\n", result.ResourceType, result.ValidCount)
	}
}

// IsValid returns true if all validations passed.
func (vr *ValidationResult) IsValid() bool {
	return vr.ErrorCount() == 0
}

// ErrorCount returns the total number of validation errors.
func (vr *ValidationResult) ErrorCount() int {
	count := 0

	if vr.Monitors.FetchError != nil || len(vr.Monitors.InvalidIDs) > 0 {
		count++
	}
	if vr.Healthchecks.FetchError != nil || len(vr.Healthchecks.InvalidIDs) > 0 {
		count++
	}
	if vr.StatusPages.FetchError != nil || len(vr.StatusPages.InvalidIDs) > 0 {
		count++
	}
	if vr.Incidents.FetchError != nil || len(vr.Incidents.InvalidIDs) > 0 {
		count++
	}
	if vr.Maintenance.FetchError != nil || len(vr.Maintenance.InvalidIDs) > 0 {
		count++
	}
	if vr.Outages.FetchError != nil || len(vr.Outages.InvalidIDs) > 0 {
		count++
	}

	return count
}
