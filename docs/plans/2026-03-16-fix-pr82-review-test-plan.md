# Test Plan: Fix PR #82 Review Findings

**Date**: 2026-03-16
**Branch**: feat/hyperping-exporter (worktree: fix-pr82-review)
**Plan**: `.trycycle/plan.md`

---

## Strategy reconciliation

The implementation plan covers three mechanical, non-logic changes:

1. **Item 1** — Delete two process-artifact files (`.trycycle/plan.md`, `docs/plans/2026-03-16-iac-usability-test-plan.md`) and add their parent directories to `.gitignore`.
2. **Item 2** — Move `testAccProviderConfig` and `tfInt` from `monitor_resource_protocols_test.go` to `monitor_resource_test_helpers.go` and remove the orphaned `"fmt"` import.
3. **Item 3** — Extract `checkPortNotSet` helper in `monitor_validate_config.go`; refactor `validatePortNotSet` and `validateHTTPProtocol` to delegate to it.

**No production logic changes.** The plan explicitly states: "No new tests (existing tests cover all modified paths)." This is correct — every code path exercised by the refactor is already covered by the existing acceptance test suite (T01–T16 in `monitor_validate_config_test.go` and the protocol tests in `monitor_resource_protocols_test.go`).

The testing strategy therefore is:

- **Regression**: `go test ./internal/provider/... -count=1` with `TF_ACC=1` must produce zero failures. This is the primary quality gate.
- **Compile invariants**: `go build ./internal/provider/...` and `go vet ./...` must pass — the refactors introduce no new imports, and the `"fmt"` removal must be verified at compile time.
- **Lint**: `make lint` must produce zero issues.
- **Structural invariants** (unit-level): verify via `grep`/compile that `testAccProviderConfig` and `tfInt` have exactly one definition in the package after the move, and that `monitor_resource_protocols_test.go` no longer imports `"fmt"`.

No new harnesses are required. All scenarios exercise the real provider surface (Terraform plan-time validation) through the existing `newMockHyperpingServer` / `newMinimalMockServer` infrastructure already in place.

---

## Test plan

### T1 — Full acceptance suite passes without regressions

- **Type**: regression
- **Harness**: existing `TF_ACC=1 go test ./internal/provider/... -count=1 -timeout 10m`
- **Preconditions**: Implementation complete (all three items applied). Working tree clean.
- **Actions**: Run `TF_ACC=1 go test ./internal/provider/... -count=1 -timeout 10m`.
- **Expected outcome**: `PASS` with 0 failures. Any failure not present before the change is a regression introduced by the refactor. Source of truth: the existing test suite — each test was green on `feat/hyperping-exporter` before the fix branch was created.
- **Interactions**: Exercises `monitor_resource_protocols_test.go` (which calls `testAccProviderConfig` and `tfInt` from their new location in `monitor_resource_test_helpers.go`), `monitor_validate_config_test.go` (which exercises `validatePortNotSet` and `validateHTTPProtocol` via their refactored `checkPortNotSet` delegation), and all other provider tests.

---

### T2 — Package compiles and vet passes

- **Type**: invariant
- **Harness**: `go build ./... && go vet ./...`
- **Preconditions**: Implementation complete.
- **Actions**: Run `go build ./... && go vet ./...`.
- **Expected outcome**: Exit code 0, no output. A non-zero exit indicates either (a) the `"fmt"` import was not removed from `monitor_resource_protocols_test.go` after `tfInt` was moved, or (b) `checkPortNotSet` was introduced with a signature that doesn't satisfy its callers. Source of truth: Go compiler.
- **Interactions**: Validates Item 2 (`"fmt"` import removal) and Item 3 (correct helper signature).

---

### T3 — `testAccProviderConfig` has exactly one definition in the package

- **Type**: invariant
- **Harness**: `grep -rn "func testAccProviderConfig" ./internal/provider/`
- **Preconditions**: `testAccProviderConfig` moved from `monitor_resource_protocols_test.go` to `monitor_resource_test_helpers.go`.
- **Actions**: Run `grep -rn "func testAccProviderConfig" ./internal/provider/` and count matches.
- **Expected outcome**: Exactly one match, located in `monitor_resource_test_helpers.go`. Zero matches means the move was incomplete (protocols test still compiled via a definition elsewhere, masking a missing symbol). Two or more matches means the function was added to helpers but not removed from protocols. Source of truth: Go's compile-time duplicate-symbol detection also catches >1 definition, but this grep provides explicit human-readable evidence for the PR review.
- **Interactions**: None; this is a static structural check.

---

### T4 — `tfInt` has exactly one definition in the package

- **Type**: invariant
- **Harness**: `grep -rn "func tfInt\b" ./internal/provider/`
- **Preconditions**: `tfInt` moved from `monitor_resource_protocols_test.go` to `monitor_resource_test_helpers.go`.
- **Actions**: Run `grep -rn "^func tfInt\b" ./internal/provider/` and count matches.
- **Expected outcome**: Exactly one match, in `monitor_resource_test_helpers.go`. Source of truth: same as T3.
- **Interactions**: None; static check.

---

### T5 — `monitor_resource_protocols_test.go` does not import `"fmt"`

- **Type**: invariant
- **Harness**: `grep -n '"fmt"' ./internal/provider/monitor_resource_protocols_test.go`
- **Preconditions**: `tfInt` (the only `fmt.Sprintf` caller in that file) has been removed.
- **Actions**: Run `grep -n '"fmt"' ./internal/provider/monitor_resource_protocols_test.go`.
- **Expected outcome**: No output (grep exits 1). Any match is a compile error in disguise — Go will reject an unused import. Source of truth: Go compiler, which rejects unused imports.
- **Interactions**: Validates Item 2, step 4 of the execution order.

---

### T6 — `checkPortNotSet` helper is present and called by both delegating functions

- **Type**: invariant
- **Harness**: `grep -n "checkPortNotSet\|func validatePortNotSet\|func validateHTTPProtocol" ./internal/provider/monitor_validate_config.go`
- **Preconditions**: Item 3 applied.
- **Actions**: Run the grep and inspect output.
- **Expected outcome**:
  - `func checkPortNotSet` appears exactly once (the new helper definition).
  - `validatePortNotSet` body contains `checkPortNotSet(` (delegates, does not duplicate the loop).
  - `validateHTTPProtocol` body contains `checkPortNotSet(` (delegates, does not duplicate the loop).
  Source of truth: implementation plan § Item 3.
- **Interactions**: None; static.

---

### T7 — `validatePortNotSet` error message preserved verbatim (regression)

- **Type**: regression
- **Harness**: `TF_ACC=1 go test ./internal/provider/ -run TestAccMonitorResource_icmpRejectsPort -v -timeout 60s`
- **Preconditions**: Item 3 applied.
- **Actions**: Run the test.
- **Expected outcome**: `PASS`. The test asserts `regexp.MustCompile(`(?i)port.*not valid.*icmp`)` against the diagnostic. If the refactor changes the error message, this test fails. Source of truth: `monitor_validate_config_test.go` T07.
- **Interactions**: Exercises `validatePortNotSet` → `checkPortNotSet` delegation path.

---

### T8 — `validateHTTPProtocol` error message preserved verbatim (regression)

- **Type**: regression
- **Harness**: `TF_ACC=1 go test ./internal/provider/ -run TestAccMonitorResource_httpRejectsPort -v -timeout 60s`
- **Preconditions**: Item 3 applied.
- **Actions**: Run the test.
- **Expected outcome**: `PASS`. The test asserts `regexp.MustCompile(`(?i)port.*not valid.*http`)`. Source of truth: `monitor_validate_config_test.go` T10.
- **Interactions**: Exercises `validateHTTPProtocol` → `checkPortNotSet` delegation path.

---

### T9 — Process-artifact files are no longer tracked by git

- **Type**: invariant
- **Harness**: `git ls-files .trycycle/ docs/plans/`
- **Preconditions**: Item 1 applied (`git rm` executed).
- **Actions**: Run `git ls-files .trycycle/ docs/plans/` from worktree root.
- **Expected outcome**: Empty output (no tracked files in those directories). Source of truth: plan § Item 1 ("git rm both files").
- **Interactions**: None; git index check.

---

### T10 — `.gitignore` prevents future accidental commits of `.trycycle/` and `docs/plans/`

- **Type**: invariant
- **Harness**: `grep -E "^\.trycycle/|^docs/plans/" .gitignore`
- **Preconditions**: Item 1 `.gitignore` update applied.
- **Actions**: Run the grep.
- **Expected outcome**: Two matching lines: `.trycycle/` and `docs/plans/`. Source of truth: plan § Item 1 decisions.
- **Interactions**: None; static check.

---

### T11 — Lint passes with zero issues

- **Type**: invariant
- **Harness**: `make lint`
- **Preconditions**: All three items applied.
- **Actions**: Run `make lint`.
- **Expected outcome**: Exit code 0, no lint warnings or errors. Source of truth: project quality gate (GNUmakefile `lint` target).
- **Interactions**: Validates that the `"fmt"` removal (T5) and the `checkPortNotSet` extraction don't introduce any linting issues.

---

## Coverage summary

### Covered

| Area | Test(s) | How covered |
|------|---------|-------------|
| Full test suite regression | T1 | All 700+ tests run; any regression from the move or refactor surfaces immediately |
| Compile / vet | T2 | Catches orphaned import, wrong function signatures |
| `testAccProviderConfig` single definition | T3 | Structural grep |
| `tfInt` single definition | T4 | Structural grep |
| `"fmt"` import removed from protocols test | T5 | Grep + implicit via compile (T2) |
| `checkPortNotSet` delegation present | T6 | Structural grep |
| `validatePortNotSet` error message | T7 | Existing acceptance test T07 |
| `validateHTTPProtocol` error message | T8 | Existing acceptance test T10 |
| Artifact files removed from git | T9 | `git ls-files` |
| `.gitignore` entries added | T10 | grep |
| Lint clean | T11 | `make lint` |

### Explicitly excluded

| Area | Reason |
|------|--------|
| New behavior tests for `checkPortNotSet` | No new behavior introduced; helper is internal. Existing T07 and T10 already exercise both call paths with the real error message. |
| Performance testing | No performance-sensitive code changed. |
| End-to-end / API integration tests | All changes are in test helpers and an internal refactor; no production CRUD paths change. |
| Tests for `.gitignore` effectiveness at git-commit time | The grep (T10) is sufficient; a full `git add` simulation would be over-engineered for a one-line `.gitignore` change. |

### Risk carried by exclusions

**Low**. The primary risk is a subtle compile error from the `"fmt"` removal, which is caught definitively by T2 (compile) and T5 (grep). The risk of a silently-wrong refactor in `checkPortNotSet` is caught by T7 and T8 (message-preserving regression tests) and T1 (full suite). No exclusion leaves a meaningful user-visible behavior untested.
