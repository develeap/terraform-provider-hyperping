// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack/betterstack"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack/converter"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack/generator"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack/report"
	"github.com/develeap/terraform-provider-hyperping/pkg/interactive"
	"github.com/develeap/terraform-provider-hyperping/pkg/recovery"
)

// interactiveConfig holds configuration collected from interactive prompts.
type interactiveConfig struct {
	betterstackToken string
	outputFile       string
	importScript     string
	reportFile       string
	manualStepsFile  string
	dryRun           bool
}

// runInteractive runs the migration tool in interactive mode.
func runInteractive(_ *recovery.Logger) int {
	prompter := interactive.NewPrompter(interactive.DefaultConfig())

	// Welcome message
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "🚀 Hyperping Migration Tool - Better Stack Edition\n")
	fmt.Fprintf(os.Stderr, "═══════════════════════════════════════════════════\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "This wizard will guide you through migrating your Better Stack\n")
	fmt.Fprintf(os.Stderr, "monitors to Hyperping.\n")
	fmt.Fprintf(os.Stderr, "\n")

	config := &interactiveConfig{}

	// Step 1: Source Platform Configuration
	prompter.PrintHeader("Step 1/5: Source Platform Configuration")
	fmt.Fprintf(os.Stderr, "\n")

	token, err := prompter.AskPassword(
		"Enter your Better Stack API token:",
		"Get it from: https://betterstack.com/team/api-tokens",
		interactive.SourceAPIKeyValidator("betterstack"),
	)
	if err != nil {
		prompter.PrintError(fmt.Sprintf("Failed to get API token: %v", err))
		return 1
	}
	config.betterstackToken = token

	// Test Better Stack API connection
	spinner := interactive.NewSpinner("Testing Better Stack API connection...", os.Stderr)
	spinner.Start()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bsClient := betterstack.NewClient(config.betterstackToken)
	monitors, err := bsClient.FetchMonitors(ctx)
	if err != nil {
		spinner.ErrorMessage(fmt.Sprintf("Connection failed: %v", err))
		prompter.PrintError("Unable to connect to Better Stack API")
		prompter.PrintInfo("Please verify your API token and try again")
		return 1
	}

	heartbeats, err := bsClient.FetchHeartbeats(ctx)
	if err != nil {
		spinner.ErrorMessage(fmt.Sprintf("Failed to fetch heartbeats: %v", err))
		prompter.PrintWarning("Unable to fetch heartbeats, continuing with monitors only")
		heartbeats = []betterstack.Heartbeat{}
	}

	totalResources := len(monitors) + len(heartbeats)
	spinner.SuccessMessage(fmt.Sprintf("Connected! Found %d monitors and %d heartbeats", len(monitors), len(heartbeats)))

	// Step 2: Destination Platform Configuration
	prompter.PrintHeader("Step 2/5: Destination Platform Configuration")
	fmt.Fprintf(os.Stderr, "\n")

	// Ask if this is a dry run
	dryRun, err := prompter.AskConfirm(
		"Perform dry run only (validate without creating files)?",
		false,
	)
	if err != nil {
		prompter.PrintError(fmt.Sprintf("Failed to get confirmation: %v", err))
		return 1
	}
	config.dryRun = dryRun

	if !dryRun {
		hyperpingKey, keyErr := prompter.AskPassword(
			"Enter your Hyperping API key:",
			"Get it from: https://app.hyperping.io/settings/api",
			interactive.HyperpingAPIKeyValidator,
		)
		if keyErr != nil {
			prompter.PrintError(fmt.Sprintf("Failed to get API key: %v", keyErr))
			return 1
		}
		_ = hyperpingKey
	}

	// Step 3: Output Configuration
	prompter.PrintHeader("Step 3/5: Output Configuration")
	fmt.Fprintf(os.Stderr, "\n")

	outputFile, err := prompter.AskString(
		"Terraform output file:",
		"migrated-resources.tf",
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

	// Step 4: Migration Preview
	prompter.PrintHeader("Step 4/5: Migration Preview")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  📊 Summary:\n")
	fmt.Fprintf(os.Stderr, "    - Total monitors: %d\n", len(monitors))
	fmt.Fprintf(os.Stderr, "    - Total heartbeats: %d\n", len(heartbeats))
	fmt.Fprintf(os.Stderr, "    - Total resources: %d\n", totalResources)
	if dryRun {
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

	// Step 5: Running Migration
	prompter.PrintHeader("Step 5/5: Running Migration")
	fmt.Fprintf(os.Stderr, "\n")

	// Convert monitors
	conv := converter.New()

	progressBar := interactive.NewProgressBar(int64(len(monitors)), "Converting monitors", os.Stderr)
	var convertedMonitors []converter.ConvertedMonitor
	var monitorIssues []converter.ConversionIssue

	for _, monitor := range monitors {
		converted, issues := conv.ConvertMonitors([]betterstack.Monitor{monitor})
		convertedMonitors = append(convertedMonitors, converted...)
		monitorIssues = append(monitorIssues, issues...)
		//nolint:errcheck
		progressBar.Add(1)
	}
	//nolint:errcheck
	progressBar.Finish()

	// Convert heartbeats
	progressBar = interactive.NewProgressBar(int64(len(heartbeats)), "Converting heartbeats", os.Stderr)
	var convertedHealthchecks []converter.ConvertedHealthcheck
	var healthcheckIssues []converter.ConversionIssue

	for _, heartbeat := range heartbeats {
		converted, issues := conv.ConvertHeartbeats([]betterstack.Heartbeat{heartbeat})
		convertedHealthchecks = append(convertedHealthchecks, converted...)
		healthcheckIssues = append(healthcheckIssues, issues...)
		//nolint:errcheck
		progressBar.Add(1)
	}
	//nolint:errcheck
	progressBar.Finish()

	// Generate files
	gen := generator.New()
	tfConfig := gen.GenerateTerraform(convertedMonitors, convertedHealthchecks)
	importScriptContent := gen.GenerateImportScript(convertedMonitors, convertedHealthchecks)
	manualSteps := gen.GenerateManualSteps(monitorIssues, healthcheckIssues)
	migrationReport := report.Generate(
		monitors,
		heartbeats,
		convertedMonitors,
		convertedHealthchecks,
		monitorIssues,
		healthcheckIssues,
	)

	// Dry run mode - just validate
	if config.dryRun {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "=== DRY RUN MODE ===\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Would create:\n")
		fmt.Fprintf(os.Stderr, "  - %s (%d bytes)\n", config.outputFile, len(tfConfig))
		fmt.Fprintf(os.Stderr, "  - %s (%d bytes)\n", config.importScript, len(importScriptContent))
		fmt.Fprintf(os.Stderr, "  - %s (%d bytes)\n", config.reportFile, len(migrationReport.JSON()))
		fmt.Fprintf(os.Stderr, "  - %s (%d bytes)\n", config.manualStepsFile, len(manualSteps))
		fmt.Fprintf(os.Stderr, "\n")
		migrationReport.PrintSummary(os.Stderr)
		return 0
	}

	// Write files
	fileSpinner := interactive.NewSpinner("Writing output files...", os.Stderr)
	fileSpinner.Start()

	if err := os.WriteFile(config.outputFile, []byte(tfConfig), 0o600); err != nil {
		fileSpinner.ErrorMessage(fmt.Sprintf("Failed to write %s", config.outputFile))
		prompter.PrintError(fmt.Sprintf("Error: %v", err))
		return 1
	}

	if err := os.WriteFile(config.importScript, []byte(importScriptContent), 0o600); err != nil {
		fileSpinner.ErrorMessage(fmt.Sprintf("Failed to write %s", config.importScript))
		prompter.PrintError(fmt.Sprintf("Error: %v", err))
		return 1
	}

	if err := os.WriteFile(config.reportFile, []byte(migrationReport.JSON()), 0o600); err != nil {
		fileSpinner.ErrorMessage(fmt.Sprintf("Failed to write %s", config.reportFile))
		prompter.PrintError(fmt.Sprintf("Error: %v", err))
		return 1
	}

	if err := os.WriteFile(config.manualStepsFile, []byte(manualSteps), 0o600); err != nil {
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
	fmt.Fprintf(os.Stderr, "  📄 %s - Terraform configuration (%d lines)\n", config.outputFile, len(convertedMonitors)+len(convertedHealthchecks))
	fmt.Fprintf(os.Stderr, "  📜 %s - Import script\n", config.importScript)
	fmt.Fprintf(os.Stderr, "  📊 %s - Migration report\n", config.reportFile)
	fmt.Fprintf(os.Stderr, "  📝 %s - Manual configuration steps\n", config.manualStepsFile)
	fmt.Fprintf(os.Stderr, "\n")

	if len(monitorIssues) > 0 || len(healthcheckIssues) > 0 {
		prompter.PrintWarning(fmt.Sprintf("%d warnings - see %s for details", len(monitorIssues)+len(healthcheckIssues), config.reportFile))
		fmt.Fprintf(os.Stderr, "\n")
	}

	fmt.Fprintf(os.Stderr, "Next steps:\n")
	fmt.Fprintf(os.Stderr, "  1. Review %s and adjust as needed\n", config.outputFile)
	fmt.Fprintf(os.Stderr, "  2. Review %s for manual configuration steps\n", config.manualStepsFile)
	fmt.Fprintf(os.Stderr, "  3. Run: terraform init && terraform plan\n")
	fmt.Fprintf(os.Stderr, "  4. Run: terraform apply\n")
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
	if *betterstackToken != "" {
		return true
	}
	if *hyperpingAPIKey != "" {
		return true
	}
	if *outputFile != "migrated-resources.tf" {
		return true
	}
	if *importScript != "import.sh" {
		return true
	}
	if *reportFile != "migration-report.json" {
		return true
	}
	if *manualStepsFile != "manual-steps.md" {
		return true
	}
	if *dryRun {
		return true
	}
	if *validateTF {
		return true
	}
	if *verbose {
		return true
	}
	if *debug {
		return true
	}
	if *resume {
		return true
	}
	if *resumeID != "" {
		return true
	}
	if *rollback {
		return true
	}
	if *rollbackID != "" {
		return true
	}
	if *rollbackForce {
		return true
	}
	if *listCheckpointsFlag {
		return true
	}

	// Also check environment variables
	if os.Getenv("BETTERSTACK_API_TOKEN") != "" {
		return true
	}
	if os.Getenv("HYPERPING_API_KEY") != "" {
		return true
	}

	return false
}
