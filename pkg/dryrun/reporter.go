// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package dryrun

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// Reporter orchestrates dry-run report generation.
type Reporter struct {
	diffFormatter    *DiffFormatter
	previewGenerator *PreviewGenerator
	compatAnalyzer   *CompatibilityAnalyzer
	useColors        bool
}

// NewReporter creates a new dry-run reporter.
func NewReporter(useColors bool) *Reporter {
	return &Reporter{
		diffFormatter:    NewDiffFormatter(useColors),
		previewGenerator: NewPreviewGenerator(useColors),
		compatAnalyzer:   NewCompatibilityAnalyzer(),
		useColors:        useColors,
	}
}

// GenerateReport creates a comprehensive dry-run report.
func (r *Reporter) GenerateReport(
	sourcePlatform string,
	comparisons []ResourceComparison,
	warnings []Warning,
	tfContent string,
	estimates PerformanceEstimates,
	summary Summary,
) Report {
	compatibility := r.compatAnalyzer.AnalyzeCompatibility(comparisons, warnings)

	preview := r.previewGenerator.GeneratePreview(tfContent, summary.ExpectedTFResources, 3)

	return Report{
		Summary:          summary,
		Comparison:       comparisons,
		TerraformPreview: preview,
		Compatibility:    compatibility,
		Warnings:         warnings,
		Estimates:        estimates,
		SourcePlatform:   sourcePlatform,
		TargetPlatform:   "Hyperping",
		Timestamp:        time.Now(),
	}
}

// PrintReport outputs the dry-run report to a writer.
func (r *Reporter) PrintReport(w io.Writer, report Report, opts Options) error {
	r.printHeader(w, report)
	r.printSummary(w, report)
	r.printCompatibility(w, report)

	if opts.Verbose || opts.ShowAllResources {
		r.printComparisons(w, report, opts)
	} else {
		r.printSampleComparisons(w, report)
	}

	r.printTerraformPreview(w, report)
	r.printWarnings(w, report)
	r.printEstimates(w, report)
	r.printNextSteps(w, report)

	return nil
}

func (r *Reporter) printHeader(w io.Writer, _ Report) {
	r.printSectionDivider(w)
	fmt.Fprintf(w, "%s\n", r.colorize("🔍 DRY RUN MODE - Migration Preview", "bold"))
	r.printSectionDivider(w)
	fmt.Fprintln(w)
}

func (r *Reporter) printSummary(w io.Writer, report Report) {
	r.printSectionHeader(w, "📊 SUMMARY")

	fmt.Fprintf(w, "Source Platform:  %s\n", report.SourcePlatform)
	fmt.Fprintf(w, "Target Platform:  %s\n", report.TargetPlatform)
	fmt.Fprintf(w, "Timestamp:        %s\n\n", report.Timestamp.Format("2006-01-02 15:04:05"))

	if report.Summary.TotalMonitors > 0 {
		fmt.Fprintf(w, "Total Monitors:   %d\n", report.Summary.TotalMonitors)
	}
	if report.Summary.TotalHealthchecks > 0 {
		fmt.Fprintf(w, "Total Heartbeats: %d\n", report.Summary.TotalHealthchecks)
	}

	fmt.Fprintln(w, "\nExpected Output:")
	for resType, count := range report.Summary.ResourceBreakdown {
		fmt.Fprintf(w, "  - %-25s: %d resources\n", resType, count)
	}

	if report.Summary.ExpectedTFLines > 0 {
		fmt.Fprintf(w, "  - Total TF size:         ~%d lines (~%s)\n",
			report.Summary.ExpectedTFLines,
			r.formatBytes(report.Summary.ExpectedTFSizeBytes))
	}

	fmt.Fprintln(w)
}

func (r *Reporter) printCompatibility(w io.Writer, report Report) {
	r.printSectionHeader(w, fmt.Sprintf("🎯 COMPATIBILITY SCORE: %.1f%% (%s)",
		report.Compatibility.OverallScore,
		r.getScoreRating(report.Compatibility.OverallScore)))

	fmt.Fprintf(w, "Migration Complexity: %s\n\n", r.colorizeComplexity(report.Compatibility.Complexity))

	fmt.Fprintln(w, "Breakdown:")
	fmt.Fprintf(w, "  %s Clean migrations:  %d/%d (%.1f%%)\n",
		r.colorize("✅", "green"),
		report.Compatibility.CleanMigrations,
		report.Compatibility.CleanMigrations+report.Compatibility.WarningCount+report.Compatibility.ErrorCount,
		report.Compatibility.OverallScore)

	if report.Compatibility.WarningCount > 0 {
		fmt.Fprintf(w, "  %s With warnings:     %d/%d (%.1f%%)\n",
			r.colorize("⚠️ ", "yellow"),
			report.Compatibility.WarningCount,
			report.Compatibility.CleanMigrations+report.Compatibility.WarningCount+report.Compatibility.ErrorCount,
			float64(report.Compatibility.WarningCount)/float64(report.Compatibility.CleanMigrations+report.Compatibility.WarningCount+report.Compatibility.ErrorCount)*100)
	}

	if report.Compatibility.ErrorCount > 0 {
		fmt.Fprintf(w, "  %s Errors:            %d/%d (%.1f%%)\n",
			r.colorize("❌", "red"),
			report.Compatibility.ErrorCount,
			report.Compatibility.CleanMigrations+report.Compatibility.WarningCount+report.Compatibility.ErrorCount,
			float64(report.Compatibility.ErrorCount)/float64(report.Compatibility.CleanMigrations+report.Compatibility.WarningCount+report.Compatibility.ErrorCount)*100)
	}

	if len(report.Compatibility.ByType) > 0 {
		fmt.Fprintln(w, "\nBy Resource Type:")
		for resType, score := range report.Compatibility.ByType {
			fmt.Fprintf(w, "  %-20s: %.1f%% compatible\n", resType, score)
		}
	}

	fmt.Fprintln(w)
}

func (r *Reporter) printSampleComparisons(w io.Writer, report Report) {
	if len(report.Comparison) == 0 {
		return
	}

	r.printSectionHeader(w, "📋 SAMPLE RESOURCE COMPARISONS")

	limit := 3
	if len(report.Comparison) < limit {
		limit = len(report.Comparison)
	}

	fmt.Fprintf(w, "Showing %d of %d resources (use --verbose for all):\n\n", limit, len(report.Comparison))

	result := r.diffFormatter.FormatComparisonList(report.Comparison, limit)
	fmt.Fprintln(w, result)
}

func (r *Reporter) printComparisons(w io.Writer, report Report, _ Options) {
	if len(report.Comparison) == 0 {
		return
	}

	r.printSectionHeader(w, "📋 RESOURCE COMPARISONS")

	result := r.diffFormatter.FormatComparisonList(report.Comparison, 0)
	fmt.Fprintln(w, result)
}

func (r *Reporter) printTerraformPreview(w io.Writer, report Report) {
	r.printSectionHeader(w, "📝 TERRAFORM PREVIEW")
	fmt.Fprintln(w, report.TerraformPreview)
	fmt.Fprintln(w)
}

func (r *Reporter) printWarnings(w io.Writer, report Report) {
	if len(report.Warnings) == 0 {
		return
	}

	r.printSectionHeader(w, fmt.Sprintf("⚠️  WARNINGS & MANUAL STEPS (%d)", len(report.Warnings)))

	categorized := r.compatAnalyzer.CategorizeWarnings(report.Warnings)

	// Critical warnings
	if critical, ok := categorized["critical"]; ok && len(critical) > 0 {
		fmt.Fprintf(w, "%s (%d):\n", r.colorize("CRITICAL", "red"), len(critical))
		for i, warning := range critical {
			r.printWarning(w, warning, i+1)
		}
		fmt.Fprintln(w)
	}

	// Warnings
	if warnings, ok := categorized["warning"]; ok && len(warnings) > 0 {
		fmt.Fprintf(w, "%s (%d):\n", r.colorize("WARNINGS", "yellow"), len(warnings))
		for i, warning := range warnings {
			r.printWarning(w, warning, i+1)
		}
		fmt.Fprintln(w)
	}

	// Info
	if info, ok := categorized["info"]; ok && len(info) > 0 {
		fmt.Fprintf(w, "INFO (%d):\n", len(info))
		for _, warning := range info {
			fmt.Fprintf(w, "  - %s\n", warning.Message)
		}
		fmt.Fprintln(w)
	}

	// Effort estimate
	minutes, desc := r.compatAnalyzer.EstimateManualEffort(report.Warnings)
	if minutes > 0 {
		fmt.Fprintf(w, "Estimated manual effort: %s\n\n", desc)
	}
}

func (r *Reporter) printWarning(w io.Writer, warning Warning, index int) {
	fmt.Fprintf(w, "  %d. %s\n", index, warning.Message)
	if warning.ResourceName != "" {
		fmt.Fprintf(w, "     Resource: %s (%s)\n", warning.ResourceName, warning.ResourceType)
	}
	if warning.Action != "" {
		fmt.Fprintf(w, "     Action: %s\n", warning.Action)
	}
	if warning.Impact != "" {
		fmt.Fprintf(w, "     Impact: %s\n", warning.Impact)
	}
	fmt.Fprintln(w)
}

func (r *Reporter) printEstimates(w io.Writer, report Report) {
	r.printSectionHeader(w, "⏱️  PERFORMANCE ESTIMATES")

	fmt.Fprintf(w, "Migration Time:        ~%s\n", r.formatDuration(report.Estimates.MigrationTime))

	fmt.Fprintln(w, "API Calls:")
	fmt.Fprintf(w, "  - %s:      %d calls\n", report.SourcePlatform, report.Estimates.SourceAPICalls)
	fmt.Fprintf(w, "  - Hyperping:         %d calls (dry-run mode)\n", report.Estimates.TargetAPICalls)

	fmt.Fprintln(w, "\nTerraform Operations:")
	if report.Estimates.TerraformPlanTime > 0 {
		fmt.Fprintf(w, "  - terraform plan:    ~%s (estimated)\n", r.formatDuration(report.Estimates.TerraformPlanTime))
	}
	if report.Estimates.TerraformApplyTime > 0 {
		fmt.Fprintf(w, "  - terraform apply:   ~%s (%d resources)\n",
			r.formatDuration(report.Estimates.TerraformApplyTime),
			report.Summary.ExpectedTFResources)
	}

	fmt.Fprintln(w, "\nGenerated Files:")
	if report.Estimates.TerraformFileSize > 0 {
		fmt.Fprintf(w, "  - Terraform config:  %s\n", r.formatBytes(report.Estimates.TerraformFileSize))
	}
	if report.Estimates.ImportScriptSize > 0 {
		fmt.Fprintf(w, "  - Import script:     %s\n", r.formatBytes(report.Estimates.ImportScriptSize))
	}
	if report.Estimates.ManualStepsSize > 0 {
		fmt.Fprintf(w, "  - Manual steps:      %s\n", r.formatBytes(report.Estimates.ManualStepsSize))
	}
	if report.Estimates.ReportSize > 0 {
		fmt.Fprintf(w, "  - Migration report:  %s\n", r.formatBytes(report.Estimates.ReportSize))
	}

	fmt.Fprintln(w)
}

func (r *Reporter) printNextSteps(w io.Writer, report Report) {
	r.printSectionHeader(w, "📚 NEXT STEPS")

	fmt.Fprintln(w, "Ready to proceed? Remove --dry-run flag to execute migration:")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  migrate-betterstack --output=hyperping.tf")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Review documentation:")
	fmt.Fprintln(w, "  - Migration guide: docs/guides/migrate-from-betterstack.md")
	fmt.Fprintln(w, "  - Dry-run guide:   docs/DRY_RUN_GUIDE.md")
	fmt.Fprintln(w)
}

func (r *Reporter) printSectionHeader(w io.Writer, title string) {
	r.printSectionDivider(w)
	fmt.Fprintf(w, "%s\n", title)
	r.printSectionDivider(w)
	fmt.Fprintln(w)
}

func (r *Reporter) printSectionDivider(w io.Writer) {
	fmt.Fprintln(w, strings.Repeat("━", 70))
}

func (r *Reporter) colorize(text, color string) string {
	if !r.useColors {
		return text
	}

	codes := map[string]string{
		"bold":   "\033[1m",
		"red":    "\033[31m",
		"green":  "\033[32m",
		"yellow": "\033[33m",
		"blue":   "\033[34m",
		"cyan":   "\033[36m",
		"reset":  "\033[0m",
	}

	code, ok := codes[color]
	if !ok {
		return text
	}

	return code + text + codes["reset"]
}

func (r *Reporter) colorizeComplexity(complexity string) string {
	switch complexity {
	case "Simple":
		return r.colorize(complexity, "green")
	case "Medium":
		return r.colorize(complexity, "yellow")
	case "Complex":
		return r.colorize(complexity, "red")
	default:
		return complexity
	}
}

func (r *Reporter) getScoreRating(score float64) string {
	switch {
	case score >= 90:
		return r.colorize("EXCELLENT", "green")
	case score >= 75:
		return r.colorize("GOOD", "green")
	case score >= 50:
		return r.colorize("FAIR", "yellow")
	default:
		return r.colorize("NEEDS REVIEW", "red")
	}
}

func (r *Reporter) formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB"}
	if exp >= len(units) {
		exp = len(units) - 1
	}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

func (r *Reporter) formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	if seconds > 0 {
		return fmt.Sprintf("%d minutes %d seconds", minutes, seconds)
	}
	return fmt.Sprintf("%d minutes", minutes)
}

// FormatJSON returns the report as JSON.
func (r *Reporter) FormatJSON(report Report) (string, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
