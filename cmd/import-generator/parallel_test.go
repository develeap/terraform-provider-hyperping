// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"testing"
	"time"
)

func TestNewParallelImporter(t *testing.T) {
	tests := []struct {
		name            string
		workers         int
		expectedWorkers int
	}{
		{
			name:            "default workers",
			workers:         5,
			expectedWorkers: 5,
		},
		{
			name:            "zero workers defaults to 5",
			workers:         0,
			expectedWorkers: 5,
		},
		{
			name:            "negative workers defaults to 5",
			workers:         -1,
			expectedWorkers: 5,
		},
		{
			name:            "max workers capped at 20",
			workers:         25,
			expectedWorkers: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			importer := NewParallelImporter(tt.workers, ".test-checkpoint")
			if importer.workers != tt.expectedWorkers {
				t.Errorf("workers = %d, want %d", importer.workers, tt.expectedWorkers)
			}
		})
	}
}

func TestImportJob_Structure(t *testing.T) {
	job := ImportJob{
		ResourceType: "hyperping_monitor",
		ResourceName: "prod_api",
		ResourceID:   "mon_123",
		Index:        0,
	}

	if job.ResourceType != "hyperping_monitor" {
		t.Errorf("ResourceType = %s, want hyperping_monitor", job.ResourceType)
	}

	if job.ResourceName != "prod_api" {
		t.Errorf("ResourceName = %s, want prod_api", job.ResourceName)
	}

	if job.ResourceID != "mon_123" {
		t.Errorf("ResourceID = %s, want mon_123", job.ResourceID)
	}

	if job.Index != 0 {
		t.Errorf("Index = %d, want 0", job.Index)
	}
}

func TestImportSummary_PrintSummary(_ *testing.T) {
	summary := &ImportSummary{
		TotalJobs:     100,
		SuccessCount:  95,
		FailureCount:  5,
		WarningCount:  2,
		TotalDuration: 5 * time.Minute,
		StartTime:     time.Now().Add(-5 * time.Minute),
		EndTime:       time.Now(),
	}

	// Just verify it doesn't panic
	summary.PrintSummary()
}

func TestSequentialImporter_EmptyJobs(t *testing.T) {
	importer := NewSequentialImporter()
	ctx := context.Background()

	summary, err := importer.Import(ctx, []ImportJob{})
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if summary.TotalJobs != 0 {
		t.Errorf("TotalJobs = %d, want 0", summary.TotalJobs)
	}

	if summary.SuccessCount != 0 {
		t.Errorf("SuccessCount = %d, want 0", summary.SuccessCount)
	}
}

func TestParallelImporter_EmptyJobs(t *testing.T) {
	importer := NewParallelImporter(5, ".test-checkpoint")
	ctx := context.Background()

	summary, err := importer.Import(ctx, []ImportJob{})
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if summary.TotalJobs != 0 {
		t.Errorf("TotalJobs = %d, want 0", summary.TotalJobs)
	}
}

func TestParallelImporter_ProgressCallback(t *testing.T) {
	importer := NewParallelImporter(2, ".test-checkpoint")

	importer.SetProgressCallback(func(completed, total int, current string) {
		// Callback implementation
	})

	// Note: We can't actually test terraform import without a real terraform env,
	// so this test verifies the callback mechanism is set up correctly
	if importer.onProgress == nil {
		t.Error("Expected progress callback to be set")
	}
}

func TestParallelImporter_CheckpointCallback(t *testing.T) {
	importer := NewParallelImporter(2, ".test-checkpoint")

	importer.SetCheckpointCallback(func(checkpoint *ImportCheckpoint) error {
		return nil
	})

	if importer.onCheckpoint == nil {
		t.Error("Expected checkpoint callback to be set")
	}
}

func TestRepeatString(t *testing.T) {
	tests := []struct {
		name  string
		s     string
		count int
		want  string
	}{
		{
			name:  "repeat dash",
			s:     "-",
			count: 5,
			want:  "-----",
		},
		{
			name:  "repeat equals",
			s:     "=",
			count: 3,
			want:  "===",
		},
		{
			name:  "repeat zero",
			s:     "x",
			count: 0,
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := repeatString(tt.s, tt.count)
			if got != tt.want {
				t.Errorf("repeatString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func BenchmarkParallelImporter_EmptyJobs(b *testing.B) {
	importer := NewParallelImporter(5, ".bench-checkpoint")
	ctx := context.Background()
	jobs := []ImportJob{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		importer.Import(ctx, jobs)
	}
}
