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
	"github.com/develeap/terraform-provider-hyperping/pkg/checkpoint"
	"github.com/develeap/terraform-provider-hyperping/pkg/migrationstate"
	"github.com/develeap/terraform-provider-hyperping/pkg/recovery"
)

var (
	pingdomAPIKey       = flag.String("pingdom-api-key", "", "Pingdom API token (or set PINGDOM_API_KEY)")
	hyperpingAPIKey     = flag.String("hyperping-api-key", "", "Hyperping API key (or set HYPERPING_API_KEY)")
	outputDir           = flag.String("output", "./pingdom-migration", "Output directory for generated files")
	prefix              = flag.String("prefix", "", "Prefix for Terraform resource names")
	pingdomBaseURL      = flag.String("pingdom-base-url", "", "Pingdom API base URL (optional)")
	hyperpingBaseURL    = flag.String("hyperping-base-url", "https://api.hyperping.io", "Hyperping API base URL")
	dryRun              = flag.Bool("dry-run", false, "Generate configs without creating resources in Hyperping")
	verbose             = flag.Bool("verbose", false, "Verbose output")
	resume              = flag.Bool("resume", false, "Resume from last checkpoint")
	resumeID            = flag.String("resume-id", "", "Resume from specific checkpoint ID")
	rollback            = flag.Bool("rollback", false, "Rollback migration (delete Hyperping resources)")
	rollbackID          = flag.String("rollback-id", "", "Rollback specific migration ID")
	rollbackForce       = flag.Bool("force", false, "Force rollback without confirmation")
	listCheckpointsFlag = flag.Bool("list-checkpoints", false, "List available checkpoints")
)

// pingdomRunner holds resolved configuration for a non-interactive run.
type pingdomRunner struct {
	pingdomKey   string
	hyperpingKey string
	ctx          context.Context
	cancel       context.CancelFunc
	state        *migrationstate.State
	migrationID  string
}

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
		fmt.Fprintf(os.Stderr, "  # Resume from last checkpoint\n")
		fmt.Fprintf(os.Stderr, "  migrate-pingdom --resume\n\n")
		fmt.Fprintf(os.Stderr, "  # Rollback migration\n")
		fmt.Fprintf(os.Stderr, "  migrate-pingdom --rollback --rollback-id=pingdom-20260213-120000\n\n")
	}

	os.Exit(run())
}

func run() int {
	flag.Parse()

	if shouldUseInteractive() {
		return runInteractive()
	}

	if *listCheckpointsFlag {
		return migrationstate.ListCheckpoints(toolName)
	}

	if *rollback {
		return handleRollback()
	}

	r, exitCode := newPingdomRunner()
	if exitCode != 0 {
		return exitCode
	}
	defer r.cancel()

	checks, results, exitCode := r.fetchAndConvert()
	if exitCode != 0 {
		return exitCode
	}

	reporter := report.NewReporter()
	migrationReport := reporter.GenerateReport(checks, results)

	if exitCode := r.writeReports(reporter, migrationReport); exitCode != 0 {
		return exitCode
	}

	createdResources := r.createHyperpingResources(checks, results)

	if exitCode := r.writeImportScript(checks, results, createdResources); exitCode != 0 {
		return exitCode
	}

	if r.state != nil {
		hasFailures := r.state.Checkpoint.Failed > 0
		r.state.Finalize(!hasFailures)
		if failureReport := r.state.GetFailureReport(); failureReport != "" {
			fmt.Fprintln(os.Stderr, failureReport)
		}
	}

	printRunSummary(migrationReport)
	return 0
}

// handleRollback resolves the migration ID and delegates to the shared rollback implementation.
func handleRollback() int {
	hpKey := *hyperpingAPIKey
	if hpKey == "" {
		hpKey = os.Getenv("HYPERPING_API_KEY")
	}
	if hpKey == "" {
		fmt.Fprintln(os.Stderr, "Error: Hyperping API key is required for rollback")
		fmt.Fprintln(os.Stderr, "Set --hyperping-api-key flag or HYPERPING_API_KEY environment variable")
		return 1
	}

	logger, err := recovery.NewLogger(false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create logger: %v\n", err)
		return 1
	}
	defer logger.Close()

	migID := *rollbackID
	if migID == "" {
		mgr, mgrErr := checkpoint.NewManager()
		if mgrErr != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to create checkpoint manager: %v\n", mgrErr)
			return 1
		}
		latest, latestErr := mgr.FindLatest(toolName)
		if latestErr != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", latestErr)
			fmt.Fprintln(os.Stderr, "Use --rollback-id to specify a checkpoint or --list-checkpoints to see available checkpoints")
			return 1
		}
		migID = latest.MigrationID
	}

	return migrationstate.PerformRollback(migID, hpKey, *rollbackForce, logger)
}

// newPingdomRunner validates flags, resolves API keys, sets up the context, and initialises state.
func newPingdomRunner() (*pingdomRunner, int) {
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
		return nil, 1
	}

	if hyperpingKey == "" && !*dryRun {
		fmt.Fprintln(os.Stderr, "Error: Hyperping API key is required (--hyperping-api-key or HYPERPING_API_KEY)")
		fmt.Fprintln(os.Stderr, "Hint: Use --dry-run to generate configs without creating resources")
		return nil, 1
	}

	if err := os.MkdirAll(*outputDir, 0o700); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		return nil, 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)

	r := &pingdomRunner{
		pingdomKey:   pingdomKey,
		hyperpingKey: hyperpingKey,
		ctx:          ctx,
		cancel:       cancel,
	}

	if err := r.initState(); err != nil {
		cancel()
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return nil, 1
	}

	return r, 0
}

// initState initialises or resumes migration state.
func (r *pingdomRunner) initState() error {
	logger, err := recovery.NewLogger(false)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	migID := *resumeID
	if *resume || migID != "" {
		if migID == "" {
			mgr, mgrErr := checkpoint.NewManager()
			if mgrErr != nil {
				logger.Close()
				return fmt.Errorf("failed to create checkpoint manager: %w", mgrErr)
			}
			latest, latestErr := mgr.FindLatest(toolName)
			if latestErr != nil {
				logger.Close()
				return fmt.Errorf("no checkpoint found to resume from")
			}
			migID = latest.MigrationID
		}
		state, stateErr := migrationstate.Resume(migID, logger)
		if stateErr != nil {
			logger.Close()
			return fmt.Errorf("failed to resume from checkpoint: %w", stateErr)
		}
		r.state = state
		r.migrationID = migID
		return nil
	}

	migID = checkpoint.GenerateMigrationID(toolName)
	// totalResources will be updated after fetch; use 0 as placeholder
	state, stateErr := migrationstate.New(toolName, migID, 0, logger)
	if stateErr != nil {
		logger.Close()
		return fmt.Errorf("failed to create migration state: %w", stateErr)
	}
	r.state = state
	r.migrationID = migID
	return nil
}

// fetchAndConvert fetches Pingdom checks and converts them to Hyperping format.
func (r *pingdomRunner) fetchAndConvert() ([]pingdom.Check, []converter.ConversionResult, int) {
	log("Fetching Pingdom checks...")
	pingdomClient := createPingdomClient(r.pingdomKey)

	checks, err := pingdomClient.ListChecks(r.ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching Pingdom checks: %v\n", err)
		return nil, nil, 1
	}
	log(fmt.Sprintf("Fetched %d checks from Pingdom", len(checks)))

	if r.state != nil {
		r.state.Checkpoint.TotalResources = len(checks)
	}

	log("Converting checks to Hyperping format...")
	checkConverter := converter.NewCheckConverter()
	results := make([]converter.ConversionResult, len(checks))
	supportedCount := 0
	for i, check := range checks {
		checkID := fmt.Sprintf("check-%d", check.ID)
		if r.state != nil && r.state.IsProcessed(checkID) {
			log(fmt.Sprintf("Skipping already processed check: %s", checkID))
			results[i] = checkConverter.Convert(check)
			if results[i].Supported {
				supportedCount++
			}
			continue
		}

		results[i] = checkConverter.Convert(check)
		if results[i].Supported {
			supportedCount++
		}

		if r.state != nil {
			if results[i].Supported {
				r.state.MarkResourceProcessed(checkID)
			} else {
				r.state.MarkResourceFailed(checkID, "check", check.Name, "unsupported check type")
			}
		}
	}
	log(fmt.Sprintf("Converted %d/%d checks (%d unsupported)", supportedCount, len(checks), len(checks)-supportedCount))

	if r.state != nil {
		r.state.SaveCheckpoint()
	}

	log("Generating Terraform configuration...")
	tfGen := generator.NewTerraformGenerator(*prefix)
	hclContent := tfGen.GenerateHCL(checks, results)

	hclPath := filepath.Join(*outputDir, "monitors.tf")
	if writeErr := os.WriteFile(hclPath, []byte(hclContent), 0o600); writeErr != nil {
		fmt.Fprintf(os.Stderr, "Error writing Terraform configuration: %v\n", writeErr)
		return nil, nil, 1
	}
	log(fmt.Sprintf("Terraform configuration written to %s", hclPath))

	return checks, results, 0
}

// writeReports generates and writes all report files.
func (r *pingdomRunner) writeReports(reporter *report.Reporter, migrationReport *report.MigrationReport) int {
	log("Generating migration report...")

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

	textReport := reporter.GenerateTextReport(migrationReport)
	textPath := filepath.Join(*outputDir, "report.txt")
	if writeErr := os.WriteFile(textPath, []byte(textReport), 0o600); writeErr != nil {
		fmt.Fprintf(os.Stderr, "Error writing text report: %v\n", writeErr)
		return 1
	}

	manualSteps := reporter.GenerateManualStepsMarkdown(migrationReport)
	manualPath := filepath.Join(*outputDir, "manual-steps.md")
	if writeErr := os.WriteFile(manualPath, []byte(manualSteps), 0o600); writeErr != nil {
		fmt.Fprintf(os.Stderr, "Error writing manual steps: %v\n", writeErr)
		return 1
	}

	log(fmt.Sprintf("Reports written to %s", *outputDir))
	return 0
}

// createHyperpingResources creates monitors in Hyperping (skipped in dry-run mode).
func (r *pingdomRunner) createHyperpingResources(checks []pingdom.Check, results []converter.ConversionResult) map[int]string {
	createdResources := make(map[int]string)
	if *dryRun {
		return createdResources
	}

	log("Creating monitors in Hyperping...")
	hyperpingClient := createHyperpingClient(r.hyperpingKey)
	createdCount := 0
	errorCount := 0

	for i, check := range checks {
		result := results[i]
		if !result.Supported || result.Monitor == nil {
			continue
		}

		monitor, err := hyperpingClient.CreateMonitor(r.ctx, *result.Monitor)
		if err != nil {
			errorCount++
			fmt.Fprintf(os.Stderr, "Warning: Failed to create monitor for check %d (%s): %v\n", check.ID, check.Name, err)
			continue
		}

		createdResources[check.ID] = monitor.UUID
		if r.state != nil {
			r.state.AddHyperpingResource(monitor.UUID, "monitor")
		}
		createdCount++

		if *verbose {
			log(fmt.Sprintf("Created monitor %s for check %d (%s)", monitor.UUID, check.ID, check.Name))
		}
	}

	log(fmt.Sprintf("Created %d monitors in Hyperping (%d errors)", createdCount, errorCount))
	return createdResources
}

// writeImportScript generates and writes the import shell script.
func (r *pingdomRunner) writeImportScript(checks []pingdom.Check, results []converter.ConversionResult, createdResources map[int]string) int {
	log("Generating import script...")
	importGen := generator.NewImportGenerator(*prefix)
	importScriptContent := importGen.GenerateImportScript(checks, results, createdResources)

	importPath := filepath.Join(*outputDir, "import.sh")
	if writeErr := os.WriteFile(importPath, []byte(importScriptContent), 0o700); writeErr != nil { //nolint:gosec // G306: import.sh must be executable (0700) to run as a shell script
		fmt.Fprintf(os.Stderr, "Error writing import script: %v\n", writeErr)
		return 1
	}

	if chmodErr := os.Chmod(importPath, 0o700); chmodErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to make import.sh executable: %v\n", chmodErr)
	}

	log(fmt.Sprintf("Import script written to %s", importPath))
	return 0
}

// printRunSummary prints the final migration summary and next steps.
func printRunSummary(migrationReport *report.MigrationReport) {
	hclPath := filepath.Join(*outputDir, "monitors.tf")
	importPath := filepath.Join(*outputDir, "import.sh")
	jsonPath := filepath.Join(*outputDir, "report.json")
	textPath := filepath.Join(*outputDir, "report.txt")
	manualPath := filepath.Join(*outputDir, "manual-steps.md")

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
	fmt.Printf("Summary: %d total checks, %d supported, %d unsupported\n",
		migrationReport.TotalChecks,
		migrationReport.SupportedChecks,
		migrationReport.UnsupportedChecks)

	if len(migrationReport.ManualSteps) > 0 {
		fmt.Printf("Manual steps required: %d (see manual-steps.md)\n", len(migrationReport.ManualSteps))
	}
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
	if *verbose {
		fmt.Fprintf(os.Stderr, "[migrate-pingdom] %s\n", msg)
	}
}
