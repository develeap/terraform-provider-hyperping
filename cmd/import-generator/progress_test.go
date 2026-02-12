// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bytes"
	"errors"
	"testing"
)

func TestProgressReporter_Disabled(t *testing.T) {
	var buf bytes.Buffer
	pr := NewProgressReporter(false)
	pr.output = &buf

	pr.SetTotal(3)
	pr.Step("monitors")
	pr.Report(10, "monitor(s)")
	pr.Complete()

	if buf.Len() != 0 {
		t.Errorf("Expected no output when disabled, got: %s", buf.String())
	}
}

func TestProgressReporter_Enabled(t *testing.T) {
	var buf bytes.Buffer
	pr := NewProgressReporter(true)
	pr.output = &buf

	pr.SetTotal(3)
	pr.Step("monitors")
	pr.Report(10, "monitor(s)")
	pr.Step("healthchecks")
	pr.Report(5, "healthcheck(s)")
	pr.Complete()

	output := buf.String()

	expectedPhrases := []string{
		"[1/3] Fetching monitors",
		"Found 10 monitor(s)",
		"[2/3] Fetching healthchecks",
		"Found 5 healthcheck(s)",
		"Resource fetching complete",
	}

	for _, phrase := range expectedPhrases {
		if !bytes.Contains(buf.Bytes(), []byte(phrase)) {
			t.Errorf("Expected output to contain %q, got: %s", phrase, output)
		}
	}
}

func TestProgressReporter_Error(t *testing.T) {
	var buf bytes.Buffer
	pr := NewProgressReporter(true)
	pr.output = &buf

	pr.SetTotal(1)
	pr.Step("monitors")
	pr.Error(errors.New("API error"))

	output := buf.String()

	if !bytes.Contains(buf.Bytes(), []byte("Error: API error")) {
		t.Errorf("Expected error message in output, got: %s", output)
	}
}

func TestProgressReporter_MultipleSteps(t *testing.T) {
	var buf bytes.Buffer
	pr := NewProgressReporter(true)
	pr.output = &buf

	pr.SetTotal(6)

	resources := []string{"monitors", "healthchecks", "statuspages", "incidents", "maintenance", "outages"}
	for i, r := range resources {
		pr.Step(r)
		pr.Report(i+1, "resource(s)")
	}

	output := buf.String()

	// Verify all steps are present
	for i, r := range resources {
		expected := "[" + string(rune('1'+i)) + "/6] Fetching " + r
		if !bytes.Contains(buf.Bytes(), []byte(expected)) {
			t.Errorf("Expected %q in output, got: %s", expected, output)
		}
	}
}
