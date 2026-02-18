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
	hyperpingAPIKey  string
	outputFile       string
	importScript     string
	reportFile       string
	manualStepsFile  string
	dryRun           bool
}

// promptSourceCredentials asks the user for their Better Stack token and tests connectivity.
// Returns the fetched monitors and heartbeats on success.
func promptSourceCredentials(prompter *interactive.Prompter) (string, []betterstack.Monitor, []betterstack.Heartbeat, error) {
	prompter.PrintHeader("Step 1/5: Source Platform Configuration")
	fmt.Fprintf(os.Stderr, "\n")

	token, err := prompter.AskPassword(
		"Enter your Better Stack API token:",
		"Get it from: https://betterstack.com/team/api-tokens",
		interactive.SourceAPIKeyValidator("betterstack"),
	)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to get API token: %w", err)
	}

	spinner := interactive.NewSpinner("Testing Better Stack API connection...", os.Stderr)
	spinner.Start()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bsClient := betterstack.NewClient(token)
	monitors, err := bsClient.FetchMonitors(ctx)
	if err != nil {
		spinner.ErrorMessage(fmt.Sprintf("Connection failed: %v", err))
		return "", nil, nil, fmt.Errorf("unable to connect to Better Stack API")
	}

	heartbeats, err := bsClient.FetchHeartbeats(ctx)
	if err != nil {
		spinner.ErrorMessage(fmt.Sprintf("Failed to fetch heartbeats: %v", err))
		prompter.PrintWarning("Unable to fetch heartbeats, continuing with monitors only")
		heartbeats = []betterstack.Heartbeat{}
	}

	spinner.SuccessMessage(fmt.Sprintf("Connected! Found %d monitors and %d heartbeats", len(monitors), len(heartbeats)))
	return token, monitors, heartbeats, nil
}

// promptDestinationConfig asks about dry-run mode and optional Hyperping API key.
// Returns (hyperpingAPIKey, dryRun, error).
func promptDestinationConfig(prompter *interactive.Prompter) (string, bool, error) {
	prompter.PrintHeader("Step 2/5: Destination Platform Configuration")
	fmt.Fprintf(os.Stderr, "\n")

	dryRunMode, err := prompter.AskConfirm(
		"Perform dry run only (validate without creating files)?",
		false,
	)
	if err != nil {
		return "", false, fmt.Errorf("failed to get confirmation: %w", err)
	}

	if !dryRunMode {
		hyperpingKey, keyErr := prompter.AskPassword(
			"Enter your Hyperping API key:",
			"Get it from: https://app.hyperping.io/settings/api",
			interactive.HyperpingAPIKeyValidator,
		)
		if keyErr != nil {
			return "", false, fmt.Errorf("failed to get API key: %w", keyErr)
		}
		return hyperpingKey, dryRunMode, nil
	}

	return "", dryRunMode, nil
}

// promptOutputConfig asks for the four output file paths.
func promptOutputConfig(prompter *interactive.Prompter) (outputF, importF, reportF, manualF string, err error) {
	prompter.PrintHeader("Step 3/5: Output Configuration")
	fmt.Fprintf(os.Stderr, "\n")

	outputF, err = prompter.AskString("Terraform output file:", "migrated-resources.tf", "Main Terraform configuration file", interactive.FilePathValidator)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to get output file: %w", err)
	}

	importF, err = prompter.AskString("Import script file:", "import.sh", "Shell script to import existing resources", interactive.FilePathValidator)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to get import script file: %w", err)
	}

	reportF, err = prompter.AskString("Migration report file:", "migration-report.json", "Detailed JSON report of the migration", interactive.FilePathValidator)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to get report file: %w", err)
	}

	manualF, err = prompter.AskString("Manual steps file:", "manual-steps.md", "Documentation of manual configuration steps", interactive.FilePathValidator)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to get manual steps file: %w", err)
	}

	return outputF, importF, reportF, manualF, nil
}

// promptMigrationPreview shows a summary and asks the user to confirm.
func promptMigrationPreview(prompter *interactive.Prompter, config *interactiveConfig, monitors []betterstack.Monitor, heartbeats []betterstack.Heartbeat) (bool, error) {
	prompter.PrintHeader("Step 4/5: Migration Preview")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  Summary:\n")
	fmt.Fprintf(os.Stderr, "    - Total monitors: %d\n", len(monitors))
	fmt.Fprintf(os.Stderr, "    - Total heartbeats: %d\n", len(heartbeats))
	fmt.Fprintf(os.Stderr, "    - Total resources: %d\n", len(monitors)+len(heartbeats))
	if config.dryRun {
		fmt.Fprintf(os.Stderr, "    - Mode: Dry run (no files will be created)\n")
	} else {
		fmt.Fprintf(os.Stderr, "    - Mode: Full migration\n")
	}
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  Output files:\n")
	fmt.Fprintf(os.Stderr, "    - %s (Terraform configuration)\n", config.outputFile)
	fmt.Fprintf(os.Stderr, "    - %s (Import script)\n", config.importScript)
	fmt.Fprintf(os.Stderr, "    - %s (Migration report)\n", config.reportFile)
	fmt.Fprintf(os.Stderr, "    - %s (Manual steps)\n", config.manualStepsFile)
	fmt.Fprintf(os.Stderr, "\n")

	proceed, err := prompter.AskConfirm("Proceed with migration?", true)
	if err != nil {
		return false, fmt.Errorf("failed to get confirmation: %w", err)
	}
	return proceed, nil
}

// runInteractiveConversion converts resources and generates output content.
func runInteractiveConversion(
	monitors []betterstack.Monitor,
	heartbeats []betterstack.Heartbeat,
) ([]converter.ConvertedMonitor, []converter.ConvertedHealthcheck, []converter.ConversionIssue, []converter.ConversionIssue) {
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

	return convertedMonitors, convertedHealthchecks, monitorIssues, healthcheckIssues
}

// printInteractiveDryRun shows the dry-run output for interactive mode.
func printInteractiveDryRun(
	config *interactiveConfig,
	tfConfig, importScriptContent, manualSteps string,
	migrationReport *report.Report,
) {
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
}

// writeInteractiveFiles writes all generated files with a spinner.
func writeInteractiveFiles(
	config *interactiveConfig,
	prompter *interactive.Prompter,
	tfConfig, importScriptContent, manualSteps string,
	migrationReport *report.Report,
) int {
	fileSpinner := interactive.NewSpinner("Writing output files...", os.Stderr)
	fileSpinner.Start()

	type fileWrite struct {
		path    string
		content []byte
	}

	writes := []fileWrite{
		{config.outputFile, []byte(tfConfig)},
		{config.importScript, []byte(importScriptContent)},
		{config.reportFile, []byte(migrationReport.JSON())},
		{config.manualStepsFile, []byte(manualSteps)},
	}

	for _, w := range writes {
		if err := os.WriteFile(w.path, w.content, 0o600); err != nil {
			fileSpinner.ErrorMessage(fmt.Sprintf("Failed to write %s", w.path))
			prompter.PrintError(fmt.Sprintf("Error: %v", err))
			return 1
		}
	}

	fileSpinner.SuccessMessage("All files written successfully")
	return 0
}

// printInteractiveSummary prints the final summary after a successful interactive migration.
func printInteractiveSummary(
	config *interactiveConfig,
	prompter *interactive.Prompter,
	convertedMonitors []converter.ConvertedMonitor,
	convertedHealthchecks []converter.ConvertedHealthcheck,
	monitorIssues, healthcheckIssues []converter.ConversionIssue,
) {
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Migration complete!\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Generated files:\n")
	fmt.Fprintf(os.Stderr, "  %s - Terraform configuration (%d lines)\n", config.outputFile, len(convertedMonitors)+len(convertedHealthchecks))
	fmt.Fprintf(os.Stderr, "  %s - Import script\n", config.importScript)
	fmt.Fprintf(os.Stderr, "  %s - Migration report\n", config.reportFile)
	fmt.Fprintf(os.Stderr, "  %s - Manual configuration steps\n", config.manualStepsFile)
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
	fmt.Fprintf(os.Stderr, "Documentation: https://github.com/develeap/terraform-provider-hyperping/tree/main/docs/guides\n")
	fmt.Fprintf(os.Stderr, "\n")
}

// runInteractive runs the migration tool in interactive mode.
func runInteractive(_ *recovery.Logger) int {
	prompter := interactive.NewPrompter(interactive.DefaultConfig())

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Hyperping Migration Tool - Better Stack Edition\n")
	fmt.Fprintf(os.Stderr, "===================================================\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "This wizard will guide you through migrating your Better Stack\n")
	fmt.Fprintf(os.Stderr, "monitors to Hyperping.\n")
	fmt.Fprintf(os.Stderr, "\n")

	token, monitors, heartbeats, err := promptSourceCredentials(prompter)
	if err != nil {
		prompter.PrintError(err.Error())
		return 1
	}

	config := &interactiveConfig{betterstackToken: token}

	hyperpingKey, dryRunMode, err := promptDestinationConfig(prompter)
	if err != nil {
		prompter.PrintError(err.Error())
		return 1
	}
	config.dryRun = dryRunMode
	config.hyperpingAPIKey = hyperpingKey

	config.outputFile, config.importScript, config.reportFile, config.manualStepsFile, err = promptOutputConfig(prompter)
	if err != nil {
		prompter.PrintError(err.Error())
		return 1
	}

	proceed, err := promptMigrationPreview(prompter, config, monitors, heartbeats)
	if err != nil {
		prompter.PrintError(err.Error())
		return 1
	}
	if !proceed {
		prompter.PrintInfo("Migration cancelled by user")
		return 0
	}

	prompter.PrintHeader("Step 5/5: Running Migration")
	fmt.Fprintf(os.Stderr, "\n")

	convertedMonitors, convertedHealthchecks, monitorIssues, healthcheckIssues := runInteractiveConversion(monitors, heartbeats)

	gen := generator.New()
	tfConfig := gen.GenerateTerraform(convertedMonitors, convertedHealthchecks)
	importScriptContent := gen.GenerateImportScript(convertedMonitors, convertedHealthchecks)
	manualSteps := gen.GenerateManualSteps(monitorIssues, healthcheckIssues)
	migrationReport := report.Generate(monitors, heartbeats, convertedMonitors, convertedHealthchecks, monitorIssues, healthcheckIssues)

	if config.dryRun {
		printInteractiveDryRun(config, tfConfig, importScriptContent, manualSteps, migrationReport)
		return 0
	}

	if code := writeInteractiveFiles(config, prompter, tfConfig, importScriptContent, manualSteps, migrationReport); code != 0 {
		return code
	}

	printInteractiveSummary(config, prompter, convertedMonitors, convertedHealthchecks, monitorIssues, healthcheckIssues)
	return 0
}

// shouldUseInteractive determines if interactive mode should be used.
func shouldUseInteractive() bool {
	if isFlagPassed() {
		return false
	}
	if !interactive.IsInteractive() {
		return false
	}
	return true
}

// isFlagPassed checks if any command-line flags were passed.
func isFlagPassed() bool {
	stringChecks := []struct {
		val  *string
		dflt string
	}{
		{betterstackToken, ""},
		{hyperpingAPIKey, ""},
		{outputFile, "migrated-resources.tf"},
		{importScript, "import.sh"},
		{reportFile, "migration-report.json"},
		{manualStepsFile, "manual-steps.md"},
		{resumeID, ""},
		{rollbackID, ""},
	}

	for _, c := range stringChecks {
		if *c.val != c.dflt {
			return true
		}
	}

	boolChecks := []*bool{
		dryRun, validateTF, verbose, debug, resume, rollback, rollbackForce, listCheckpointsFlag,
	}

	for _, b := range boolChecks {
		if *b {
			return true
		}
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
