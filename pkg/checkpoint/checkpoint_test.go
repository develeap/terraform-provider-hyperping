// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package checkpoint

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if mgr.checkpointDir == "" {
		t.Error("Expected checkpoint directory to be set")
	}

	// Verify directory was created
	if _, err := os.Stat(mgr.checkpointDir); os.IsNotExist(err) {
		t.Errorf("Checkpoint directory was not created: %s", mgr.checkpointDir)
	}
}

func TestSaveAndLoad(t *testing.T) {
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create test checkpoint
	cp := &Checkpoint{
		MigrationID:    "test-migration-001",
		Tool:           "test-tool",
		Status:         StatusInProgress,
		TotalResources: 100,
		Processed:      50,
		Failed:         5,
		ProcessedIDs:   []string{"id1", "id2", "id3"},
		FailedResources: []FailedResource{
			{ID: "fail1", Type: "monitor", Error: "test error"},
		},
		HyperpingCreated: []string{"uuid1", "uuid2"},
		Metadata:         map[string]string{"key": "value"},
	}

	// Save checkpoint
	//nolint:govet
	if err := mgr.Save(cp); err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}

	// Load checkpoint
	loaded, err := mgr.Load("test-migration-001")
	if err != nil {
		t.Fatalf("Failed to load checkpoint: %v", err)
	}

	// Verify fields
	if loaded.MigrationID != cp.MigrationID {
		t.Errorf("Expected MigrationID %s, got %s", cp.MigrationID, loaded.MigrationID)
	}
	if loaded.Tool != cp.Tool {
		t.Errorf("Expected Tool %s, got %s", cp.Tool, loaded.Tool)
	}
	if loaded.TotalResources != cp.TotalResources {
		t.Errorf("Expected TotalResources %d, got %d", cp.TotalResources, loaded.TotalResources)
	}
	if loaded.Processed != cp.Processed {
		t.Errorf("Expected Processed %d, got %d", cp.Processed, loaded.Processed)
	}
	if loaded.Failed != cp.Failed {
		t.Errorf("Expected Failed %d, got %d", cp.Failed, loaded.Failed)
	}
	if len(loaded.ProcessedIDs) != len(cp.ProcessedIDs) {
		t.Errorf("Expected %d ProcessedIDs, got %d", len(cp.ProcessedIDs), len(loaded.ProcessedIDs))
	}
	if len(loaded.HyperpingCreated) != len(cp.HyperpingCreated) {
		t.Errorf("Expected %d HyperpingCreated, got %d", len(cp.HyperpingCreated), len(loaded.HyperpingCreated))
	}

	// Cleanup
	mgr.Delete("test-migration-001")
}

func TestAtomicSave(t *testing.T) {
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	cp := &Checkpoint{
		MigrationID:    "test-atomic-001",
		Tool:           "test-tool",
		Status:         StatusInProgress,
		TotalResources: 10,
	}

	// Save checkpoint
	//nolint:govet
	if err := mgr.Save(cp); err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}

	// Verify no temp file exists
	tempFile := filepath.Join(mgr.checkpointDir, "test-atomic-001.json.tmp")
	if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
		t.Error("Temporary file should not exist after save")
	}

	// Verify checkpoint file exists
	checkpointFile := filepath.Join(mgr.checkpointDir, "test-atomic-001.json")
	if _, err := os.Stat(checkpointFile); os.IsNotExist(err) {
		t.Error("Checkpoint file should exist")
	}

	// Cleanup
	mgr.Delete("test-atomic-001")
}

func TestList(t *testing.T) {
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create multiple checkpoints
	createdIDs := []string{}
	for i := 1; i <= 3; i++ {
		migrationID := fmt.Sprintf("test-list-%d-%d", time.Now().UnixNano(), i)
		cp := &Checkpoint{
			MigrationID:    migrationID,
			Tool:           "test-list-tool",
			Status:         StatusInProgress,
			TotalResources: i * 10,
		}
		if saveErr := mgr.Save(cp); saveErr != nil {
			t.Fatalf("Failed to save checkpoint %d: %v", i, saveErr)
		}
		createdIDs = append(createdIDs, migrationID)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// List checkpoints
	checkpoints, err := mgr.List()
	if err != nil {
		t.Fatalf("Failed to list checkpoints: %v", err)
	}

	// Count how many of our test checkpoints are present
	testCount := 0
	for _, cp := range checkpoints {
		if cp.Tool == "test-list-tool" {
			testCount++
		}
	}

	if testCount < 3 {
		t.Errorf("Expected at least 3 test checkpoints, got %d", testCount)
	}

	// Cleanup
	for _, id := range createdIDs {
		mgr.Delete(id)
	}
}

func TestFindLatest(t *testing.T) {
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create checkpoints with different timestamps
	migrationIDs := []string{}
	for i := 1; i <= 3; i++ {
		cp := &Checkpoint{
			MigrationID:    GenerateMigrationID("findlatest"),
			Tool:           "findlatest-tool",
			Status:         StatusInProgress,
			TotalResources: i * 10,
		}
		if saveErr := mgr.Save(cp); saveErr != nil {
			t.Fatalf("Failed to save checkpoint %d: %v", i, saveErr)
		}
		migrationIDs = append(migrationIDs, cp.MigrationID)
		time.Sleep(10 * time.Millisecond)
	}

	// Find latest
	latest, err := mgr.FindLatest("findlatest-tool")
	if err != nil {
		t.Fatalf("Failed to find latest: %v", err)
	}

	// Should be the last one created
	if latest.MigrationID != migrationIDs[len(migrationIDs)-1] {
		t.Errorf("Expected latest migration ID %s, got %s", migrationIDs[len(migrationIDs)-1], latest.MigrationID)
	}

	// Cleanup
	for _, id := range migrationIDs {
		mgr.Delete(id)
	}
}

func TestExists(t *testing.T) {
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test non-existent checkpoint
	if mgr.Exists("non-existent-migration") {
		t.Error("Expected checkpoint to not exist")
	}

	// Create checkpoint
	cp := &Checkpoint{
		MigrationID:    "test-exists-001",
		Tool:           "test-tool",
		Status:         StatusInProgress,
		TotalResources: 10,
	}
	if err := mgr.Save(cp); err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}

	// Test existing checkpoint
	if !mgr.Exists("test-exists-001") {
		t.Error("Expected checkpoint to exist")
	}

	// Cleanup
	mgr.Delete("test-exists-001")
}

func TestMarkProcessed(t *testing.T) {
	cp := &Checkpoint{
		MigrationID:    "test-mark-001",
		Tool:           "test-tool",
		Status:         StatusInProgress,
		TotalResources: 10,
		ProcessedIDs:   []string{},
	}

	// Mark resources as processed
	cp.MarkProcessed("id1")
	cp.MarkProcessed("id2")
	cp.MarkProcessed("id3")

	if cp.Processed != 3 {
		t.Errorf("Expected Processed to be 3, got %d", cp.Processed)
	}

	if len(cp.ProcessedIDs) != 3 {
		t.Errorf("Expected 3 ProcessedIDs, got %d", len(cp.ProcessedIDs))
	}

	// Test IsProcessed
	if !cp.IsProcessed("id2") {
		t.Error("Expected id2 to be processed")
	}

	if cp.IsProcessed("id999") {
		t.Error("Expected id999 to not be processed")
	}

	// Test duplicate
	cp.MarkProcessed("id1")
	if cp.Processed != 3 {
		t.Errorf("Expected Processed to still be 3 after duplicate, got %d", cp.Processed)
	}
}

func TestMarkFailed(t *testing.T) {
	cp := &Checkpoint{
		MigrationID:     "test-mark-fail-001",
		Tool:            "test-tool",
		Status:          StatusInProgress,
		TotalResources:  10,
		FailedResources: []FailedResource{},
	}

	// Mark resources as failed
	cp.MarkFailed(FailedResource{
		ID:    "fail1",
		Type:  "monitor",
		Name:  "Test Monitor",
		Error: "Connection timeout",
	})

	if cp.Failed != 1 {
		t.Errorf("Expected Failed to be 1, got %d", cp.Failed)
	}

	if len(cp.FailedResources) != 1 {
		t.Errorf("Expected 1 FailedResource, got %d", len(cp.FailedResources))
	}

	if cp.FailedResources[0].Error != "Connection timeout" {
		t.Errorf("Expected error message 'Connection timeout', got '%s'", cp.FailedResources[0].Error)
	}
}

func TestGenerateMigrationID(t *testing.T) {
	id1 := GenerateMigrationID("test-tool")
	time.Sleep(1100 * time.Millisecond) // Sleep longer than 1 second to ensure different timestamps
	id2 := GenerateMigrationID("test-tool")

	if id1 == id2 {
		t.Errorf("Expected different migration IDs, both were: %s", id1)
	}

	if !strings.HasPrefix(id1, "test-tool") {
		t.Errorf("Expected migration ID to start with 'test-tool', got %s", id1)
	}

	if !strings.HasPrefix(id2, "test-tool") {
		t.Errorf("Expected migration ID to start with 'test-tool', got %s", id2)
	}
}

func TestDelete(t *testing.T) {
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create checkpoint
	cp := &Checkpoint{
		MigrationID:    "test-delete-001",
		Tool:           "test-tool",
		Status:         StatusInProgress,
		TotalResources: 10,
	}
	if err := mgr.Save(cp); err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}

	// Verify exists
	if !mgr.Exists("test-delete-001") {
		t.Error("Expected checkpoint to exist before deletion")
	}

	// Delete
	if err := mgr.Delete("test-delete-001"); err != nil {
		t.Fatalf("Failed to delete checkpoint: %v", err)
	}

	// Verify deleted
	if mgr.Exists("test-delete-001") {
		t.Error("Expected checkpoint to not exist after deletion")
	}

	// Delete non-existent should not error
	if err := mgr.Delete("non-existent"); err != nil {
		t.Errorf("Expected no error deleting non-existent checkpoint, got: %v", err)
	}
}
