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

	// Create maps for easier lookup
	oldMap := make(map[string]extractor.APIParameter)
	newMap := make(map[string]extractor.APIParameter)

	for _, p := range oldParams {
		oldMap[p.Name] = p
	}
	for _, p := range newParams {
		newMap[p.Name] = p
	}

	// Find added parameters
	for name, newParam := range newMap {
		if _, exists := oldMap[name]; !exists {
			diff.AddedParams = append(diff.AddedParams, newParam)

			// Breaking change if added parameter is required
			if newParam.Required {
				diff.Breaking = true
			}
		}
	}

	// Find removed parameters
	for name, oldParam := range oldMap {
		if _, exists := newMap[name]; !exists {
			diff.RemovedParams = append(diff.RemovedParams, oldParam)

			// Removing any parameter is breaking
			diff.Breaking = true
		}
	}

	// Find modified parameters
	for name, oldParam := range oldMap {
		if newParam, exists := newMap[name]; exists {
			// Check if anything changed
			if oldParam.Type != newParam.Type ||
				oldParam.Required != newParam.Required ||
				oldParam.Default != newParam.Default {

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

				// Breaking changes:
				// - Type changed
				// - Optional became required
				if oldParam.Type != newParam.Type {
					diff.Breaking = true
				}
				if !oldParam.Required && newParam.Required {
					diff.Breaking = true
				}
			}
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

	// Group diffs by section
	sections := make(map[string][]APIDiff)
	for _, diff := range report.APIDiffs {
		if len(diff.AddedParams) > 0 || len(diff.RemovedParams) > 0 || len(diff.ModifiedParams) > 0 {
			sections[diff.Section] = append(sections[diff.Section], diff)
		}
	}

	if len(sections) == 0 {
		md.WriteString("✅ No changes detected.\n")
		return md.String()
	}

	// Write each section
	for section, diffs := range sections {
		md.WriteString(fmt.Sprintf("## %s API\n\n", strings.Title(section)))

		for _, diff := range diffs {
			// Endpoint header
			endpointName := diff.Method
			if endpointName == "" {
				endpointName = "Overview"
			}
			md.WriteString(fmt.Sprintf("### %s\n\n", strings.Title(endpointName)))

			if diff.Breaking {
				md.WriteString("⚠️ **BREAKING CHANGE**\n\n")
			}

			// Added parameters
			if len(diff.AddedParams) > 0 {
				md.WriteString("#### 🆕 Added Parameters\n\n")
				for _, p := range diff.AddedParams {
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
			}

			// Removed parameters
			if len(diff.RemovedParams) > 0 {
				md.WriteString("#### ❌ Removed Parameters\n\n")
				for _, p := range diff.RemovedParams {
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
			}

			// Modified parameters
			if len(diff.ModifiedParams) > 0 {
				md.WriteString("#### 📝 Modified Parameters\n\n")
				for _, change := range diff.ModifiedParams {
					md.WriteString(fmt.Sprintf("- **%s**\n", change.Name))

					if change.OldType != change.NewType {
						md.WriteString(fmt.Sprintf("  - Type: `%s` → `%s`\n", change.OldType, change.NewType))
					}

					if change.OldRequired != change.NewRequired {
						oldReq := "optional"
						newReq := "optional"
						if change.OldRequired {
							oldReq = "required"
						}
						if change.NewRequired {
							newReq = "required"
						}
						md.WriteString(fmt.Sprintf("  - Status: `%s` → `%s`", oldReq, newReq))
						if !change.OldRequired && change.NewRequired {
							md.WriteString(" ⚠️ **BREAKING**")
						}
						md.WriteString("\n")
					}

					if change.OldDefault != change.NewDefault {
						md.WriteString(fmt.Sprintf("  - Default: `%v` → `%v`\n", change.OldDefault, change.NewDefault))
					}
				}
				md.WriteString("\n")
			}

			// Impact assessment
			md.WriteString("#### Impact\n\n")
			if diff.Breaking {
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

			md.WriteString("---\n\n")
		}
	}

	// Footer
	md.WriteString("## Action Items\n\n")
	md.WriteString("- [ ] Review all changes\n")
	md.WriteString("- [ ] Update Terraform provider schema\n")
	md.WriteString("- [ ] Update validation logic\n")
	md.WriteString("- [ ] Update tests\n")
	md.WriteString("- [ ] Update documentation\n")

	if report.Breaking {
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
