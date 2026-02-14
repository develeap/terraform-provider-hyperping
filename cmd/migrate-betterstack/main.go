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
	"github.com/develeap/terraform-provider-hyperping/pkg/checkpoint"
	"github.com/develeap/terraform-provider-hyperping/pkg/recovery"
)

var (
	betterstackToken    = flag.String("betterstack-token", "", "Better Stack API token (or set BETTERSTACK_API_TOKEN)")
	hyperpingAPIKey     = flag.String("hyperping-api-key", "", "Hyperping API key (or set HYPERPING_API_KEY)")
	outputFile          = flag.String("output", "migrated-resources.tf", "Output Terraform configuration file")
	importScript        = flag.String("import-script", "import.sh", "Output import script file")
	reportFile          = flag.String("report", "migration-report.json", "Output migration report file")
	manualStepsFile     = flag.String("manual-steps", "manual-steps.md", "Output manual steps documentation")
	dryRun              = flag.Bool("dry-run", false, "Validate without creating files")
	validateTF          = flag.Bool("validate", false, "Run terraform validate on output")
	verbose             = flag.Bool("verbose", false, "Enable verbose logging")
	debug               = flag.Bool("debug", false, "Enable debug mode with detailed logging")
	resume              = flag.Bool("resume", false, "Resume from last checkpoint")
	resumeID            = flag.String("resume-id", "", "Resume from specific checkpoint ID")
	rollback            = flag.Bool("rollback", false, "Rollback migration (delete Hyperping resources)")
	rollbackID          = flag.String("rollback-id", "", "Rollback specific migration ID")
	rollbackForce       = flag.Bool("force", false, "Force rollback without confirmation")
	listCheckpointsFlag = flag.Bool("list-checkpoints", false, "List available checkpoints")
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
		fmt.Fprintf(os.Stderr, "  # Resume from last checkpoint\n")
		fmt.Fprintf(os.Stderr, "  migrate-betterstack --resume\n\n")
		fmt.Fprintf(os.Stderr, "  # Rollback migration\n")
		fmt.Fprintf(os.Stderr, "  migrate-betterstack --rollback --rollback-id=betterstack-20260213-120000\n\n")
		fmt.Fprintf(os.Stderr, "  # Debug mode with detailed logging\n")
		fmt.Fprintf(os.Stderr, "  migrate-betterstack --debug\n\n")
	}
	os.Exit(run())
}

func run() int {
	flag.Parse()

	// Create logger
	logger, err := recovery.NewLogger(*debug || *verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create logger: %v\n", err)
		return 1
	}
	defer logger.Close()

	if *debug && logger.GetLogPath() != "" {
		logger.Info("Debug logging enabled, log file: %s", logger.GetLogPath())
	}

	// Handle list checkpoints
	if *listCheckpointsFlag {
		return listCheckpoints(toolName)
	}

	// Get credentials from flags or environment
	bsToken := *betterstackToken
	if bsToken == "" {
		bsToken = os.Getenv("BETTERSTACK_API_TOKEN")
	}
	hpKey := *hyperpingAPIKey
	if hpKey == "" {
		hpKey = os.Getenv("HYPERPING_API_KEY")
	}

	// Handle rollback
	if *rollback {
		if hpKey == "" {
			fmt.Fprintln(os.Stderr, "Error: Hyperping API key is required for rollback")
			fmt.Fprintln(os.Stderr, "Set --hyperping-api-key flag or HYPERPING_API_KEY environment variable")
			return 1
		}

		migrationID := *rollbackID
		if migrationID == "" {
			// Find latest checkpoint
			//nolint:govet
			mgr, err := checkpoint.NewManager()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to create checkpoint manager: %v\n", err)
				return 1
			}
			//nolint:govet
			latest, err := mgr.FindLatest(toolName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				fmt.Fprintln(os.Stderr, "Use --rollback-id to specify a checkpoint or --list-checkpoints to see available checkpoints")
				return 1
			}
			migrationID = latest.MigrationID
		}

		return performRollback(migrationID, hpKey, *rollbackForce, logger)
	}

	if bsToken == "" {
		fmt.Fprintln(os.Stderr, "Error: Better Stack API token is required")
		fmt.Fprintln(os.Stderr, "Set --betterstack-token flag or BETTERSTACK_API_TOKEN environment variable")
		return 1
	}

	if hpKey == "" && !*dryRun {
		fmt.Fprintln(os.Stderr, "Error: Hyperping API key is required")
		fmt.Fprintln(os.Stderr, "Set --hyperping-api-key flag or HYPERPING_API_KEY environment variable")
		return 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Enhanced dry-run: validate API connectivity
	if *dryRun {
		logger.Info("Dry run mode: validating API connectivity...")
		validator := recovery.NewAPIValidator(logger)

		// Validate Better Stack API
		//nolint:govet
		bsValidation := validator.ValidateSourceAPI(ctx, "Better Stack", func(ctx context.Context) error {
			bsClient := betterstack.NewClient(bsToken)
			_, err := bsClient.FetchMonitors(ctx)
			return err
		})

		if !bsValidation.Valid {
			fmt.Fprintf(os.Stderr, "Error: %s\n", bsValidation.ErrorMessage)
			return 1
		}

		fmt.Fprintln(os.Stderr, "API validation successful")
	}

	// Initialize or resume migration state
	var state *migrationState
	migrationID := *resumeID

	if *resume || migrationID != "" {
		// Resume from checkpoint
		if migrationID == "" {
			//nolint:govet
			mgr, err := checkpoint.NewManager()
			if err != nil {
				logger.Error("Failed to create checkpoint manager: %v", err)
				return 1
			}
			//nolint:govet
			latest, err := mgr.FindLatest(toolName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: No checkpoint found to resume from\n")
				fmt.Fprintln(os.Stderr, "Use --list-checkpoints to see available checkpoints")
				return 1
			}
			migrationID = latest.MigrationID
		}

		//nolint:govet
		state, err = resumeFromCheckpoint(migrationID, logger)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to resume from checkpoint: %v\n", err)
			return 1
		}

		logger.Info("Resuming migration from checkpoint: %s", migrationID)
	}

	logger.Info("Starting Better Stack to Hyperping migration...")

	// Create Better Stack client
	bsClient := betterstack.NewClient(bsToken)

	// Note: Hyperping client not needed - tool generates Terraform files
	// Users will apply the Terraform config to create resources in Hyperping

	// Fetch Better Stack resources
	logger.Info("Fetching Better Stack monitors...")
	monitors, err := bsClient.FetchMonitors(ctx)
	if err != nil {
		logger.Error("Failed to fetch monitors: %v", err)
		fmt.Fprintf(os.Stderr, "Error fetching Better Stack monitors: %v\n", err)
		return 1
	}
	logger.Info("Found %d monitors", len(monitors))

	logger.Info("Fetching Better Stack heartbeats...")
	heartbeats, err := bsClient.FetchHeartbeats(ctx)
	if err != nil {
		logger.Error("Failed to fetch heartbeats: %v", err)
		fmt.Fprintf(os.Stderr, "Error fetching Better Stack heartbeats: %v\n", err)
		return 1
	}
	logger.Info("Found %d heartbeats", len(heartbeats))

	// Initialize state if not resuming
	totalResources := len(monitors) + len(heartbeats)
	if state == nil {
		migrationID = checkpoint.GenerateMigrationID(toolName)
		state, err = newMigrationState(migrationID, totalResources, logger)
		if err != nil {
			logger.Error("Failed to create migration state: %v", err)
			return 1
		}
		logger.Info("Created new migration: %s", migrationID)
	}

	// Convert resources with error handling
	logger.Info("Converting monitors to Hyperping format...")
	conv := converter.New()

	var convertedMonitors []converter.ConvertedMonitor
	var monitorIssues []converter.ConversionIssue

	for _, monitor := range monitors {
		monitorID := fmt.Sprintf("monitor-%s", monitor.ID)

		// Skip if already processed
		if state.isProcessed(monitorID) {
			logger.Debug("Skipping already processed monitor: %s", monitorID)
			continue
		}

		// Convert monitor
		converted, issues := conv.ConvertMonitors([]betterstack.Monitor{monitor})

		if len(converted) > 0 {
			convertedMonitors = append(convertedMonitors, converted...)
			state.markResourceProcessed(monitorID)
		} else {
			// Mark as failed
			errorMsg := "conversion failed"
			if len(issues) > 0 {
				errorMsg = issues[0].Message
			}
			state.markResourceFailed(monitorID, "monitor", monitor.Attributes.URL, errorMsg)
		}

		monitorIssues = append(monitorIssues, issues...)
	}

	logger.Info("Converting heartbeats to Hyperping format...")
	var convertedHealthchecks []converter.ConvertedHealthcheck
	var healthcheckIssues []converter.ConversionIssue

	for _, heartbeat := range heartbeats {
		heartbeatID := fmt.Sprintf("heartbeat-%s", heartbeat.ID)

		// Skip if already processed
		if state.isProcessed(heartbeatID) {
			logger.Debug("Skipping already processed heartbeat: %s", heartbeatID)
			continue
		}

		// Convert heartbeat
		converted, issues := conv.ConvertHeartbeats([]betterstack.Heartbeat{heartbeat})

		if len(converted) > 0 {
			convertedHealthchecks = append(convertedHealthchecks, converted...)
			state.markResourceProcessed(heartbeatID)
		} else {
			// Mark as failed
			errorMsg := "conversion failed"
			if len(issues) > 0 {
				errorMsg = issues[0].Message
			}
			state.markResourceFailed(heartbeatID, "heartbeat", heartbeat.Attributes.Name, errorMsg)
		}

		healthcheckIssues = append(healthcheckIssues, issues...)
	}

	// Save checkpoint after conversion
	state.saveCheckpoint()

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

		// Print failure report if any
		if failureReport := state.getFailureReport(); failureReport != "" {
			fmt.Fprintln(os.Stderr, failureReport)
		}

		return 0
	}

	// Write Terraform configuration
	logger.Debug("Writing Terraform configuration to %s", *outputFile)
	if err := os.WriteFile(*outputFile, []byte(tfConfig), 0o600); err != nil {
		logger.Error("Failed to write Terraform config: %v", err)
		fmt.Fprintf(os.Stderr, "Error writing Terraform config: %v\n", err)
		state.finalize(false)
		return 1
	}
	logger.Info("Generated %s", *outputFile)

	// Write import script
	logger.Debug("Writing import script to %s", *importScript)
	if err := os.WriteFile(*importScript, []byte(importScriptContent), 0o600); err != nil {
		logger.Error("Failed to write import script: %v", err)
		fmt.Fprintf(os.Stderr, "Error writing import script: %v\n", err)
		state.finalize(false)
		return 1
	}
	logger.Info("Generated %s", *importScript)

	// Write migration report
	logger.Debug("Writing migration report to %s", *reportFile)
	if err := os.WriteFile(*reportFile, []byte(migrationReport.JSON()), 0o600); err != nil {
		logger.Error("Failed to write migration report: %v", err)
		fmt.Fprintf(os.Stderr, "Error writing migration report: %v\n", err)
		state.finalize(false)
		return 1
	}
	logger.Info("Generated %s", *reportFile)

	// Write manual steps
	logger.Debug("Writing manual steps to %s", *manualStepsFile)
	if err := os.WriteFile(*manualStepsFile, []byte(manualSteps), 0o600); err != nil {
		logger.Error("Failed to write manual steps: %v", err)
		fmt.Fprintf(os.Stderr, "Error writing manual steps: %v\n", err)
		state.finalize(false)
		return 1
	}
	logger.Info("Generated %s", *manualStepsFile)

	// Finalize checkpoint
	hasFailures := state.checkpoint.Failed > 0
	state.finalize(!hasFailures)

	// Print summary
	fmt.Fprintln(os.Stderr, "\n=== Migration Complete ===")
	migrationReport.PrintSummary(os.Stderr)

	// Print failure report if any
	if failureReport := state.getFailureReport(); failureReport != "" {
		fmt.Fprintln(os.Stderr, failureReport)
		fmt.Fprintf(os.Stderr, "\nSome resources failed to convert. See details above.\n")
		fmt.Fprintf(os.Stderr, "Migration ID: %s\n", migrationID)
	}

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
		logger.Info("Validating Terraform configuration...")
		if err := gen.ValidateTerraform(*outputFile); err != nil {
			logger.Error("Terraform validation failed: %v", err)
			fmt.Fprintf(os.Stderr, "Terraform validation failed: %v\n", err)
			return 1
		}
		logger.Info("Terraform validation passed")
		fmt.Fprintln(os.Stderr, "Terraform validation passed!")
	}

	// Determine exit code based on failures
	if hasFailures {
		logger.Warn("Migration completed with %d failures", state.checkpoint.Failed)
		return 1
	}

	logger.Info("Migration completed successfully")
	return 0
}
