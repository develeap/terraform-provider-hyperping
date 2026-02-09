# ISS-007: Status Page Resource State Drift Issues

**Date:** 2026-02-09
**Status:** CRITICAL - Multiple API inconsistencies causing "inconsistent result after apply"

---

## Summary

The Status Page resource has FOUR distinct issues causing state drift:

| Issue # | Problem | Root Cause | v1.0.4 Fix Status | Additional Fix Needed |
|---------|---------|------------|-------------------|----------------------|
| 1 | Localized fields auto-populate | API adds all languages | ✅ **FIXED** (language filtering) | None |
| 2 | Subdomain gets domain appended | API appends `.hyperping.app` | ✅ **FIXED** (normalization) | None |
| 3 | `settings.name` overridden | API returns resource name | ❌ **NOT FIXED** | Plan preservation |
| 4 | `show_response_times` flips | API returns wrong value | ❌ **NOT FIXED** | Plan preservation or API fix |

---

## Issue 1: Localized Field Auto-Population ✅ FIXED

### Problem
```hcl
# User configures:
sections = [{
  name = { en = "API Services" }  # Only English
}]

# After apply, API returns:
sections = [{
  name = {
    en = "API Services"
    de = "API-Dienste"      # Auto-generated!
    fr = "Services API"     # Auto-generated!
    ru = "API-сервисы"     # Auto-generated!
  }
}]

# Terraform sees drift
```

### Fix in v1.0.4
```go
// statuspage_mapping.go
func filterLocalizedMap(m map[string]string, configuredLangs []string) map[string]string {
    if len(configuredLangs) == 0 || len(m) == 0 {
        return m
    }
    // Only return languages user actually configured
    filtered := make(map[string]string)
    for _, lang := range configuredLangs {
        if val, ok := m[lang]; ok {
            filtered[lang] = val
        }
    }
    return filtered
}
```

**Status:** ✅ Working correctly

---

## Issue 2: Subdomain Domain Suffix ✅ FIXED

### Problem
```hcl
# User configures:
hosted_subdomain = "my-status"

# API returns:
hosted_subdomain = "my-status.hyperping.app"

# Terraform sees drift
```

### Fix in v1.0.4
```go
// statuspage_mapping.go
func normalizeSubdomain(subdomain string) string {
    const HyperpingSubdomainSuffix = ".hyperping.app"
    if strings.HasSuffix(subdomain, HyperpingSubdomainSuffix) {
        return strings.TrimSuffix(subdomain, HyperpingSubdomainSuffix)
    }
    return subdomain
}
```

**Status:** ✅ Working correctly

---

## Issue 3: settings.name Override ❌ NOT FIXED

### Problem
```hcl
resource "hyperping_statuspage" "test" {
  name = "v1.0.4 Test Status Page"  # Resource-level name

  settings = {
    name = "Custom Display Name"     # Settings-level name
  }
}

# After apply:
# settings.name changes from "Custom Display Name" to "v1.0.4 Test Status Page"
# API returns resource.name in settings.name field
```

### Root Cause
The Hyperping API has a design quirk where `settings.name` always returns the same value as the resource-level `name` field, regardless of what was sent during creation.

### Current Mapping (WRONG)
```go
// statuspage_mapping.go - mapSettingsToTFWithFilter()
settingsObj, settingsDiags := types.ObjectValue(StatusPageSettingsAttrTypes(), map[string]attr.Value{
    "name": types.StringValue(settings.Name),  // ❌ Overwrites with API value
    // ...
})
```

### Proposed Fix: Plan Value Preservation
```go
func (r *StatusPageResource) mapStatusPageToModel(sp *client.StatusPage, model *StatusPageResourceModel, diags *diag.Diagnostics) {
    // Extract configured languages
    configuredLangs := r.extractConfiguredLanguages(model.Settings, diags)

    // Map with language filtering
    commonFields := MapStatusPageCommonFieldsWithFilter(sp, configuredLangs, diags)

    // Standard mappings
    model.ID = commonFields.ID
    model.Name = commonFields.Name
    model.Hostname = commonFields.Hostname
    model.HostedSubdomain = commonFields.HostedSubdomain
    model.URL = commonFields.URL
    model.Sections = commonFields.Sections

    // ⚠️ SPECIAL HANDLING: settings.name
    // API always returns resource.name in settings.name field
    // We need to preserve the plan value to avoid drift
    if !model.Settings.IsNull() && !model.Settings.IsUnknown() {
        // Extract settings.name from PLAN (model already has it)
        planSettingsAttrs := model.Settings.Attributes()
        if planName, ok := planSettingsAttrs["name"].(types.String); ok && !planName.IsNull() {
            // Preserve the plan value - don't overwrite with API value
            // Build new settings object with plan name but API values for everything else
            model.Settings = buildSettingsWithPreservedName(commonFields.Settings, planName, diags)
        } else {
            model.Settings = commonFields.Settings
        }
    } else {
        model.Settings = commonFields.Settings
    }
}
```

**Alternative:** Document as API limitation and tell users to use resource-level `name` only.

---

## Issue 4: show_response_times Boolean Flip ❌ NOT FIXED

### Problem
```hcl
sections = [{
  services = [{
    show_response_times = true
  }]
}]

# After apply:
# show_response_times changes from true to false
```

### Root Cause (Hypothesis)
One of these scenarios:
1. **API Bug:** API doesn't persist the value correctly
2. **Default Value:** API uses `false` as default when field is omitted
3. **Parsing Error:** Provider incorrectly parses the response

### Investigation Needed
```bash
# Test with actual API:
curl -X POST https://api.hyperping.io/v2/statuspages \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "name": "Test",
    "sections": [{
      "services": [{
        "monitor_uuid": "mon_xxx",
        "show_response_times": true
      }]
    }]
  }'

# Then GET and check response:
curl -X GET https://api.hyperping.io/v2/statuspages/{uuid} \
  -H "Authorization: Bearer $API_KEY" | jq '.sections[0].services[0].show_response_times'
```

### Current Mapping
```go
// statuspage_mapping.go - mapServiceToTF()
"show_response_times": types.BoolValue(service.ShowResponseTimes),
```

This looks correct. The issue is likely API-side.

### Potential Workarounds

**Option A: Plan Preservation (Same as text field)**
```go
// Only set if API returns true; preserve plan value if false
if service.ShowResponseTimes {
    serviceMap["show_response_times"] = types.BoolValue(true)
}
// If false and plan had true, keep plan value
```

**Option B: Default Value in Schema**
```go
"show_response_times": schema.BoolAttribute{
    Optional:            true,
    Computed:            true,
    Default:             booldefault.StaticBool(true),  // Default to true
}
```

**Option C: Contact Hyperping Support**
This is an API bug that needs upstream fix.

---

## Implementation Priority

### High Priority (Blocks v1.0.5)
1. ✅ ~~Localized field filtering~~ (Already in v1.0.4)
2. ✅ ~~Subdomain normalization~~ (Already in v1.0.4)
3. ❌ **Fix settings.name preservation** (Implement plan preservation)

### Medium Priority (Can wait for investigation)
4. ❌ **Investigate show_response_times** (Need API testing)

---

## Testing Checklist

### For settings.name Fix
```bash
# Create status page with custom settings.name
resource "hyperping_statuspage" "test" {
  name = "Resource Name"
  settings = {
    name = "Custom Settings Name"
  }
}

# After apply:
terraform state show hyperping_statuspage.test
# settings.name should be "Custom Settings Name", NOT "Resource Name"

# Plan should show no changes:
terraform plan
# No changes. Infrastructure is up-to-date.
```

### For show_response_times Fix
```bash
# Create with show_response_times = true
# After apply, verify it stays true
# Update to false, verify it changes
# Update back to true, verify it changes
```

---

## Long-Term Solutions

### Upstream API Improvements (Hyperping)
1. Stop overriding `settings.name` with resource-level `name`
2. Fix `show_response_times` persistence bug
3. Document which fields are computed vs user-settable

### Provider Enhancements
1. Add validation to warn users about API quirks
2. Comprehensive API response logging in DEBUG mode
3. Contract tests that verify API behavior hasn't changed

---

## Related Issues
- ISS-005: Incident text field (write-only) ✅ Fixed
- ISS-006: Maintenance text field (write-only) ✅ Fixed
- ISS-007: Status page multiple issues (partially fixed)
