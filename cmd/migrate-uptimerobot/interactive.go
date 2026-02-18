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

// interactiveWizardUR manages the state and steps of the interactive wizard.
type interactiveWizardUR struct {
	prompter      *interactive.Prompter
	config        *interactiveConfigUR
	monitors      []uptimerobot.Monitor
	alertContacts []uptimerobot.AlertContact
}

// newInteractiveWizardUR creates a new wizard instance.
func newInteractiveWizardUR() *interactiveWizardUR {
	return &interactiveWizardUR{
		prompter: interactive.NewPrompter(interactive.DefaultConfig()),
		config:   &interactiveConfigUR{},
	}
}

// collectCredentials handles Step 1: connecting to UptimeRobot and fetching monitors.
func (w *interactiveWizardUR) collectCredentials() error {
	w.prompter.PrintHeader("Step 1/5: Source Platform Configuration")
	fmt.Fprintf(os.Stderr, "\n")

	apiKey, err := w.prompter.AskPassword(
		"Enter your UptimeRobot API key:",
		"Get it from: https://uptimerobot.com/dashboard#mySettings",
		interactive.SourceAPIKeyValidator("uptimerobot"),
	)
	if err != nil {
		return fmt.Errorf("failed to get API key: %w", err)
	}
	w.config.uptimerobotAPIKey = apiKey

	spinner := interactive.NewSpinner("Testing UptimeRobot API connection...", os.Stderr)
	spinner.Start()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	urClient := uptimerobot.NewClient(w.config.uptimerobotAPIKey)
	monitors, err := urClient.GetMonitors(ctx)
	if err != nil {
		spinner.ErrorMessage(fmt.Sprintf("Connection failed: %v", err))
		w.prompter.PrintError("Unable to connect to UptimeRobot API")
		w.prompter.PrintInfo("Please verify your API key and try again")
		return fmt.Errorf("connection failed: %w", err)
	}

	alertContacts, err := urClient.GetAlertContacts(ctx)
	if err != nil {
		spinner.ErrorMessage(fmt.Sprintf("Failed to fetch alert contacts: %v", err))
		w.prompter.PrintWarning("Unable to fetch alert contacts, continuing with monitors only")
		alertContacts = []uptimerobot.AlertContact{}
	}

	spinner.SuccessMessage(fmt.Sprintf("Connected! Found %d monitors and %d alert contacts", len(monitors), len(alertContacts)))
	w.monitors = monitors
	w.alertContacts = alertContacts

	printMonitorTypeBreakdown(monitors)
	return nil
}

// printMonitorTypeBreakdown prints the breakdown of monitor types to stderr.
func printMonitorTypeBreakdown(monitors []uptimerobot.Monitor) {
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
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  Monitor types:\n")
	for typeID, count := range typeCounts {
		name := typeNames[typeID]
		if name == "" {
			name = fmt.Sprintf("Unknown (type %d)", typeID)
		}
		fmt.Fprintf(os.Stderr, "    - %s: %d\n", name, count)
	}
}

// selectMode handles Step 2: choosing migration mode.
// Returns true if the caller should delegate to validation mode.
func (w *interactiveWizardUR) selectMode() (bool, error) {
	w.prompter.PrintHeader("Step 2/5: Migration Mode")
	fmt.Fprintf(os.Stderr, "\n")

	mode, err := w.prompter.AskSelect(
		"Select migration mode:",
		[]string{
			"Full migration (generate files and proceed)",
			"Dry run (validate without creating files)",
			"Validate only (check monitors)",
		},
		"Full migration (generate files and proceed)",
	)
	if err != nil {
		return false, fmt.Errorf("failed to select mode: %w", err)
	}

	switch mode {
	case "Validate only (check monitors)":
		w.config.validate = true
		return true, nil
	case "Dry run (validate without creating files)":
		w.config.dryRun = true
	}
	return false, nil
}

// collectHyperpingKey handles Step 3: collecting the Hyperping API key (full migration only).
func (w *interactiveWizardUR) collectHyperpingKey() error {
	if w.config.dryRun || w.config.validate {
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

// configureOutput handles Step 4: collecting output file paths.
func (w *interactiveWizardUR) configureOutput() error {
	w.prompter.PrintHeader(fmt.Sprintf("Step %d/5: Output Configuration", getStepNumber(w.config)))
	fmt.Fprintf(os.Stderr, "\n")

	outputFile, err := w.prompter.AskString(
		"Terraform output file:",
		"hyperping.tf",
		"Main Terraform configuration file",
		interactive.FilePathValidator,
	)
	if err != nil {
		return fmt.Errorf("failed to get output file: %w", err)
	}
	w.config.outputFile = outputFile

	importScriptFile, err := w.prompter.AskString(
		"Import script file:",
		"import.sh",
		"Shell script to import existing resources",
		interactive.FilePathValidator,
	)
	if err != nil {
		return fmt.Errorf("failed to get import script file: %w", err)
	}
	w.config.importScript = importScriptFile

	reportFileValue, err := w.prompter.AskString(
		"Migration report file:",
		"migration-report.json",
		"Detailed JSON report of the migration",
		interactive.FilePathValidator,
	)
	if err != nil {
		return fmt.Errorf("failed to get report file: %w", err)
	}
	w.config.reportFile = reportFileValue

	manualStepsFileValue, err := w.prompter.AskString(
		"Manual steps file:",
		"manual-steps.md",
		"Documentation of manual configuration steps",
		interactive.FilePathValidator,
	)
	if err != nil {
		return fmt.Errorf("failed to get manual steps file: %w", err)
	}
	w.config.manualStepsFile = manualStepsFileValue
	return nil
}

// previewAndConfirm handles Step 5: showing the migration preview and asking for confirmation.
// Returns true if the user wants to proceed.
func (w *interactiveWizardUR) previewAndConfirm() (bool, error) {
	w.prompter.PrintHeader(fmt.Sprintf("Step %d/5: Migration Preview", getStepNumber(w.config)+1))
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  📊 Summary:\n")
	fmt.Fprintf(os.Stderr, "    - Total monitors: %d\n", len(w.monitors))
	fmt.Fprintf(os.Stderr, "    - Alert contacts: %d\n", len(w.alertContacts))
	if w.config.dryRun {
		fmt.Fprintf(os.Stderr, "    - Mode: Dry run (no files will be created)\n")
	} else {
		fmt.Fprintf(os.Stderr, "    - Mode: Full migration\n")
	}
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  📁 Output files:\n")
	fmt.Fprintf(os.Stderr, "    - %s (Terraform configuration)\n", w.config.outputFile)
	fmt.Fprintf(os.Stderr, "    - %s (Import script)\n", w.config.importScript)
	fmt.Fprintf(os.Stderr, "    - %s (Migration report)\n", w.config.reportFile)
	fmt.Fprintf(os.Stderr, "    - %s (Manual steps)\n", w.config.manualStepsFile)
	fmt.Fprintf(os.Stderr, "\n")

	proceed, err := w.prompter.AskConfirm("Proceed with migration?", true)
	if err != nil {
		return false, fmt.Errorf("failed to get confirmation: %w", err)
	}
	return proceed, nil
}

// executeMigration handles Step 6: running the conversion and writing all output files.
func (w *interactiveWizardUR) executeMigration() int {
	w.prompter.PrintHeader(fmt.Sprintf("Step %d/5: Running Migration", getStepNumber(w.config)+2))
	fmt.Fprintf(os.Stderr, "\n")

	conversionSpinner := interactive.NewSpinner("Converting monitors to Hyperping format...", os.Stderr)
	conversionSpinner.Start()

	conv := converter.NewConverter()
	conversionResult := conv.Convert(w.monitors, w.alertContacts)
	conversionSpinner.SuccessMessage("Conversion complete")

	migrationReport := report.Generate(w.monitors, w.alertContacts, conversionResult)
	printMigrationSummary(w.prompter, migrationReport)

	if w.config.dryRun {
		fmt.Fprintf(os.Stderr, "=== DRY RUN MODE ===\n\n")
		fmt.Fprintf(os.Stderr, "No files were created. Migration preview complete.\n")
		return 0
	}

	return w.writeOutputFiles(conversionResult, migrationReport)
}

// printMigrationSummary prints the migration summary and any warnings/errors.
func printMigrationSummary(prompter *interactive.Prompter, migrationReport *report.Report) {
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
		for i, wn := range migrationReport.Warnings {
			if i >= 5 {
				fmt.Fprintf(os.Stderr, "    ... and %d more (see report for details)\n", len(migrationReport.Warnings)-5)
				break
			}
			fmt.Fprintf(os.Stderr, "    - %s: %s\n", wn.Resource, wn.Message)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}
}

// writeOutputFiles writes all generated files and prints final summary.
func (w *interactiveWizardUR) writeOutputFiles(conversionResult *converter.ConversionResult, migrationReport *report.Report) int {
	fileSpinner := interactive.NewSpinner("Writing output files...", os.Stderr)
	fileSpinner.Start()

	tfConfig := generator.GenerateTerraform(conversionResult)
	if writeErr := os.WriteFile(w.config.outputFile, []byte(tfConfig), 0o600); writeErr != nil {
		fileSpinner.ErrorMessage(fmt.Sprintf("Failed to write %s", w.config.outputFile))
		w.prompter.PrintError(fmt.Sprintf("Error: %v", writeErr))
		return 1
	}

	importScriptContent := generator.GenerateImportScript(conversionResult)
	if writeErr := os.WriteFile(w.config.importScript, []byte(importScriptContent), 0o600); writeErr != nil {
		fileSpinner.ErrorMessage(fmt.Sprintf("Failed to write %s", w.config.importScript))
		w.prompter.PrintError(fmt.Sprintf("Error: %v", writeErr))
		return 1
	}

	reportJSON, marshalErr := json.MarshalIndent(migrationReport, "", "  ")
	if marshalErr != nil {
		fileSpinner.ErrorMessage("Failed to marshal report")
		w.prompter.PrintError(fmt.Sprintf("Error: %v", marshalErr))
		return 1
	}
	if writeErr := os.WriteFile(w.config.reportFile, reportJSON, 0o600); writeErr != nil {
		fileSpinner.ErrorMessage(fmt.Sprintf("Failed to write %s", w.config.reportFile))
		w.prompter.PrintError(fmt.Sprintf("Error: %v", writeErr))
		return 1
	}

	manualStepsContent := generator.GenerateManualSteps(conversionResult, w.alertContacts)
	if writeErr := os.WriteFile(w.config.manualStepsFile, []byte(manualStepsContent), 0o600); writeErr != nil {
		fileSpinner.ErrorMessage(fmt.Sprintf("Failed to write %s", w.config.manualStepsFile))
		w.prompter.PrintError(fmt.Sprintf("Error: %v", writeErr))
		return 1
	}

	fileSpinner.SuccessMessage("All files written successfully")
	printFinalSummary(w.prompter, w.config, migrationReport)
	return 0
}

// printFinalSummary prints the completion message and next steps.
func printFinalSummary(prompter *interactive.Prompter, config *interactiveConfigUR, migrationReport *report.Report) {
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
}

// runInteractive runs the migration tool in interactive mode.
func runInteractive() int {
	wizard := newInteractiveWizardUR()

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "🚀 Hyperping Migration Tool - UptimeRobot Edition\n")
	fmt.Fprintf(os.Stderr, "═════════════════════════════════════════════════\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "This wizard will guide you through migrating your UptimeRobot\n")
	fmt.Fprintf(os.Stderr, "monitors to Hyperping.\n")
	fmt.Fprintf(os.Stderr, "\n")

	if err := wizard.collectCredentials(); err != nil {
		wizard.prompter.PrintError(err.Error())
		return 1
	}

	isValidateOnly, err := wizard.selectMode()
	if err != nil {
		wizard.prompter.PrintError(err.Error())
		return 1
	}
	if isValidateOnly {
		return runValidationInteractive(wizard.monitors, wizard.alertContacts, wizard.prompter)
	}

	if hpErr := wizard.collectHyperpingKey(); hpErr != nil {
		wizard.prompter.PrintError(hpErr.Error())
		return 1
	}

	if cfgErr := wizard.configureOutput(); cfgErr != nil {
		wizard.prompter.PrintError(cfgErr.Error())
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

// runValidationInteractive runs validation mode interactively.
func runValidationInteractive(monitors []uptimerobot.Monitor, alertContacts []uptimerobot.AlertContact, prompter *interactive.Prompter) int {
	prompter.PrintHeader("Validation Results")
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
	if *dryRun || *validate || *verbose || *resume || *rollback || *rollbackForce || *listCheckpointsFlag {
		return true
	}
	if *resumeID != "" || *rollbackID != "" {
		return true
	}
	if os.Getenv("UPTIMEROBOT_API_KEY") != "" {
		return true
	}
	if os.Getenv("HYPERPING_API_KEY") != "" {
		return true
	}
	return false
}
