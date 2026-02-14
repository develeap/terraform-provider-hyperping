// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package dryrun

import (
	"fmt"
	"strings"
)

// CompatibilityAnalyzer calculates migration compatibility scores.
type CompatibilityAnalyzer struct{}

// NewCompatibilityAnalyzer creates a new compatibility analyzer.
func NewCompatibilityAnalyzer() *CompatibilityAnalyzer {
	return &CompatibilityAnalyzer{}
}

// AnalyzeCompatibility calculates overall compatibility score.
func (a *CompatibilityAnalyzer) AnalyzeCompatibility(
	comparisons []ResourceComparison,
	warnings []Warning,
) CompatibilityScore {
	if len(comparisons) == 0 {
		return CompatibilityScore{
			OverallScore: 100.0,
			Complexity:   "Simple",
			Details:      "No resources to migrate",
		}
	}

	clean := 0
	withWarnings := 0
	withErrors := 0
	byType := make(map[string]*typeScore)

	for _, comp := range comparisons {
		typeStats, exists := byType[comp.ResourceType]
		if !exists {
			typeStats = &typeScore{}
			byType[comp.ResourceType] = typeStats
		}
		typeStats.total++

		switch {
		case comp.HasErrors:
			withErrors++
			typeStats.errors++
		case comp.HasWarnings:
			withWarnings++
			typeStats.warnings++
		default:
			clean++
			typeStats.clean++
		}
	}

	total := len(comparisons)
	score := (float64(clean) / float64(total)) * 100.0

	// Calculate complexity based on warnings and errors
	complexity := a.calculateComplexity(total, withWarnings, withErrors)

	// Build type breakdown
	typeScores := make(map[string]float64)
	for resType, stats := range byType {
		if stats.total > 0 {
			typeScores[resType] = (float64(stats.clean) / float64(stats.total)) * 100.0
		}
	}

	// Build details
	details := a.buildDetails(clean, withWarnings, withErrors, total)

	return CompatibilityScore{
		OverallScore:    score,
		ByType:          typeScores,
		CleanMigrations: clean,
		WarningCount:    withWarnings,
		ErrorCount:      withErrors,
		Complexity:      complexity,
		Details:         details,
	}
}

type typeScore struct {
	total    int
	clean    int
	warnings int
	errors   int
}

func (a *CompatibilityAnalyzer) calculateComplexity(total, warnings, errors int) string {
	if total == 0 {
		return "Simple"
	}

	errorRate := float64(errors) / float64(total)
	warningRate := float64(warnings) / float64(total)

	if errorRate > 0.1 || warningRate > 0.5 {
		return "Complex"
	}
	if errorRate > 0 || warningRate > 0.25 {
		return "Medium"
	}
	return "Simple"
}

func (a *CompatibilityAnalyzer) buildDetails(clean, warnings, errors, total int) string {
	var parts []string

	if clean == total {
		parts = append(parts, "All resources convert cleanly")
	} else {
		parts = append(parts, fmt.Sprintf("%d of %d resources convert cleanly", clean, total))
	}

	if warnings > 0 {
		parts = append(parts, fmt.Sprintf("%d resources require review", warnings))
	}

	if errors > 0 {
		parts = append(parts, fmt.Sprintf("%d resources have critical issues", errors))
	}

	return strings.Join(parts, ". ")
}

// CategorizeWarnings groups warnings by severity and category.
func (a *CompatibilityAnalyzer) CategorizeWarnings(warnings []Warning) map[string][]Warning {
	categorized := make(map[string][]Warning)

	for _, w := range warnings {
		key := w.Severity
		categorized[key] = append(categorized[key], w)
	}

	return categorized
}

// EstimateManualEffort calculates estimated time for manual steps.
func (a *CompatibilityAnalyzer) EstimateManualEffort(warnings []Warning) (minutes int, description string) {
	baseMinutes := 0

	critical := 0
	warningCount := 0
	info := 0

	for _, w := range warnings {
		switch w.Severity {
		case "critical":
			critical++
			baseMinutes += 10
		case "warning":
			warningCount++
			baseMinutes += 5
		case "info":
			info++
			baseMinutes += 2
		}
	}

	if baseMinutes == 0 {
		return 0, "No manual steps required"
	}

	high := baseMinutes
	low := baseMinutes / 2

	return low, fmt.Sprintf("%d-%d minutes (%d critical, %d warnings, %d info items)",
		low, high, critical, warningCount, info)
}
