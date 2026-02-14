// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package dryrun

import (
	"fmt"
	"strings"
)

// PreviewGenerator creates Terraform configuration previews.
type PreviewGenerator struct {
	useColors bool
}

// NewPreviewGenerator creates a new preview generator.
func NewPreviewGenerator(useColors bool) *PreviewGenerator {
	return &PreviewGenerator{
		useColors: useColors,
	}
}

// GeneratePreview creates a preview showing sample resources.
func (g *PreviewGenerator) GeneratePreview(
	tfContent string,
	resourceCount int,
	previewLimit int,
) string {
	var sb strings.Builder

	// Parse resources from TF content
	resources := g.extractResources(tfContent)

	// Show header
	sb.WriteString(g.colorize("📄 Terraform Preview", "bold"))
	if len(resources) > previewLimit {
		sb.WriteString(fmt.Sprintf(" (showing %d of %d resources):\n\n", previewLimit, len(resources)))
	} else {
		sb.WriteString(fmt.Sprintf(" (%d resources):\n\n", len(resources)))
	}

	// Show sample resources
	count := previewLimit
	if count > len(resources) {
		count = len(resources)
	}

	for i := 0; i < count; i++ {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(g.highlightResource(resources[i]))
	}

	// Show summary if truncated
	if len(resources) > previewLimit {
		remaining := len(resources) - previewLimit
		sb.WriteString(fmt.Sprintf("\n... (%d more resources)\n", remaining))
	}

	// Resource breakdown
	breakdown := g.analyzeResources(resources)
	sb.WriteString("\n")
	sb.WriteString(g.colorize("📊 Resource Breakdown:\n", "bold"))
	for resType, count := range breakdown {
		sb.WriteString(fmt.Sprintf("  - %s: %d\n", resType, count))
	}

	// Estimate file size
	lines := strings.Count(tfContent, "\n")
	size := len(tfContent)
	sb.WriteString(fmt.Sprintf("\n  Total size: ~%d lines (~%s)\n", lines, g.formatBytes(int64(size))))

	return sb.String()
}

// extractResources parses resource blocks from Terraform content.
func (g *PreviewGenerator) extractResources(tfContent string) []string {
	var resources []string
	var current strings.Builder
	inResource := false
	braceCount := 0

	lines := strings.Split(tfContent, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "resource ") {
			inResource = true
			braceCount = 0
			current.Reset()
		}

		if inResource {
			current.WriteString(line)
			current.WriteString("\n")

			braceCount += strings.Count(line, "{")
			braceCount -= strings.Count(line, "}")

			if braceCount == 0 && strings.Contains(line, "}") {
				resources = append(resources, current.String())
				inResource = false
			}
		}
	}

	return resources
}

// highlightResource adds syntax highlighting to a resource block.
func (g *PreviewGenerator) highlightResource(resource string) string {
	if !g.useColors {
		return resource
	}

	lines := strings.Split(resource, "\n")
	var highlighted []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(trimmed, "#"):
			// Comment
			highlighted = append(highlighted, g.colorize(line, "gray"))
		case strings.HasPrefix(trimmed, "resource "):
			// Resource declaration
			highlighted = append(highlighted, g.colorize(line, "blue"))
		case strings.Contains(trimmed, "=") && !strings.HasPrefix(trimmed, "}"):
			// Attribute assignment
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := g.colorize(parts[0], "cyan")
				highlighted = append(highlighted, key+"="+parts[1])
			} else {
				highlighted = append(highlighted, line)
			}
		default:
			highlighted = append(highlighted, line)
		}
	}

	return strings.Join(highlighted, "\n")
}

// analyzeResources counts resources by type.
func (g *PreviewGenerator) analyzeResources(resources []string) map[string]int {
	breakdown := make(map[string]int)

	for _, res := range resources {
		resType := g.extractResourceType(res)
		if resType != "" {
			breakdown[resType]++
		}
	}

	return breakdown
}

// extractResourceType extracts the resource type from a resource block.
func (g *PreviewGenerator) extractResourceType(resource string) string {
	lines := strings.Split(resource, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "resource ") {
			// Format: resource "type" "name" {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				resType := strings.Trim(parts[1], `"`)
				return resType
			}
		}
	}
	return ""
}

func (g *PreviewGenerator) colorize(text, style string) string {
	if !g.useColors {
		return text
	}

	codes := map[string]string{
		"bold":  "\033[1m",
		"blue":  "\033[34m",
		"cyan":  "\033[36m",
		"gray":  "\033[90m",
		"reset": "\033[0m",
	}

	code, ok := codes[style]
	if !ok {
		return text
	}

	return code + text + codes["reset"]
}

func (g *PreviewGenerator) formatBytes(bytes int64) string {
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

// GenerateResourceSummary creates a detailed resource breakdown.
func (g *PreviewGenerator) GenerateResourceSummary(
	monitorCount int,
	healthcheckCount int,
	frequencyDist map[int]int,
	regionDist map[string]int,
) string {
	var sb strings.Builder

	sb.WriteString(g.colorize("📊 Resource Distribution:\n", "bold"))
	sb.WriteString(fmt.Sprintf("  Monitors:     %d\n", monitorCount))
	sb.WriteString(fmt.Sprintf("  Healthchecks: %d\n", healthcheckCount))
	sb.WriteString(fmt.Sprintf("  Total:        %d\n", monitorCount+healthcheckCount))

	if len(frequencyDist) > 0 {
		sb.WriteString("\n")
		sb.WriteString(g.colorize("⏱️  Check Frequency Distribution:\n", "bold"))
		for freq, count := range frequencyDist {
			sb.WriteString(fmt.Sprintf("  %6ds: %d monitors\n", freq, count))
		}
	}

	if len(regionDist) > 0 {
		sb.WriteString("\n")
		sb.WriteString(g.colorize("🌍 Region Distribution:\n", "bold"))
		for region, count := range regionDist {
			sb.WriteString(fmt.Sprintf("  %-12s: %d monitors\n", region, count))
		}
	}

	return sb.String()
}
