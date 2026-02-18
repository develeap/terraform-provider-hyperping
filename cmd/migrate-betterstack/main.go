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
	"github.com/develeap/terraform-provider-hyperping/pkg/migrationstate"
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
	formatJSON          = flag.Bool("format", false, "Output dry-run report as JSON (use with --dry-run)")
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

// migrationResult holds the generated content from a migration run.
type migrationResult struct {
	tfConfig              string
	importScriptContent   string
	manualSteps           string
	migrationReport       *report.Report
	monitorIssues         []converter.ConversionIssue
	healthcheckIssues     []converter.ConversionIssue
	convertedMonitors     []converter.ConvertedMonitor
	convertedHealthchecks []converter.ConvertedHealthcheck
}

// validateCredentials checks that required API credentials are present and returns
// the resolved bsToken and hpKey values (from flags or env vars).
func validateCredentials() (bsToken, hpKey string, code int) {
	bsToken = *betterstackToken
	if bsToken == "" {
		bsToken = os.Getenv("BETTERSTACK_API_TOKEN")
	}
	hpKey = *hyperpingAPIKey
	if hpKey == "" {
		hpKey = os.Getenv("HYPERPING_API_KEY")
	}
	return bsToken, hpKey, 0
}

// handleRollbackMode performs rollback when the --rollback flag is set.
// Returns (exitCode, handled). If handled is false, rollback was not requested.
func handleRollbackMode(hpKey string, logger *recovery.Logger) (int, bool) {
	if !*rollback {
		return 0, false
	}

	if hpKey == "" {
		fmt.Fprintln(os.Stderr, "Error: Hyperping API key is required for rollback")
		fmt.Fprintln(os.Stderr, "Set --hyperping-api-key flag or HYPERPING_API_KEY environment variable")
		return 1, true
	}

	migrationID := *rollbackID
	if migrationID == "" {
		mgr, err := checkpoint.NewManager()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to create checkpoint manager: %v\n", err)
			return 1, true
		}
		latest, err := mgr.FindLatest(toolName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			fmt.Fprintln(os.Stderr, "Use --rollback-id to specify a checkpoint or --list-checkpoints to see available checkpoints")
			return 1, true
		}
		migrationID = latest.MigrationID
	}

	return migrationstate.PerformRollback(migrationID, hpKey, *rollbackForce, logger), true
}

// validateSourceCredentials checks that source/dest credentials exist before a full migration.
func validateSourceCredentials(bsToken, hpKey string) int {
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
	return 0
}

// runDryValidation validates API connectivity in dry-run mode.
func runDryValidation(ctx context.Context, bsToken string, logger *recovery.Logger) int {
	logger.Info("Dry run mode: validating API connectivity...")
	validator := recovery.NewAPIValidator(logger)

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
	return 0
}

// resolveOrCreateState initialises a migration state, resuming from a checkpoint if requested.
func resolveOrCreateState(ctx context.Context, totalResources int, logger *recovery.Logger) (*migrationstate.State, string, error) {
	migID := *resumeID

	if *resume || migID != "" {
		if migID == "" {
			mgr, err := checkpoint.NewManager()
			if err != nil {
				return nil, "", fmt.Errorf("failed to create checkpoint manager: %w", err)
			}
			latest, err := mgr.FindLatest(toolName)
			if err != nil {
				return nil, "", fmt.Errorf("no checkpoint found to resume from")
			}
			migID = latest.MigrationID
		}

		state, err := migrationstate.Resume(migID, logger)
		if err != nil {
			return nil, "", fmt.Errorf("failed to resume from checkpoint: %w", err)
		}
		logger.Info("Resuming migration from checkpoint: %s", migID)
		return state, migID, nil
	}

	_ = ctx
	migID = checkpoint.GenerateMigrationID(toolName)
	state, err := migrationstate.New(toolName, migID, totalResources, logger)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create migration state: %w", err)
	}
	logger.Info("Created new migration: %s", migID)
	return state, migID, nil
}

// fetchBetterStackResources fetches monitors and heartbeats from Better Stack.
func fetchBetterStackResources(ctx context.Context, bsToken string, logger *recovery.Logger) ([]betterstack.Monitor, []betterstack.Heartbeat, error) {
	bsClient := betterstack.NewClient(bsToken)

	logger.Info("Fetching Better Stack monitors...")
	monitors, err := bsClient.FetchMonitors(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("error fetching Better Stack monitors: %w", err)
	}
	logger.Info("Found %d monitors", len(monitors))

	logger.Info("Fetching Better Stack heartbeats...")
	heartbeats, err := bsClient.FetchHeartbeats(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("error fetching Better Stack heartbeats: %w", err)
	}
	logger.Info("Found %d heartbeats", len(heartbeats))

	return monitors, heartbeats, nil
}

// convertResources converts all Better Stack resources to Hyperping format.
func convertResources(
	monitors []betterstack.Monitor,
	heartbeats []betterstack.Heartbeat,
	state *migrationstate.State,
	logger *recovery.Logger,
) ([]converter.ConvertedMonitor, []converter.ConvertedHealthcheck, []converter.ConversionIssue, []converter.ConversionIssue) {
	conv := converter.New()

	logger.Info("Converting monitors to Hyperping format...")
	convertedMonitors, monitorIssues := convertMonitorList(monitors, conv, state, logger)

	logger.Info("Converting heartbeats to Hyperping format...")
	convertedHealthchecks, healthcheckIssues := convertHeartbeatList(heartbeats, conv, state, logger)

	return convertedMonitors, convertedHealthchecks, monitorIssues, healthcheckIssues
}

// convertMonitorList converts a slice of monitors, updating state as it goes.
func convertMonitorList(
	monitors []betterstack.Monitor,
	conv *converter.Converter,
	state *migrationstate.State,
	logger *recovery.Logger,
) ([]converter.ConvertedMonitor, []converter.ConversionIssue) {
	var convertedMonitors []converter.ConvertedMonitor
	var monitorIssues []converter.ConversionIssue

	for _, monitor := range monitors {
		monitorID := fmt.Sprintf("monitor-%s", monitor.ID)
		if state.IsProcessed(monitorID) {
			logger.Debug("Skipping already processed monitor: %s", monitorID)
			continue
		}

		converted, issues := conv.ConvertMonitors([]betterstack.Monitor{monitor})
		if len(converted) > 0 {
			convertedMonitors = append(convertedMonitors, converted...)
			state.MarkResourceProcessed(monitorID)
		} else {
			errorMsg := "conversion failed"
			if len(issues) > 0 {
				errorMsg = issues[0].Message
			}
			state.MarkResourceFailed(monitorID, "monitor", monitor.Attributes.URL, errorMsg)
		}
		monitorIssues = append(monitorIssues, issues...)
	}

	return convertedMonitors, monitorIssues
}

// convertHeartbeatList converts a slice of heartbeats, updating state as it goes.
func convertHeartbeatList(
	heartbeats []betterstack.Heartbeat,
	conv *converter.Converter,
	state *migrationstate.State,
	logger *recovery.Logger,
) ([]converter.ConvertedHealthcheck, []converter.ConversionIssue) {
	var convertedHealthchecks []converter.ConvertedHealthcheck
	var healthcheckIssues []converter.ConversionIssue

	for _, heartbeat := range heartbeats {
		heartbeatID := fmt.Sprintf("heartbeat-%s", heartbeat.ID)
		if state.IsProcessed(heartbeatID) {
			logger.Debug("Skipping already processed heartbeat: %s", heartbeatID)
			continue
		}

		converted, issues := conv.ConvertHeartbeats([]betterstack.Heartbeat{heartbeat})
		if len(converted) > 0 {
			convertedHealthchecks = append(convertedHealthchecks, converted...)
			state.MarkResourceProcessed(heartbeatID)
		} else {
			errorMsg := "conversion failed"
			if len(issues) > 0 {
				errorMsg = issues[0].Message
			}
			state.MarkResourceFailed(heartbeatID, "heartbeat", heartbeat.Attributes.Name, errorMsg)
		}
		healthcheckIssues = append(healthcheckIssues, issues...)
	}

	return convertedHealthchecks, healthcheckIssues
}

// buildMigrationResult generates all output content from converted resources.
func buildMigrationResult(
	monitors []betterstack.Monitor,
	heartbeats []betterstack.Heartbeat,
	convertedMonitors []converter.ConvertedMonitor,
	convertedHealthchecks []converter.ConvertedHealthcheck,
	monitorIssues []converter.ConversionIssue,
	healthcheckIssues []converter.ConversionIssue,
) *migrationResult {
	gen := generator.New()
	return &migrationResult{
		tfConfig:              gen.GenerateTerraform(convertedMonitors, convertedHealthchecks),
		importScriptContent:   gen.GenerateImportScript(convertedMonitors, convertedHealthchecks),
		manualSteps:           gen.GenerateManualSteps(monitorIssues, healthcheckIssues),
		migrationReport:       report.Generate(monitors, heartbeats, convertedMonitors, convertedHealthchecks, monitorIssues, healthcheckIssues),
		monitorIssues:         monitorIssues,
		healthcheckIssues:     healthcheckIssues,
		convertedMonitors:     convertedMonitors,
		convertedHealthchecks: convertedHealthchecks,
	}
}

// runDryRunOutput prints dry-run preview and returns the exit code.
func runDryRunOutput(
	monitors []betterstack.Monitor,
	heartbeats []betterstack.Heartbeat,
	result *migrationResult,
	state *migrationstate.State,
) int {
	dryRunReport := buildDryRunReport(
		monitors,
		heartbeats,
		result.convertedMonitors,
		result.convertedHealthchecks,
		result.monitorIssues,
		result.healthcheckIssues,
		result.tfConfig,
		result.importScriptContent,
		result.manualSteps,
	)

	printEnhancedDryRun(dryRunReport, *verbose, *formatJSON)

	if failureReport := state.GetFailureReport(); failureReport != "" {
		fmt.Fprintln(os.Stderr, "\n"+failureReport)
	}

	return 0
}

// writeOutputFiles writes all generated files to disk.
func writeOutputFiles(result *migrationResult, logger *recovery.Logger) (int, error) {
	type fileWrite struct {
		path    string
		content []byte
		logMsg  string
	}

	writes := []fileWrite{
		{*outputFile, []byte(result.tfConfig), *outputFile},
		{*importScript, []byte(result.importScriptContent), *importScript},
		{*reportFile, []byte(result.migrationReport.JSON()), *reportFile},
		{*manualStepsFile, []byte(result.manualSteps), *manualStepsFile},
	}

	for _, w := range writes {
		logger.Debug("Writing %s", w.path)
		if err := os.WriteFile(w.path, w.content, 0o600); err != nil {
			logger.Error("Failed to write %s: %v", w.path, err)
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", w.path, err)
			return 1, err
		}
		logger.Info("Generated %s", w.logMsg)
	}
	return 0, nil
}

// printSuccessSummary prints the post-migration summary to stderr.
func printSuccessSummary(result *migrationResult, state *migrationstate.State, migrationID string) {
	fmt.Fprintln(os.Stderr, "\n=== Migration Complete ===")
	result.migrationReport.PrintSummary(os.Stderr)

	if failureReport := state.GetFailureReport(); failureReport != "" {
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
}

// runTerraformValidation optionally validates the written Terraform file.
func runTerraformValidation(logger *recovery.Logger) int {
	if !*validateTF {
		return 0
	}
	gen := generator.New()
	logger.Info("Validating Terraform configuration...")
	if err := gen.ValidateTerraform(*outputFile); err != nil {
		logger.Error("Terraform validation failed: %v", err)
		fmt.Fprintf(os.Stderr, "Terraform validation failed: %v\n", err)
		return 1
	}
	logger.Info("Terraform validation passed")
	fmt.Fprintln(os.Stderr, "Terraform validation passed!")
	return 0
}

// logFatalErr logs an error to both the structured logger and stderr, returning exit code 1.
func logFatalErr(logger *recovery.Logger, err error) int {
	logger.Error("%v", err)
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	return 1
}

// logDebugPath logs the debug log path when debug mode is active.
func logDebugPath(logger *recovery.Logger) {
	if *debug && logger.GetLogPath() != "" {
		logger.Info("Debug logging enabled, log file: %s", logger.GetLogPath())
	}
}

// runConversionAndOutput converts resources and writes output files, returning an exit code.
func runConversionAndOutput(
	monitors []betterstack.Monitor,
	heartbeats []betterstack.Heartbeat,
	state *migrationstate.State,
	migrationID string,
	logger *recovery.Logger,
) int {
	logger.Info("Starting Better Stack to Hyperping migration...")
	convertedMonitors, convertedHealthchecks, monitorIssues, healthcheckIssues := convertResources(monitors, heartbeats, state, logger)
	state.SaveCheckpoint()

	result := buildMigrationResult(monitors, heartbeats, convertedMonitors, convertedHealthchecks, monitorIssues, healthcheckIssues)

	if *dryRun {
		return runDryRunOutput(monitors, heartbeats, result, state)
	}

	if code, writeErr := writeOutputFiles(result, logger); writeErr != nil {
		state.Finalize(false)
		return code
	}

	hasFailures := state.Checkpoint.Failed > 0
	state.Finalize(!hasFailures)
	printSuccessSummary(result, state, migrationID)
	return finalizeMigration(hasFailures, state, logger)
}

// finalizeMigration runs optional validation and returns the final exit code.
func finalizeMigration(hasFailures bool, state *migrationstate.State, logger *recovery.Logger) int {
	if code := runTerraformValidation(logger); code != 0 {
		return code
	}
	if hasFailures {
		logger.Warn("Migration completed with %d failures", state.Checkpoint.Failed)
		return 1
	}
	logger.Info("Migration completed successfully")
	return 0
}

func run() int {
	flag.Parse()

	logger, err := recovery.NewLogger(*debug || *verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create logger: %v\n", err)
		return 1
	}
	defer logger.Close()

	logDebugPath(logger)

	if shouldUseInteractive() {
		return runInteractive(logger)
	}

	if *listCheckpointsFlag {
		return migrationstate.ListCheckpoints(toolName)
	}

	bsToken, hpKey, _ := validateCredentials()

	if code, handled := handleRollbackMode(hpKey, logger); handled {
		return code
	}

	if code := validateSourceCredentials(bsToken, hpKey); code != 0 {
		return code
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if *dryRun {
		if code := runDryValidation(ctx, bsToken, logger); code != 0 {
			return code
		}
	}

	monitors, heartbeats, err := fetchBetterStackResources(ctx, bsToken, logger)
	if err != nil {
		return logFatalErr(logger, err)
	}

	state, migrationID, err := resolveOrCreateState(ctx, len(monitors)+len(heartbeats), logger)
	if err != nil {
		return logFatalErr(logger, err)
	}

	return runConversionAndOutput(monitors, heartbeats, state, migrationID, logger)
}
