// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"os"
	"testing"
)

func TestImportCheckpoint_AddImported(t *testing.T) {
	checkpoint := NewImportCheckpoint(10)

	checkpoint.AddImported("mon_123", "hyperping_monitor", "prod_api")
	checkpoint.AddImported("hc_456", "hyperping_healthcheck", "prod_heartbeat")

	if len(checkpoint.ImportedIDs) != 2 {
		t.Errorf("Expected 2 imported resources, got %d", len(checkpoint.ImportedIDs))
	}

	if !checkpoint.IsImported("mon_123") {
		t.Error("Expected mon_123 to be marked as imported")
	}

	if !checkpoint.IsImported("hc_456") {
		t.Error("Expected hc_456 to be marked as imported")
	}
}

func TestImportCheckpoint_AddFailed(t *testing.T) {
	checkpoint := NewImportCheckpoint(10)

	checkpoint.AddFailed("mon_789")
	checkpoint.AddFailed("hc_012")

	if len(checkpoint.FailedIDs) != 2 {
		t.Errorf("Expected 2 failed resources, got %d", len(checkpoint.FailedIDs))
	}

	if !checkpoint.IsFailed("mon_789") {
		t.Error("Expected mon_789 to be marked as failed")
	}
}

func TestImportCheckpoint_SaveLoad(t *testing.T) {
	tempFile := ".test-checkpoint"
	defer os.Remove(tempFile)

	// Create and save checkpoint
	original := NewImportCheckpoint(100)
	original.AddImported("mon_123", "hyperping_monitor", "prod_api")
	original.AddFailed("mon_456")
	original.CurrentIndex = 50
	original.FilterSummary = "Name pattern: PROD-.*"

	if err := original.Save(tempFile); err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}

	// Load checkpoint
	loaded, err := LoadCheckpoint(tempFile)
	if err != nil {
		t.Fatalf("Failed to load checkpoint: %v", err)
	}

	// Verify data
	if loaded.TotalResources != original.TotalResources {
		t.Errorf("TotalResources = %d, want %d", loaded.TotalResources, original.TotalResources)
	}

	if loaded.CurrentIndex != original.CurrentIndex {
		t.Errorf("CurrentIndex = %d, want %d", loaded.CurrentIndex, original.CurrentIndex)
	}

	if len(loaded.ImportedIDs) != len(original.ImportedIDs) {
		t.Errorf("ImportedIDs length = %d, want %d", len(loaded.ImportedIDs), len(original.ImportedIDs))
	}

	if len(loaded.FailedIDs) != len(original.FailedIDs) {
		t.Errorf("FailedIDs length = %d, want %d", len(loaded.FailedIDs), len(original.FailedIDs))
	}

	if loaded.FilterSummary != original.FilterSummary {
		t.Errorf("FilterSummary = %q, want %q", loaded.FilterSummary, original.FilterSummary)
	}
}

func TestCheckpointExists(t *testing.T) {
	tempFile := ".test-checkpoint-exists"

	// Should not exist initially
	if CheckpointExists(tempFile) {
		t.Error("Expected checkpoint to not exist")
	}

	// Create checkpoint
	checkpoint := NewImportCheckpoint(10)
	if err := checkpoint.Save(tempFile); err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}
	defer os.Remove(tempFile)

	// Should exist now
	if !CheckpointExists(tempFile) {
		t.Error("Expected checkpoint to exist")
	}
}

func TestDeleteCheckpoint(t *testing.T) {
	tempFile := ".test-checkpoint-delete"

	// Create checkpoint
	checkpoint := NewImportCheckpoint(10)
	if err := checkpoint.Save(tempFile); err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}

	// Delete checkpoint
	if err := DeleteCheckpoint(tempFile); err != nil {
		t.Fatalf("Failed to delete checkpoint: %v", err)
	}

	// Should not exist
	if CheckpointExists(tempFile) {
		t.Error("Expected checkpoint to be deleted")
	}
}

func TestFilterJobsForResume(t *testing.T) {
	jobs := []ImportJob{
		{ResourceID: "mon_1", ResourceType: "hyperping_monitor", ResourceName: "api"},
		{ResourceID: "mon_2", ResourceType: "hyperping_monitor", ResourceName: "web"},
		{ResourceID: "mon_3", ResourceType: "hyperping_monitor", ResourceName: "db"},
		{ResourceID: "mon_4", ResourceType: "hyperping_monitor", ResourceName: "cache"},
	}

	checkpoint := NewImportCheckpoint(4)
	checkpoint.AddImported("mon_1", "hyperping_monitor", "api")
	checkpoint.AddFailed("mon_3")

	filtered := FilterJobsForResume(jobs, checkpoint)

	// Should only have mon_2 and mon_4 (skip imported and failed)
	if len(filtered) != 2 {
		t.Errorf("Expected 2 jobs after filtering, got %d", len(filtered))
	}

	for _, job := range filtered {
		if job.ResourceID == "mon_1" || job.ResourceID == "mon_3" {
			t.Errorf("Job %s should have been filtered out", job.ResourceID)
		}
	}
}

func TestCheckpointManager(t *testing.T) {
	tempFile := ".test-checkpoint-manager"
	defer os.Remove(tempFile)

	mgr := NewCheckpointManager(tempFile, true)

	// Save checkpoint
	checkpoint := NewImportCheckpoint(10)
	checkpoint.AddImported("mon_123", "hyperping_monitor", "api")

	if err := mgr.Save(checkpoint); err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}

	// Verify exists
	if !mgr.Exists() {
		t.Error("Expected checkpoint to exist")
	}

	// Load checkpoint
	loaded, err := mgr.Load()
	if err != nil {
		t.Fatalf("Failed to load checkpoint: %v", err)
	}

	if len(loaded.ImportedIDs) != 1 {
		t.Errorf("Expected 1 imported resource, got %d", len(loaded.ImportedIDs))
	}

	// Delete checkpoint
	if err := mgr.Delete(); err != nil {
		t.Fatalf("Failed to delete checkpoint: %v", err)
	}

	if mgr.Exists() {
		t.Error("Expected checkpoint to be deleted")
	}
}

func TestCheckpointManager_Disabled(t *testing.T) {
	tempFile := ".test-checkpoint-disabled"
	defer os.Remove(tempFile)

	mgr := NewCheckpointManager(tempFile, false)

	// Save should do nothing when disabled
	checkpoint := NewImportCheckpoint(10)
	if err := mgr.Save(checkpoint); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// File should not exist
	if mgr.Exists() {
		t.Error("Expected checkpoint file to not exist when disabled")
	}
}

func BenchmarkCheckpointSave(b *testing.B) {
	tempFile := ".bench-checkpoint"
	defer os.Remove(tempFile)

	checkpoint := NewImportCheckpoint(1000)
	for i := 0; i < 500; i++ {
		checkpoint.AddImported("mon_"+string(rune(i)), "hyperping_monitor", "resource")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checkpoint.Save(tempFile)
	}
}

func BenchmarkCheckpointLoad(b *testing.B) {
	tempFile := ".bench-checkpoint-load"
	defer os.Remove(tempFile)

	checkpoint := NewImportCheckpoint(1000)
	for i := 0; i < 500; i++ {
		checkpoint.AddImported("mon_"+string(rune(i)), "hyperping_monitor", "resource")
	}
	checkpoint.Save(tempFile)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LoadCheckpoint(tempFile)
	}
}
