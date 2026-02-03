# Semantic API Diff - Required Enhancement

## Problem

Current hash-based diff only detects "something changed" but not WHAT changed.

### False Positives
- Timestamps in HTML
- CSS/JavaScript version updates
- Navigation menu changes
- Footer copyright year
- Ad content rotation

### Missing Intelligence
- Can't identify new API parameters
- Can't detect breaking changes
- Can't classify change severity
- Can't generate useful changelogs

---

## Solution: Extract Structured API Data

### Phase 1: Parse API Documentation into Schema

For each endpoint page, extract:

```json
{
  "endpoint": "/v1/monitors",
  "method": "POST",
  "title": "Create Monitor",
  "parameters": [
    {
      "name": "test_mode",
      "type": "boolean",
      "required": false,
      "default": false,
      "description": "Enable test mode",
      "valid_values": null
    },
    {
      "name": "timeout",
      "type": "integer",
      "required": false,
      "default": 10,
      "valid_values": [5, 10, 15, 20, 30]
    }
  ],
  "response": {
    "status": 200,
    "schema": { ... }
  }
}
```

### Phase 2: Compare Schemas (Not Raw Text)

```go
type APIDiff struct {
    Endpoint string
    Method   string

    AddedParameters    []Parameter
    RemovedParameters  []Parameter
    ModifiedParameters []ParameterChange

    Breaking bool
    Severity string // "major", "minor", "patch"
}

type ParameterChange struct {
    Name        string
    OldRequired bool
    NewRequired bool
    OldType     string
    NewType     string
    OldDefault  interface{}
    NewDefault  interface{}
}
```

### Phase 3: Generate Actionable Reports

```markdown
# API Changes Detected: 2026-02-03

## Breaking Changes (Action Required)

### POST /v1/monitors
- ⚠️ **BREAKING**: Parameter 'url' validation stricter (now requires https://)
- **Impact**: Existing http:// URLs will be rejected
- **Migration**: Update all monitor URLs to use https://

## New Features

### POST /v1/monitors
- ✨ **NEW**: test_mode parameter (boolean, optional)
  - Enables dry-run testing without actually creating monitor
  - Default: false

## Deprecations

### POST /v1/monitors
- ⚠️ **DEPRECATED**: legacy_mode parameter (will be removed in v2)
  - Use new 'mode' parameter instead
```

---

## Implementation Strategy

### Immediate (Add to Plan A)

**Option 1: Simple Parameter Extraction**
- Extract parameter names from HTML
- Compare parameter lists (added/removed)
- Flag any changes for manual review
- **Time**: +1 day

**Option 2: Full Schema Parsing**
- Parse complete parameter definitions
- Extract types, defaults, validations
- Semantic diff with breaking change detection
- **Time**: +3-4 days

**Option 3: Hybrid (Recommended)**
- Extract key fields (name, type, required)
- Simple diff (added/removed/type changed)
- Manual review for complex changes
- **Time**: +2 days

### Future Enhancements

1. **LLM-Assisted Analysis**
   - Use Claude to analyze diffs
   - Classify breaking vs non-breaking
   - Generate migration guides

2. **OpenAPI Generation**
   - Convert scraped docs to OpenAPI spec
   - Use standard OpenAPI diff tools
   - Generate client code automatically

3. **Regression Testing**
   - Detect when documented behavior changes
   - Alert if response format differs
   - Validate against provider implementation

---

## Recommendation for Plan A

### Original Plan A
1. Dynamic URL discovery
2. Content-hash caching (detects "changed")
3. Resource blocking
4. Diff reports (show hash changed)

### Enhanced Plan A (Recommended)
1. Dynamic URL discovery
2. **Structured data extraction** (parse API schemas)
3. **Semantic comparison** (what changed, not just that it changed)
4. Resource blocking
5. **Actionable diff reports** (added/removed params, breaking changes)

### Trade-off

| Approach | Pros | Cons | Time |
|----------|------|------|------|
| **Original (hash-only)** | Fast to implement | Useless diffs, false positives | 2 days |
| **Enhanced (semantic)** | Actionable intelligence | More complex parsing | 4-5 days |
| **Hybrid** | Good balance | Some manual review needed | 3 days |

---

## Decision Point

**Question for user**: How detailed do you need the diff reports?

**Option A**: Simple detection (hash changed) + manual review
- Fastest (2 days)
- You review changes manually when alerted
- Good for infrequent API changes

**Option B**: Structured diff (parameter-level changes)
- Medium (3 days)
- Shows added/removed parameters
- Still need manual review for breaking changes

**Option C**: Full semantic analysis (breaking change detection)
- Comprehensive (5 days)
- Automated breaking change alerts
- Generated migration guides
- Best for production

---

## Current Status

✅ **Validated**: Hash-based detection works (98% time savings)
❌ **Gap Identified**: Doesn't show WHAT changed
🎯 **Next Decision**: Which diff approach for Plan A?
