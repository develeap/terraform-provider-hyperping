// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/analyzer"
)

// SyncStatus represents the overall sync state
type SyncStatus struct {
	InSync   bool
	Issues   []SyncIssue
	Warnings []string
}

// SyncIssue represents a single sync problem that needs fixing
type SyncIssue struct {
	Category   string // "endpoint_version", "missing_field", "type_mismatch"
	Severity   string // "error", "warning"
	Resource   string
	Message    string
	Fix        string // Actionable fix instruction
	File       string
	Line       int
}

// RunSyncCheck performs a complete sync check and returns status
func RunSyncCheck(providerDir, snapshotDir string) (*SyncStatus, error) {
	status := &SyncStatus{InSync: true}

	// 1. Check endpoint versions
	log.Println("Checking endpoint versions...")
	endpointIssues, err := checkEndpointVersions(snapshotDir, providerDir)
	if err != nil {
		return nil, fmt.Errorf("endpoint check failed: %w", err)
	}
	status.Issues = append(status.Issues, endpointIssues...)

	// 2. Check field coverage
	log.Println("Checking field coverage...")
	coverageIssues, err := checkFieldCoverage(providerDir, snapshotDir)
	if err != nil {
		return nil, fmt.Errorf("coverage check failed: %w", err)
	}
	status.Issues = append(status.Issues, coverageIssues...)

	// Determine if in sync
	for _, issue := range status.Issues {
		if issue.Severity == "error" {
			status.InSync = false
			break
		}
	}

	return status, nil
}

// checkEndpointVersions compares API endpoint versions
func checkEndpointVersions(snapshotDir, providerDir string) ([]SyncIssue, error) {
	var issues []SyncIssue

	// Find latest snapshot subdirectory
	snapshotMgr := NewSnapshotManager(snapshotDir)
	latestSnapshot, err := snapshotMgr.GetLatestSnapshot()
	if err != nil {
		return nil, fmt.Errorf("no snapshots found: %w", err)
	}

	docsEndpoints, err := analyzer.ExtractEndpointsFromDocs(latestSnapshot)
	if err != nil {
		return nil, err
	}

	providerEndpoints, err := analyzer.ExtractEndpointsFromProvider(providerDir)
	if err != nil {
		return nil, err
	}

	report := analyzer.CompareEndpointVersions(docsEndpoints, providerEndpoints)

	for _, m := range report.Mismatches {
		issues = append(issues, SyncIssue{
			Category: "endpoint_version",
			Severity: "error",
			Resource: m.Resource,
			Message:  fmt.Sprintf("API version mismatch: docs=%s, provider=%s", m.DocsVersion, m.ProviderVersion),
			Fix:      m.Suggestion,
			File:     m.FilePath,
			Line:     m.LineNumber,
		})
	}

	return issues, nil
}

// checkFieldCoverage compares API fields with provider schema
func checkFieldCoverage(providerDir, snapshotDir string) ([]SyncIssue, error) {
	var issues []SyncIssue

	// Find latest snapshot
	snapshotMgr := NewSnapshotManager(snapshotDir)
	latestSnapshot, err := snapshotMgr.GetLatestSnapshot()
	if err != nil {
		return nil, fmt.Errorf("no snapshots found: %w", err)
	}

	// Initialize analyzer
	a, err := analyzer.NewAnalyzer(providerDir)
	if err != nil {
		return nil, err
	}

	// Load API params
	apiParams, err := analyzer.LoadAPIParamsFromSnapshot(latestSnapshot)
	if err != nil {
		return nil, err
	}

	// Run analysis
	report := a.AnalyzeCoverage(apiParams)

	// Convert gaps to issues
	for _, gap := range report.Gaps {
		severity := "warning"
		if gap.Type == analyzer.GapMissing {
			severity = "error" // Missing fields are errors
		}

		issues = append(issues, SyncIssue{
			Category: string(gap.Type),
			Severity: severity,
			Resource: gap.Resource,
			Message:  gap.Details,
			Fix:      gap.Suggestion,
			File:     gap.FilePath,
		})
	}

	return issues, nil
}

// FormatSyncStatus formats the sync status for output
func FormatSyncStatus(status *SyncStatus) string {
	var sb strings.Builder

	if status.InSync {
		sb.WriteString("✅ PROVIDER IN SYNC\n\n")
		sb.WriteString("All checks passed. Provider matches API documentation.\n")
		return sb.String()
	}

	sb.WriteString("❌ PROVIDER OUT OF SYNC\n\n")

	// Group by category
	byCategory := make(map[string][]SyncIssue)
	for _, issue := range status.Issues {
		byCategory[issue.Category] = append(byCategory[issue.Category], issue)
	}

	// Endpoint version issues
	if issues, ok := byCategory["endpoint_version"]; ok && len(issues) > 0 {
		sb.WriteString("## Endpoint Version Mismatches\n\n")
		for _, issue := range issues {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", issue.Resource, issue.Message))
			sb.WriteString(fmt.Sprintf("  - Fix: `%s`\n", issue.Fix))
		}
		sb.WriteString("\n")
	}

	// Missing field issues
	if issues, ok := byCategory["missing"]; ok && len(issues) > 0 {
		sb.WriteString("## Missing Fields\n\n")
		for _, issue := range issues {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", issue.Resource, issue.Message))
			sb.WriteString(fmt.Sprintf("  - Fix: %s\n", issue.Fix))
		}
		sb.WriteString("\n")
	}

	// Summary
	errorCount := 0
	for _, issue := range status.Issues {
		if issue.Severity == "error" {
			errorCount++
		}
	}
	sb.WriteString(fmt.Sprintf("---\n**Total issues: %d (%d errors)**\n", len(status.Issues), errorCount))

	return sb.String()
}

// PrintSyncStatus prints a concise sync status to stdout
func PrintSyncStatus(status *SyncStatus) {
	if status.InSync {
		fmt.Println("✅ SYNC OK - Provider matches API documentation")
		return
	}

	fmt.Println("❌ SYNC FAILED - Provider is out of sync with API")
	fmt.Println()

	errorCount := 0
	for _, issue := range status.Issues {
		icon := "⚠️"
		if issue.Severity == "error" {
			icon = "❌"
			errorCount++
		}
		fmt.Printf("%s [%s] %s: %s\n", icon, issue.Category, issue.Resource, issue.Message)
		fmt.Printf("   Fix: %s\n", issue.Fix)
	}

	fmt.Printf("\nTotal: %d issues (%d errors)\n", len(status.Issues), errorCount)
}
