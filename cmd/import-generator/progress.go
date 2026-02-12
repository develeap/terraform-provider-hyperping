// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"io"
	"os"
)

// ProgressReporter handles progress output during resource fetching.
type ProgressReporter struct {
	enabled bool
	output  io.Writer
	current int
	total   int
}

// NewProgressReporter creates a new progress reporter.
func NewProgressReporter(enabled bool) *ProgressReporter {
	return &ProgressReporter{
		enabled: enabled,
		output:  os.Stderr,
	}
}

// SetTotal sets the total number of steps.
func (p *ProgressReporter) SetTotal(total int) {
	p.total = total
	p.current = 0
}

// Step reports progress for a single step.
func (p *ProgressReporter) Step(resource string) {
	if !p.enabled {
		return
	}
	p.current++
	fmt.Fprintf(p.output, "[%d/%d] Fetching %s...\n", p.current, p.total, resource)
}

// Report reports the result of a step.
func (p *ProgressReporter) Report(count int, resourceType string) {
	if !p.enabled {
		return
	}
	fmt.Fprintf(p.output, "  Found %d %s\n", count, resourceType)
}

// Error reports an error during a step.
func (p *ProgressReporter) Error(err error) {
	if !p.enabled {
		return
	}
	fmt.Fprintf(p.output, "  Error: %v\n", err)
}

// Complete reports completion.
func (p *ProgressReporter) Complete() {
	if !p.enabled {
		return
	}
	fmt.Fprintln(p.output, "\nResource fetching complete")
}
