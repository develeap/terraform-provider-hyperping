// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/converter"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/generator"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/report"
	"github.com/develeap/terraform-provider-hyperping/pkg/interactive"
)

// interactiveConfigPD holds configuration collected from interactive prompts.
type interactiveConfigPD struct {
	pingdomAPIKey   string
	hyperpingAPIKey string
	outputDir       string
	prefix          string
	dryRun          bool
}

// runInteractive runs the migration tool in interactive mode.
func runInteractive() int {
	prompter := interactive.NewPrompter(interactive.DefaultConfig())

	// Welcome message
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "🚀 Hyperping Migration Tool - Pingdom Edition\n")
	fmt.Fprintf(os.Stderr, "═════════════════════════════════════════════════\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "This wizard will guide you through migrating your Pingdom\n")
	fmt.Fprintf(os.Stderr, "checks to Hyperping.\n")
	fmt.Fprintf(os.Stderr, "\n")

	config := &interactiveConfigPD{}

	// Step 1: Source Platform Configuration
	prompter.PrintHeader("Step 1/5: Source Platform Configuration")
	fmt.Fprintf(os.Stderr, "\n")

	apiKey, err := prompter.AskPassword(
		"Enter your Pingdom API token:",
		"Get it from: https://my.pingdom.com/app/api-tokens",
		interactive.SourceAPIKeyValidator("pingdom"),
	)
	if err != nil {
		prompter.PrintError(fmt.Sprintf("Failed to get API token: %v", err))
		return 1
	}
	config.pingdomAPIKey = apiKey

	// Test Pingdom API connection
	spinner := interactive.NewSpinner("Testing Pingdom API connection...", os.Stderr)
	spinner.Start()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pingdomClient := createPingdomClient(config.pingdomAPIKey)
	checks, err := pingdomClient.ListChecks(ctx)
	if err != nil {
		spinner.ErrorMessage(fmt.Sprintf("Connection failed: %v", err))
		prompter.PrintError("Unable to connect to Pingdom API")
		prompter.PrintInfo("Please verify your API token and try again")
		return 1
	}

	spinner.SuccessMessage(fmt.Sprintf("Connected! Found %d checks", len(checks)))

	// Show check type breakdown
	fmt.Fprintf(os.Stderr, "\n")
	typeCounts := make(map[string]int)
	for _, c := range checks {
		typeCounts[c.Type]++
	}
	fmt.Fprintf(os.Stderr, "  Check types:\n")
	for checkType, count := range typeCounts {
		fmt.Fprintf(os.Stderr, "    - %s: %d\n", checkType, count)
	}

	// Step 2: Migration Mode
	prompter.PrintHeader("Step 2/5: Migration Mode")
	fmt.Fprintf(os.Stderr, "\n")

	mode, err := prompter.AskSelect(
		"Select migration mode:",
		[]string{
			"Full migration (create resources in Hyperping)",
			"Dry run (generate configs only)",
		},
		"Full migration (create resources in Hyperping)",
	)
	if err != nil {
		prompter.PrintError(fmt.Sprintf("Failed to select mode: %v", err))
		return 1
	}

	config.dryRun = mode == "Dry run (generate configs only)"

	// Step 3: Destination Platform Configuration
	if !config.dryRun {
		prompter.PrintHeader("Step 3/5: Destination Platform Configuration")
		fmt.Fprintf(os.Stderr, "\n")

		hyperpingKey, keyErr := prompter.AskPassword(
			"Enter your Hyperping API key:",
			"Get it from: https://app.hyperping.io/settings/api",
			interactive.HyperpingAPIKeyValidator,
		)
		if keyErr != nil {
			prompter.PrintError(fmt.Sprintf("Failed to get API key: %v", keyErr))
			return 1
		}
		config.hyperpingAPIKey = hyperpingKey
	}

	// Step 4: Output Configuration
	stepNum := 3
	if !config.dryRun {
		stepNum = 4
	}
	prompter.PrintHeader(fmt.Sprintf("Step %d/5: Output Configuration", stepNum))
	fmt.Fprintf(os.Stderr, "\n")

	outputDir, err := prompter.AskString(
		"Output directory for migration files:",
		"./pingdom-migration",
		"Directory where all migration files will be saved",
		interactive.FilePathValidator,
	)
	if err != nil {
		prompter.PrintError(fmt.Sprintf("Failed to get output directory: %v", err))
		return 1
	}
	config.outputDir = outputDir

	prefixValue, err := prompter.AskString(
		"Resource name prefix (optional):",
		"",
		"Prefix for Terraform resource names (e.g., 'pingdom_')",
		nil,
	)
	if err != nil {
		prompter.PrintError(fmt.Sprintf("Failed to get prefix: %v", err))
		return 1
	}
	config.prefix = prefixValue

	// Step 5: Migration Preview
	stepNum++
	prompter.PrintHeader(fmt.Sprintf("Step %d/5: Migration Preview", stepNum))
	fmt.Fprintf(os.Stderr, "\n")

	// Perform conversion to get counts
	checkConverter := converter.NewCheckConverter()
	results := make([]converter.ConversionResult, len(checks))
	supportedCount := 0
	for i, check := range checks {
		results[i] = checkConverter.Convert(check)
		if results[i].Supported {
			supportedCount++
		}
	}

	fmt.Fprintf(os.Stderr, "  📊 Summary:\n")
	fmt.Fprintf(os.Stderr, "    - Total checks: %d\n", len(checks))
	fmt.Fprintf(os.Stderr, "    - Supported checks: %d\n", supportedCount)
	fmt.Fprintf(os.Stderr, "    - Unsupported checks: %d\n", len(checks)-supportedCount)
	if config.dryRun {
		fmt.Fprintf(os.Stderr, "    - Mode: Dry run (configs only, no resources created)\n")
	} else {
		fmt.Fprintf(os.Stderr, "    - Mode: Full migration (resources will be created)\n")
	}
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  📁 Output directory: %s\n", config.outputDir)
	if config.prefix != "" {
		fmt.Fprintf(os.Stderr, "  🏷️  Resource prefix: %s\n", config.prefix)
	}
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  Files to be generated:\n")
	fmt.Fprintf(os.Stderr, "    - monitors.tf (Terraform configuration)\n")
	fmt.Fprintf(os.Stderr, "    - import.sh (Import script)\n")
	fmt.Fprintf(os.Stderr, "    - report.json (Detailed migration report)\n")
	fmt.Fprintf(os.Stderr, "    - report.txt (Human-readable report)\n")
	fmt.Fprintf(os.Stderr, "    - manual-steps.md (Manual configuration steps)\n")
	fmt.Fprintf(os.Stderr, "\n")

	if len(checks)-supportedCount > 0 {
		prompter.PrintWarning(fmt.Sprintf("%d unsupported checks will require manual migration", len(checks)-supportedCount))
		fmt.Fprintf(os.Stderr, "\n")
	}

	proceed, err := prompter.AskConfirm(
		"Proceed with migration?",
		true,
	)
	if err != nil {
		prompter.PrintError(fmt.Sprintf("Failed to get confirmation: %v", err))
		return 1
	}

	if !proceed {
		prompter.PrintInfo("Migration cancelled by user")
		return 0
	}

	// Step 6: Running Migration
	stepNum++
	prompter.PrintHeader(fmt.Sprintf("Step %d/5: Running Migration", stepNum))
	fmt.Fprintf(os.Stderr, "\n")

	// Create output directory
	if mkdirErr := os.MkdirAll(config.outputDir, 0o755); mkdirErr != nil {
		prompter.PrintError(fmt.Sprintf("Failed to create output directory: %v", mkdirErr))
		return 1
	}

	// Generate files
	progressBar := interactive.NewProgressBar(5, "Generating files", os.Stderr)

	// Terraform configuration
	tfGen := generator.NewTerraformGenerator(config.prefix)
	hclContent := tfGen.GenerateHCL(checks, results)
	hclPath := filepath.Join(config.outputDir, "monitors.tf")
	if writeErr := os.WriteFile(hclPath, []byte(hclContent), 0o600); writeErr != nil {
		prompter.PrintError(fmt.Sprintf("Failed to write Terraform config: %v", writeErr))
		return 1
	}
	//nolint:errcheck
	progressBar.Add(1)

	// Reports
	reporter := report.NewReporter()
	migrationReport := reporter.GenerateReport(checks, results)

	// JSON report
	jsonReport, err := reporter.GenerateJSONReport(migrationReport)
	if err != nil {
		prompter.PrintError(fmt.Sprintf("Failed to generate JSON report: %v", err))
		return 1
	}
	jsonPath := filepath.Join(config.outputDir, "report.json")
	if writeErr := os.WriteFile(jsonPath, []byte(jsonReport), 0o600); writeErr != nil {
		prompter.PrintError(fmt.Sprintf("Failed to write JSON report: %v", writeErr))
		return 1
	}
	//nolint:errcheck
	progressBar.Add(1)

	// Text report
	textReport := reporter.GenerateTextReport(migrationReport)
	textPath := filepath.Join(config.outputDir, "report.txt")
	if writeErr := os.WriteFile(textPath, []byte(textReport), 0o600); writeErr != nil {
		prompter.PrintError(fmt.Sprintf("Failed to write text report: %v", writeErr))
		return 1
	}
	//nolint:errcheck
	progressBar.Add(1)

	// Manual steps
	manualSteps := reporter.GenerateManualStepsMarkdown(migrationReport)
	manualPath := filepath.Join(config.outputDir, "manual-steps.md")
	if writeErr := os.WriteFile(manualPath, []byte(manualSteps), 0o600); writeErr != nil {
		prompter.PrintError(fmt.Sprintf("Failed to write manual steps: %v", writeErr))
		return 1
	}
	//nolint:errcheck
	progressBar.Add(1)

	// Create resources in Hyperping (if not dry run)
	createdResources := make(map[int]string)

	if !config.dryRun {
		createSpinner := interactive.NewSpinner("Creating monitors in Hyperping...", os.Stderr)
		createSpinner.Start()

		hyperpingClient := createHyperpingClient(config.hyperpingAPIKey)

		createdCount := 0
		errorCount := 0

		for i, check := range checks {
			result := results[i]

			if !result.Supported || result.Monitor == nil {
				continue
			}

			monitor, err := hyperpingClient.CreateMonitor(ctx, *result.Monitor)
			if err != nil {
				errorCount++
				if *verbose {
					fmt.Fprintf(os.Stderr, "\nWarning: Failed to create monitor for check %d (%s): %v\n", check.ID, check.Name, err)
				}
				continue
			}

			createdResources[check.ID] = monitor.UUID
			createdCount++
		}

		if errorCount > 0 {
			createSpinner.ErrorMessage(fmt.Sprintf("Created %d monitors with %d errors", createdCount, errorCount))
		} else {
			createSpinner.SuccessMessage(fmt.Sprintf("Created %d monitors in Hyperping", createdCount))
		}
	}
	//nolint:errcheck
	progressBar.Add(1)

	// Generate import script
	importGen := generator.NewImportGenerator(config.prefix)
	importScript := importGen.GenerateImportScript(checks, results, createdResources)
	importPath := filepath.Join(config.outputDir, "import.sh")
	if writeErr := os.WriteFile(importPath, []byte(importScript), 0o600); writeErr != nil {
		prompter.PrintError(fmt.Sprintf("Failed to write import script: %v", writeErr))
		return 1
	}

	//nolint:errcheck
	progressBar.Finish()

	// Final summary
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "✅ Migration complete!\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Generated files in %s:\n", config.outputDir)
	fmt.Fprintf(os.Stderr, "  📄 monitors.tf - Terraform configuration\n")
	fmt.Fprintf(os.Stderr, "  📜 import.sh - Import script\n")
	fmt.Fprintf(os.Stderr, "  📊 report.json - Detailed migration report\n")
	fmt.Fprintf(os.Stderr, "  📝 report.txt - Human-readable report\n")
	fmt.Fprintf(os.Stderr, "  📋 manual-steps.md - Manual configuration steps\n")
	fmt.Fprintf(os.Stderr, "\n")

	if len(migrationReport.ManualSteps) > 0 {
		prompter.PrintWarning(fmt.Sprintf("%d checks require manual steps - see manual-steps.md", len(migrationReport.ManualSteps)))
		fmt.Fprintf(os.Stderr, "\n")
	}

	if config.dryRun {
		fmt.Fprintf(os.Stderr, "Next steps:\n")
		fmt.Fprintf(os.Stderr, "  1. Review monitors.tf and adjust as needed\n")
		fmt.Fprintf(os.Stderr, "  2. Review manual-steps.md for unsupported checks\n")
		fmt.Fprintf(os.Stderr, "  3. Run without --dry-run to create resources in Hyperping\n")
	} else {
		fmt.Fprintf(os.Stderr, "Next steps:\n")
		fmt.Fprintf(os.Stderr, "  1. Review monitors.tf and adjust as needed\n")
		fmt.Fprintf(os.Stderr, "  2. Run: cd %s && terraform init\n", config.outputDir)
		fmt.Fprintf(os.Stderr, "  3. Run: terraform plan\n")
		fmt.Fprintf(os.Stderr, "  4. Run: ./import.sh to import resources into Terraform state\n")
		fmt.Fprintf(os.Stderr, "  5. Review manual-steps.md for unsupported checks\n")
	}

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Summary: %d total checks, %d supported, %d unsupported\n",
		migrationReport.TotalChecks,
		migrationReport.SupportedChecks,
		migrationReport.UnsupportedChecks)

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "📚 Documentation: https://github.com/develeap/terraform-provider-hyperping/tree/main/docs/guides\n")
	fmt.Fprintf(os.Stderr, "\n")

	return 0
}

// shouldUseInteractive determines if interactive mode should be used.
func shouldUseInteractive() bool {
	// Don't use interactive mode if any flags are set
	if isFlagPassed() {
		return false
	}

	// Check if we're in a TTY
	if !interactive.IsInteractive() {
		return false
	}

	return true
}

// isFlagPassed checks if any command-line flags were passed.
func isFlagPassed() bool {
	// Check if any relevant flags were explicitly set
	if *pingdomAPIKey != "" {
		return true
	}
	if *hyperpingAPIKey != "" {
		return true
	}
	if *outputDir != "./pingdom-migration" {
		return true
	}
	if *prefix != "" {
		return true
	}
	if *pingdomBaseURL != "" {
		return true
	}
	if *hyperpingBaseURL != "https://api.hyperping.io" {
		return true
	}
	if *dryRun {
		return true
	}
	if *verbose {
		return true
	}

	// Also check environment variables
	if os.Getenv("PINGDOM_API_KEY") != "" || os.Getenv("PINGDOM_API_TOKEN") != "" {
		return true
	}
	if os.Getenv("HYPERPING_API_KEY") != "" {
		return true
	}

	return false
}
