// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

const (
	defaultWorkers = 5
	maxWorkers     = 20
)

// ImportJob represents a single import operation.
type ImportJob struct {
	ResourceType string
	ResourceName string
	ResourceID   string
	Index        int
}

// ImportResult holds the result of an import operation.
type ImportResult struct {
	Job       ImportJob
	Success   bool
	Output    string
	Error     error
	Duration  time.Duration
	StartTime time.Time
}

// ImportSummary holds aggregate statistics for import operations.
type ImportSummary struct {
	TotalJobs     int
	SuccessCount  int
	FailureCount  int
	WarningCount  int
	TotalDuration time.Duration
	FailedJobs    []ImportResult
	WarningJobs   []ImportResult
	StartTime     time.Time
	EndTime       time.Time
}

// ParallelImporter manages concurrent import operations.
type ParallelImporter struct {
	workers         int
	checkpointEvery int
	checkpointFile  string
	importLog       *ImportLog
	onProgress      func(completed, total int, current string)
	onCheckpoint    func(checkpoint *ImportCheckpoint) error
}

// NewParallelImporter creates a new parallel importer.
func NewParallelImporter(workers int, checkpointFile string) *ParallelImporter {
	if workers <= 0 {
		workers = defaultWorkers
	}
	if workers > maxWorkers {
		workers = maxWorkers
	}

	return &ParallelImporter{
		workers:         workers,
		checkpointEvery: 10,
		checkpointFile:  checkpointFile,
		importLog:       NewImportLog(),
	}
}

// SetProgressCallback sets a callback for progress updates.
func (pi *ParallelImporter) SetProgressCallback(fn func(completed, total int, current string)) {
	pi.onProgress = fn
}

// SetCheckpointCallback sets a callback for checkpoint saves.
func (pi *ParallelImporter) SetCheckpointCallback(fn func(checkpoint *ImportCheckpoint) error) {
	pi.onCheckpoint = fn
}

// Import executes import jobs in parallel.
//
//nolint:unparam
func (pi *ParallelImporter) Import(ctx context.Context, jobs []ImportJob) (*ImportSummary, error) {
	if len(jobs) == 0 {
		return &ImportSummary{}, nil
	}

	summary := &ImportSummary{
		TotalJobs: len(jobs),
		StartTime: time.Now(),
	}

	jobChan := make(chan ImportJob, len(jobs))
	resultChan := make(chan ImportResult, len(jobs))

	// Start worker pool
	var wg sync.WaitGroup
	for i := 0; i < pi.workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			pi.worker(ctx, workerID, jobChan, resultChan)
		}(i)
	}

	// Send jobs to workers
	for _, job := range jobs {
		jobChan <- job
	}
	close(jobChan)

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	completed := 0
	checkpoint := NewImportCheckpoint(len(jobs))

	for result := range resultChan {
		completed++

		// Update summary
		if result.Success {
			summary.SuccessCount++
			checkpoint.AddImported(result.Job.ResourceID, result.Job.ResourceType, result.Job.ResourceName)
			pi.importLog.AddImport(result.Job.ResourceType, result.Job.ResourceName, result.Job.ResourceID)
		} else {
			summary.FailureCount++
			summary.FailedJobs = append(summary.FailedJobs, result)
			checkpoint.AddFailed(result.Job.ResourceID)
		}

		summary.TotalDuration += result.Duration
		checkpoint.CurrentIndex = completed

		// Progress callback
		if pi.onProgress != nil {
			currentResource := fmt.Sprintf("%s.%s", result.Job.ResourceType, result.Job.ResourceName)
			pi.onProgress(completed, len(jobs), currentResource)
		}

		// Checkpoint every N imports
		if completed%pi.checkpointEvery == 0 && pi.onCheckpoint != nil {
			if err := pi.onCheckpoint(checkpoint); err != nil {
				fmt.Printf("Warning: Failed to save checkpoint: %v\n", err)
			}
		}
	}

	summary.EndTime = time.Now()

	// Final checkpoint
	if pi.onCheckpoint != nil {
		checkpoint.Completed = true
		if err := pi.onCheckpoint(checkpoint); err != nil {
			fmt.Printf("Warning: Failed to save final checkpoint: %v\n", err)
		}
	}

	return summary, nil
}

// worker processes import jobs from the channel.
func (pi *ParallelImporter) worker(ctx context.Context, _ int, jobs <-chan ImportJob, results chan<- ImportResult) {
	for job := range jobs {
		select {
		case <-ctx.Done():
			results <- ImportResult{
				Job:     job,
				Success: false,
				Error:   ctx.Err(),
			}
			return
		default:
			result := pi.executeImport(ctx, job)
			results <- result
		}
	}
}

// executeImport runs a single terraform import command.
func (pi *ParallelImporter) executeImport(ctx context.Context, job ImportJob) ImportResult {
	startTime := time.Now()

	result := ImportResult{
		Job:       job,
		StartTime: startTime,
	}

	// Build terraform import command
	resourceAddress := fmt.Sprintf("%s.%s", job.ResourceType, job.ResourceName)
	cmd := exec.CommandContext(ctx, "terraform", "import", resourceAddress, job.ResourceID)

	// Execute command
	output, err := cmd.CombinedOutput()
	result.Output = string(output)
	result.Duration = time.Since(startTime)

	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("import failed: %w", err)
	} else {
		result.Success = true
	}

	return result
}

// GetImportLog returns the import log.
func (pi *ParallelImporter) GetImportLog() *ImportLog {
	return pi.importLog
}

// SequentialImporter provides a fallback sequential import implementation.
type SequentialImporter struct {
	importLog  *ImportLog
	onProgress func(completed, total int, current string)
}

// NewSequentialImporter creates a new sequential importer.
func NewSequentialImporter() *SequentialImporter {
	return &SequentialImporter{
		importLog: NewImportLog(),
	}
}

// SetProgressCallback sets a callback for progress updates.
func (si *SequentialImporter) SetProgressCallback(fn func(completed, total int, current string)) {
	si.onProgress = fn
}

// Import executes import jobs sequentially.
func (si *SequentialImporter) Import(ctx context.Context, jobs []ImportJob) (*ImportSummary, error) {
	if len(jobs) == 0 {
		return &ImportSummary{}, nil
	}

	summary := &ImportSummary{
		TotalJobs: len(jobs),
		StartTime: time.Now(),
	}

	for i, job := range jobs {
		select {
		case <-ctx.Done():
			return summary, ctx.Err()
		default:
			result := si.executeImport(ctx, job)

			if result.Success {
				summary.SuccessCount++
				si.importLog.AddImport(result.Job.ResourceType, result.Job.ResourceName, result.Job.ResourceID)
			} else {
				summary.FailureCount++
				summary.FailedJobs = append(summary.FailedJobs, result)
			}

			summary.TotalDuration += result.Duration

			// Progress callback
			if si.onProgress != nil {
				currentResource := fmt.Sprintf("%s.%s", result.Job.ResourceType, result.Job.ResourceName)
				si.onProgress(i+1, len(jobs), currentResource)
			}
		}
	}

	summary.EndTime = time.Now()
	return summary, nil
}

// executeImport runs a single terraform import command.
func (si *SequentialImporter) executeImport(ctx context.Context, job ImportJob) ImportResult {
	startTime := time.Now()

	result := ImportResult{
		Job:       job,
		StartTime: startTime,
	}

	resourceAddress := fmt.Sprintf("%s.%s", job.ResourceType, job.ResourceName)
	cmd := exec.CommandContext(ctx, "terraform", "import", resourceAddress, job.ResourceID)

	output, err := cmd.CombinedOutput()
	result.Output = string(output)
	result.Duration = time.Since(startTime)

	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("import failed: %w", err)
	} else {
		result.Success = true
	}

	return result
}

// GetImportLog returns the import log.
func (si *SequentialImporter) GetImportLog() *ImportLog {
	return si.importLog
}

// PrintSummary prints a formatted import summary.
func (s *ImportSummary) PrintSummary() {
	fmt.Println("\n" + repeatString("=", 80))
	fmt.Println("IMPORT SUMMARY")
	fmt.Println(repeatString("=", 80))

	fmt.Printf("Total resources:      %d\n", s.TotalJobs)
	fmt.Printf("Successfully imported: %d\n", s.SuccessCount)
	fmt.Printf("Failed:               %d\n", s.FailureCount)
	fmt.Printf("Warnings:             %d\n", s.WarningCount)

	if s.TotalJobs > 0 {
		successRate := float64(s.SuccessCount) / float64(s.TotalJobs) * 100
		fmt.Printf("Success rate:         %.1f%%\n", successRate)
	}

	fmt.Printf("Total time:           %s\n", s.EndTime.Sub(s.StartTime).Round(time.Second))
	if s.TotalJobs > 0 {
		avgTime := s.TotalDuration / time.Duration(s.TotalJobs)
		fmt.Printf("Average time/import:  %s\n", avgTime.Round(time.Millisecond))
	}

	if len(s.FailedJobs) > 0 {
		fmt.Println("\n" + repeatString("-", 80))
		fmt.Println("FAILED IMPORTS:")
		fmt.Println(repeatString("-", 80))
		for _, job := range s.FailedJobs {
			fmt.Printf("  %s.%s (ID: %s)\n", job.Job.ResourceType, job.Job.ResourceName, job.Job.ResourceID)
			if job.Error != nil {
				fmt.Printf("    Error: %v\n", job.Error)
			}
		}
	}

	if len(s.WarningJobs) > 0 {
		fmt.Println("\n" + repeatString("-", 80))
		fmt.Println("WARNINGS:")
		fmt.Println(repeatString("-", 80))
		for _, job := range s.WarningJobs {
			fmt.Printf("  %s.%s (ID: %s)\n", job.Job.ResourceType, job.Job.ResourceName, job.Job.ResourceID)
		}
	}

	fmt.Println(repeatString("=", 80))
}

func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
