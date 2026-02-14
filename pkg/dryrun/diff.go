// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package dryrun

import (
	"fmt"
	"strings"
)

// DiffFormatter creates side-by-side comparisons.
type DiffFormatter struct {
	useColors bool
}

// NewDiffFormatter creates a new diff formatter.
func NewDiffFormatter(useColors bool) *DiffFormatter {
	return &DiffFormatter{
		useColors: useColors,
	}
}

// FormatComparison creates a side-by-side comparison view.
func (f *DiffFormatter) FormatComparison(comp ResourceComparison, maxWidth int) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("Resource: %q (%s)\n", comp.ResourceName, comp.ResourceType))

	// Determine column width
	colWidth := (maxWidth - 3) / 2
	if colWidth < 30 {
		colWidth = 30
	}

	// Top border
	sb.WriteString(f.horizontalLine(colWidth, "┌", "┬", "┐"))
	sb.WriteString("\n")

	// Column headers
	sourceHeader := f.centerText("Source Platform", colWidth)
	targetHeader := f.centerText("Hyperping", colWidth)
	sb.WriteString(fmt.Sprintf("│%s│%s│\n", sourceHeader, targetHeader))

	// Separator
	sb.WriteString(f.horizontalLine(colWidth, "├", "┼", "┤"))
	sb.WriteString("\n")

	// Show transformations
	for _, trans := range comp.Transformations {
		sourceVal := f.formatValue(trans.SourceValue)
		targetVal := f.formatValue(trans.TargetValue)

		sourceField := trans.SourceField
		targetField := trans.TargetField

		sourceLine := fmt.Sprintf("%s: %s", sourceField, sourceVal)
		targetLine := fmt.Sprintf("%s: %s", targetField, targetVal)

		// Add action indicator
		indicator := f.getActionIndicator(trans.Action)
		targetLine = fmt.Sprintf("%s %s", indicator, targetLine)

		// Add notes if present
		if trans.Notes != "" {
			targetLine = fmt.Sprintf("%s (%s)", targetLine, trans.Notes)
		}

		// Pad and wrap
		sourceLines := f.wrapText(sourceLine, colWidth-2)
		targetLines := f.wrapText(targetLine, colWidth-2)

		maxLines := len(sourceLines)
		if len(targetLines) > maxLines {
			maxLines = len(targetLines)
		}

		for i := 0; i < maxLines; i++ {
			sLine := ""
			if i < len(sourceLines) {
				sLine = sourceLines[i]
			}

			tLine := ""
			if i < len(targetLines) {
				tLine = targetLines[i]
			}

			sb.WriteString(fmt.Sprintf("│ %-*s │ %-*s │\n",
				colWidth-2, sLine, colWidth-2, tLine))
		}
	}

	// Show unsupported features
	if len(comp.Unsupported) > 0 {
		sb.WriteString(f.horizontalLine(colWidth, "├", "┼", "┤"))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("│ %-*s │ %-*s │\n",
			colWidth-2, "Unsupported features:", colWidth-2, ""))

		for _, feature := range comp.Unsupported {
			sb.WriteString(fmt.Sprintf("│ %-*s │ %-*s │\n",
				colWidth-2, "  - "+feature, colWidth-2, f.colorize("⚠ Not available", "yellow")))
		}
	}

	// Bottom border
	sb.WriteString(f.horizontalLine(colWidth, "└", "┴", "┘"))
	sb.WriteString("\n")

	// Transformation summary
	if len(comp.Transformations) > 0 {
		sb.WriteString("\nTransformations:\n")
		for _, trans := range comp.Transformations {
			icon := f.getActionIcon(trans.Action)
			sb.WriteString(fmt.Sprintf("  %s %s\n", icon, trans.Notes))
		}
	}

	return sb.String()
}

func (f *DiffFormatter) horizontalLine(colWidth int, left, mid, right string) string {
	line := strings.Repeat("─", colWidth)
	return fmt.Sprintf("%s%s%s%s%s", left, line, mid, line, right)
}

func (f *DiffFormatter) centerText(text string, width int) string {
	if len(text) >= width {
		return text[:width]
	}

	padding := width - len(text)
	leftPad := padding / 2
	rightPad := padding - leftPad

	return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
}

func (f *DiffFormatter) formatValue(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case []string:
		return "[" + strings.Join(v, ", ") + "]"
	case []int:
		strs := make([]string, len(v))
		for i, n := range v {
			strs[i] = fmt.Sprintf("%d", n)
		}
		return "[" + strings.Join(strs, ", ") + "]"
	case int:
		return fmt.Sprintf("%d", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (f *DiffFormatter) getActionIndicator(action string) string {
	switch action {
	case "preserved":
		return f.colorize("→", "green")
	case "mapped":
		return f.colorize("→", "blue")
	case "rounded":
		return f.colorize("~", "yellow")
	case "defaulted":
		return f.colorize("+", "cyan")
	case "dropped":
		return f.colorize("✗", "red")
	default:
		return "→"
	}
}

func (f *DiffFormatter) getActionIcon(action string) string {
	switch action {
	case "preserved":
		return f.colorize("✓", "green")
	case "mapped":
		return f.colorize("✓", "green")
	case "rounded":
		return f.colorize("⚠", "yellow")
	case "defaulted":
		return f.colorize("ⓘ", "cyan")
	case "dropped":
		return f.colorize("✗", "red")
	default:
		return "•"
	}
}

func (f *DiffFormatter) colorize(text, color string) string {
	if !f.useColors {
		return text
	}

	codes := map[string]string{
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

func (f *DiffFormatter) wrapText(text string, width int) []string {
	if len(text) <= width {
		return []string{text}
	}

	var lines []string
	current := text

	for len(current) > width {
		// Find last space before width
		breakPoint := width
		for i := width; i > 0; i-- {
			if current[i] == ' ' {
				breakPoint = i
				break
			}
		}

		lines = append(lines, current[:breakPoint])
		current = strings.TrimSpace(current[breakPoint:])
	}

	if current != "" {
		lines = append(lines, current)
	}

	return lines
}

// FormatComparisonList creates a compact list of comparisons.
func (f *DiffFormatter) FormatComparisonList(comparisons []ResourceComparison, limit int) string {
	var sb strings.Builder

	count := len(comparisons)
	if limit > 0 && limit < count {
		count = limit
	}

	for i := 0; i < count; i++ {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(f.FormatComparison(comparisons[i], 80))
	}

	if limit > 0 && limit < len(comparisons) {
		remaining := len(comparisons) - limit
		sb.WriteString(fmt.Sprintf("\n... (%d more resources)\n", remaining))
	}

	return sb.String()
}
