package main

import (
	"fmt"
	"strings"
)

// PreviewGitHubIssue shows what the issue would look like without creating it
func PreviewGitHubIssue(report DiffReport, snapshotURL string) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("📋 GITHUB ISSUE PREVIEW")
	fmt.Println(strings.Repeat("=", 70))

	// Generate title
	title := fmt.Sprintf("[API Change] %s", report.Summary)
	fmt.Printf("\n📌 TITLE:\n%s\n", title)

	// Generate labels
	labels := determineLabelsForPreview(report)
	fmt.Printf("\n🏷️  LABELS:\n%s\n", strings.Join(labels, ", "))

	// Generate body
	client := &GitHubClient{}
	body := client.formatIssueBody(report, snapshotURL)

	fmt.Printf("\n📄 BODY:\n")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Println(body)
	fmt.Println(strings.Repeat("-", 70))

	// Summary
	fmt.Println("\n✅ Preview generated successfully!")
	fmt.Println("\nTo create this issue for real:")
	fmt.Println("  1. Set GITHUB_TOKEN environment variable")
	fmt.Println("  2. Set GITHUB_REPOSITORY=owner/repo")
	fmt.Println("  3. Run scraper again")
}

// determineLabelsForPreview generates labels without needing a client
func determineLabelsForPreview(report DiffReport) []string {
	labels := []string{"api-change", "automated"}

	// Breaking vs non-breaking
	if report.Breaking {
		labels = append(labels, "breaking-change")
	} else {
		labels = append(labels, "non-breaking")
	}

	// Per-API section labels
	sections := make(map[string]bool)
	for _, diff := range report.APIDiffs {
		sections[diff.Section] = true
	}

	for section := range sections {
		labels = append(labels, section+"-api")
	}

	return labels
}
