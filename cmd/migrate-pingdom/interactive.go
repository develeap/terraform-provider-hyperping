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
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/pingdom"
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

// interactiveWizardPD manages the state and steps of the Pingdom interactive wizard.
type interactiveWizardPD struct {
	prompter *interactive.Prompter
	config   *interactiveConfigPD
	checks   []pingdom.Check
	results  []converter.ConversionResult
	ctx      context.Context
}

// newInteractiveWizardPD creates a new wizard instance.
func newInteractiveWizardPD() *interactiveWizardPD {
	return &interactiveWizardPD{
		prompter: interactive.NewPrompter(interactive.DefaultConfig()),
		config:   &interactiveConfigPD{},
	}
}

// collectCredentials handles Step 1: connecting to Pingdom and fetching checks.
func (w *interactiveWizardPD) collectCredentials() error {
	w.prompter.PrintHeader("Step 1/5: Source Platform Configuration")
	fmt.Fprintf(os.Stderr, "\n")

	apiKey, err := w.prompter.AskPassword(
		"Enter your Pingdom API token:",
		"Get it from: https://my.pingdom.com/app/api-tokens",
		interactive.SourceAPIKeyValidator("pingdom"),
	)
	if err != nil {
		return fmt.Errorf("failed to get API token: %w", err)
	}
	w.config.pingdomAPIKey = apiKey

	spinner := interactive.NewSpinner("Testing Pingdom API connection...", os.Stderr)
	spinner.Start()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	w.ctx = ctx

	pingdomClient := createPingdomClient(w.config.pingdomAPIKey)
	checks, err := pingdomClient.ListChecks(ctx)
	if err != nil {
		spinner.ErrorMessage(fmt.Sprintf("Connection failed: %v", err))
		w.prompter.PrintError("Unable to connect to Pingdom API")
		w.prompter.PrintInfo("Please verify your API token and try again")
		return fmt.Errorf("connection failed: %w", err)
	}

	spinner.SuccessMessage(fmt.Sprintf("Connected! Found %d checks", len(checks)))
	w.checks = checks

	printCheckTypeBreakdown(checks)
	return nil
}

// printCheckTypeBreakdown prints the breakdown of check types to stderr.
func printCheckTypeBreakdown(checks []pingdom.Check) {
	typeCounts := make(map[string]int)
	for _, c := range checks {
		typeCounts[c.Type]++
	}
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  Check types:\n")
	for checkType, count := range typeCounts {
		fmt.Fprintf(os.Stderr, "    - %s: %d\n", checkType, count)
	}
}

// selectMode handles Step 2: choosing migration mode.
func (w *interactiveWizardPD) selectMode() error {
	w.prompter.PrintHeader("Step 2/5: Migration Mode")
	fmt.Fprintf(os.Stderr, "\n")

	mode, err := w.prompter.AskSelect(
		"Select migration mode:",
		[]string{
			"Full migration (create resources in Hyperping)",
			"Dry run (generate configs only)",
		},
		"Full migration (create resources in Hyperping)",
	)
	if err != nil {
		return fmt.Errorf("failed to select mode: %w", err)
	}

	w.config.dryRun = mode == "Dry run (generate configs only)"
	return nil
}

// collectHyperpingKey handles Step 3: collecting the Hyperping API key (full migration only).
func (w *interactiveWizardPD) collectHyperpingKey() error {
	if w.config.dryRun {
		return nil
	}
	w.prompter.PrintHeader("Step 3/5: Destination Platform Configuration")
	fmt.Fprintf(os.Stderr, "\n")

	hyperpingKey, err := w.prompter.AskPassword(
		"Enter your Hyperping API key:",
		"Get it from: https://app.hyperping.io/settings/api",
		interactive.HyperpingAPIKeyValidator,
	)
	if err != nil {
		return fmt.Errorf("failed to get API key: %w", err)
	}
	w.config.hyperpingAPIKey = hyperpingKey
	return nil
}

// configureOutput handles Step 4: collecting output directory and prefix.
func (w *interactiveWizardPD) configureOutput() error {
	stepNum := 3
	if !w.config.dryRun {
		stepNum = 4
	}
	w.prompter.PrintHeader(fmt.Sprintf("Step %d/5: Output Configuration", stepNum))
	fmt.Fprintf(os.Stderr, "\n")

	outputDir, err := w.prompter.AskString(
		"Output directory for migration files:",
		"./pingdom-migration",
		"Directory where all migration files will be saved",
		interactive.FilePathValidator,
	)
	if err != nil {
		return fmt.Errorf("failed to get output directory: %w", err)
	}
	w.config.outputDir = outputDir

	prefixValue, err := w.prompter.AskString(
		"Resource name prefix (optional):",
		"",
		"Prefix for Terraform resource names (e.g., 'pingdom_')",
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to get prefix: %w", err)
	}
	w.config.prefix = prefixValue
	return nil
}

// previewAndConfirm handles Step 5: conversion preview and user confirmation.
// Returns true if the user wants to proceed.
func (w *interactiveWizardPD) previewAndConfirm() (bool, error) {
	stepNum := 4
	if w.config.dryRun {
		stepNum = 3
	}
	w.prompter.PrintHeader(fmt.Sprintf("Step %d/5: Migration Preview", stepNum+1))
	fmt.Fprintf(os.Stderr, "\n")

	checkConverter := converter.NewCheckConverter()
	results := make([]converter.ConversionResult, len(w.checks))
	supportedCount := 0
	for i, check := range w.checks {
		results[i] = checkConverter.Convert(check)
		if results[i].Supported {
			supportedCount++
		}
	}
	w.results = results

	fmt.Fprintf(os.Stderr, "  📊 Summary:\n")
	fmt.Fprintf(os.Stderr, "    - Total checks: %d\n", len(w.checks))
	fmt.Fprintf(os.Stderr, "    - Supported checks: %d\n", supportedCount)
	fmt.Fprintf(os.Stderr, "    - Unsupported checks: %d\n", len(w.checks)-supportedCount)
	if w.config.dryRun {
		fmt.Fprintf(os.Stderr, "    - Mode: Dry run (configs only, no resources created)\n")
	} else {
		fmt.Fprintf(os.Stderr, "    - Mode: Full migration (resources will be created)\n")
	}
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  📁 Output directory: %s\n", w.config.outputDir)
	if w.config.prefix != "" {
		fmt.Fprintf(os.Stderr, "  🏷️  Resource prefix: %s\n", w.config.prefix)
	}
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  Files to be generated:\n")
	fmt.Fprintf(os.Stderr, "    - monitors.tf (Terraform configuration)\n")
	fmt.Fprintf(os.Stderr, "    - import.sh (Import script)\n")
	fmt.Fprintf(os.Stderr, "    - report.json (Detailed migration report)\n")
	fmt.Fprintf(os.Stderr, "    - report.txt (Human-readable report)\n")
	fmt.Fprintf(os.Stderr, "    - manual-steps.md (Manual configuration steps)\n")
	fmt.Fprintf(os.Stderr, "\n")

	if len(w.checks)-supportedCount > 0 {
		w.prompter.PrintWarning(fmt.Sprintf("%d unsupported checks will require manual migration", len(w.checks)-supportedCount))
		fmt.Fprintf(os.Stderr, "\n")
	}

	proceed, err := w.prompter.AskConfirm("Proceed with migration?", true)
	if err != nil {
		return false, fmt.Errorf("failed to get confirmation: %w", err)
	}
	return proceed, nil
}

// executeMigration handles Step 6: generating files and optionally creating Hyperping resources.
func (w *interactiveWizardPD) executeMigration() int {
	stepNum := 5
	if w.config.dryRun {
		stepNum = 4
	}
	w.prompter.PrintHeader(fmt.Sprintf("Step %d/5: Running Migration", stepNum+1))
	fmt.Fprintf(os.Stderr, "\n")

	if mkdirErr := os.MkdirAll(w.config.outputDir, 0o755); mkdirErr != nil {
		w.prompter.PrintError(fmt.Sprintf("Failed to create output directory: %v", mkdirErr))
		return 1
	}

	progressBar := interactive.NewProgressBar(5, "Generating files", os.Stderr)

	reporter := report.NewReporter()
	migrationReport := reporter.GenerateReport(w.checks, w.results)

	if exitCode := w.writeGeneratedFiles(reporter, migrationReport, progressBar); exitCode != 0 {
		return exitCode
	}

	createdResources := w.createHyperpingResources(progressBar)

	//nolint:errcheck // best-effort progress display; Add error does not affect migration correctness
	progressBar.Add(1)

	if exitCode := w.writeImportScript(createdResources); exitCode != 0 {
		return exitCode
	}

	//nolint:errcheck // best-effort progress display; Finish error does not affect migration correctness
	progressBar.Finish()

	w.printFinalSummary(migrationReport)
	return 0
}

// writeGeneratedFiles writes TF config, JSON report, text report, and manual steps.
func (w *interactiveWizardPD) writeGeneratedFiles(reporter *report.Reporter, migrationReport *report.MigrationReport, progressBar *interactive.ProgressBar) int {
	tfGen := generator.NewTerraformGenerator(w.config.prefix)
	hclContent := tfGen.GenerateHCL(w.checks, w.results)
	hclPath := filepath.Join(w.config.outputDir, "monitors.tf")
	if writeErr := os.WriteFile(hclPath, []byte(hclContent), 0o600); writeErr != nil {
		w.prompter.PrintError(fmt.Sprintf("Failed to write Terraform config: %v", writeErr))
		return 1
	}
	//nolint:errcheck // best-effort progress display; Add error does not affect migration correctness
	progressBar.Add(1)

	jsonReport, err := reporter.GenerateJSONReport(migrationReport)
	if err != nil {
		w.prompter.PrintError(fmt.Sprintf("Failed to generate JSON report: %v", err))
		return 1
	}
	jsonPath := filepath.Join(w.config.outputDir, "report.json")                        //nolint:gosec // G703: outputDir is a CLI flag, operator-controlled
	if writeErr := os.WriteFile(jsonPath, []byte(jsonReport), 0o600); writeErr != nil { //nolint:gosec // G703: jsonPath derived from operator-controlled CLI flag
		w.prompter.PrintError(fmt.Sprintf("Failed to write JSON report: %v", writeErr))
		return 1
	}
	//nolint:errcheck // best-effort progress display; Add error does not affect migration correctness
	progressBar.Add(1)

	textReport := reporter.GenerateTextReport(migrationReport)
	textPath := filepath.Join(w.config.outputDir, "report.txt")                         //nolint:gosec // G703: outputDir is a CLI flag, operator-controlled
	if writeErr := os.WriteFile(textPath, []byte(textReport), 0o600); writeErr != nil { //nolint:gosec // G703: textPath derived from operator-controlled CLI flag
		w.prompter.PrintError(fmt.Sprintf("Failed to write text report: %v", writeErr))
		return 1
	}
	//nolint:errcheck // best-effort progress display; Add error does not affect migration correctness
	progressBar.Add(1)

	manualSteps := reporter.GenerateManualStepsMarkdown(migrationReport)
	manualPath := filepath.Join(w.config.outputDir, "manual-steps.md")                     //nolint:gosec // G703: outputDir is a CLI flag, operator-controlled
	if writeErr := os.WriteFile(manualPath, []byte(manualSteps), 0o600); writeErr != nil { //nolint:gosec // G703: manualPath derived from operator-controlled CLI flag
		w.prompter.PrintError(fmt.Sprintf("Failed to write manual steps: %v", writeErr))
		return 1
	}
	//nolint:errcheck // best-effort progress display; Add error does not affect migration correctness
	progressBar.Add(1)

	return 0
}

// createHyperpingResources creates monitors in Hyperping when not in dry-run mode.
func (w *interactiveWizardPD) createHyperpingResources(progressBar *interactive.ProgressBar) map[int]string {
	_ = progressBar
	createdResources := make(map[int]string)
	if w.config.dryRun {
		return createdResources
	}

	createSpinner := interactive.NewSpinner("Creating monitors in Hyperping...", os.Stderr)
	createSpinner.Start()

	hyperpingClient := createHyperpingClient(w.config.hyperpingAPIKey)
	createdCount := 0
	errorCount := 0

	for i, check := range w.checks {
		result := w.results[i]
		if !result.Supported || result.Monitor == nil {
			continue
		}

		monitor, err := hyperpingClient.CreateMonitor(w.ctx, *result.Monitor)
		if err != nil {
			errorCount++
			if *verbose {
				fmt.Fprintf(os.Stderr, "\nWarning: Failed to create monitor for check %d (%s): %v\n", check.ID, check.Name, err) //nolint:gosec // G705: writing to stderr, not an HTTP response
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

	return createdResources
}

// writeImportScript generates and writes the import shell script.
func (w *interactiveWizardPD) writeImportScript(createdResources map[int]string) int {
	importGen := generator.NewImportGenerator(w.config.prefix)
	importScript := importGen.GenerateImportScript(w.checks, w.results, createdResources)
	importPath := filepath.Join(w.config.outputDir, "import.sh")                            //nolint:gosec // G703: outputDir is a CLI flag, operator-controlled
	if writeErr := os.WriteFile(importPath, []byte(importScript), 0o700); writeErr != nil { //nolint:gosec // G306: import.sh must be executable (0700); G703: importPath derived from operator-controlled CLI flag
		w.prompter.PrintError(fmt.Sprintf("Failed to write import script: %v", writeErr))
		return 1
	}
	return 0
}

// printFinalSummary prints the completion message and next steps.
func (w *interactiveWizardPD) printFinalSummary(migrationReport *report.MigrationReport) {
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "✅ Migration complete!\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Generated files in %s:\n", w.config.outputDir)
	fmt.Fprintf(os.Stderr, "  📄 monitors.tf - Terraform configuration\n")
	fmt.Fprintf(os.Stderr, "  📜 import.sh - Import script\n")
	fmt.Fprintf(os.Stderr, "  📊 report.json - Detailed migration report\n")
	fmt.Fprintf(os.Stderr, "  📝 report.txt - Human-readable report\n")
	fmt.Fprintf(os.Stderr, "  📋 manual-steps.md - Manual configuration steps\n")
	fmt.Fprintf(os.Stderr, "\n")

	if len(migrationReport.ManualSteps) > 0 {
		w.prompter.PrintWarning(fmt.Sprintf("%d checks require manual steps - see manual-steps.md", len(migrationReport.ManualSteps)))
		fmt.Fprintf(os.Stderr, "\n")
	}

	if w.config.dryRun {
		fmt.Fprintf(os.Stderr, "Next steps:\n")
		fmt.Fprintf(os.Stderr, "  1. Review monitors.tf and adjust as needed\n")
		fmt.Fprintf(os.Stderr, "  2. Review manual-steps.md for unsupported checks\n")
		fmt.Fprintf(os.Stderr, "  3. Run without --dry-run to create resources in Hyperping\n")
	} else {
		fmt.Fprintf(os.Stderr, "Next steps:\n")
		fmt.Fprintf(os.Stderr, "  1. Review monitors.tf and adjust as needed\n")
		fmt.Fprintf(os.Stderr, "  2. Run: cd %s && terraform init\n", w.config.outputDir)
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
}

// runInteractive runs the migration tool in interactive mode.
func runInteractive() int {
	wizard := newInteractiveWizardPD()

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "🚀 Hyperping Migration Tool - Pingdom Edition\n")
	fmt.Fprintf(os.Stderr, "═════════════════════════════════════════════════\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "This wizard will guide you through migrating your Pingdom\n")
	fmt.Fprintf(os.Stderr, "checks to Hyperping.\n")
	fmt.Fprintf(os.Stderr, "\n")

	if err := wizard.collectCredentials(); err != nil {
		wizard.prompter.PrintError(err.Error())
		return 1
	}

	if err := wizard.selectMode(); err != nil {
		wizard.prompter.PrintError(err.Error())
		return 1
	}

	if err := wizard.collectHyperpingKey(); err != nil {
		wizard.prompter.PrintError(err.Error())
		return 1
	}

	if err := wizard.configureOutput(); err != nil {
		wizard.prompter.PrintError(err.Error())
		return 1
	}

	proceed, err := wizard.previewAndConfirm()
	if err != nil {
		wizard.prompter.PrintError(err.Error())
		return 1
	}
	if !proceed {
		wizard.prompter.PrintInfo("Migration cancelled by user")
		return 0
	}

	return wizard.executeMigration()
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
	if *dryRun || *verbose || *resume || *rollback || *rollbackForce || *listCheckpointsFlag {
		return true
	}
	if *resumeID != "" || *rollbackID != "" {
		return true
	}
	if os.Getenv("PINGDOM_API_KEY") != "" || os.Getenv("PINGDOM_API_TOKEN") != "" {
		return true
	}
	if os.Getenv("HYPERPING_API_KEY") != "" {
		return true
	}
	return false
}
