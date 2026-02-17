// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

// migrate-uptimerobot migrates monitoring configurations from UptimeRobot to Hyperping.
//
// Usage:
//
//	export UPTIMEROBOT_API_KEY="u1234567-abc..."
//	export HYPERPING_API_KEY="sk_your_key"
//	go run ./cmd/migrate-uptimerobot
//
// Or with flags:
//
//	go run ./cmd/migrate-uptimerobot -output=hyperping.tf -import-script=import.sh -report=report.json
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/converter"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/generator"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/report"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/uptimerobot"
)

var (
	uptimerobotAPIKey = flag.String("uptimerobot-api-key", "", "UptimeRobot API key (or set UPTIMEROBOT_API_KEY)")
	hyperpingAPIKey   = flag.String("hyperping-api-key", "", "Hyperping API key (or set HYPERPING_API_KEY)")
	output            = flag.String("output", "hyperping.tf", "Output Terraform configuration file")
	importScript      = flag.String("import-script", "import.sh", "Output import script file")
	reportFile        = flag.String("report", "migration-report.json", "Output migration report file")
	manualSteps       = flag.String("manual-steps", "manual-steps.md", "Output manual steps documentation")
	dryRun            = flag.Bool("dry-run", false, "Perform dry run without creating output files")
	validate          = flag.Bool("validate", false, "Validate UptimeRobot resources without generating output")
	verbose           = flag.Bool("verbose", false, "Enable verbose output")
)

// runner holds the resolved configuration for a non-interactive run.
type runner struct {
	urAPIKey string
	hpAPIKey string
	ctx      context.Context
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: migrate-uptimerobot [options]\n\n")
		fmt.Fprintf(os.Stderr, "Migrates monitoring configurations from UptimeRobot to Hyperping.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Validate UptimeRobot monitors\n")
		fmt.Fprintf(os.Stderr, "  migrate-uptimerobot -validate\n\n")
		fmt.Fprintf(os.Stderr, "  # Perform dry run\n")
		fmt.Fprintf(os.Stderr, "  migrate-uptimerobot -dry-run -verbose\n\n")
		fmt.Fprintf(os.Stderr, "  # Generate migration files\n")
		fmt.Fprintf(os.Stderr, "  migrate-uptimerobot -output=hyperping.tf -import-script=import.sh\n\n")
	}

	os.Exit(run())
}

func run() int {
	flag.Parse()

	if shouldUseInteractive() {
		return runInteractive()
	}

	r, exitCode := newRunner()
	if exitCode != 0 {
		return exitCode
	}
	if cancel, ok := r.ctx.Value(cancelKey{}).(context.CancelFunc); ok {
		defer cancel()
	}

	monitors, alertContacts, exitCode := r.fetchMonitors()
	if exitCode != 0 {
		return exitCode
	}

	if *validate {
		return runValidation(monitors, alertContacts)
	}

	conversionResult, migrationReport := r.convertAndReport(monitors, alertContacts)

	if *dryRun {
		fmt.Fprintln(os.Stderr, "\nDry run complete. No files written.")
		return 0
	}

	return r.writeFiles(conversionResult, migrationReport, alertContacts)
}

// cancelKey is an unexported type used as a context key to avoid collisions.
type cancelKey struct{}

// newRunner validates flags, resolves API keys, and sets up the context.
func newRunner() (*runner, int) {
	urAPIKey := *uptimerobotAPIKey
	if urAPIKey == "" {
		urAPIKey = os.Getenv("UPTIMEROBOT_API_KEY")
	}

	hpAPIKey := *hyperpingAPIKey
	if hpAPIKey == "" {
		hpAPIKey = os.Getenv("HYPERPING_API_KEY")
	}

	if urAPIKey == "" {
		fmt.Fprintln(os.Stderr, "Error: UPTIMEROBOT_API_KEY is required")
		fmt.Fprintln(os.Stderr, "Set via environment variable or -uptimerobot-api-key flag")
		return nil, 1
	}

	if !*validate && !*dryRun && hpAPIKey == "" {
		fmt.Fprintln(os.Stderr, "Error: HYPERPING_API_KEY is required for migration")
		fmt.Fprintln(os.Stderr, "Set via environment variable or -hyperping-api-key flag")
		return nil, 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	ctx = context.WithValue(ctx, cancelKey{}, cancel)
	return &runner{urAPIKey: urAPIKey, hpAPIKey: hpAPIKey, ctx: ctx}, 0
}

// fetchMonitors fetches monitors and alert contacts from UptimeRobot.
func (r *runner) fetchMonitors() ([]uptimerobot.Monitor, []uptimerobot.AlertContact, int) {
	urClient := uptimerobot.NewClient(r.urAPIKey)

	if *verbose {
		fmt.Fprintln(os.Stderr, "Fetching monitors from UptimeRobot...")
	}

	monitors, err := urClient.GetMonitors(r.ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching monitors: %v\n", err)
		return nil, nil, 1
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Fetched %d monitors\n", len(monitors))
		fmt.Fprintln(os.Stderr, "Fetching alert contacts from UptimeRobot...")
	}

	alertContacts, err := urClient.GetAlertContacts(r.ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Error fetching alert contacts: %v\n", err)
		alertContacts = []uptimerobot.AlertContact{}
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Fetched %d alert contacts\n", len(alertContacts))
	}

	return monitors, alertContacts, 0
}

// convertAndReport converts monitors and prints the migration summary.
func (r *runner) convertAndReport(monitors []uptimerobot.Monitor, alertContacts []uptimerobot.AlertContact) (*converter.ConversionResult, *report.Report) {
	if *verbose {
		fmt.Fprintln(os.Stderr, "Converting monitors to Hyperping resources...")
	}

	conv := converter.NewConverter()
	conversionResult := conv.Convert(monitors, alertContacts)
	migrationReport := report.Generate(monitors, alertContacts, conversionResult)

	fmt.Fprintln(os.Stderr, "\nMigration Summary:")
	fmt.Fprintf(os.Stderr, "  Total monitors: %d\n", migrationReport.Summary.TotalMonitors)
	fmt.Fprintf(os.Stderr, "  Migrated monitors: %d\n", migrationReport.Summary.MigratedMonitors)
	fmt.Fprintf(os.Stderr, "  Migrated healthchecks: %d\n", migrationReport.Summary.MigratedHealthchecks)
	fmt.Fprintf(os.Stderr, "  Warnings: %d\n", len(migrationReport.Warnings))
	fmt.Fprintf(os.Stderr, "  Errors: %d\n", len(migrationReport.Errors))

	if len(migrationReport.Errors) > 0 {
		fmt.Fprintln(os.Stderr, "\nErrors encountered:")
		for _, e := range migrationReport.Errors {
			fmt.Fprintf(os.Stderr, "  - %s: %s\n", e.Resource, e.Message)
		}
	}

	if len(migrationReport.Warnings) > 0 {
		fmt.Fprintln(os.Stderr, "\nWarnings:")
		for _, w := range migrationReport.Warnings {
			fmt.Fprintf(os.Stderr, "  - %s: %s\n", w.Resource, w.Message)
		}
	}

	return conversionResult, migrationReport
}

// writeFiles writes all generated output files.
func (r *runner) writeFiles(conversionResult *converter.ConversionResult, migrationReport *report.Report, alertContacts []uptimerobot.AlertContact) int {
	if exitCode := r.writeTerraformConfig(conversionResult); exitCode != 0 {
		return exitCode
	}
	if exitCode := r.writeImportScript(conversionResult); exitCode != 0 {
		return exitCode
	}
	if exitCode := r.writeMigrationReport(migrationReport); exitCode != 0 {
		return exitCode
	}
	if exitCode := r.writeManualSteps(conversionResult, alertContacts); exitCode != 0 {
		return exitCode
	}

	fmt.Fprintln(os.Stderr, "\nMigration files generated successfully!")
	fmt.Fprintln(os.Stderr, "\nNext steps:")
	fmt.Fprintf(os.Stderr, "  1. Review %s and adjust as needed\n", *output)
	fmt.Fprintf(os.Stderr, "  2. Run: terraform init && terraform plan\n")
	fmt.Fprintf(os.Stderr, "  3. Run: terraform apply\n")
	fmt.Fprintf(os.Stderr, "  4. Review %s for manual configuration steps\n", *manualSteps)
	return 0
}

// writeTerraformConfig generates and writes the Terraform configuration file.
func (r *runner) writeTerraformConfig(conversionResult *converter.ConversionResult) int {
	if *verbose {
		fmt.Fprintln(os.Stderr, "\nGenerating Terraform configuration...")
	}
	tfConfig := generator.GenerateTerraform(conversionResult)
	if err := os.WriteFile(*output, []byte(tfConfig), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing Terraform config: %v\n", err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "  ✓ Terraform configuration written to %s\n", *output)
	return 0
}

// writeImportScript generates and writes the import script file.
func (r *runner) writeImportScript(conversionResult *converter.ConversionResult) int {
	if *verbose {
		fmt.Fprintln(os.Stderr, "Generating import script...")
	}
	importScriptContent := generator.GenerateImportScript(conversionResult)
	if err := os.WriteFile(*importScript, []byte(importScriptContent), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing import script: %v\n", err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "  ✓ Import script written to %s\n", *importScript)
	return 0
}

// writeMigrationReport marshals and writes the JSON migration report.
func (r *runner) writeMigrationReport(migrationReport *report.Report) int {
	if *verbose {
		fmt.Fprintln(os.Stderr, "Writing migration report...")
	}
	reportJSON, err := json.MarshalIndent(migrationReport, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling report: %v\n", err)
		return 1
	}
	if err := os.WriteFile(*reportFile, reportJSON, 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing report: %v\n", err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "  ✓ Migration report written to %s\n", *reportFile)
	return 0
}

// writeManualSteps generates and writes the manual steps documentation.
func (r *runner) writeManualSteps(conversionResult *converter.ConversionResult, alertContacts []uptimerobot.AlertContact) int {
	if *verbose {
		fmt.Fprintln(os.Stderr, "Generating manual steps documentation...")
	}

	manualStepsContent := generator.GenerateManualSteps(conversionResult, alertContacts)
	if err := os.WriteFile(*manualSteps, []byte(manualStepsContent), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing manual steps: %v\n", err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "  ✓ Manual steps written to %s\n", *manualSteps)
	return 0
}

func runValidation(monitors []uptimerobot.Monitor, alertContacts []uptimerobot.AlertContact) int {
	fmt.Fprintln(os.Stderr, "Validating UptimeRobot monitors...")

	typeCounts := make(map[int]int)
	for _, m := range monitors {
		typeCounts[m.Type]++
	}

	fmt.Fprintln(os.Stderr, "\nMonitor Types:")
	typeNames := map[int]string{
		1: "HTTP/HTTPS",
		2: "Keyword",
		3: "Ping (ICMP)",
		4: "Port",
		5: "Heartbeat",
	}

	for typeID, count := range typeCounts {
		name := typeNames[typeID]
		if name == "" {
			name = fmt.Sprintf("Unknown (type %d)", typeID)
		}
		fmt.Fprintf(os.Stderr, "  %s: %d\n", name, count)
	}

	fmt.Fprintf(os.Stderr, "\nTotal monitors: %d\n", len(monitors))
	fmt.Fprintf(os.Stderr, "Alert contacts: %d\n", len(alertContacts))

	fmt.Fprintln(os.Stderr, "\nValidation complete.")
	return 0
}
