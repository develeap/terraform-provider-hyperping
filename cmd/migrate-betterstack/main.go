// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

// migrate-betterstack migrates monitoring resources from Better Stack to Hyperping
//
// Usage:
//
//	export BETTERSTACK_API_TOKEN="your_betterstack_token"
//	export HYPERPING_API_KEY="sk_your_hyperping_key"
//	go run ./cmd/migrate-betterstack
//
// Or with flags:
//
//	go run ./cmd/migrate-betterstack \
//	  --betterstack-token="token" \
//	  --hyperping-api-key="sk_key" \
//	  --output=migrated-resources.tf \
//	  --import-script=import.sh \
//	  --report=migration-report.json
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack/betterstack"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack/converter"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack/generator"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack/report"
)

var (
	betterstackToken = flag.String("betterstack-token", "", "Better Stack API token (or set BETTERSTACK_API_TOKEN)")
	hyperpingAPIKey  = flag.String("hyperping-api-key", "", "Hyperping API key (or set HYPERPING_API_KEY)")
	outputFile       = flag.String("output", "migrated-resources.tf", "Output Terraform configuration file")
	importScript     = flag.String("import-script", "import.sh", "Output import script file")
	reportFile       = flag.String("report", "migration-report.json", "Output migration report file")
	manualStepsFile  = flag.String("manual-steps", "manual-steps.md", "Output manual steps documentation")
	dryRun           = flag.Bool("dry-run", false, "Validate without creating files")
	validateTF       = flag.Bool("validate", false, "Run terraform validate on output")
	verbose          = flag.Bool("verbose", false, "Enable verbose logging")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: migrate-betterstack [options]\n\n")
		fmt.Fprintf(os.Stderr, "Migrates monitoring resources from Better Stack to Hyperping.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Basic migration\n")
		fmt.Fprintf(os.Stderr, "  export BETTERSTACK_API_TOKEN=\"your_token\"\n")
		fmt.Fprintf(os.Stderr, "  export HYPERPING_API_KEY=\"sk_your_key\"\n")
		fmt.Fprintf(os.Stderr, "  migrate-betterstack\n\n")
		fmt.Fprintf(os.Stderr, "  # Dry run to validate\n")
		fmt.Fprintf(os.Stderr, "  migrate-betterstack --dry-run --verbose\n\n")
		fmt.Fprintf(os.Stderr, "  # Custom output files\n")
		fmt.Fprintf(os.Stderr, "  migrate-betterstack --output=production.tf --import-script=import-prod.sh\n\n")
	}
	os.Exit(run())
}

func run() int {
	flag.Parse()

	// Get credentials from flags or environment
	bsToken := *betterstackToken
	if bsToken == "" {
		bsToken = os.Getenv("BETTERSTACK_API_TOKEN")
	}
	hpKey := *hyperpingAPIKey
	if hpKey == "" {
		hpKey = os.Getenv("HYPERPING_API_KEY")
	}

	if bsToken == "" {
		fmt.Fprintln(os.Stderr, "Error: Better Stack API token is required")
		fmt.Fprintln(os.Stderr, "Set --betterstack-token flag or BETTERSTACK_API_TOKEN environment variable")
		return 1
	}

	if hpKey == "" {
		fmt.Fprintln(os.Stderr, "Error: Hyperping API key is required")
		fmt.Fprintln(os.Stderr, "Set --hyperping-api-key flag or HYPERPING_API_KEY environment variable")
		return 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if *verbose {
		fmt.Fprintln(os.Stderr, "Starting Better Stack to Hyperping migration...")
	}

	// Create Better Stack client
	bsClient := betterstack.NewClient(bsToken)

	// Note: Hyperping client not needed - tool generates Terraform files
	// Users will apply the Terraform config to create resources in Hyperping

	// Fetch Better Stack resources
	if *verbose {
		fmt.Fprintln(os.Stderr, "Fetching Better Stack monitors...")
	}
	monitors, err := bsClient.FetchMonitors(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching Better Stack monitors: %v\n", err)
		return 1
	}
	if *verbose {
		fmt.Fprintf(os.Stderr, "Found %d monitors\n", len(monitors))
	}

	if *verbose {
		fmt.Fprintln(os.Stderr, "Fetching Better Stack heartbeats...")
	}
	heartbeats, err := bsClient.FetchHeartbeats(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching Better Stack heartbeats: %v\n", err)
		return 1
	}
	if *verbose {
		fmt.Fprintf(os.Stderr, "Found %d heartbeats\n", len(heartbeats))
	}

	// Convert resources
	if *verbose {
		fmt.Fprintln(os.Stderr, "Converting monitors to Hyperping format...")
	}
	conv := converter.New()
	convertedMonitors, monitorIssues := conv.ConvertMonitors(monitors)
	convertedHealthchecks, healthcheckIssues := conv.ConvertHeartbeats(heartbeats)

	// Generate Terraform configuration
	gen := generator.New()
	tfConfig := gen.GenerateTerraform(convertedMonitors, convertedHealthchecks)

	// Generate import script
	importScriptContent := gen.GenerateImportScript(convertedMonitors, convertedHealthchecks)

	// Generate manual steps
	manualSteps := gen.GenerateManualSteps(monitorIssues, healthcheckIssues)

	// Generate migration report
	migrationReport := report.Generate(
		monitors,
		heartbeats,
		convertedMonitors,
		convertedHealthchecks,
		monitorIssues,
		healthcheckIssues,
	)

	// Dry run mode - just validate
	if *dryRun {
		fmt.Fprintln(os.Stderr, "\n=== DRY RUN MODE ===")
		fmt.Fprintf(os.Stderr, "Would create %s (%d bytes)\n", *outputFile, len(tfConfig))
		fmt.Fprintf(os.Stderr, "Would create %s (%d bytes)\n", *importScript, len(importScriptContent))
		fmt.Fprintf(os.Stderr, "Would create %s (%d bytes)\n", *reportFile, len(migrationReport.JSON()))
		fmt.Fprintf(os.Stderr, "Would create %s (%d bytes)\n", *manualStepsFile, len(manualSteps))
		fmt.Fprintln(os.Stderr, "\nSummary:")
		migrationReport.PrintSummary(os.Stderr)
		return 0
	}

	// Write Terraform configuration
	if err := os.WriteFile(*outputFile, []byte(tfConfig), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing Terraform config: %v\n", err)
		return 1
	}
	if *verbose {
		fmt.Fprintf(os.Stderr, "Generated %s\n", *outputFile)
	}

	// Write import script
	if err := os.WriteFile(*importScript, []byte(importScriptContent), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing import script: %v\n", err)
		return 1
	}
	if *verbose {
		fmt.Fprintf(os.Stderr, "Generated %s\n", *importScript)
	}

	// Write migration report
	if err := os.WriteFile(*reportFile, []byte(migrationReport.JSON()), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing migration report: %v\n", err)
		return 1
	}
	if *verbose {
		fmt.Fprintf(os.Stderr, "Generated %s\n", *reportFile)
	}

	// Write manual steps
	if err := os.WriteFile(*manualStepsFile, []byte(manualSteps), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing manual steps: %v\n", err)
		return 1
	}
	if *verbose {
		fmt.Fprintf(os.Stderr, "Generated %s\n", *manualStepsFile)
	}

	// Print summary
	fmt.Fprintln(os.Stderr, "\n=== Migration Complete ===")
	migrationReport.PrintSummary(os.Stderr)

	fmt.Fprintf(os.Stderr, "\nGenerated files:\n")
	fmt.Fprintf(os.Stderr, "  - %s (Terraform configuration)\n", *outputFile)
	fmt.Fprintf(os.Stderr, "  - %s (import script)\n", *importScript)
	fmt.Fprintf(os.Stderr, "  - %s (migration report)\n", *reportFile)
	fmt.Fprintf(os.Stderr, "  - %s (manual steps)\n", *manualStepsFile)

	fmt.Fprintf(os.Stderr, "\nNext steps:\n")
	fmt.Fprintf(os.Stderr, "  1. Review %s and adjust as needed\n", *outputFile)
	fmt.Fprintf(os.Stderr, "  2. Review %s for any manual actions\n", *manualStepsFile)
	fmt.Fprintf(os.Stderr, "  3. Run: terraform init\n")
	fmt.Fprintf(os.Stderr, "  4. Run: terraform plan\n")
	fmt.Fprintf(os.Stderr, "  5. Run: terraform apply\n")

	// Validate Terraform if requested
	if *validateTF {
		if *verbose {
			fmt.Fprintln(os.Stderr, "\nValidating Terraform configuration...")
		}
		if err := gen.ValidateTerraform(*outputFile); err != nil {
			fmt.Fprintf(os.Stderr, "Terraform validation failed: %v\n", err)
			return 1
		}
		fmt.Fprintln(os.Stderr, "Terraform validation passed!")
	}

	return 0
}
