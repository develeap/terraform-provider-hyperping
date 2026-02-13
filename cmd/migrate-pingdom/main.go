// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

// migrate-pingdom migrates Pingdom checks to Hyperping monitors.
//
// Usage:
//
//	export PINGDOM_API_KEY="your_pingdom_token"
//	export HYPERPING_API_KEY="sk_your_hyperping_key"
//	go run ./cmd/migrate-pingdom --output=./migration-output
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/converter"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/generator"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/pingdom"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/report"
	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

var (
	pingdomAPIKey    = flag.String("pingdom-api-key", "", "Pingdom API token (or set PINGDOM_API_KEY)")
	hyperpingAPIKey  = flag.String("hyperping-api-key", "", "Hyperping API key (or set HYPERPING_API_KEY)")
	outputDir        = flag.String("output", "./pingdom-migration", "Output directory for generated files")
	prefix           = flag.String("prefix", "", "Prefix for Terraform resource names")
	pingdomBaseURL   = flag.String("pingdom-base-url", "", "Pingdom API base URL (optional)")
	hyperpingBaseURL = flag.String("hyperping-base-url", "https://api.hyperping.io", "Hyperping API base URL")
	dryRun           = flag.Bool("dry-run", false, "Generate configs without creating resources in Hyperping")
	verbose          = flag.Bool("verbose", false, "Verbose output")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: migrate-pingdom [options]\n\n")
		fmt.Fprintf(os.Stderr, "Migrates Pingdom checks to Hyperping monitors.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Dry run (generate configs only)\n")
		fmt.Fprintf(os.Stderr, "  migrate-pingdom --dry-run --output=./migration\n\n")
		fmt.Fprintf(os.Stderr, "  # Full migration\n")
		fmt.Fprintf(os.Stderr, "  migrate-pingdom --output=./migration\n\n")
		fmt.Fprintf(os.Stderr, "  # With resource name prefix\n")
		fmt.Fprintf(os.Stderr, "  migrate-pingdom --prefix=pingdom_ --output=./migration\n\n")
	}

	os.Exit(run())
}

func run() int {
	flag.Parse()

	// Get API keys from flags or environment
	pingdomKey := *pingdomAPIKey
	if pingdomKey == "" {
		pingdomKey = os.Getenv("PINGDOM_API_KEY")
	}
	if pingdomKey == "" {
		pingdomKey = os.Getenv("PINGDOM_API_TOKEN")
	}

	hyperpingKey := *hyperpingAPIKey
	if hyperpingKey == "" {
		hyperpingKey = os.Getenv("HYPERPING_API_KEY")
	}

	if pingdomKey == "" {
		fmt.Fprintln(os.Stderr, "Error: Pingdom API key is required (--pingdom-api-key or PINGDOM_API_KEY)")
		return 1
	}

	if hyperpingKey == "" && !*dryRun {
		fmt.Fprintln(os.Stderr, "Error: Hyperping API key is required (--hyperping-api-key or HYPERPING_API_KEY)")
		fmt.Fprintln(os.Stderr, "Hint: Use --dry-run to generate configs without creating resources")
		return 1
	}

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		return 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Step 1: Fetch Pingdom checks
	log("Fetching Pingdom checks...")
	pingdomClient := createPingdomClient(pingdomKey)

	checks, err := pingdomClient.ListChecks(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching Pingdom checks: %v\n", err)
		return 1
	}

	log(fmt.Sprintf("Fetched %d checks from Pingdom", len(checks)))

	// Step 2: Convert checks
	log("Converting checks to Hyperping format...")
	checkConverter := converter.NewCheckConverter()
	results := make([]converter.ConversionResult, len(checks))

	supportedCount := 0
	for i, check := range checks {
		results[i] = checkConverter.Convert(check)
		if results[i].Supported {
			supportedCount++
		}
	}

	log(fmt.Sprintf("Converted %d/%d checks (%d unsupported)", supportedCount, len(checks), len(checks)-supportedCount))

	// Step 3: Generate Terraform configuration
	log("Generating Terraform configuration...")
	tfGen := generator.NewTerraformGenerator(*prefix)
	hclContent := tfGen.GenerateHCL(checks, results)

	hclPath := filepath.Join(*outputDir, "monitors.tf")
	if writeErr := os.WriteFile(hclPath, []byte(hclContent), 0o600); writeErr != nil {
		fmt.Fprintf(os.Stderr, "Error writing Terraform configuration: %v\n", writeErr)
		return 1
	}

	log(fmt.Sprintf("Terraform configuration written to %s", hclPath))

	// Step 4: Generate reports
	log("Generating migration report...")
	reporter := report.NewReporter()
	migrationReport := reporter.GenerateReport(checks, results)

	// JSON report
	jsonReport, err := reporter.GenerateJSONReport(migrationReport)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating JSON report: %v\n", err)
		return 1
	}

	jsonPath := filepath.Join(*outputDir, "report.json")
	if writeErr := os.WriteFile(jsonPath, []byte(jsonReport), 0o600); writeErr != nil {
		fmt.Fprintf(os.Stderr, "Error writing JSON report: %v\n", writeErr)
		return 1
	}

	// Text report
	textReport := reporter.GenerateTextReport(migrationReport)
	textPath := filepath.Join(*outputDir, "report.txt")
	if writeErr := os.WriteFile(textPath, []byte(textReport), 0o600); writeErr != nil {
		fmt.Fprintf(os.Stderr, "Error writing text report: %v\n", writeErr)
		return 1
	}

	// Manual steps markdown
	manualSteps := reporter.GenerateManualStepsMarkdown(migrationReport)
	manualPath := filepath.Join(*outputDir, "manual-steps.md")
	if writeErr := os.WriteFile(manualPath, []byte(manualSteps), 0o600); writeErr != nil {
		fmt.Fprintf(os.Stderr, "Error writing manual steps: %v\n", writeErr)
		return 1
	}

	log(fmt.Sprintf("Reports written to %s", *outputDir))

	// Step 5: Create resources in Hyperping (if not dry run)
	createdResources := make(map[int]string)

	if !*dryRun {
		log("Creating monitors in Hyperping...")
		hyperpingClient := createHyperpingClient(hyperpingKey)

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
				fmt.Fprintf(os.Stderr, "Warning: Failed to create monitor for check %d (%s): %v\n", check.ID, check.Name, err)
				continue
			}

			createdResources[check.ID] = monitor.UUID
			createdCount++

			if *verbose {
				log(fmt.Sprintf("Created monitor %s for check %d (%s)", monitor.UUID, check.ID, check.Name))
			}
		}

		log(fmt.Sprintf("Created %d monitors in Hyperping (%d errors)", createdCount, errorCount))
	}

	// Step 6: Generate import script
	log("Generating import script...")
	importGen := generator.NewImportGenerator(*prefix)
	importScript := importGen.GenerateImportScript(checks, results, createdResources)

	importPath := filepath.Join(*outputDir, "import.sh")
	if writeErr := os.WriteFile(importPath, []byte(importScript), 0o600); writeErr != nil {
		fmt.Fprintf(os.Stderr, "Error writing import script: %v\n", writeErr)
		return 1
	}

	// Make script executable
	if chmodErr := os.Chmod(importPath, 0o600); chmodErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to make import.sh executable: %v\n", chmodErr)
	}

	log(fmt.Sprintf("Import script written to %s", importPath))

	// Summary
	fmt.Println()
	fmt.Println("=================================================================")
	fmt.Println("Migration Complete!")
	fmt.Println("=================================================================")
	fmt.Println()
	fmt.Printf("Output directory: %s\n", *outputDir)
	fmt.Println()
	fmt.Println("Generated files:")
	fmt.Printf("  - %s (Terraform configuration)\n", filepath.Base(hclPath))
	fmt.Printf("  - %s (import script)\n", filepath.Base(importPath))
	fmt.Printf("  - %s (JSON report)\n", filepath.Base(jsonPath))
	fmt.Printf("  - %s (text report)\n", filepath.Base(textPath))
	fmt.Printf("  - %s (manual steps)\n", filepath.Base(manualPath))
	fmt.Println()

	if *dryRun {
		fmt.Println("DRY RUN: No resources were created in Hyperping")
		fmt.Println("Review the generated files and run without --dry-run to create resources")
	} else {
		fmt.Println("Next steps:")
		fmt.Println("  1. Review monitors.tf and adjust as needed")
		fmt.Println("  2. Run 'terraform init' and 'terraform plan'")
		fmt.Println("  3. Run './import.sh' to import resources into Terraform state")
		fmt.Println("  4. Review manual-steps.md for unsupported checks")
	}

	fmt.Println()

	// Print summary from report
	fmt.Printf("Summary: %d total checks, %d supported, %d unsupported\n",
		migrationReport.TotalChecks,
		migrationReport.SupportedChecks,
		migrationReport.UnsupportedChecks)

	if len(migrationReport.ManualSteps) > 0 {
		fmt.Printf("Manual steps required: %d (see manual-steps.md)\n", len(migrationReport.ManualSteps))
	}

	return 0
}

func createPingdomClient(apiKey string) *pingdom.Client {
	options := []pingdom.Option{}
	if *pingdomBaseURL != "" {
		options = append(options, pingdom.WithBaseURL(*pingdomBaseURL))
	}

	return pingdom.NewClient(apiKey, options...)
}

func createHyperpingClient(apiKey string) *client.Client {
	return client.NewClient(apiKey, client.WithBaseURL(*hyperpingBaseURL))
}

func log(msg string) {
	if *verbose || true {
		fmt.Fprintf(os.Stderr, "[migrate-pingdom] %s\n", msg)
	}
}
