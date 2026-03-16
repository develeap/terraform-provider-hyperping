# Fix PR #82 Review Findings

## Goal

Address three categories of code-review findings on the `feat/hyperping-exporter` branch (worktree: `fix-pr82-review`) that were raised after the IaC usability improvements were committed:

1. Delete process-artifact files that were accidentally committed.
2. Move `testAccProviderConfig` to the canonical shared test helpers file.
3. DRY up the overlapping `validateHTTPProtocol` and `validatePortNotSet` functions.

All three changes are mechanical and non-breaking. No logic changes to production code are required beyond the refactor in item 3.

---

## Item 1 — Remove Process Artifacts

### What to delete

Two files were committed that are trycycle planning/testing artifacts, not code that belongs in the repository:

| File | Reason |
|------|--------|
| `.trycycle/plan.md` | Trycycle planning artifact; not source code |
| `docs/plans/2026-03-16-iac-usability-test-plan.md` | Trycycle test-plan artifact; not source code |

### Verification

Confirmed by `git ls-files .trycycle/ docs/plans/`:
```
.trycycle/plan.md
docs/plans/2026-03-16-iac-usability-test-plan.md
```

Both files are present and are purely process artifacts with no bearing on compilation, tests, or documentation.

The `.trycycle/` directory contains only `plan.md`. The test-plan artifact lives under `docs/plans/`, not under `.trycycle/`.

### Action

`git rm` both files. This removes them from the index and working tree simultaneously, producing a clean deletion commit entry.

**Decision — add `.trycycle/` to `.gitignore`**: The `.trycycle/` directory is not mentioned in the current `.gitignore`. The correct fix is to delete the committed file and add `.trycycle/` to `.gitignore` to prevent future accidents. This eliminates the root cause and is trivially safe.

**Decision — add `docs/plans/` to `.gitignore`**: Similarly, `docs/plans/` is a trycycle process-artifact directory (not part of the provider's user-facing docs). Adding it to `.gitignore` prevents recurrence for this location too. The provider's generated docs live in `docs/resources/`, `docs/data-sources/`, etc. — not in `docs/plans/`.

---

## Item 2 — Consolidate `testAccProviderConfig` into Shared Helpers

### Current state

`testAccProviderConfig(baseURL string) string` is defined at line 278 of `internal/provider/monitor_resource_protocols_test.go`. It is used only within that same file (lines 193, 204, 214, 227, 243, 255, 265).

There is no other definition of this function anywhere in the test suite. The PR review flagged it as a potential redefinition risk and asked that it be moved to a shared location.

### Correct shared location

`internal/provider/monitor_resource_test_helpers.go` is the established shared file for monitor test helpers. It already contains:
- All `testAccMonitorResourceConfig*` functions (config generators)
- `newMockHyperpingServer` and `newMockHyperpingServerWithErrors` (shared mock infrastructure)
- Helper functions like `getOrDefault*`

Moving `testAccProviderConfig` there follows the established pattern exactly. `provider_test.go` is not the right home because it contains provider unit tests and uses inline `fmt.Sprintf` provider blocks, not a shared config generator.

**Decision — move to `monitor_resource_test_helpers.go`, not `provider_test.go`**: The function signature `testAccProviderConfig(baseURL string) string` is specific to monitor acceptance tests (it takes a `baseURL` for mock server injection). Provider-level acceptance tests in `provider_test.go` use inline provider blocks. Moving to `monitor_resource_test_helpers.go` co-locates it with the functions that call it and follows the single-responsibility principle of that file.

### Also move `tfInt`

`tfInt(val int) string` (line 287 of `monitor_resource_protocols_test.go`) is a simple format helper specific to Terraform integer rendering in HCL configs. It is used in the same file as `testAccProviderConfig`. Move it to `monitor_resource_test_helpers.go` alongside `testAccProviderConfig` to complete the cleanup.

**Decision — move `tfInt` together with `testAccProviderConfig`**: Both are pure helper functions with no side effects. Moving them together leaves `monitor_resource_protocols_test.go` containing only test cases and config generators that call the shared helpers, which is architecturally clean.

### Steps

1. Add `testAccProviderConfig` and `tfInt` to `internal/provider/monitor_resource_test_helpers.go` (append at end of file).
2. Remove those two function definitions from `internal/provider/monitor_resource_protocols_test.go`.
3. Remove the `"fmt"` import from `monitor_resource_protocols_test.go` since `tfInt` was the only caller of `fmt.Sprintf` in that file.
4. Verify the package still compiles: `go build ./internal/provider/...` (or `make test`).

**Note on `"fmt"` import**: After removing `tfInt`, the `fmt` package is no longer used in `monitor_resource_protocols_test.go`. The import must be removed to avoid a compile error. `monitor_resource_test_helpers.go` already imports `"fmt"` so no new import is needed there.

**Note on file size**: `monitor_resource_test_helpers.go` is currently 673 lines. Adding `testAccProviderConfig` (7 lines) and `tfInt` (3 lines) brings it to ~683 lines, well within the 800-line limit.

---

## Item 3 — DRY Up `validateHTTPProtocol` and `validatePortNotSet`

### Current state (`internal/provider/monitor_validate_config.go`)

`validatePortNotSet` (lines 72-86):
```go
func validatePortNotSet(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse, protocol string) {
    var port types.Int64
    resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("port"), &port)...)
    if resp.Diagnostics.HasError() { return }
    if !port.IsNull() && !port.IsUnknown() {
        resp.Diagnostics.AddAttributeError(
            path.Root("port"),
            "Invalid Attribute Combination",
            fmt.Sprintf("port is not valid when protocol is %q. Remove port or change protocol to \"port\".", protocol),
        )
    }
}
```

`validateHTTPProtocol` (lines 88-103):
```go
func validateHTTPProtocol(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
    var port types.Int64
    resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("port"), &port)...)
    if resp.Diagnostics.HasError() { return }
    if !port.IsNull() && !port.IsUnknown() {
        resp.Diagnostics.AddAttributeError(
            path.Root("port"),
            "Invalid Attribute Combination",
            "port is not valid when protocol is \"http\". The URL contains the port for HTTP monitors. Remove port or change protocol to \"port\".",
        )
    }
}
```

The two functions are structurally identical — read `port`, check null/unknown, emit `AddAttributeError` — and differ only in the error message string. `validateHTTPProtocol` is called for `protocol = "http"`, `validatePortNotSet` for `protocol = "icmp"` (passing the protocol name as the parameter).

### Solution

Extract the shared logic into a private helper `checkPortNotSet(ctx, req, resp, errorDetail string)` that accepts the pre-composed detail string. Both callers construct their detail string and delegate to this helper.

**New helper signature**:
```go
// checkPortNotSet reads the port attribute and adds an error if it is explicitly set.
// errorDetail is the full human-readable detail message to use in the diagnostic.
func checkPortNotSet(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse, errorDetail string) {
    var port types.Int64
    resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("port"), &port)...)
    if resp.Diagnostics.HasError() {
        return
    }
    if !port.IsNull() && !port.IsUnknown() {
        resp.Diagnostics.AddAttributeError(
            path.Root("port"),
            "Invalid Attribute Combination",
            errorDetail,
        )
    }
}
```

**Updated `validatePortNotSet`**:
```go
func validatePortNotSet(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse, protocol string) {
    checkPortNotSet(ctx, req, resp,
        fmt.Sprintf("port is not valid when protocol is %q. Remove port or change protocol to \"port\".", protocol),
    )
}
```

**Updated `validateHTTPProtocol`**:
```go
func validateHTTPProtocol(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
    checkPortNotSet(ctx, req, resp,
        "port is not valid when protocol is \"http\". The URL contains the port for HTTP monitors. Remove port or change protocol to \"port\".",
    )
}
```

**Decision — keep `validatePortNotSet` and `validateHTTPProtocol` as named functions**: The call sites in `ValidateConfig` (lines 35 and 40) use these named functions. Keeping them as thin wrappers preserves readability of the switch statement and allows future divergence (e.g., if the HTTP-specific message needs to change) without touching the call sites. The duplication is eliminated at the implementation level while the public interface of the validate logic remains expressive.

**Decision — do not change error messages**: Both error messages are tested by existing tests (`TestAccMonitorResource_icmpRejectsPort` matches `port.*not valid.*icmp`, `TestAccMonitorResource_httpRejectsPort` matches `port.*not valid.*http`). The refactor preserves these messages verbatim.

**Note — consistent with existing `check*NotSet` helpers**: The file already uses this pattern for `checkStringNotSet`, `checkBoolNotSet`, and `checkListNotSet`. Adding `checkPortNotSet` is idiomatic within the file's own structure.

---

## Execution Order

1. `git rm .trycycle/plan.md docs/plans/2026-03-16-iac-usability-test-plan.md`
2. Add `.trycycle/` and `docs/plans/` to `.gitignore`
3. Move `testAccProviderConfig` and `tfInt` from `monitor_resource_protocols_test.go` to `monitor_resource_test_helpers.go`
4. Remove `"fmt"` import from `monitor_resource_protocols_test.go`
5. Add `checkPortNotSet` helper to `monitor_validate_config.go`; refactor `validatePortNotSet` and `validateHTTPProtocol` to use it
6. Run `make test` — all existing tests must pass (no new tests required; the behavior is unchanged)
7. Run `make lint` — zero issues required

---

## File Change Summary

| File | Change |
|------|--------|
| `.trycycle/plan.md` | DELETE (git rm) |
| `docs/plans/2026-03-16-iac-usability-test-plan.md` | DELETE (git rm) |
| `.gitignore` | Add `.trycycle/` and `docs/plans/` entries |
| `internal/provider/monitor_resource_test_helpers.go` | Add `testAccProviderConfig` and `tfInt` at end of file |
| `internal/provider/monitor_resource_protocols_test.go` | Remove `testAccProviderConfig`, `tfInt`, and unused `"fmt"` import |
| `internal/provider/monitor_validate_config.go` | Add `checkPortNotSet` helper; refactor `validatePortNotSet` and `validateHTTPProtocol` |

---

## Risks and Mitigations

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Compile error from orphaned `"fmt"` import | Low | Plan explicitly calls out removing the import; `make test` will catch it immediately |
| Test behavior changed by refactor | None | Error messages are preserved verbatim; only the internal structure of the validate functions changes |
| Future `.trycycle/` files accidentally committed | Low after fix | `.gitignore` entry prevents recurrence |
| Future `docs/plans/` files accidentally committed | Low after fix | `.gitignore` entry prevents recurrence |
| `testAccProviderConfig` redefinition if another file adds it | None after move | Single definition in `monitor_resource_test_helpers.go`; Go compile-time duplicate detection covers any future accidents |

---

## Non-Goals

- No changes to production logic (CRUD, schema, mapping)
- No new tests (existing tests cover all modified paths)
- No changes to any other file beyond those listed above
- No changes to the client package
