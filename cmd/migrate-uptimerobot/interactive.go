// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/converter"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/generator"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/report"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/uptimerobot"
	"github.com/develeap/terraform-provider-hyperping/pkg/interactive"
)

// interactiveConfigUR holds configuration collected from interactive prompts.
type interactiveConfigUR struct {
	uptimerobotAPIKey string
	hyperpingAPIKey   string
	outputFile        string
	importScript      string
	reportFile        string
	manualStepsFile   string
	dryRun            bool
	validate          bool
}

// runInteractive runs the migration tool in interactive mode.
func runInteractive() int {
	prompter := interactive.NewPrompter(interactive.DefaultConfig())

	// Welcome message
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "🚀 Hyperping Migration Tool - UptimeRobot Edition\n")
	fmt.Fprintf(os.Stderr, "═════════════════════════════════════════════════\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "This wizard will guide you through migrating your UptimeRobot\n")
	fmt.Fprintf(os.Stderr, "monitors to Hyperping.\n")
	fmt.Fprintf(os.Stderr, "\n")

	config := &interactiveConfigUR{}

	// Step 1: Source Platform Configuration
	prompter.PrintHeader("Step 1/5: Source Platform Configuration")
	fmt.Fprintf(os.Stderr, "\n")

	apiKey, err := prompter.AskPassword(
		"Enter your UptimeRobot API key:",
		"Get it from: https://uptimerobot.com/dashboard#mySettings",
		interactive.SourceAPIKeyValidator("uptimerobot"),
	)
	if err != nil {
		prompter.PrintError(fmt.Sprintf("Failed to get API key: %v", err))
		return 1
	}
	config.uptimerobotAPIKey = apiKey

	// Test UptimeRobot API connection
	spinner := interactive.NewSpinner("Testing UptimeRobot API connection...", os.Stderr)
	spinner.Start()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	urClient := uptimerobot.NewClient(config.uptimerobotAPIKey)
	monitors, err := urClient.GetMonitors(ctx)
	if err != nil {
		spinner.ErrorMessage(fmt.Sprintf("Connection failed: %v", err))
		prompter.PrintError("Unable to connect to UptimeRobot API")
		prompter.PrintInfo("Please verify your API key and try again")
		return 1
	}

	alertContacts, err := urClient.GetAlertContacts(ctx)
	if err != nil {
		spinner.ErrorMessage(fmt.Sprintf("Failed to fetch alert contacts: %v", err))
		prompter.PrintWarning("Unable to fetch alert contacts, continuing with monitors only")
		alertContacts = []uptimerobot.AlertContact{}
	}

	spinner.SuccessMessage(fmt.Sprintf("Connected! Found %d monitors and %d alert contacts", len(monitors), len(alertContacts)))

	// Show monitor type breakdown
	fmt.Fprintf(os.Stderr, "\n")
	typeCounts := make(map[int]int)
	for _, m := range monitors {
		typeCounts[m.Type]++
	}
	typeNames := map[int]string{
		1: "HTTP/HTTPS",
		2: "Keyword",
		3: "Ping (ICMP)",
		4: "Port",
		5: "Heartbeat",
	}
	fmt.Fprintf(os.Stderr, "  Monitor types:\n")
	for typeID, count := range typeCounts {
		name := typeNames[typeID]
		if name == "" {
			name = fmt.Sprintf("Unknown (type %d)", typeID)
		}
		fmt.Fprintf(os.Stderr, "    - %s: %d\n", name, count)
	}

	// Step 2: Mode Selection
	prompter.PrintHeader("Step 2/5: Migration Mode")
	fmt.Fprintf(os.Stderr, "\n")

	mode, err := prompter.AskSelect(
		"Select migration mode:",
		[]string{
			"Full migration (generate files and proceed)",
			"Dry run (validate without creating files)",
			"Validate only (check monitors)",
		},
		"Full migration (generate files and proceed)",
	)
	if err != nil {
		prompter.PrintError(fmt.Sprintf("Failed to select mode: %v", err))
		return 1
	}

	switch mode {
	case "Validate only (check monitors)":
		config.validate = true
		return runValidationInteractive(monitors, alertContacts, prompter)
	case "Dry run (validate without creating files)":
		config.dryRun = true
	}

	// Step 3: Destination Platform Configuration
	if !config.dryRun && !config.validate {
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
	prompter.PrintHeader(fmt.Sprintf("Step %d/5: Output Configuration", getStepNumber(config)))
	fmt.Fprintf(os.Stderr, "\n")

	outputFile, err := prompter.AskString(
		"Terraform output file:",
		"hyperping.tf",
		"Main Terraform configuration file",
		interactive.FilePathValidator,
	)
	if err != nil {
		prompter.PrintError(fmt.Sprintf("Failed to get output file: %v", err))
		return 1
	}
	config.outputFile = outputFile

	importScriptFile, err := prompter.AskString(
		"Import script file:",
		"import.sh",
		"Shell script to import existing resources",
		interactive.FilePathValidator,
	)
	if err != nil {
		prompter.PrintError(fmt.Sprintf("Failed to get import script file: %v", err))
		return 1
	}
	config.importScript = importScriptFile

	reportFileValue, err := prompter.AskString(
		"Migration report file:",
		"migration-report.json",
		"Detailed JSON report of the migration",
		interactive.FilePathValidator,
	)
	if err != nil {
		prompter.PrintError(fmt.Sprintf("Failed to get report file: %v", err))
		return 1
	}
	config.reportFile = reportFileValue

	manualStepsFileValue, err := prompter.AskString(
		"Manual steps file:",
		"manual-steps.md",
		"Documentation of manual configuration steps",
		interactive.FilePathValidator,
	)
	if err != nil {
		prompter.PrintError(fmt.Sprintf("Failed to get manual steps file: %v", err))
		return 1
	}
	config.manualStepsFile = manualStepsFileValue

	// Step 5: Migration Preview
	prompter.PrintHeader(fmt.Sprintf("Step %d/5: Migration Preview", getStepNumber(config)+1))
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  📊 Summary:\n")
	fmt.Fprintf(os.Stderr, "    - Total monitors: %d\n", len(monitors))
	fmt.Fprintf(os.Stderr, "    - Alert contacts: %d\n", len(alertContacts))
	if config.dryRun {
		fmt.Fprintf(os.Stderr, "    - Mode: Dry run (no files will be created)\n")
	} else {
		fmt.Fprintf(os.Stderr, "    - Mode: Full migration\n")
	}
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  📁 Output files:\n")
	fmt.Fprintf(os.Stderr, "    - %s (Terraform configuration)\n", config.outputFile)
	fmt.Fprintf(os.Stderr, "    - %s (Import script)\n", config.importScript)
	fmt.Fprintf(os.Stderr, "    - %s (Migration report)\n", config.reportFile)
	fmt.Fprintf(os.Stderr, "    - %s (Manual steps)\n", config.manualStepsFile)
	fmt.Fprintf(os.Stderr, "\n")

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
	prompter.PrintHeader(fmt.Sprintf("Step %d/5: Running Migration", getStepNumber(config)+2))
	fmt.Fprintf(os.Stderr, "\n")

	// Convert monitors
	conversionSpinner := interactive.NewSpinner("Converting monitors to Hyperping format...", os.Stderr)
	conversionSpinner.Start()

	conv := converter.NewConverter()
	conversionResult := conv.Convert(monitors, alertContacts)

	conversionSpinner.SuccessMessage("Conversion complete")

	// Generate migration report
	migrationReport := report.Generate(monitors, alertContacts, conversionResult)

	// Show summary
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  Migration Summary:\n")
	fmt.Fprintf(os.Stderr, "    - Total monitors: %d\n", migrationReport.Summary.TotalMonitors)
	fmt.Fprintf(os.Stderr, "    - Migrated monitors: %d\n", migrationReport.Summary.MigratedMonitors)
	fmt.Fprintf(os.Stderr, "    - Migrated healthchecks: %d\n", migrationReport.Summary.MigratedHealthchecks)
	fmt.Fprintf(os.Stderr, "    - Warnings: %d\n", len(migrationReport.Warnings))
	fmt.Fprintf(os.Stderr, "    - Errors: %d\n", len(migrationReport.Errors))
	fmt.Fprintf(os.Stderr, "\n")

	if len(migrationReport.Errors) > 0 {
		prompter.PrintWarning("Errors encountered during conversion:")
		for i, e := range migrationReport.Errors {
			if i >= 5 {
				fmt.Fprintf(os.Stderr, "    ... and %d more (see report for details)\n", len(migrationReport.Errors)-5)
				break
			}
			fmt.Fprintf(os.Stderr, "    - %s: %s\n", e.Resource, e.Message)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	if len(migrationReport.Warnings) > 0 {
		prompter.PrintWarning("Warnings:")
		for i, w := range migrationReport.Warnings {
			if i >= 5 {
				fmt.Fprintf(os.Stderr, "    ... and %d more (see report for details)\n", len(migrationReport.Warnings)-5)
				break
			}
			fmt.Fprintf(os.Stderr, "    - %s: %s\n", w.Resource, w.Message)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	// Dry run mode - just validate
	if config.dryRun {
		fmt.Fprintf(os.Stderr, "=== DRY RUN MODE ===\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "No files were created. Migration preview complete.\n")
		return 0
	}

	// Write files
	fileSpinner := interactive.NewSpinner("Writing output files...", os.Stderr)
	fileSpinner.Start()

	// Generate Terraform configuration
	tfConfig := generator.GenerateTerraform(conversionResult)
	if writeErr := os.WriteFile(config.outputFile, []byte(tfConfig), 0o600); writeErr != nil {
		fileSpinner.ErrorMessage(fmt.Sprintf("Failed to write %s", config.outputFile))
		prompter.PrintError(fmt.Sprintf("Error: %v", writeErr))
		return 1
	}

	// Generate import script
	importScriptContent := generator.GenerateImportScript(conversionResult)
	if writeErr := os.WriteFile(config.importScript, []byte(importScriptContent), 0o600); writeErr != nil {
		fileSpinner.ErrorMessage(fmt.Sprintf("Failed to write %s", config.importScript))
		prompter.PrintError(fmt.Sprintf("Error: %v", writeErr))
		return 1
	}

	// Generate migration report
	reportJSON, err := json.MarshalIndent(migrationReport, "", "  ")
	if err != nil {
		fileSpinner.ErrorMessage("Failed to marshal report")
		prompter.PrintError(fmt.Sprintf("Error: %v", err))
		return 1
	}
	if err := os.WriteFile(config.reportFile, reportJSON, 0o600); err != nil {
		fileSpinner.ErrorMessage(fmt.Sprintf("Failed to write %s", config.reportFile))
		prompter.PrintError(fmt.Sprintf("Error: %v", err))
		return 1
	}

	// Generate manual steps
	manualStepsContent := generator.GenerateManualSteps(conversionResult, alertContacts)
	if err := os.WriteFile(config.manualStepsFile, []byte(manualStepsContent), 0o600); err != nil {
		fileSpinner.ErrorMessage(fmt.Sprintf("Failed to write %s", config.manualStepsFile))
		prompter.PrintError(fmt.Sprintf("Error: %v", err))
		return 1
	}

	fileSpinner.SuccessMessage("All files written successfully")

	// Final summary
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "✅ Migration complete!\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Generated files:\n")
	fmt.Fprintf(os.Stderr, "  📄 %s - Terraform configuration\n", config.outputFile)
	fmt.Fprintf(os.Stderr, "  📜 %s - Import script\n", config.importScript)
	fmt.Fprintf(os.Stderr, "  📊 %s - Migration report\n", config.reportFile)
	fmt.Fprintf(os.Stderr, "  📝 %s - Manual configuration steps\n", config.manualStepsFile)
	fmt.Fprintf(os.Stderr, "\n")

	if len(migrationReport.Warnings) > 0 || len(migrationReport.Errors) > 0 {
		prompter.PrintWarning(fmt.Sprintf("%d warnings/errors - see %s for details", len(migrationReport.Warnings)+len(migrationReport.Errors), config.reportFile))
		fmt.Fprintf(os.Stderr, "\n")
	}

	fmt.Fprintf(os.Stderr, "Next steps:\n")
	fmt.Fprintf(os.Stderr, "  1. Review %s and adjust as needed\n", config.outputFile)
	fmt.Fprintf(os.Stderr, "  2. Run: terraform init && terraform plan\n")
	fmt.Fprintf(os.Stderr, "  3. Run: terraform apply\n")
	fmt.Fprintf(os.Stderr, "  4. Review %s for manual configuration steps\n", config.manualStepsFile)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "📚 Documentation: https://github.com/develeap/terraform-provider-hyperping/tree/main/docs/guides\n")
	fmt.Fprintf(os.Stderr, "\n")

	return 0
}

// runValidationInteractive runs validation mode interactively.
func runValidationInteractive(monitors []uptimerobot.Monitor, alertContacts []uptimerobot.AlertContact, prompter *interactive.Prompter) int {
	prompter.PrintHeader("Validation Results")
	fmt.Fprintf(os.Stderr, "\n")

	// Count by type
	typeCounts := make(map[int]int)
	for _, m := range monitors {
		typeCounts[m.Type]++
	}

	typeNames := map[int]string{
		1: "HTTP/HTTPS",
		2: "Keyword",
		3: "Ping (ICMP)",
		4: "Port",
		5: "Heartbeat",
	}

	fmt.Fprintf(os.Stderr, "Monitor Types:\n")
	for typeID, count := range typeCounts {
		name := typeNames[typeID]
		if name == "" {
			name = fmt.Sprintf("Unknown (type %d)", typeID)
		}
		fmt.Fprintf(os.Stderr, "  %s: %d\n", name, count)
	}

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Total monitors: %d\n", len(monitors))
	fmt.Fprintf(os.Stderr, "Alert contacts: %d\n", len(alertContacts))
	fmt.Fprintf(os.Stderr, "\n")

	prompter.PrintSuccess("Validation complete")
	return 0
}

// getStepNumber returns the current step number based on configuration.
func getStepNumber(config *interactiveConfigUR) int {
	if config.dryRun || config.validate {
		return 3
	}
	return 4
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
	if *uptimerobotAPIKey != "" {
		return true
	}
	if *hyperpingAPIKey != "" {
		return true
	}
	if *output != "hyperping.tf" {
		return true
	}
	if *importScript != "import.sh" {
		return true
	}
	if *reportFile != "migration-report.json" {
		return true
	}
	if *manualSteps != "manual-steps.md" {
		return true
	}
	if *dryRun {
		return true
	}
	if *validate {
		return true
	}
	if *verbose {
		return true
	}

	// Also check environment variables
	if os.Getenv("UPTIMEROBOT_API_KEY") != "" {
		return true
	}
	if os.Getenv("HYPERPING_API_KEY") != "" {
		return true
	}

	return false
}
