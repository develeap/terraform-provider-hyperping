# API Schema Mismatch Analysis - Incident & Maintenance Resources

**Date:** 2026-02-09
**Issue:** ISS-005, ISS-006, ISS-007 fixes didn't resolve "inconsistent result" errors
**Root Cause:** API response format doesn't match provider schema expectations

---

## The Problem

The read-after-create pattern IS working correctly in v1.0.4. However, the API returns data in a format that causes state inconsistencies.

### Issue 1: `text` Field is Write-Only

**What Happens:**
```
CREATE Request:  {"text": {"en": "Description"}}  ✅ Accepted
GET Response:    { /* text field missing */ }    ❌ Not returned
```

**Impact:**
```go
// In mapIncidentToModel():
model.Text = types.StringValue(incident.Text.En)  // incident.Text.En is ""

// Terraform comlondonon:
Plan:  text = "Testing..."
State: text = ""
Result: INCONSISTENT RESULT AFTER APPLY ❌
```

### Issue 2: `title` Field Works But Could Be Improved

The API returns `{"title": {"en": "value"}}` and we extract `.En`, which works. However, users might want multi-language support in the future.

---

## The Fix: Write-Only Fields Pattern

For fields that the API accepts but doesn't return, we need to use **plan value preservation**.

### Solution 1: Preserve Plan Value (Recommended)

```go
func (r *IncidentResource) mapIncidentToModel(incident *client.Incident, plan *IncidentResourceModel, diags *diag.Diagnostics) {
    // ... other fields ...

    // Text field: API doesn't return it, so preserve plan value
    // This tells Terraform "I created what you asked for, trust me"
    if !plan.Text.IsNull() && !plan.Text.IsUnknown() {
        // Keep the plan value as-is (don't overwrite from API)
        // model.Text already has the plan value, so do nothing
    } else {
        // Only set to null if it was null in the plan
        model.Text = types.StringNull()
    }
}
```

### Solution 2: Mark as Write-Only in Docs

Update the schema description:
```go
"text": schema.StringAttribute{
    MarkdownDescription: "The description text (English). **Note:** This field is write-only and will not appear in state after creation.",
    Required:            true,
},
```

But this is bad UX because Terraform will always show drift.

---

## Implementation Plan

### Step 1: Fix Incident Resource

**File:** `internal/provider/incident_resource.go`

```go
func (r *IncidentResource) mapIncidentToModel(incident *client.Incident, plan *IncidentResourceModel, diags *diag.Diagnostics) {
    model.ID = types.StringValue(incident.UUID)
    model.Title = types.StringValue(incident.Title.En)

    // Handle date
    if incident.Date != "" {
        model.Date = types.StringValue(incident.Date)
    } else {
        model.Date = types.StringNull()
    }

    // ⚠️ CRITICAL: Text field is write-only in API
    // API accepts it during CREATE but never returns it in GET
    // We must preserve the plan value to avoid state drift
    // DO NOT set model.Text here - it should retain the plan value

    model.Type = types.StringValue(incident.Type)
    model.AffectedComponents = mapStringSliceToList(incident.AffectedComponents, diags)
    model.StatusPages = mapStringSliceToList(incident.StatusPages, diags)
}
```

### Step 2: Fix Maintenance Resource

**File:** `internal/provider/maintenance_resource.go`

Same pattern - the `text`/`message` field is likely write-only.

### Step 3: Update Documentation

Add to resource documentation:
```markdown
## Known API Limitations

- The `text` field is write-only. After creation, Terraform will preserve the value you specified, but the API does not return this field in subsequent reads.
```

---

## Testing Verification

### Before Fix
```
$ terraform apply
...
Error: Inconsistent result after apply

  .text: "Testing..." → ""
```

### After Fix
```
$ terraform apply
Apply complete! Resources: 1 added, 0 changed, 0 destroyed.

$ terraform plan
No changes. Your infrastructure matches the configuration.
```

---

## Alternative Approaches Considered

### ❌ Option A: Remove `text` field entirely
- **Pro:** No state drift
- **Con:** Users can't set incident descriptions (major feature loss)

### ❌ Option B: Make `text` optional + computed
- **Pro:** No drift error
- **Con:** Users think they can read it back, but it's always null (confusing)

### ✅ Option C: Preserve plan value (CHOSEN)
- **Pro:** Field works as expected, no drift
- **Con:** State contains plan value, not API value (acceptable for write-only fields)

---

## Related Issues

This same pattern likely affects:
- ❓ `maintenance_resource.go` - Check if `message` field is returned
- ❓ `statuspage_resource.go` - Check localized fields behavior
- ❓ `incident_update_resource.go` - Check if updates are returned

---

## Long-Term Solution

**Upstream API Fix (Hyperping):**
- Have GET endpoints return ALL fields that POST/PUT accept
- This makes the API consistent and HATEOAS-compliant

**Provider Enhancement:**
- Add multi-language support using nested attributes:
  ```hcl
  title = {
    en = "Incident Title"
    fr = "Titre de l'incident"
  }
  ```
