# Contract Testing Implementation Plan

## Goal
Discover undocumented API fields by recording real API responses, extracting schemas from cassettes, and comparing against both scraped docs and provider implementation.

## Phase 1: Response Recording with go-vcr

### Objective
Add go-vcr to acceptance tests to record real API request/response pairs.

### DOD (Definition of Done)
- [x] go-vcr dependency added to go.mod
- [x] Test helper created for VCR recording mode
- [x] At least one acceptance test modified to use VCR
- [ ] Cassette file generated with real API interaction
- [ ] Verification: cassette contains request/response JSON

### Tasks
1. ~~Add go-vcr/v2 dependency~~ ✅ Added gopkg.in/dnaeon/go-vcr.v3
2. ~~Create `internal/provider/testutil/vcr.go` helper~~ ✅
3. ~~Create VCR-based live contract tests~~ ✅ Created `internal/client/live_contract_test.go`
4. Run test with real API to generate cassette (requires HYPERPING_API_KEY)
5. Verify cassette structure

### Files Created
- `internal/provider/testutil/vcr.go` - VCR test helper with mode detection
- `internal/provider/testutil/vcr_test.go` - Unit tests for VCR helper
- `internal/client/live_contract_test.go` - Live contract tests for CRUD operations

### Usage
```bash
# Replay mode (default) - uses existing cassettes, skips if none exist
go test ./internal/client/... -run "TestLiveContract" -v

# Record mode - records real API interactions
RECORD_MODE=true HYPERPING_API_KEY=xxx go test ./internal/client/... -run "TestLiveContract" -v

# Field discovery mode - lists all fields from real API
DISCOVER_FIELDS=true HYPERPING_API_KEY=xxx go test ./internal/client/... -run "TestLiveContract_DiscoverFields" -v
```

---

## Phase 2: Schema Extraction from Cassettes

### Objective
Build extractor that reads cassette YAML files and extracts API field schemas.

### DOD (Definition of Done)
- [x] `contract/` package created with types
- [x] Parser for cassette YAML format
- [x] Type inference from JSON values
- [x] Unit tests with sample cassettes
- [x] Verification: extracted schema matches expected

### Tasks
1. ~~Create `tools/cmd/scraper/contract/types.go`~~ ✅
2. ~~Create `tools/cmd/scraper/contract/extractor.go`~~ ✅
3. ~~Create `tools/cmd/scraper/contract/extractor_test.go`~~ ✅
4. ~~Test with sample cassettes~~ ✅

### Files Created
- `tools/cmd/scraper/contract/types.go` - Schema types (APIFieldSchema, EndpointSchema, CassetteSchema)
- `tools/cmd/scraper/contract/extractor.go` - Cassette parser and field extractor
- `tools/cmd/scraper/contract/extractor_test.go` - Unit tests
- `tools/cmd/scraper/contract/testdata/sample_cassette.yaml` - Test fixture

### Capabilities
- Parses go-vcr cassette YAML format
- Extracts request and response fields with types
- Infers types: string, integer, number, boolean, array, object, null
- Tracks nullable fields (when value is null in any response)
- Normalizes paths by replacing IDs with `{id}` placeholders
- Extracts nested object and array element fields
- Tracks observed HTTP status codes

---

## Phase 3: Three-Way Comparison

### Objective
Enhance analyzer to compare: API docs (scraped) vs Provider (AST) vs Actual API (cassettes).

### DOD (Definition of Done)
- [ ] Analyzer accepts cassette directory parameter
- [ ] Three-way comparison logic implemented
- [ ] Report shows documentation gaps discovered via contract testing
- [ ] Verification: correctly identifies undocumented fields

### Tasks
1. Add cassette data source to analyzer
2. Implement three-way comparison
3. Enhance coverage report format
4. Test with real cassettes

---

## Phase 4: CI Integration

### Objective
Automate contract testing in CI/CD pipeline.

### DOD (Definition of Done)
- [ ] CI job runs acceptance tests with recording
- [ ] Cassettes stored as artifacts
- [ ] Schema extraction runs after recording
- [ ] Analysis runs with cassette input
- [ ] Verification: pipeline completes and reports gaps

### Tasks
1. Add record mode environment variable
2. Update scraper workflow
3. Store cassettes as artifacts
4. Add analysis with cassettes

---

## Current Status

**Phase**: 2 - Schema Extraction
**Status**: Complete

### Phase 1 Summary (Response Recording)
- Added go-vcr v3 dependency for HTTP recording/replay
- Created VCR test helper with three modes:
  - **ModeReplay**: Uses existing cassettes, fails if none exist
  - **ModeRecord**: Always records new interactions
  - **ModeAuto**: Skips if no cassette exists (safe for CI)
- Created live contract tests for:
  - Healthcheck CRUD
  - Monitor CRUD
  - Outage CRUD
  - Field discovery (logs all API response fields)
- Sensitive data (API keys, auth headers) automatically masked in cassettes
- Tests skip gracefully when no cassettes exist and not in record mode

### Phase 2 Summary (Schema Extraction)
- Created `contract/` package for schema extraction
- Parses go-vcr cassette YAML format
- Extracts request/response fields with inferred types
- Handles nested objects, arrays, nullable fields
- All unit tests passing

### Next Steps
1. Proceed to Phase 3: Three-way comparison in analyzer
2. Integrate cassette schema with existing coverage analyzer
