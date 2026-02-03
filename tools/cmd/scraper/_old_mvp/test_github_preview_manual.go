package main

import (
	"fmt"
	"time"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/extractor"
)

func main() {
	fmt.Println("🧪 GitHub Issue Preview Test")
	fmt.Println("=============================\n")

	// Create a sample diff report with a breaking change
	report := DiffReport{
		Timestamp:      time.Now(),
		TotalPages:     50,
		ChangedPages:   1,
		UnchangedPages: 49,
		Breaking:       true,
		Summary:        "1 endpoint(s) changed (1 breaking)",
		APIDiffs: []APIDiff{
			{
				Section:  "monitors",
				Endpoint: "https://hyperping.com/docs/api/monitors/create",
				Method:   "create",
				RemovedParams: []extractor.APIParameter{
					{
						Name:        "test_mode",
						Type:        "boolean",
						Required:    false,
						Description: "Enable test mode for dry-run testing",
					},
				},
				Breaking: true,
			},
		},
	}

	// Preview the issue
	PreviewGitHubIssue(report, "https://github.com/develeap/terraform-provider-hyperping/tree/api-snapshots/2026-02-03_14-51-20")
}
