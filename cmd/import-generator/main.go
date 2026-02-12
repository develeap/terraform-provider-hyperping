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
	outputFormat    = flag.String("format", "both", "Output format: import, hcl, both, or script")
	outputFile      = flag.String("output", "", "Output file (default: stdout)")
	resources       = flag.String("resources", "all", "Resources to import: all, monitors, healthchecks, statuspages, incidents, maintenance, outages")
	prefix          = flag.String("prefix", "", "Prefix for Terraform resource names (e.g., 'prod_')")
	baseURL         = flag.String("base-url", "https://api.hyperping.io", "Hyperping API base URL")
	validate        = flag.Bool("validate", false, "Validate resources without generating output")
	progress        = flag.Bool("progress", false, "Show progress indicators")
	continueOnError = flag.Bool("continue-on-error", false, "Continue on errors instead of failing")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: import-generator [options]\n\n")
		fmt.Fprintf(os.Stderr, "Generates Terraform import commands and HCL configurations from existing Hyperping resources.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Validate resources\n")
		fmt.Fprintf(os.Stderr, "  import-generator -validate\n\n")
		fmt.Fprintf(os.Stderr, "  # Generate with progress indicators\n")
		fmt.Fprintf(os.Stderr, "  import-generator -progress -format=both\n\n")
		fmt.Fprintf(os.Stderr, "  # Generate executable script\n")
		fmt.Fprintf(os.Stderr, "  import-generator -format=script -output=import.sh\n\n")
		fmt.Fprintf(os.Stderr, "  # Continue on errors\n")
		fmt.Fprintf(os.Stderr, "  import-generator -continue-on-error -format=import\n\n")
	}
	os.Exit(run())
}

func run() int {
	flag.Parse()

	apiKey := os.Getenv("HYPERPING_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: HYPERPING_API_KEY environment variable is required")
		return 1
	}

	// Create client
	c := client.NewClient(apiKey, client.WithBaseURL(*baseURL))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create generator
	gen := &Generator{
		client:          c,
		prefix:          *prefix,
		resources:       parseResources(*resources),
		showProgress:    *progress,
		continueOnError: *continueOnError,
	}

	// Handle validation mode
	if *validate {
		return runValidation(ctx, gen)
	}

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
