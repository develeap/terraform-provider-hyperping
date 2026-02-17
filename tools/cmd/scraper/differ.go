package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/extractor"
	"github.com/develeap/terraform-provider-hyperping/tools/scraper/utils"
)

// CompareParameters compares two sets of parameters and returns a diff
func CompareParameters(section, endpoint, method string, oldParams, newParams []extractor.APIParameter) APIDiff {
	diff := APIDiff{
		Section:  section,
		Endpoint: endpoint,
		Method:   method,
		Breaking: false,
	}

	oldMap := buildParamMap(oldParams)
	newMap := buildParamMap(newParams)

	diff = detectAddedParams(diff, oldMap, newMap)
	diff = detectRemovedParams(diff, oldMap, newMap)
	diff = detectModifiedParams(diff, oldMap, newMap)

	return diff
}

// buildParamMap creates a name-keyed lookup map from a parameter slice.
func buildParamMap(params []extractor.APIParameter) map[string]extractor.APIParameter {
	m := make(map[string]extractor.APIParameter, len(params))
	for _, p := range params {
		m[p.Name] = p
	}
	return m
}

// detectAddedParams appends newly added parameters to diff.
func detectAddedParams(diff APIDiff, oldMap, newMap map[string]extractor.APIParameter) APIDiff {
	for name, newParam := range newMap {
		if _, exists := oldMap[name]; !exists {
			diff.AddedParams = append(diff.AddedParams, newParam)
			if newParam.Required {
				diff.Breaking = true
			}
		}
	}
	return diff
}

// detectRemovedParams appends removed parameters to diff.
func detectRemovedParams(diff APIDiff, oldMap, newMap map[string]extractor.APIParameter) APIDiff {
	for name, oldParam := range oldMap {
		if _, exists := newMap[name]; !exists {
			diff.RemovedParams = append(diff.RemovedParams, oldParam)
			diff.Breaking = true
		}
	}
	return diff
}

// detectModifiedParams appends changed parameters to diff.
func detectModifiedParams(diff APIDiff, oldMap, newMap map[string]extractor.APIParameter) APIDiff {
	for name, oldParam := range oldMap {
		newParam, exists := newMap[name]
		if !exists {
			continue
		}
		if oldParam.Type == newParam.Type &&
			oldParam.Required == newParam.Required &&
			oldParam.Default == newParam.Default {
			continue
		}
		change := ParameterChange{
			Name:        name,
			OldType:     oldParam.Type,
			NewType:     newParam.Type,
			OldRequired: oldParam.Required,
			NewRequired: newParam.Required,
			OldDefault:  oldParam.Default,
			NewDefault:  newParam.Default,
		}
		diff.ModifiedParams = append(diff.ModifiedParams, change)
		if oldParam.Type != newParam.Type {
			diff.Breaking = true
		}
		if !oldParam.Required && newParam.Required {
			diff.Breaking = true
		}
	}
	return diff
}

// GenerateDiffReport creates a comprehensive diff report for all changes
func GenerateDiffReport(diffs []APIDiff, timestamp time.Time) DiffReport {
	report := DiffReport{
		Timestamp: timestamp,
		APIDiffs:  diffs,
	}

	// Calculate statistics
	changedPages := 0
	breakingCount := 0

	for _, diff := range diffs {
		if len(diff.AddedParams) > 0 || len(diff.RemovedParams) > 0 || len(diff.ModifiedParams) > 0 {
			changedPages++
		}
		if diff.Breaking {
			breakingCount++
			report.Breaking = true
		}
	}

	report.ChangedPages = changedPages

	// Generate summary
	if changedPages == 0 {
		report.Summary = "No API changes detected"
	} else {
		report.Summary = fmt.Sprintf("%d endpoint(s) changed", changedPages)
		if breakingCount > 0 {
			report.Summary += fmt.Sprintf(" (%d breaking)", breakingCount)
		}
	}

	return report
}

// FormatDiffAsMarkdown converts a diff report to a markdown document
func FormatDiffAsMarkdown(report DiffReport) string {
	var md strings.Builder

	md.WriteString("# API Changes Detected\n\n")
	md.WriteString(fmt.Sprintf("**Date:** %s\n\n", report.Timestamp.Format("2006-01-02 15:04:05 MST")))
	md.WriteString(fmt.Sprintf("**Summary:** %s\n\n", report.Summary))

	if report.Breaking {
		md.WriteString("⚠️ **WARNING: Breaking changes detected!**\n\n")
	}

	md.WriteString("---\n\n")

	sections := groupDiffsBySection(report.APIDiffs)

	if len(sections) == 0 {
		md.WriteString("✅ No changes detected.\n")
		return md.String()
	}

	for section, diffs := range sections {
		md.WriteString(fmt.Sprintf("## %s API\n\n", strings.Title(section)))
		for _, diff := range diffs {
			md.WriteString(formatSingleDiff(diff))
		}
	}

	md.WriteString(formatDiffActionItems(report.Breaking))

	return md.String()
}

// groupDiffsBySection groups non-empty diffs by their section name.
func groupDiffsBySection(diffs []APIDiff) map[string][]APIDiff {
	sections := make(map[string][]APIDiff)
	for _, diff := range diffs {
		if len(diff.AddedParams) > 0 || len(diff.RemovedParams) > 0 || len(diff.ModifiedParams) > 0 {
			sections[diff.Section] = append(sections[diff.Section], diff)
		}
	}
	return sections
}

// formatSingleDiff renders one APIDiff entry as markdown.
func formatSingleDiff(diff APIDiff) string {
	var md strings.Builder

	endpointName := diff.Method
	if endpointName == "" {
		endpointName = "Overview"
	}
	md.WriteString(fmt.Sprintf("### %s\n\n", strings.Title(endpointName)))

	if diff.Breaking {
		md.WriteString("⚠️ **BREAKING CHANGE**\n\n")
	}

	md.WriteString(formatAddedParams(diff.AddedParams))
	md.WriteString(formatRemovedParams(diff.RemovedParams))
	md.WriteString(formatModifiedParams(diff.ModifiedParams))
	md.WriteString(formatImpactSection(diff.Breaking))
	md.WriteString("---\n\n")

	return md.String()
}

// formatAddedParams renders the added-parameters section.
func formatAddedParams(params []extractor.APIParameter) string {
	if len(params) == 0 {
		return ""
	}
	var md strings.Builder
	md.WriteString("#### 🆕 Added Parameters\n\n")
	for _, p := range params {
		required := "optional"
		if p.Required {
			required = "**required**"
		}
		md.WriteString(fmt.Sprintf("- **%s** (%s, %s)\n", p.Name, p.Type, required))
		if p.Description != "" {
			md.WriteString(fmt.Sprintf("  - %s\n", p.Description))
		}
		if p.Default != nil {
			md.WriteString(fmt.Sprintf("  - Default: `%v`\n", p.Default))
		}
	}
	md.WriteString("\n")
	return md.String()
}

// formatRemovedParams renders the removed-parameters section.
func formatRemovedParams(params []extractor.APIParameter) string {
	if len(params) == 0 {
		return ""
	}
	var md strings.Builder
	md.WriteString("#### ❌ Removed Parameters\n\n")
	for _, p := range params {
		required := "optional"
		if p.Required {
			required = "**required**"
		}
		md.WriteString(fmt.Sprintf("- **%s** (%s, %s)\n", p.Name, p.Type, required))
		if p.Description != "" {
			md.WriteString(fmt.Sprintf("  - %s\n", p.Description))
		}
	}
	md.WriteString("\n")
	return md.String()
}

// formatModifiedParams renders the modified-parameters section.
func formatModifiedParams(changes []ParameterChange) string {
	if len(changes) == 0 {
		return ""
	}
	var md strings.Builder
	md.WriteString("#### 📝 Modified Parameters\n\n")
	for _, change := range changes {
		md.WriteString(fmt.Sprintf("- **%s**\n", change.Name))
		if change.OldType != change.NewType {
			md.WriteString(fmt.Sprintf("  - Type: `%s` → `%s`\n", change.OldType, change.NewType))
		}
		if change.OldRequired != change.NewRequired {
			md.WriteString(formatRequiredChange(change))
		}
		if change.OldDefault != change.NewDefault {
			md.WriteString(fmt.Sprintf("  - Default: `%v` → `%v`\n", change.OldDefault, change.NewDefault))
		}
	}
	md.WriteString("\n")
	return md.String()
}

// formatRequiredChange renders the required/optional status change line.
func formatRequiredChange(change ParameterChange) string {
	oldReq := "optional"
	newReq := "optional"
	if change.OldRequired {
		oldReq = "required"
	}
	if change.NewRequired {
		newReq = "required"
	}
	line := fmt.Sprintf("  - Status: `%s` → `%s`", oldReq, newReq)
	if !change.OldRequired && change.NewRequired {
		line += " ⚠️ **BREAKING**"
	}
	return line + "\n"
}

// formatImpactSection renders the impact assessment block.
func formatImpactSection(breaking bool) string {
	var md strings.Builder
	md.WriteString("#### Impact\n\n")
	if breaking {
		md.WriteString("❌ **Breaking change** - Requires immediate attention\n")
		md.WriteString("- Update provider schema\n")
		md.WriteString("- Update validation logic\n")
		md.WriteString("- Test backward compatibility\n")
		md.WriteString("- Update documentation\n")
	} else {
		md.WriteString("✅ **Non-breaking change** - Can be added incrementally\n")
		md.WriteString("- Consider adding new optional fields to provider\n")
		md.WriteString("- Update documentation\n")
	}
	md.WriteString("\n")
	return md.String()
}

// formatDiffActionItems renders the trailing action-items checklist.
func formatDiffActionItems(breaking bool) string {
	var md strings.Builder
	md.WriteString("## Action Items\n\n")
	md.WriteString("- [ ] Review all changes\n")
	md.WriteString("- [ ] Update Terraform provider schema\n")
	md.WriteString("- [ ] Update validation logic\n")
	md.WriteString("- [ ] Update tests\n")
	md.WriteString("- [ ] Update documentation\n")
	if breaking {
		md.WriteString("- [ ] ⚠️ Plan migration strategy for breaking changes\n")
		md.WriteString("- [ ] Communicate breaking changes to users\n")
	}
	return md.String()
}

// SaveDiffReport writes the diff report to a markdown file
func SaveDiffReport(report DiffReport, filename string) error {
	markdown := FormatDiffAsMarkdown(report)
	return utils.SaveToFile(filename, markdown)
}

// CompareCachedPages compares parameters from two cache entries
// Returns a diff if parameters changed, or nil if no semantic changes
func CompareCachedPages(oldPage, newPage *extractor.PageData) *APIDiff {
	oldParams := extractor.ExtractAPIParameters(oldPage)
	newParams := extractor.ExtractAPIParameters(newPage)

	// Parse section and method from URL
	section, method := parseURLComponents(newPage.URL)

	// Compare parameters
	diff := CompareParameters(section, newPage.URL, method, oldParams, newParams)

	// Only return diff if there are actual changes
	if len(diff.AddedParams) > 0 || len(diff.RemovedParams) > 0 || len(diff.ModifiedParams) > 0 {
		return &diff
	}

	return nil
}

// parseURLComponents extracts section and method from a URL
// Example: https://hyperping.com/docs/api/monitors/create → section=monitors, method=create
func parseURLComponents(url string) (section, method string) {
	// Remove base URL
	path := strings.TrimPrefix(url, "https://hyperping.com/docs/api/")

	// Split by /
	parts := strings.Split(path, "/")

	if len(parts) >= 1 {
		section = parts[0]
	}
	if len(parts) >= 2 {
		method = parts[1]
	}

	return
}
