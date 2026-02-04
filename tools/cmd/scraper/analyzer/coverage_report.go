package analyzer

import (
	"fmt"
	"strings"
	"time"
)

// FormatCoverageReportMarkdown generates a markdown coverage report
func FormatCoverageReportMarkdown(report *CoverageReport) string {
	var md strings.Builder

	md.WriteString("# Provider Coverage Report\n\n")
	md.WriteString(fmt.Sprintf("**Generated:** %s\n\n", report.Timestamp.Format("2006-01-02 15:04:05 UTC")))
	md.WriteString(fmt.Sprintf("**Overall Coverage:** %.1f%%\n\n", report.CoveragePercent))

	// Summary table
	md.WriteString("## Summary\n\n")
	md.WriteString("| Resource | API Fields | Implemented | Missing | Stale | Coverage |\n")
	md.WriteString("|----------|-----------|-------------|---------|-------|----------|\n")

	for _, rc := range report.Resources {
		md.WriteString(fmt.Sprintf("| %s | %d | %d | %d | %d | %.1f%% |\n",
			rc.Resource, rc.APIFields, rc.ImplementedFields, rc.MissingFields, rc.StaleFields, rc.CoveragePercent))
	}

	md.WriteString("\n")

	// Gaps by resource
	if len(report.Gaps) > 0 {
		md.WriteString("## Coverage Gaps\n\n")

		// Group gaps by resource
		gapsByResource := make(map[string][]CoverageGap)
		for _, gap := range report.Gaps {
			gapsByResource[gap.Resource] = append(gapsByResource[gap.Resource], gap)
		}

		for resource, gaps := range gapsByResource {
			md.WriteString(fmt.Sprintf("### %s\n\n", resource))

			// Group by type
			missing := filterGapsByType(gaps, GapMissing)
			typeMismatch := filterGapsByType(gaps, GapTypeMismatch)
			stale := filterGapsByType(gaps, GapStale)

			if len(missing) > 0 {
				md.WriteString("#### Missing Fields (API has, provider doesn't)\n\n")
				for _, gap := range missing {
					md.WriteString(fmt.Sprintf("- **%s** (`%s`)\n", gap.APIField, gap.APIType))
					if gap.Suggestion != "" {
						md.WriteString(fmt.Sprintf("  - %s\n", gap.Suggestion))
					}
					if gap.FilePath != "" {
						md.WriteString(fmt.Sprintf("  - File: `%s`\n", gap.FilePath))
					}
				}
				md.WriteString("\n")
			}

			if len(typeMismatch) > 0 {
				md.WriteString("#### Type Mismatches\n\n")
				for _, gap := range typeMismatch {
					md.WriteString(fmt.Sprintf("- **%s**: API=`%s`, TF=`%s`\n", gap.TFField, gap.APIType, gap.TFType))
				}
				md.WriteString("\n")
			}

			if len(stale) > 0 {
				md.WriteString("#### Stale Fields (provider has, API doesn't document)\n\n")
				for _, gap := range stale {
					md.WriteString(fmt.Sprintf("- **%s**\n", gap.TFField))
				}
				md.WriteString("\n")
			}
		}
	} else {
		md.WriteString("## Coverage Gaps\n\n")
		md.WriteString("✅ No coverage gaps detected!\n\n")
	}

	// Action items
	md.WriteString("## Action Items\n\n")
	if report.MissingFields > 0 {
		md.WriteString(fmt.Sprintf("- [ ] Implement %d missing fields\n", report.MissingFields))
	}
	if report.StaleFields > 0 {
		md.WriteString(fmt.Sprintf("- [ ] Review %d potentially stale fields\n", report.StaleFields))
	}
	if report.MissingFields == 0 && report.StaleFields == 0 {
		md.WriteString("- ✅ No action items - full coverage!\n")
	}

	return md.String()
}

// FormatGitHubIssueBody generates a GitHub issue body for a resource's gaps
func FormatGitHubIssueBody(resource string, tfResource string, coverage ResourceCoverage, gaps []CoverageGap) string {
	var md strings.Builder

	md.WriteString(fmt.Sprintf("## Coverage Analysis for %s\n\n", tfResource))
	md.WriteString(fmt.Sprintf("**Coverage:** %.1f%% (%d/%d API fields implemented)\n",
		coverage.CoveragePercent, coverage.ImplementedFields, coverage.APIFields))
	md.WriteString(fmt.Sprintf("**Last checked:** %s\n\n", time.Now().Format("2006-01-02")))

	// Missing fields
	missingGaps := filterGapsByType(gaps, GapMissing)
	if len(missingGaps) > 0 {
		md.WriteString("### Missing Fields (API has, provider doesn't)\n\n")

		for i, gap := range missingGaps {
			required := "optional"
			if strings.Contains(strings.ToLower(gap.Details), "required") {
				required = "required"
			}

			md.WriteString(fmt.Sprintf("#### %d. `%s` (%s, %s)\n", i+1, gap.APIField, gap.APIType, required))

			if gap.FilePath != "" {
				md.WriteString(fmt.Sprintf("- **File to modify:** `%s`\n", gap.FilePath))
			}

			if gap.CodeHint != "" {
				md.WriteString("- **Suggested addition:**\n")
				md.WriteString("```go\n")
				md.WriteString(gap.CodeHint)
				md.WriteString("\n```\n")
			}

			md.WriteString("\n")
		}
	}

	// Type mismatches
	typeMismatches := filterGapsByType(gaps, GapTypeMismatch)
	if len(typeMismatches) > 0 {
		md.WriteString("### Type Mismatches\n\n")
		md.WriteString("| Field | API Type | TF Type | Notes |\n")
		md.WriteString("|-------|----------|---------|-------|\n")

		for _, gap := range typeMismatches {
			md.WriteString(fmt.Sprintf("| `%s` | %s | %s | Verify intentional |\n",
				gap.TFField, gap.APIType, gap.TFType))
		}
		md.WriteString("\n")
	}

	// Action items
	md.WriteString("### Action Items\n\n")

	for i, gap := range missingGaps {
		tfName := gap.TFField
		if tfName == "" {
			tfName = CamelToSnake(gap.APIField)
		}
		md.WriteString(fmt.Sprintf("- [ ] %d. Add `%s` to resource model\n", i+1, tfName))
	}

	if len(missingGaps) > 0 {
		md.WriteString(fmt.Sprintf("- [ ] Add schema attribute definitions\n"))
		md.WriteString(fmt.Sprintf("- [ ] Add to client request/response mapping\n"))
		md.WriteString(fmt.Sprintf("- [ ] Update documentation\n"))
	}

	md.WriteString("\n---\n")
	md.WriteString("*This issue was automatically generated by the API coverage analyzer.*\n")

	return md.String()
}

// FormatGitHubIssueTitle generates a GitHub issue title for a resource
func FormatGitHubIssueTitle(tfResource string, missingCount int) string {
	if missingCount == 1 {
		return fmt.Sprintf("[Coverage Gap] %s: 1 field needs attention", tfResource)
	}
	return fmt.Sprintf("[Coverage Gap] %s: %d fields need attention", tfResource, missingCount)
}

// GetGitHubIssueLabels returns labels for coverage gap issues
func GetGitHubIssueLabels() []string {
	return []string{"coverage-gap", "enhancement", "automated"}
}

// filterGapsByType filters gaps by their type
func filterGapsByType(gaps []CoverageGap, gapType CoverageGapType) []CoverageGap {
	var filtered []CoverageGap
	for _, gap := range gaps {
		if gap.Type == gapType {
			filtered = append(filtered, gap)
		}
	}
	return filtered
}

// GroupGapsByResource groups gaps by their resource
func GroupGapsByResource(gaps []CoverageGap) map[string][]CoverageGap {
	result := make(map[string][]CoverageGap)
	for _, gap := range gaps {
		result[gap.Resource] = append(result[gap.Resource], gap)
	}
	return result
}

// GetResourcesWithGaps returns resources that have coverage gaps
func GetResourcesWithGaps(report *CoverageReport) []string {
	gapsByResource := GroupGapsByResource(report.Gaps)
	var resources []string
	for resource := range gapsByResource {
		resources = append(resources, resource)
	}
	return resources
}

// GetCoverageForResource returns coverage stats for a specific resource
func GetCoverageForResource(report *CoverageReport, resource string) *ResourceCoverage {
	for i := range report.Resources {
		if report.Resources[i].Resource == resource {
			return &report.Resources[i]
		}
	}
	return nil
}
