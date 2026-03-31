// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package diff

import (
	"os"
	"strings"
	"testing"

	"github.com/oasdiff/oasdiff/diff"
)

func TestFormatMarkdown_NilDiff(t *testing.T) {
	result := FormatMarkdown(nil)
	if result != "No API changes detected." {
		t.Errorf("expected no changes message, got: %s", result)
	}
}

func TestFormatMarkdown_EmptyDiff(t *testing.T) {
	d := &diff.Diff{}
	result := FormatMarkdown(d)
	if result != "No API changes detected." {
		t.Errorf("expected no changes message, got: %s", result)
	}
}

func TestFormatMarkdown_AddedEndpoints(t *testing.T) {
	d := &diff.Diff{
		PathsDiff: &diff.PathsDiff{
			Added: []string{"/v1/new-endpoint"},
		},
	}

	result := FormatMarkdown(d)

	if !strings.Contains(result, "### New Endpoints") {
		t.Error("expected New Endpoints section")
	}
	if !strings.Contains(result, "`/v1/new-endpoint`") {
		t.Error("expected new endpoint path in output")
	}
}

func TestFormatMarkdown_DeletedEndpoints(t *testing.T) {
	d := &diff.Diff{
		PathsDiff: &diff.PathsDiff{
			Deleted: []string{"/v1/removed-endpoint"},
		},
	}

	result := FormatMarkdown(d)

	if !strings.Contains(result, "### Removed Endpoints (BREAKING)") {
		t.Error("expected Removed Endpoints section")
	}
	if !strings.Contains(result, "`/v1/removed-endpoint`") {
		t.Error("expected removed endpoint path in output")
	}
}

func TestFormatMarkdown_ModifiedWithPropertyDescriptionChange(t *testing.T) {
	d := &diff.Diff{
		PathsDiff: &diff.PathsDiff{
			Modified: map[string]*diff.PathDiff{
				"/v1/monitors": {
					OperationsDiff: &diff.OperationsDiff{
						Modified: map[string]*diff.MethodDiff{
							"post": {
								RequestBodyDiff: &diff.RequestBodyDiff{
									ContentDiff: &diff.ContentDiff{
										MediaTypeModified: map[string]*diff.MediaTypeDiff{
											"application/json": {
												SchemaDiff: &diff.SchemaDiff{
													PropertiesDiff: &diff.SchemasDiff{
														Modified: map[string]*diff.SchemaDiff{
															"expected_status_code": {
																DescriptionDiff: &diff.ValueDiff{
																	From: "old description",
																	To:   "new description",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result := FormatMarkdown(d)

	if !strings.Contains(result, "### Modified Endpoints") {
		t.Error("expected Modified Endpoints section")
	}
	if !strings.Contains(result, "#### `POST /v1/monitors`") {
		t.Error("expected POST /v1/monitors heading")
	}
	if !strings.Contains(result, "`~ expected_status_code`: description updated") {
		t.Error("expected field-level description change detail")
	}
}

func TestFormatMarkdown_ModifiedWithAddedAndDeletedProperties(t *testing.T) {
	d := &diff.Diff{
		PathsDiff: &diff.PathsDiff{
			Modified: map[string]*diff.PathDiff{
				"/v1/monitors/{uuid}": {
					OperationsDiff: &diff.OperationsDiff{
						Modified: map[string]*diff.MethodDiff{
							"put": {
								RequestBodyDiff: &diff.RequestBodyDiff{
									ContentDiff: &diff.ContentDiff{
										MediaTypeModified: map[string]*diff.MediaTypeDiff{
											"application/json": {
												SchemaDiff: &diff.SchemaDiff{
													PropertiesDiff: &diff.SchemasDiff{
														Added:   []string{"new_field"},
														Deleted: []string{"old_field"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result := FormatMarkdown(d)

	if !strings.Contains(result, "`+ new_field` (new property)") {
		t.Error("expected added property detail")
	}
	if !strings.Contains(result, "`- old_field` (removed property)") {
		t.Error("expected deleted property detail")
	}
}

func TestFormatMarkdown_ModifiedWithTypeChange(t *testing.T) {
	d := &diff.Diff{
		PathsDiff: &diff.PathsDiff{
			Modified: map[string]*diff.PathDiff{
				"/v1/monitors": {
					OperationsDiff: &diff.OperationsDiff{
						Modified: map[string]*diff.MethodDiff{
							"get": {
								ResponsesDiff: &diff.ResponsesDiff{
									Modified: map[string]*diff.ResponseDiff{
										"200": {
											ContentDiff: &diff.ContentDiff{
												MediaTypeModified: map[string]*diff.MediaTypeDiff{
													"application/json": {
														SchemaDiff: &diff.SchemaDiff{
															PropertiesDiff: &diff.SchemasDiff{
																Modified: map[string]*diff.SchemaDiff{
																	"count": {
																		TypeDiff: &diff.StringsDiff{
																			Added:   []string{"integer"},
																			Deleted: []string{"string"},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result := FormatMarkdown(d)

	if !strings.Contains(result, "`~ count`: type changed") {
		t.Error("expected type change detail")
	}
}

func TestFormatMarkdown_OperationMetadataOnly(t *testing.T) {
	d := &diff.Diff{
		PathsDiff: &diff.PathsDiff{
			Modified: map[string]*diff.PathDiff{
				"/v1/monitors": {
					OperationsDiff: &diff.OperationsDiff{
						Modified: map[string]*diff.MethodDiff{
							"get": {
								TagsDiff: &diff.StringsDiff{
									Added: []string{"new-tag"},
								},
							},
						},
					},
				},
			},
		},
	}

	result := FormatMarkdown(d)

	if !strings.Contains(result, "_operation metadata changed_") {
		t.Error("expected fallback metadata message when no property changes")
	}
}

func TestFormatMarkdown_AllSections(t *testing.T) {
	d := &diff.Diff{
		PathsDiff: &diff.PathsDiff{
			Added:   []string{"/v1/new"},
			Deleted: []string{"/v1/removed"},
			Modified: map[string]*diff.PathDiff{
				"/v1/monitors": {
					OperationsDiff: &diff.OperationsDiff{
						Modified: map[string]*diff.MethodDiff{
							"post": {
								RequestBodyDiff: &diff.RequestBodyDiff{
									ContentDiff: &diff.ContentDiff{
										MediaTypeModified: map[string]*diff.MediaTypeDiff{
											"application/json": {
												SchemaDiff: &diff.SchemaDiff{
													PropertiesDiff: &diff.SchemasDiff{
														Modified: map[string]*diff.SchemaDiff{
															"field": {
																DescriptionDiff: &diff.ValueDiff{
																	From: "old",
																	To:   "new",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result := FormatMarkdown(d)

	if !strings.Contains(result, "### New Endpoints") {
		t.Error("expected New Endpoints section")
	}
	if !strings.Contains(result, "### Removed Endpoints (BREAKING)") {
		t.Error("expected Removed Endpoints section")
	}
	if !strings.Contains(result, "### Modified Endpoints") {
		t.Error("expected Modified Endpoints section")
	}
	if !strings.Contains(result, "`~ field`: description updated") {
		t.Error("expected field-level detail in Modified section")
	}
}

func TestDescribePropertyChange_MultipleChanges(t *testing.T) {
	sd := &diff.SchemaDiff{
		DescriptionDiff: &diff.ValueDiff{From: "old", To: "new"},
		DefaultDiff:     &diff.ValueDiff{From: "2xx", To: "200"},
	}

	result := describePropertyChange(sd)

	if !strings.Contains(result, "description updated") {
		t.Error("expected description updated")
	}
	if !strings.Contains(result, "default 2xx → 200") {
		t.Error("expected default change")
	}
}

func TestDescribePropertyChange_NoKnownChanges(t *testing.T) {
	sd := &diff.SchemaDiff{
		CircularRefDiff: true,
	}

	result := describePropertyChange(sd)
	if result != "schema changed" {
		t.Errorf("expected 'schema changed' fallback, got: %s", result)
	}
}

func TestSortedKeys(t *testing.T) {
	m := map[string]int{"c": 3, "a": 1, "b": 2}
	keys := sortedKeys(m)

	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
	if keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Errorf("expected [a b c], got %v", keys)
	}
}

func TestSortedKeys_EmptyMap(t *testing.T) {
	m := map[string]int{}
	keys := sortedKeys(m)
	if len(keys) != 0 {
		t.Errorf("expected empty slice, got %v", keys)
	}
}

func TestDescribePropertyChange_Nil(t *testing.T) {
	result := describePropertyChange(nil)
	if result != "schema changed" {
		t.Errorf("expected 'schema changed' for nil, got: %s", result)
	}
}

func TestDescribePropertyChange_EnumDiff(t *testing.T) {
	sd := &diff.SchemaDiff{
		EnumDiff: &diff.EnumDiff{
			Added:   diff.EnumValues{"active"},
			Deleted: diff.EnumValues{"disabled"},
		},
	}
	result := describePropertyChange(sd)
	if result != "enum values changed" {
		t.Errorf("expected 'enum values changed', got: %s", result)
	}
}

func TestFormatPathDiff_NilOperationsDiff(t *testing.T) {
	d := &diff.Diff{
		PathsDiff: &diff.PathsDiff{
			Modified: map[string]*diff.PathDiff{
				"/v1/test": {OperationsDiff: nil},
			},
		},
	}
	result := FormatMarkdown(d)
	if !strings.Contains(result, "_No operation changes._") {
		t.Error("expected 'No operation changes' message")
	}
}

func TestFormatPathDiff_AddedAndDeletedOperations(t *testing.T) {
	d := &diff.Diff{
		PathsDiff: &diff.PathsDiff{
			Modified: map[string]*diff.PathDiff{
				"/v1/monitors/{uuid}": {
					OperationsDiff: &diff.OperationsDiff{
						Added:   []string{"delete"},
						Deleted: []string{"patch"},
					},
				},
			},
		},
	}

	result := FormatMarkdown(d)

	if !strings.Contains(result, "`DELETE /v1/monitors/{uuid}` (new method)") {
		t.Error("expected new method listing for added operation")
	}
	if !strings.Contains(result, "`PATCH /v1/monitors/{uuid}` (removed method, BREAKING)") {
		t.Error("expected removed method listing for deleted operation")
	}
}

func TestFormatMethodDiff_ParameterChanges(t *testing.T) {
	d := &diff.Diff{
		PathsDiff: &diff.PathsDiff{
			Modified: map[string]*diff.PathDiff{
				"/v1/monitors": {
					OperationsDiff: &diff.OperationsDiff{
						Modified: map[string]*diff.MethodDiff{
							"get": {
								ParametersDiff: &diff.ParametersDiffByLocation{
									Added: diff.ParamNamesByLocation{
										"query": []string{"page", "per_page"},
									},
									Deleted: diff.ParamNamesByLocation{
										"query": []string{"offset"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result := FormatMarkdown(d)

	if !strings.Contains(result, "`+ page` (query parameter, new)") {
		t.Error("expected added query parameter")
	}
	if !strings.Contains(result, "`+ per_page` (query parameter, new)") {
		t.Error("expected added per_page parameter")
	}
	if !strings.Contains(result, "`- offset` (query parameter, removed)") {
		t.Error("expected removed query parameter")
	}
}

func TestFormatMarkdown_MultiplePathsSorted(t *testing.T) {
	d := &diff.Diff{
		PathsDiff: &diff.PathsDiff{
			Modified: map[string]*diff.PathDiff{
				"/v1/monitors/{uuid}": {OperationsDiff: nil},
				"/v1/healthchecks":    {OperationsDiff: nil},
				"/v1/monitors":        {OperationsDiff: nil},
			},
		},
	}

	result := FormatMarkdown(d)

	// Verify deterministic ordering
	idx1 := strings.Index(result, "/v1/healthchecks")
	idx2 := strings.Index(result, "/v1/monitors`")
	idx3 := strings.Index(result, "/v1/monitors/{uuid}")

	if idx1 < 0 || idx2 < 0 || idx3 < 0 {
		t.Fatalf("expected all three paths in output, got: %s", result)
	}
	if idx1 >= idx2 || idx2 >= idx3 {
		t.Errorf("expected paths in alphabetical order, got indices: healthchecks=%d, monitors=%d, monitors/{uuid}=%d", idx1, idx2, idx3)
	}
}

// TestCompare_MetadataOnlyChange verifies that HasPathChanges is false when
// diffs are metadata-only (e.g. info/description changed, no path changes).
// This was the root cause of false positive "API Change Detected" issues.
func TestCompare_MetadataOnlyChange(t *testing.T) {
	base := t.TempDir() + "/base.yaml"
	curr := t.TempDir() + "/curr.yaml"

	baseSpec := `openapi: "3.0.0"
info:
  title: "Test API"
  version: "1.0.0"
  description: "Original description"
paths:
  /monitors:
    get:
      summary: "List monitors"
      responses:
        "200":
          description: "OK"
`
	// Only the info description changed — no path changes.
	currSpec := `openapi: "3.0.0"
info:
  title: "Test API"
  version: "1.0.1"
  description: "Updated description"
paths:
  /monitors:
    get:
      summary: "List monitors"
      responses:
        "200":
          description: "OK"
`
	if err := writeFile(base, baseSpec); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(curr, currSpec); err != nil {
		t.Fatal(err)
	}

	result, err := Compare(base, curr)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	if !result.HasChanges {
		t.Error("expected HasChanges=true for metadata diff")
	}
	if result.HasPathChanges {
		t.Error("expected HasPathChanges=false for metadata-only diff")
	}
}

// TestCompare_PathLevelChange verifies HasPathChanges is true when endpoints change.
func TestCompare_PathLevelChange(t *testing.T) {
	base := t.TempDir() + "/base.yaml"
	curr := t.TempDir() + "/curr.yaml"

	baseSpec := `openapi: "3.0.0"
info:
  title: "Test API"
  version: "1.0.0"
paths:
  /monitors:
    get:
      summary: "List monitors"
      responses:
        "200":
          description: "OK"
`
	// Added a new endpoint.
	currSpec := `openapi: "3.0.0"
info:
  title: "Test API"
  version: "1.0.0"
paths:
  /monitors:
    get:
      summary: "List monitors"
      responses:
        "200":
          description: "OK"
  /healthchecks:
    get:
      summary: "List healthchecks"
      responses:
        "200":
          description: "OK"
`
	if err := writeFile(base, baseSpec); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(curr, currSpec); err != nil {
		t.Fatal(err)
	}

	result, err := Compare(base, curr)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	if !result.HasChanges {
		t.Error("expected HasChanges=true")
	}
	if !result.HasPathChanges {
		t.Error("expected HasPathChanges=true for added endpoint")
	}
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0600)
}
