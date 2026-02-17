// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

// import-generator generates Terraform import commands and HCL configurations
// from existing Hyperping resources.
//
// Usage:
//
//	export HYPERPING_API_KEY="sk_your_api_key"
//	go run ./tools/cmd/import-generator
//
// Or with flags:
//
//	go run ./tools/cmd/import-generator -format=hcl -output=imported.tf
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

var (
	// Original flags
	outputFormat    = flag.String("format", "both", "Output format: import, hcl, both, or script")
	outputFile      = flag.String("output", "", "Output file (default: stdout)")
	resources       = flag.String("resources", "all", "Resources to import: all, monitors, healthchecks, statuspages, incidents, maintenance, outages")
	prefix          = flag.String("prefix", "", "Prefix for Terraform resource names (e.g., 'prod_')")
	baseURL         = flag.String("base-url", "https://api.hyperping.io", "Hyperping API base URL")
	validate        = flag.Bool("validate", false, "Validate resources without generating output")
	progress        = flag.Bool("progress", false, "Show progress indicators")
	continueOnError = flag.Bool("continue-on-error", false, "Continue on errors instead of failing")

	// Filtering flags
	filterName    = flag.String("filter-name", "", "Filter resources by name (regex pattern)")
	filterExclude = flag.String("filter-exclude", "", "Exclude resources by name (regex pattern)")
	filterType    = flag.String("filter-type", "", "Filter by resource type (e.g., hyperping_monitor)")
	dryRun        = flag.Bool("dry-run", false, "Show what would be imported without executing")

	// Parallel execution flags
	parallel   = flag.Int("parallel", 5, "Number of concurrent import workers (0=sequential, max=20)")
	sequential = flag.Bool("sequential", false, "Disable parallel execution (same as --parallel=0)")

	// Drift detection flags
	detectDrift     = flag.Bool("detect-drift", false, "Run terraform plan before import to detect drift")
	abortOnDrift    = flag.Bool("abort-on-drift", false, "Abort if drift is detected (requires --detect-drift)")
	refreshFirst    = flag.Bool("refresh-first", false, "Refresh state before drift detection")
	postImportCheck = flag.Bool("post-import-check", false, "Verify no drift after import")

	// Checkpoint/resume flags
	checkpointFile = flag.String("checkpoint-file", ".import-checkpoint", "Path to checkpoint file")
	resume         = flag.Bool("resume", false, "Resume from last checkpoint")
	noCheckpoint   = flag.Bool("no-checkpoint", false, "Disable checkpointing")

	// Rollback flags
	rollback     = flag.Bool("rollback", false, "Rollback previous import (remove from state)")
	rollbackFile = flag.String("rollback-file", ".import-log", "Path to import log for rollback")
	rollbackPlan = flag.Bool("rollback-plan", false, "Show rollback plan without executing")

	// Output flags
	verbose = flag.Bool("verbose", false, "Enable verbose output")
	quiet   = flag.Bool("quiet", false, "Minimal output (errors only)")

	// Execution mode flag
	execute = flag.Bool("execute", false, "Execute terraform imports (default: generate commands only)")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: import-generator [options]\n\n")
		fmt.Fprintf(os.Stderr, "Generates and optionally executes Terraform import commands from Hyperping resources.\n\n")
		fmt.Fprintf(os.Stderr, "MODES:\n")
		fmt.Fprintf(os.Stderr, "  Generation (default): Generate import commands and HCL\n")
		fmt.Fprintf(os.Stderr, "  Execution: Execute imports with --execute flag\n")
		fmt.Fprintf(os.Stderr, "  Rollback: Remove imported resources with --rollback\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEXAMPLES:\n")
		fmt.Fprintf(os.Stderr, "  # Generate import commands for PROD resources\n")
		fmt.Fprintf(os.Stderr, "  import-generator --filter-name=\"PROD-.*\"\n\n")
		fmt.Fprintf(os.Stderr, "  # Execute parallel import with drift detection\n")
		fmt.Fprintf(os.Stderr, "  import-generator --execute --parallel=10 --detect-drift\n\n")
		fmt.Fprintf(os.Stderr, "  # Resume interrupted import\n")
		fmt.Fprintf(os.Stderr, "  import-generator --execute --resume\n\n")
		fmt.Fprintf(os.Stderr, "  # Rollback previous import\n")
		fmt.Fprintf(os.Stderr, "  import-generator --rollback\n\n")
		fmt.Fprintf(os.Stderr, "  # Dry run to see what would be imported\n")
		fmt.Fprintf(os.Stderr, "  import-generator --dry-run --filter-type=hyperping_monitor\n\n")
	}
	os.Exit(run())
}

func run() int {
	flag.Parse()

	// Handle rollback mode first (doesn't need API key)
	if *rollback || *rollbackPlan {
		return runRollback()
	}

	// Validate flags
	if err := validateFlags(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Check API key
	apiKey := os.Getenv("HYPERPING_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: HYPERPING_API_KEY environment variable is required")
		return 1
	}

	// Create client
	c := client.NewClient(apiKey, client.WithBaseURL(*baseURL))

	// Set timeout based on execution mode
	timeout := 5 * time.Minute
	if *execute {
		timeout = 30 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create filter config
	filterConfig, err := NewFilterConfig(*filterName, *filterExclude, *filterType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating filter: %v\n", err)
		return 1
	}

	// Create generator
	gen := &Generator{
		client:          c,
		prefix:          *prefix,
		resources:       parseResources(*resources),
		showProgress:    *progress || *execute,
		continueOnError: *continueOnError,
		filterConfig:    filterConfig,
	}

	// Handle validation mode
	if *validate {
		return runValidation(ctx, gen)
	}

	// Handle execution mode
	if *execute {
		return runExecution(ctx, gen, filterConfig)
	}

	// Generate output (default mode)
	return runGeneration(ctx, gen)
}

func runValidation(ctx context.Context, gen *Generator) int {
	fmt.Fprintln(os.Stderr, "Validating resources...")

	result := gen.Validate(ctx)

	// Print results
	result.Print(os.Stderr)

	if !result.IsValid() {
		return 1
	}

	return 0
}

func parseResources(s string) []string {
	if s == "all" {
		return []string{"monitors", "healthchecks", "statuspages", "incidents", "maintenance", "outages"}
	}
	return strings.Split(s, ",")
}

func validateFlags() error {
	if *abortOnDrift && !*detectDrift {
		return fmt.Errorf("--abort-on-drift requires --detect-drift")
	}

	if *resume && *noCheckpoint {
		return fmt.Errorf("--resume and --no-checkpoint are mutually exclusive")
	}

	if *parallel < 0 || *parallel > maxWorkers {
		return fmt.Errorf("--parallel must be between 0 and %d", maxWorkers)
	}

	if *sequential && *parallel != 5 {
		return fmt.Errorf("--sequential and --parallel are mutually exclusive")
	}

	if *quiet && *verbose {
		return fmt.Errorf("--quiet and --verbose are mutually exclusive")
	}

	return nil
}

func runGeneration(ctx context.Context, gen *Generator) int {
	// Generate output
	output, err := gen.Generate(ctx, *outputFormat)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating output: %v\n", err)
		return 1
	}

	// Write output
	if *outputFile != "" {
		if err := os.WriteFile(*outputFile, []byte(output), 0o600); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			return 1
		}
		fmt.Fprintf(os.Stderr, "Output written to %s\n", *outputFile)

		// Make script executable if format is script
		if *outputFormat == "script" {
			if err := os.Chmod(*outputFile, 0o755); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to make script executable: %v\n", err)
			}
		}
	} else {
		fmt.Print(output)
	}

	return 0
}

func runExecution(ctx context.Context, gen *Generator, filterConfig *FilterConfig) int {
	if !*quiet {
		printBanner()
	}

	// Resolve checkpoint and optionally resume
	jobs, code := prepareImportJobs(ctx, gen, filterConfig)
	if code != 0 {
		return code
	}
	if jobs == nil {
		// nil jobs means "done early" (zero jobs or cancelled)
		return 0
	}

	if !*quiet {
		printImportPlan(jobs, filterConfig)
	}

	if *dryRun {
		fmt.Println("\n[DRY RUN] No imports will be executed")
		return 0
	}

	summary, err := executeImports(ctx, jobs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Import failed: %v\n", err)
		return 1
	}

	return finalizeExecution(ctx, summary)
}

// prepareImportJobs handles checkpoint/resume, drift detection, resource fetch,
// and job filtering. Returns (nil, 0) when execution should stop with success.
func prepareImportJobs(ctx context.Context, gen *Generator, filterConfig *FilterConfig) ([]ImportJob, int) {
	checkpointMgr := NewCheckpointManager(*checkpointFile, !*noCheckpoint)

	if code := runDriftDetection(ctx); code != 0 {
		return nil, code
	}

	checkpoint, code := loadCheckpointIfResuming(checkpointMgr)
	if code != 0 {
		return nil, code
	}

	data, err := gen.fetchResources(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching resources: %v\n", err)
		return nil, 1
	}

	jobs := buildImportJobs(data, gen.prefix, filterConfig)
	if len(jobs) == 0 {
		fmt.Println("No resources to import")
		return nil, 0
	}

	if checkpoint != nil {
		jobs = FilterJobsForResume(jobs, checkpoint)
		if len(jobs) == 0 {
			fmt.Println("All resources already imported")
			return nil, 0
		}
	}

	return jobs, 0
}

// runDriftDetection performs pre-import drift check when enabled.
func runDriftDetection(ctx context.Context) int {
	if !*detectDrift {
		return 0
	}
	opts := DriftDetectionOptions{
		Enabled:      true,
		AbortOnDrift: *abortOnDrift,
		Verbose:      *verbose,
		RefreshFirst: *refreshFirst,
	}
	if err := RunPreImportDriftCheck(ctx, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Drift check failed: %v\n", err)
		return 1
	}
	return 0
}

// loadCheckpointIfResuming loads an existing checkpoint when --resume is set.
func loadCheckpointIfResuming(mgr *CheckpointManager) (*ImportCheckpoint, int) {
	if !*resume {
		return nil, 0
	}
	if !mgr.Exists() {
		fmt.Fprintln(os.Stderr, "Error: No checkpoint found to resume from")
		return nil, 1
	}
	checkpoint, err := mgr.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading checkpoint: %v\n", err)
		return nil, 1
	}
	if !*quiet {
		if !PromptForResume(checkpoint) {
			fmt.Println("Resume cancelled")
			return nil, 0
		}
	}
	return checkpoint, 0
}

// executeImports runs either sequential or parallel import and returns the summary.
func executeImports(ctx context.Context, jobs []ImportJob) (*ImportSummary, error) {
	workers := *parallel
	if *sequential {
		workers = 0
	}

	if workers == 0 {
		return executeSequential(ctx, jobs)
	}
	return executeParallel(ctx, jobs, workers)
}

// executeSequential runs imports one at a time.
func executeSequential(ctx context.Context, jobs []ImportJob) (*ImportSummary, error) {
	importer := NewSequentialImporter()
	importer.SetProgressCallback(createProgressCallback())
	return importer.Import(ctx, jobs)
}

// executeParallel runs imports with the given number of workers.
func executeParallel(ctx context.Context, jobs []ImportJob, workers int) (*ImportSummary, error) {
	checkpointMgr := NewCheckpointManager(*checkpointFile, !*noCheckpoint)
	importer := NewParallelImporter(workers, *checkpointFile)
	importer.SetProgressCallback(createProgressCallback())
	if !*noCheckpoint {
		importer.SetCheckpointCallback(checkpointMgr.Save)
	}
	summary, err := importer.Import(ctx, jobs)
	if saveErr := importer.GetImportLog().Save(*rollbackFile); saveErr != nil {
		fmt.Printf("Warning: Failed to save import log: %v\n", saveErr)
	}
	return summary, err
}

// finalizeExecution handles post-import checks, cleanup, and next steps output.
func finalizeExecution(ctx context.Context, summary *ImportSummary) int {
	if !*quiet {
		summary.PrintSummary()
	}

	if *postImportCheck {
		dd := NewDriftDetector(*verbose)
		if err := dd.PostImportDriftCheck(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
	}

	checkpointMgr := NewCheckpointManager(*checkpointFile, !*noCheckpoint)
	if summary.FailureCount == 0 && !*noCheckpoint {
		//nolint:errcheck
		checkpointMgr.Delete()
	}

	if !*quiet {
		printNextSteps(summary)
	}

	if summary.FailureCount > 0 {
		return 1
	}
	return 0
}

func runRollback() int {
	mgr := NewRollbackManager(*rollbackFile, *verbose, *rollbackPlan)

	// Verify preconditions
	if err := mgr.VerifyRollbackPreconditions(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Show plan only
	if *rollbackPlan {
		if err := mgr.ShowRollbackPlan(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		return 0
	}

	// Execute rollback
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := mgr.Rollback(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Rollback failed: %v\n", err)
		return 1
	}

	return 0
}

func buildImportJobs(data *ResourceData, prefix string, filter *FilterConfig) []ImportJob {
	estimatedCapacity := len(data.Monitors) + len(data.Healthchecks) + len(data.StatusPages) + len(data.Incidents) + len(data.Maintenance) + len(data.Outages)
	jobs := make([]ImportJob, 0, estimatedCapacity)
	index := 0

	// Helper to create terraform name
	terraformName := func(name string) string {
		gen := &Generator{prefix: prefix}
		return gen.terraformName(name)
	}

	// Monitors
	monitors := filter.FilterMonitors(data.Monitors)
	for _, m := range monitors {
		jobs = append(jobs, ImportJob{
			ResourceType: "hyperping_monitor",
			ResourceName: terraformName(m.Name),
			ResourceID:   m.UUID,
			Index:        index,
		})
		index++
	}

	// Healthchecks
	healthchecks := filter.FilterHealthchecks(data.Healthchecks)
	for _, h := range healthchecks {
		jobs = append(jobs, ImportJob{
			ResourceType: "hyperping_healthcheck",
			ResourceName: terraformName(h.Name),
			ResourceID:   h.UUID,
			Index:        index,
		})
		index++
	}

	// Status Pages
	pages := filter.FilterStatusPages(data.StatusPages)
	for _, sp := range pages {
		jobs = append(jobs, ImportJob{
			ResourceType: "hyperping_statuspage",
			ResourceName: terraformName(sp.Name),
			ResourceID:   sp.UUID,
			Index:        index,
		})
		index++
	}

	// Incidents
	incidents := filter.FilterIncidents(data.Incidents)
	for _, i := range incidents {
		jobs = append(jobs, ImportJob{
			ResourceType: "hyperping_incident",
			ResourceName: terraformName(i.Title.En),
			ResourceID:   i.UUID,
			Index:        index,
		})
		index++
	}

	// Maintenance
	maintenance := filter.FilterMaintenance(data.Maintenance)
	for _, m := range maintenance {
		titleText := m.Title.En
		if titleText == "" {
			titleText = m.Name
		}
		jobs = append(jobs, ImportJob{
			ResourceType: "hyperping_maintenance",
			ResourceName: terraformName(titleText),
			ResourceID:   m.UUID,
			Index:        index,
		})
		index++
	}

	// Outages
	outages := filter.FilterOutages(data.Outages)
	for _, o := range outages {
		jobs = append(jobs, ImportJob{
			ResourceType: "hyperping_outage",
			ResourceName: terraformName(o.Monitor.Name),
			ResourceID:   o.UUID,
			Index:        index,
		})
		index++
	}

	return jobs
}

func printBanner() {
	fmt.Println("\n" + repeatString("=", 80))
	fmt.Println("HYPERPING IMPORT GENERATOR v2.0")
	fmt.Println(repeatString("=", 80))
	fmt.Println()
}

func printImportPlan(jobs []ImportJob, filter *FilterConfig) {
	fmt.Println(repeatString("=", 80))
	fmt.Println("IMPORT PLAN")
	fmt.Println(repeatString("=", 80))
	fmt.Println()

	if !filter.IsEmpty() {
		fmt.Println("Filters:")
		fmt.Println("  " + filter.Summary())
		fmt.Println()
	}

	// Group by resource type
	byType := make(map[string]int)
	for _, job := range jobs {
		byType[job.ResourceType]++
	}

	fmt.Println("Resources to import:")
	for resourceType, count := range byType {
		fmt.Printf("  %-25s %d resource(s)\n", resourceType+":", count)
	}
	fmt.Printf("\n  Total: %d resources\n\n", len(jobs))

	workers := *parallel
	if *sequential {
		workers = 0
	}

	if workers > 0 {
		fmt.Printf("Execution mode: Parallel (%d workers)\n", workers)
	} else {
		fmt.Println("Execution mode: Sequential")
	}

	if !*noCheckpoint {
		fmt.Printf("Checkpoint: Enabled (%s)\n", *checkpointFile)
	} else {
		fmt.Println("Checkpoint: Disabled")
	}

	fmt.Println()
}

func printNextSteps(summary *ImportSummary) {
	fmt.Println("\n" + repeatString("=", 80))
	fmt.Println("NEXT STEPS")
	fmt.Println(repeatString("=", 80))
	fmt.Println()
	fmt.Println("Import complete! Recommended next steps:")
	fmt.Println()
	fmt.Println("1. Review import results above")
	fmt.Println("2. Run: terraform plan")
	fmt.Println("3. Verify no unexpected changes")
	fmt.Println("4. Run: terraform apply (if needed)")
	fmt.Println()

	if summary.FailureCount > 0 {
		fmt.Println("⚠️  Some imports failed. Review errors above and retry failed resources.")
		fmt.Println()
	}

	if ImportLogExists(*rollbackFile) {
		fmt.Println("To rollback this import:")
		fmt.Printf("  import-generator --rollback --rollback-file=%s\n", *rollbackFile)
		fmt.Println()
	}
}

func createProgressCallback() func(completed, total int, current string) {
	if *quiet {
		return nil
	}

	return func(completed, total int, current string) {
		percentage := float64(completed) / float64(total) * 100
		fmt.Printf("\rProgress: %d/%d (%.1f%%) - Current: %s", completed, total, percentage, current)
		if completed == total {
			fmt.Println()
		}
	}
}
