// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"os"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/openapi"
)

// makeMonitorSpec builds a minimal spec with POST /v1/monitors whose request body
// contains a "regions" array field with the given enum values.
// Passing nil for regions omits the field entirely (simulates field removal).
func makeMonitorSpec(regions []string) *openapi.Spec {
	props := map[string]openapi.Schema{
		"name": {Type: "string", Description: "Monitor name"},
	}
	if regions != nil {
		props["regions"] = openapi.Schema{
			Type:        "array",
			Description: "Array of region names",
			Enum:        regions,
			Items:       &openapi.Schema{Type: "string"},
		}
	}

	return &openapi.Spec{
		OpenAPI: "3.0.3",
		Info:    openapi.Info{Title: "Hyperping API", Version: "test"},
		Servers: []openapi.Server{{URL: "https://api.hyperping.io"}},
		Paths: map[string]openapi.PathItem{
			"/v1/monitors": {
				Post: &openapi.Operation{
					OperationID: "POST_v1_monitors",
					Tags:        []string{"monitors"},
					RequestBody: &openapi.RequestBody{
						Required: true,
						Content: map[string]openapi.MediaType{
							"application/json": {
								Schema: openapi.Schema{
									Type:       "object",
									Properties: props,
								},
							},
						},
					},
					Responses: map[string]openapi.Response{
						"200": {Description: "OK"},
					},
				},
			},
		},
	}
}

// writeSpec serialises spec to a temp YAML file and returns its path.
func writeSpec(t *testing.T, spec *openapi.Spec) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "spec-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	f.Close()
	if err := openapi.Save(spec, f.Name()); err != nil {
		t.Fatalf("save spec: %v", err)
	}
	return f.Name()
}

func TestDetectEnumRegression(t *testing.T) {
	allRegions := []string{
		"london", "frankfurt", "paris", "amsterdam",
		"singapore", "sydney", "tokyo", "seoul", "mumbai", "bangalore",
		"virginia", "california", "sanfrancisco", "nyc", "toronto",
		"saopaulo", "bahrain", "capetown",
	}
	withoutCapetown := allRegions[:len(allRegions)-1]

	tests := []struct {
		name        string
		prevRegions []string
		newRegions  []string
		wantCount   int    // expected number of regressions
		wantMissing string // comma-joined missing values (for first regression, if any)
	}{
		{
			name:        "equal enums — no regression",
			prevRegions: allRegions,
			newRegions:  allRegions,
			wantCount:   0,
		},
		{
			name:        "enum grew — no regression",
			prevRegions: withoutCapetown,
			newRegions:  allRegions,
			wantCount:   0,
		},
		{
			name:        "one value removed",
			prevRegions: allRegions,
			newRegions:  withoutCapetown,
			wantCount:   1,
			wantMissing: "capetown",
		},
		{
			name:        "enum collapsed to empty",
			prevRegions: allRegions,
			newRegions:  []string{},
			wantCount:   1,
		},
		{
			name:        "field removed entirely from new spec",
			prevRegions: allRegions,
			newRegions:  nil, // nil → field absent in makeMonitorSpec
			wantCount:   1,
		},
		{
			name:        "prev has no enum — nothing to regress",
			prevRegions: nil,
			newRegions:  allRegions,
			wantCount:   0,
		},
		{
			name:        "both empty enums — no regression",
			prevRegions: []string{},
			newRegions:  []string{},
			wantCount:   0,
		},
		{
			name:        "multiple values removed",
			prevRegions: allRegions,
			newRegions:  []string{"london", "frankfurt", "paris"},
			wantCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prevSpec := makeMonitorSpec(tt.prevRegions)
			newSpec := makeMonitorSpec(tt.newRegions)
			prevPath := writeSpec(t, prevSpec)

			regressions, err := DetectEnumRegression(prevPath, newSpec)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(regressions) != tt.wantCount {
				t.Errorf("regression count: want %d, got %d (regressions: %+v)",
					tt.wantCount, len(regressions), regressions)
				return
			}

			if tt.wantCount > 0 {
				r := regressions[0]
				if r.Field != "regions" {
					t.Errorf("field: want %q, got %q", "regions", r.Field)
				}
				if r.Method != "POST" {
					t.Errorf("method: want %q, got %q", "POST", r.Method)
				}
				if r.Path != "/v1/monitors" {
					t.Errorf("path: want %q, got %q", "/v1/monitors", r.Path)
				}
				if tt.wantMissing != "" {
					missing := missingEnumValues(r.OldValues, r.NewValues)
					if len(missing) != 1 || missing[0] != tt.wantMissing {
						t.Errorf("missing values: want [%s], got %v", tt.wantMissing, missing)
					}
				}
			}
		})
	}
}

func TestDetectEnumRegression_BadPrevPath(t *testing.T) {
	newSpec := makeMonitorSpec([]string{"london"})
	_, err := DetectEnumRegression("/nonexistent/path.yaml", newSpec)
	if err == nil {
		t.Error("expected error for missing prev spec, got nil")
	}
}

func TestRegressionSetsMatch(t *testing.T) {
	r1 := EnumRegression{Path: "/v1/monitors", Method: "POST", Field: "regions",
		OldValues: []string{"a", "b", "c"}, NewValues: []string{"a", "b"}}
	r2 := EnumRegression{Path: "/v1/monitors", Method: "PUT", Field: "regions",
		OldValues: []string{"a", "b", "c"}, NewValues: []string{"a", "b"}}

	tests := []struct {
		name  string
		prev  []EnumRegression
		curr  []EnumRegression
		match bool
	}{
		{"both empty", nil, nil, true},
		{"same single regression", []EnumRegression{r1}, []EnumRegression{r1}, true},
		{"different method", []EnumRegression{r1}, []EnumRegression{r2}, false},
		{"prev empty, curr not", nil, []EnumRegression{r1}, false},
		{"same two regressions", []EnumRegression{r1, r2}, []EnumRegression{r2, r1}, true},
		{"different lengths", []EnumRegression{r1}, []EnumRegression{r1, r2}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := regressionSetsMatch(tt.prev, tt.curr)
			if got != tt.match {
				t.Errorf("regressionSetsMatch() = %v, want %v", got, tt.match)
			}
		})
	}
}

func TestDegradedState_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	sm := NewSnapshotManager(dir)

	// Initial load returns empty state.
	state, err := sm.LoadDegradedState()
	if err != nil {
		t.Fatalf("LoadDegradedState on empty dir: %v", err)
	}
	if state.ConsecutiveCount != 0 {
		t.Errorf("fresh state: want count 0, got %d", state.ConsecutiveCount)
	}

	// Save a state and reload it.
	state.ConsecutiveCount = 2
	state.Regressions = []EnumRegression{
		{Path: "/v1/monitors", Method: "POST", Field: "regions",
			OldValues: []string{"a", "b"}, NewValues: []string{"a"}},
	}
	err = sm.SaveDegradedState(state)
	if err != nil {
		t.Fatalf("SaveDegradedState: %v", err)
	}

	loaded, err := sm.LoadDegradedState()
	if err != nil {
		t.Fatalf("LoadDegradedState after save: %v", err)
	}
	if loaded.ConsecutiveCount != 2 {
		t.Errorf("loaded count: want 2, got %d", loaded.ConsecutiveCount)
	}
	if len(loaded.Regressions) != 1 {
		t.Errorf("loaded regressions: want 1, got %d", len(loaded.Regressions))
	}

	// Reset clears the file.
	err = sm.ResetDegradedState()
	if err != nil {
		t.Fatalf("ResetDegradedState: %v", err)
	}
	after, err := sm.LoadDegradedState()
	if err != nil {
		t.Fatalf("LoadDegradedState after reset: %v", err)
	}
	if after.ConsecutiveCount != 0 {
		t.Errorf("after reset: want count 0, got %d", after.ConsecutiveCount)
	}

	// Double-reset is idempotent (file already gone).
	if err := sm.ResetDegradedState(); err != nil {
		t.Errorf("double reset should be idempotent, got: %v", err)
	}
}
