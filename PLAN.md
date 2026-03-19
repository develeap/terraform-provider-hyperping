# Implementation Plan: 3 API Gaps

## Overview

Close three gaps between the Hyperping API surface and the provider's client/schema layers:

1. **Incident create/update missing `updates` array** (client-only)
2. **StatusPage missing `sso_connection_uuid`** (client + TF schema)
3. **StatusPage service missing `description`** (client + TF schema)

---

## Gap 1: Incident `updates` array in create/update requests

### Scope
Client layer only. No Terraform schema changes (the `incident_update` resource handles TF-level management).

### Write Shape
The API accepts updates as an array of objects without `uuid` (the API assigns UUIDs on creation):
```json
{"text": {"en": "Investigating..."}, "type": "investigating", "date": "2026-03-20T10:00:00Z"}
```

### Files to Modify

#### `internal/client/models_incident.go`

1. Add a new struct `CreateIncidentUpdate` (write shape, no UUID):
   ```go
   type CreateIncidentUpdate struct {
       Text LocalizedText `json:"text"`
       Type string        `json:"type"`
       Date string        `json:"date"`
   }
   ```
   Note: This is identical to `AddIncidentUpdateRequest` (line 64) but used inside the create/update request body as an array element. We reuse the existing type or create an alias -- since `AddIncidentUpdateRequest` already has the exact same shape, we can use it directly as the element type.

   **Decision**: Reuse `AddIncidentUpdateRequest` as the element type. Rename is not needed; the struct name is descriptive enough for both contexts.

2. Add `Updates` field to `CreateIncidentRequest` (line 33):
   ```go
   Updates []AddIncidentUpdateRequest `json:"updates,omitempty"`
   ```

3. Add `Updates` field to `UpdateIncidentRequest` (line 55):
   ```go
   Updates []AddIncidentUpdateRequest `json:"updates,omitempty"`
   ```

### Files to Test

#### `internal/client/models_test.go`

Add serialization tests:
- `TestCreateIncidentRequest_WithUpdates_JSON` -- marshal a `CreateIncidentRequest` with `Updates` populated and verify the JSON shape.
- `TestUpdateIncidentRequest_WithUpdates_JSON` -- same for `UpdateIncidentRequest`.
- `TestCreateIncidentRequest_WithoutUpdates_OmitsField` -- verify `updates` is omitted when nil/empty.

---

## Gap 2: StatusPage `sso_connection_uuid`

### Scope
Client models + Terraform schema + mapping layer.

### API Behavior
- Read: `"authentication": {"sso_connection_uuid": null}` (or a string UUID when configured)
- Write: `"authentication": {"sso_connection_uuid": "uuid-string"}` (optional)

### Files to Modify

#### 1. `internal/client/models_statuspage.go`

- **`StatusPageAuthenticationSettings`** (line 55): Add field:
  ```go
  SSOConnectionUUID *string `json:"sso_connection_uuid"`
  ```

- **`CreateStatusPageAuthenticationSettings`** (line 137): Add field:
  ```go
  SSOConnectionUUID *string `json:"sso_connection_uuid,omitempty"`
  ```

#### 2. `internal/provider/statuspage_mapping_types.go`

- **`AuthenticationSettingsAttrTypes()`** (line 114): Add entry:
  ```go
  "sso_connection_uuid": types.StringType,
  ```

#### 3. `internal/provider/statuspage_resource_schema.go`

- **authentication nested attributes** (line 205): Add attribute:
  ```go
  "sso_connection_uuid": schema.StringAttribute{
      MarkdownDescription: "SSO connection UUID for SAML SSO integration",
      Optional:            true,
      Computed:            true,
  },
  ```

#### 4. `internal/provider/statuspage_data_source.go`

- **authentication nested attributes** (line 166): Add computed attribute:
  ```go
  "sso_connection_uuid": schema.StringAttribute{
      MarkdownDescription: "SSO connection UUID",
      Computed:            true,
  },
  ```

#### 5. `internal/provider/statuspages_data_source.go`

- **authentication nested attributes** (line 179): Add computed attribute:
  ```go
  "sso_connection_uuid": schema.StringAttribute{
      MarkdownDescription: "SSO connection UUID",
      Computed:            true,
  },
  ```

#### 6. `internal/provider/statuspage_mapping.go`

- **`mapSettingsToTFWithFilter`** (line 122): Add `sso_connection_uuid` to the auth object value map:
  ```go
  // Handle optional sso_connection_uuid
  var ssoConnectionUUIDValue types.String
  if settings.Authentication.SSOConnectionUUID != nil {
      ssoConnectionUUIDValue = types.StringValue(*settings.Authentication.SSOConnectionUUID)
  } else {
      ssoConnectionUUIDValue = types.StringNull()
  }
  ```
  Then add to the `types.ObjectValue(AuthenticationSettingsAttrTypes(), ...)` call:
  ```go
  "sso_connection_uuid": ssoConnectionUUIDValue,
  ```

- **`extractAuthSettings`** (line 243): Add extraction logic:
  ```go
  if ssoUUID, ok := attrs["sso_connection_uuid"].(types.String); ok && !ssoUUID.IsNull() && !ssoUUID.IsUnknown() {
      val := ssoUUID.ValueString()
      authentication.SSOConnectionUUID = &val
  }
  ```

### Files to Test

#### `internal/client/models_test.go` (or new file)
- `TestStatusPageAuthenticationSettings_SSOConnectionUUID_JSON` -- unmarshal with `sso_connection_uuid: null` and with a string value.

#### `internal/provider/statuspage_mapping_coverage_test.go`
- Update `TestMapTFToSettings_WithValues` -- add `sso_connection_uuid` to auth object values.
- Update `TestMapSettingsToTFWithFilter_EdgeCases` -- verify null sso_connection_uuid maps correctly.
- Update `TestMapSettingsToTFWithFilter_WithOptionalValues` -- verify non-null sso_connection_uuid.
- Add `TestAuthenticationSettingsAttrTypes_Count` -- verify 5 keys (was 4).

---

## Gap 3: StatusPage service `description`

### Scope
Client models + Terraform schema + mapping layer.

### API Behavior
- Read: `"description": {"en": "Service description"}` (localized map, same as section name)
- Write: `"description": "Service description"` (plain string, follows write-plain/read-localized pattern)

### Files to Modify

#### 1. `internal/client/models_statuspage.go`

- **`StatusPageService`** (line 79): Add field:
  ```go
  Description map[string]string `json:"description,omitempty"`
  ```

- **`CreateStatusPageService`** (line 156): Add field:
  ```go
  Description *string `json:"description,omitempty"`
  ```

#### 2. `internal/provider/statuspage_mapping_types.go`

- **`NestedServiceAttrTypes()`** (line 134): Add entry:
  ```go
  "description": types.MapType{ElemType: types.StringType},
  ```

- **`ServiceAttrTypes()`** (line 147): Add entry:
  ```go
  "description": types.MapType{ElemType: types.StringType},
  ```

#### 3. `internal/provider/statuspage_resource_schema.go`

- **Top-level service nested attributes** (around line 256, inside `"services"` NestedObject Attributes): Add:
  ```go
  "description": schema.MapAttribute{
      MarkdownDescription: "Localized service description (language code -> text). On write, only the default language value is sent as a plain string.",
      ElementType:         types.StringType,
      Optional:            true,
      Computed:            true,
  },
  ```

- **Nested service attributes** (around line 292, inside the nested `"services"` NestedObject Attributes): Add the same attribute:
  ```go
  "description": schema.MapAttribute{
      MarkdownDescription: "Localized service description (language code -> text)",
      ElementType:         types.StringType,
      Optional:            true,
      Computed:            true,
  },
  ```

#### 4. `internal/provider/statuspage_data_source.go`

- **Top-level service attributes** and **nested service attributes**: Add computed `description` map attribute to both levels.

#### 5. `internal/provider/statuspages_data_source.go`

- **Top-level service attributes** and **nested service attributes**: Add computed `description` map attribute to both levels.

#### 6. `internal/provider/statuspage_mapping.go`

- **`mapServiceToTFWithFilter`** (line 350): Add description mapping:
  ```go
  filteredDesc := filterLocalizedMap(service.Description, configuredLangs)
  descMap := mapStringMapToTF(filteredDesc, diags)
  ```
  Add `"description": descMap` to the `types.ObjectValue(ServiceAttrTypes(), ...)` call.

- **`mapNestedServicesToTF`** (line 377): Add description mapping for nested services:
  ```go
  filteredDesc := filterLocalizedMap(svc.Description, configuredLangs)
  descMap := mapStringMapToTF(filteredDesc, diags)
  ```
  Add `"description": descMap` to the `types.ObjectValue(NestedServiceAttrTypes(), ...)` call.

- **`mapTFToService`** (line 497): Add description extraction (write as plain string):
  ```go
  if descMap, ok := attrs["description"].(types.Map); ok && !descMap.IsNull() {
      descStrMap := mapTFToStringMap(descMap, diags)
      if enDesc, ok := descStrMap["en"]; ok && enDesc != "" {
          service.Description = &enDesc
      } else {
          for _, v := range descStrMap {
              if v != "" {
                  service.Description = &v
                  break
              }
          }
      }
  }
  ```

- **`mapTFToNestedServices`** (line 560): Add same description extraction for nested services:
  ```go
  if descMap, ok := attrs["description"].(types.Map); ok && !descMap.IsNull() {
      descStrMap := mapTFToStringMap(descMap, diags)
      if enDesc, ok := descStrMap["en"]; ok && enDesc != "" {
          svc.Description = &enDesc
      } else {
          for _, v := range descStrMap {
              if v != "" {
                  svc.Description = &v
                  break
              }
          }
      }
  }
  ```

### Files to Test

#### `internal/client/models_test.go` (or new file)
- `TestStatusPageService_Description_JSON` -- unmarshal with `description` map and verify.
- `TestCreateStatusPageService_Description_JSON` -- marshal with `description` string.

#### `internal/provider/statuspage_mapping_coverage_test.go`
- Update `TestMapTFToSections_WithValues` (line 653) -- add `description` to service objects.
- Update `TestMapTFToServices_WithValues` (line 701) -- add `description` to service and nested service objects.
- Update `TestMapTFToServices_NonGroupWithoutUUID` (line 1156) -- add `description` to attr map.
- Update `TestMapTFToServices_GroupWithEmptyServices` (line 1182) -- add `description` to attr map.
- Add `TestServiceAttrTypes_Count` -- verify 8 keys (was 7).

#### `internal/provider/statuspage_mapping_sections_test.go`
- Update `buildFullServiceObj()` (line 261) -- add `"description"` key.
- Update `buildMinimalServiceObj()` (line 275) -- add `"description"` key.

#### `internal/provider/statuspage_nested_service_test.go`
- Update `TestNestedServiceAttrTypes` (line 70) -- update expected keys from 6 to 7, add `"description"` to `expectedKeys`.
- Update `buildNestedServiceObj()` (line 566) -- add `"description"` key.
- Update `buildTopLevelServiceObj()` (line 605) -- add `"description"` key.

#### `internal/provider/statuspage_mapping_test.go` or `statuspage_mapping_settings_test.go`
- Add test for `mapServiceToTFWithFilter` with description populated.
- Add test for `mapNestedServicesToTF` with description populated.
- Update `buildAuthObj()` helper (statuspage_mapping_test.go:101) to default `sso_connection_uuid` to `types.StringNull()` when not provided.

---

## Implementation Phases

### Phase 1: Client Models (no TF dependency)
**Files**: `internal/client/models_incident.go`, `internal/client/models_statuspage.go`
**Tests**: `internal/client/models_test.go`

1. [ ] Add `Updates []AddIncidentUpdateRequest` to `CreateIncidentRequest` after `Date` field (models_incident.go:33-40)
2. [ ] Add `Updates []AddIncidentUpdateRequest` to `UpdateIncidentRequest` after `StatusPages` field (models_incident.go:55-60)
3. [ ] Add `SSOConnectionUUID *string` to `StatusPageAuthenticationSettings` after `AllowedDomains` (models_statuspage.go:55-60)
4. [ ] Add `SSOConnectionUUID *string` to `CreateStatusPageAuthenticationSettings` after `AllowedDomains` (models_statuspage.go:137-142)
5. [ ] Add `Description map[string]string` to `StatusPageService` after `Services` field (models_statuspage.go:79-87)
6. [ ] Add `Description *string` to `CreateStatusPageService` after `Services` field (models_statuspage.go:156-165)
7. [ ] Write client model serialization tests

### Phase 2: AttrTypes (mapping types layer)
**Files**: `internal/provider/statuspage_mapping_types.go`

8. [ ] Add `"sso_connection_uuid": types.StringType` to `AuthenticationSettingsAttrTypes()` (mapping_types.go:114-121)
9. [ ] Add `"description": types.MapType{ElemType: types.StringType}` to `NestedServiceAttrTypes()` (mapping_types.go:134-143)
10. [ ] Add `"description": types.MapType{ElemType: types.StringType}` to `ServiceAttrTypes()` (mapping_types.go:147-157)

### Phase 3: Terraform Schemas
**Files**: `statuspage_resource_schema.go`, `statuspage_data_source.go`, `statuspages_data_source.go`

11. [ ] Add `sso_connection_uuid` attribute to resource schema authentication block
12. [ ] Add `description` map attribute to resource schema service blocks (both top-level and nested)
13. [ ] Add `sso_connection_uuid` attribute to data source schema authentication blocks
14. [ ] Add `description` map attribute to data source schema service blocks (both top-level and nested)

### Phase 4: Mapping Functions
**Files**: `internal/provider/statuspage_mapping.go`

15. [ ] Update `mapSettingsToTFWithFilter` -- map `SSOConnectionUUID` to auth object
16. [ ] Update `extractAuthSettings` -- extract `sso_connection_uuid` from TF object
17. [ ] Update `mapServiceToTFWithFilter` -- map `Description` to service object
18. [ ] Update `mapNestedServicesToTF` -- map `Description` to nested service object
19. [ ] Update `mapTFToService` -- extract `description` from TF service object (write as plain string)
20. [ ] Update `mapTFToNestedServices` -- extract `description` from TF nested service object

### Phase 5: Tests
**Files**: Various `*_test.go`

21. [ ] Client model serialization tests (incident updates, sso_connection_uuid, service description)
22. [ ] Update `TestNestedServiceAttrTypes` expected keys (6->7); add `TestServiceAttrTypes_Count` (7->8 keys)
23. [ ] Update `buildAuthObj` helper to auto-inject `sso_connection_uuid: types.StringNull()` default
24. [ ] Update `buildNestedServiceObj`, `buildTopLevelServiceObj`, `buildFullServiceObj`, `buildMinimalServiceObj` helpers to include `"description"` key
25. [ ] Update all inline service/auth object constructors in coverage tests to include new keys
26. [ ] Add mapping tests for sso_connection_uuid read/write round-trip
27. [ ] Add mapping tests for service description read/write round-trip

---

## Testing Strategy

### Unit Tests (Phase 5)
1. **Client model JSON serialization** -- verify `json.Marshal`/`Unmarshal` for all new fields
2. **AttrTypes count assertions** -- catch accidental omissions when types and schemas diverge
3. **Mapping round-trip tests** -- TF -> API -> TF for sso_connection_uuid and description

### Existing Test Updates
All tests that construct `ServiceAttrTypes()` or `NestedServiceAttrTypes()` objects will fail until `description` is added. These must be updated in the same commit as the AttrTypes change.

**Service/NestedService objects -- files that need `"description"` added:**
- `statuspage_mapping_coverage_test.go` -- multiple inline object constructors (lines 654, 702, 714, 726, 1158, 1184)
- `statuspage_mapping_sections_test.go` -- `buildFullServiceObj()` (line 261) and `buildMinimalServiceObj()` (line 275), plus inline constructions (line 160)
- `statuspage_nested_service_test.go` -- `buildNestedServiceObj()` (line 566) and `buildTopLevelServiceObj()` (line 605) helpers; all callers get fixed automatically

Note: `statuspage_nested_groups_test.go` does NOT construct typed objects directly -- it uses API client structs, so it needs no changes.

**Authentication objects -- files that need `"sso_connection_uuid"` added:**
- `statuspage_mapping_coverage_test.go` -- `TestMapTFToSettings_WithValues` (line 589)
- `statuspage_mapping_settings_test.go` -- `buildAuthObj()` helper (line 101); must default `sso_connection_uuid` to `types.StringNull()` when not in caller's attrs map, OR every caller (lines 120, 188, 217, 239) must add `"sso_connection_uuid"`. Best approach: update `buildAuthObj` to inject a default `types.StringNull()` for `sso_connection_uuid` when not provided, so all existing callers keep working.

### Verification Commands
```bash
# Unit tests
go test ./internal/client/... -v -run "TestCreate.*Updates|TestStatusPage.*SSO|TestStatusPage.*Description"
go test ./internal/provider/... -v -run "TestMap.*Settings|TestMap.*Service|TestMap.*Section|TestAttrTypes"

# Full test suite (must pass)
make test

# Lint
make lint
```

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| AttrTypes / schema mismatch causes runtime panics | AttrTypes count tests catch this at unit-test time |
| Existing tests break from missing `description` in object constructors | Phase 2 + Phase 5 (steps 22-25) must be applied atomically -- `types.ObjectValue` panics if any AttrTypes key is missing from the value map |
| `buildAuthObj` helper breaks when `sso_connection_uuid` added to AttrTypes | Update `buildAuthObj` to inject `types.StringNull()` default for `sso_connection_uuid` when caller omits it -- all 4 existing callers keep working |
| `sso_connection_uuid: null` causes type assertion panic | Use `*string` with explicit nil check, same pattern as existing optional fields |
| Service description write-plain/read-localized mismatch | Follow existing pattern used by statuspage settings `description` field |
| Incident `updates` array breaks existing create/update flows | `omitempty` ensures nil/empty arrays are not sent, preserving backward compatibility |

---

## Success Criteria

1. `make test` passes with 0 failures
2. `make lint` passes with 0 issues
3. All 3 API gaps are closed:
   - `CreateIncidentRequest` and `UpdateIncidentRequest` support `updates` array
   - `StatusPageAuthenticationSettings` supports `sso_connection_uuid` in client, schema, and mapping
   - `StatusPageService` supports `description` in client, schema, and mapping (both read and write)
4. No regressions in existing acceptance tests (verified by CI)
