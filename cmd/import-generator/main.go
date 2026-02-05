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
	outputFormat = flag.String("format", "both", "Output format: import, hcl, or both")
	outputFile   = flag.String("output", "", "Output file (default: stdout)")
	resources    = flag.String("resources", "all", "Resources to import: all, monitors, healthchecks, statuspages, incidents, maintenance, outages")
	prefix       = flag.String("prefix", "", "Prefix for Terraform resource names (e.g., 'prod_')")
	baseURL      = flag.String("base-url", "https://api.hyperping.io", "Hyperping API base URL")
)

func main() {
	flag.Parse()

	apiKey := os.Getenv("HYPERPING_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: HYPERPING_API_KEY environment variable is required")
		os.Exit(1)
	}

	// Create client
	c := client.NewClient(apiKey, client.WithBaseURL(*baseURL))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Generate output
	gen := &Generator{
		client:    c,
		prefix:    *prefix,
		resources: parseResources(*resources),
	}

	output, err := gen.Generate(ctx, *outputFormat)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating output: %v\n", err)
		os.Exit(1)
	}

	// Write output
	if *outputFile != "" {
		if err := os.WriteFile(*outputFile, []byte(output), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Output written to %s\n", *outputFile)
	} else {
		fmt.Print(output)
	}
}

func parseResources(s string) []string {
	if s == "all" {
		return []string{"monitors", "healthchecks", "statuspages", "incidents", "maintenance", "outages"}
	}
	return strings.Split(s, ",")
}
