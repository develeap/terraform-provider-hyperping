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

	// Check if interactive mode should be used
	if shouldUseInteractive() {
		return runInteractive()
	}

	// Get API keys from flags or environment
	urAPIKey := *uptimerobotAPIKey
	if urAPIKey == "" {
		urAPIKey = os.Getenv("UPTIMEROBOT_API_KEY")
	}

	hpAPIKey := *hyperpingAPIKey
	if hpAPIKey == "" {
		hpAPIKey = os.Getenv("HYPERPING_API_KEY")
	}

	// Validate required parameters
	if urAPIKey == "" {
		fmt.Fprintln(os.Stderr, "Error: UPTIMEROBOT_API_KEY is required")
		fmt.Fprintln(os.Stderr, "Set via environment variable or -uptimerobot-api-key flag")
		return 1
	}

	if !*validate && !*dryRun && hpAPIKey == "" {
		fmt.Fprintln(os.Stderr, "Error: HYPERPING_API_KEY is required for migration")
		fmt.Fprintln(os.Stderr, "Set via environment variable or -hyperping-api-key flag")
		return 1
	}

	// Create UptimeRobot client
	urClient := uptimerobot.NewClient(urAPIKey)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Fetch monitors from UptimeRobot
	if *verbose {
		fmt.Fprintln(os.Stderr, "Fetching monitors from UptimeRobot...")
	}

	monitors, err := urClient.GetMonitors(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching monitors: %v\n", err)
		return 1
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Fetched %d monitors\n", len(monitors))
	}

	// Fetch alert contacts
	if *verbose {
		fmt.Fprintln(os.Stderr, "Fetching alert contacts from UptimeRobot...")
	}

	alertContacts, err := urClient.GetAlertContacts(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Error fetching alert contacts: %v\n", err)
		alertContacts = []uptimerobot.AlertContact{}
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Fetched %d alert contacts\n", len(alertContacts))
	}

	// Validate mode
	if *validate {
		return runValidation(monitors, alertContacts)
	}

	// Convert monitors to Hyperping resources
	if *verbose {
		fmt.Fprintln(os.Stderr, "Converting monitors to Hyperping resources...")
	}

	conv := converter.NewConverter()
	conversionResult := conv.Convert(monitors, alertContacts)

	// Generate migration report
	migrationReport := report.Generate(monitors, alertContacts, conversionResult)

	// Print summary
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

	// Dry run mode
	if *dryRun {
		fmt.Fprintln(os.Stderr, "\nDry run complete. No files written.")
		return 0
	}

	// Generate Terraform configuration
	if *verbose {
		fmt.Fprintln(os.Stderr, "\nGenerating Terraform configuration...")
	}

	tfConfig := generator.GenerateTerraform(conversionResult)
	if err = os.WriteFile(*output, []byte(tfConfig), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing Terraform config: %v\n", err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "  ✓ Terraform configuration written to %s\n", *output)

	// Generate import script
	if *verbose {
		fmt.Fprintln(os.Stderr, "Generating import script...")
	}

	importScriptContent := generator.GenerateImportScript(conversionResult)
	if err = os.WriteFile(*importScript, []byte(importScriptContent), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing import script: %v\n", err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "  ✓ Import script written to %s\n", *importScript)

	// Generate migration report
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

	// Generate manual steps documentation
	if *verbose {
		fmt.Fprintln(os.Stderr, "Generating manual steps documentation...")
	}

	manualStepsContent := generator.GenerateManualSteps(conversionResult, alertContacts)
	if err := os.WriteFile(*manualSteps, []byte(manualStepsContent), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing manual steps: %v\n", err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "  ✓ Manual steps written to %s\n", *manualSteps)

	fmt.Fprintln(os.Stderr, "\nMigration files generated successfully!")
	fmt.Fprintln(os.Stderr, "\nNext steps:")
	fmt.Fprintf(os.Stderr, "  1. Review %s and adjust as needed\n", *output)
	fmt.Fprintf(os.Stderr, "  2. Run: terraform init && terraform plan\n")
	fmt.Fprintf(os.Stderr, "  3. Run: terraform apply\n")
	fmt.Fprintf(os.Stderr, "  4. Review %s for manual configuration steps\n", *manualSteps)

	return 0
}

func runValidation(monitors []uptimerobot.Monitor, alertContacts []uptimerobot.AlertContact) int {
	fmt.Fprintln(os.Stderr, "Validating UptimeRobot monitors...")

	// Count by type
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
