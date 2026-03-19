# Test Plan: 3 API Gaps

## Layer 1: Client Model Unit Tests (JSON Serialization/Deserialization)

File: `internal/client/models_test.go`

### Test 1.1: `TestCreateIncidentRequest_WithUpdates_JSON`
- **Verifies**: `CreateIncidentRequest` serializes the `updates` array correctly when populated
- **Key assertions**:
  - `json.Marshal` output contains `"updates"` key
  - Each element has `text`, `type`, and `date` fields
  - `text` has the localized `en` key
- **Layer**: 1

### Test 1.2: `TestUpdateIncidentRequest_WithUpdates_JSON`
- **Verifies**: `UpdateIncidentRequest` serializes the `updates` array correctly when populated
- **Key assertions**:
  - `json.Marshal` output contains `"updates"` key with correct element structure
  - Elements match `AddIncidentUpdateRequest` shape (`text`, `type`, `date`)
- **Layer**: 1

### Test 1.3: `TestCreateIncidentRequest_WithoutUpdates_OmitsField`
- **Verifies**: `updates` field is omitted from JSON when nil/empty (via `omitempty`)
- **Key assertions**:
  - `json.Marshal` output does NOT contain `"updates"` substring when `Updates` is nil
  - `json.Marshal` output does NOT contain `"updates"` substring when `Updates` is `[]AddIncidentUpdateRequest{}`
- **Layer**: 1

### Test 1.4: `TestStatusPageAuthenticationSettings_SSOConnectionUUID_JSON`
- **Verifies**: `StatusPageAuthenticationSettings` correctly deserializes `sso_connection_uuid` from API responses
- **Key assertions**:
  - Unmarshal with `"sso_connection_uuid": null` produces `nil` pointer
  - Unmarshal with `"sso_connection_uuid": "uuid-abc"` produces `*string` pointing to `"uuid-abc"`
  - Unmarshal with field absent produces `nil` pointer
- **Layer**: 1

### Test 1.5: `TestCreateStatusPageAuthenticationSettings_SSOConnectionUUID_JSON`
- **Verifies**: `CreateStatusPageAuthenticationSettings` serializes `sso_connection_uuid` correctly for write
- **Key assertions**:
  - When `SSOConnectionUUID` is nil, field is omitted from JSON (`omitempty`)
  - When `SSOConnectionUUID` is `&"uuid-abc"`, JSON contains `"sso_connection_uuid":"uuid-abc"`
- **Layer**: 1

### Test 1.6: `TestStatusPageService_Description_JSON`
- **Verifies**: `StatusPageService` correctly deserializes the localized `description` map from API responses
- **Key assertions**:
  - Unmarshal with `"description": {"en": "My service"}` produces `map[string]string{"en": "My service"}`
  - Unmarshal with `"description": {"en": "English", "fr": "French"}` produces both keys
  - Unmarshal with field absent produces nil map
- **Layer**: 1

### Test 1.7: `TestCreateStatusPageService_Description_JSON`
- **Verifies**: `CreateStatusPageService` serializes `description` as a plain string for write
- **Key assertions**:
  - When `Description` is nil, field is omitted from JSON
  - When `Description` is `&"Service desc"`, JSON contains `"description":"Service desc"`
- **Layer**: 1

---

## Layer 2: AttrTypes Count Assertions

File: `internal/provider/statuspage_nested_service_test.go` (existing test update + new tests)

### Test 2.1: Update `TestNestedServiceAttrTypes` (existing, line 70)
- **Verifies**: `NestedServiceAttrTypes()` returns exactly 7 keys (was 6)
- **Key assertions**:
  - `expectedKeys` updated to include `"description"`
  - Length check: `len(attrs) == 7`
  - `"description"` key exists in the returned map
- **Layer**: 2

### Test 2.2: `TestServiceAttrTypes_Count` (new)
- **Verifies**: `ServiceAttrTypes()` returns exactly 8 keys (was 7)
- **Key assertions**:
  - Expected keys: `id`, `uuid`, `name`, `is_group`, `show_uptime`, `show_response_times`, `services`, `description`
  - Length check: `len(attrs) == 8`
  - `"description"` key exists with `types.MapType{ElemType: types.StringType}`
- **Layer**: 2

### Test 2.3: `TestAuthenticationSettingsAttrTypes_Count` (new)
- **Verifies**: `AuthenticationSettingsAttrTypes()` returns exactly 5 keys (was 4)
- **Key assertions**:
  - Expected keys: `password_protection`, `google_sso`, `saml_sso`, `allowed_domains`, `sso_connection_uuid`
  - Length check: `len(attrs) == 5`
  - `"sso_connection_uuid"` key exists with `types.StringType`
- **Layer**: 2

File for new tests: `internal/provider/statuspage_nested_service_test.go` (Test 2.2, 2.3 can go here alongside Test 2.1, or in `statuspage_mapping_coverage_test.go`)

---

## Layer 3: Provider Mapping Tests (TF-to-API and API-to-TF)

### Existing test updates (required to prevent panics)

#### Update 3.1: `buildAuthObj` helper (`statuspage_mapping_test.go:101`)
- **What changes**: Add default `"sso_connection_uuid": types.StringNull()` when caller omits it
- **Verifies**: All 4 existing callers (lines 120, 188, 217, 239) continue to work without modification
- **Key assertions**: No new assertions; prevents `objectValue missing key` panics
- **Layer**: 3

#### Update 3.2: `buildFullServiceObj` (`statuspage_mapping_sections_test.go:261`)
- **What changes**: Add `"description": types.MapNull(types.StringType)` to the `ServiceAttrTypes()` object
- **Verifies**: Existing section mapping tests continue to pass
- **Layer**: 3

#### Update 3.3: `buildMinimalServiceObj` (`statuspage_mapping_sections_test.go:275`)
- **What changes**: Add `"description": types.MapNull(types.StringType)` to the `ServiceAttrTypes()` object
- **Verifies**: Existing section mapping tests continue to pass
- **Layer**: 3

#### Update 3.4: `buildNestedServiceObj` (`statuspage_nested_service_test.go:566`)
- **What changes**: Add `"description": types.MapNull(types.StringType)` to the `NestedServiceAttrTypes()` object
- **Verifies**: All nested service tests continue to pass
- **Layer**: 3

#### Update 3.5: `buildTopLevelServiceObj` (`statuspage_nested_service_test.go:605`)
- **What changes**: Add `"description": types.MapNull(types.StringType)` to the `ServiceAttrTypes()` object
- **Verifies**: All top-level service tests continue to pass
- **Layer**: 3

#### Update 3.6: Inline auth object in `TestMapTFToSettings_WithValues` (`statuspage_mapping_coverage_test.go:589`)
- **What changes**: Add `"sso_connection_uuid": types.StringNull()` to inline auth object
- **Verifies**: Settings mapping test continues to pass
- **Layer**: 3

#### Update 3.7: Inline service object in `TestMapTFToSections_WithValues` (`statuspage_mapping_coverage_test.go:654`)
- **What changes**: Add `"description": types.MapNull(types.StringType)` to inline service object
- **Verifies**: Section mapping test continues to pass
- **Layer**: 3

#### Update 3.8: Inline objects in `TestMapTFToServices_WithValues` (`statuspage_mapping_coverage_test.go:702, 714, 726`)
- **What changes**: Add `"description": types.MapNull(types.StringType)` to all 3 inline service/nested-service objects
- **Verifies**: Services mapping test continues to pass
- **Layer**: 3

#### Update 3.9: Inline service in `TestMapTFToServices_NonGroupWithoutUUID` (`statuspage_mapping_coverage_test.go:1158`)
- **What changes**: Add `"description": types.MapNull(types.StringType)` to inline service object
- **Verifies**: Validation error test continues to pass
- **Layer**: 3

#### Update 3.10: Inline service in `TestMapTFToServices_GroupWithEmptyServices` (`statuspage_mapping_coverage_test.go:1184`)
- **What changes**: Add `"description": types.MapNull(types.StringType)` to inline service object
- **Verifies**: Validation error test continues to pass
- **Layer**: 3

### New mapping tests

File: `internal/provider/statuspage_mapping_test.go` or `statuspage_mapping_coverage_test.go`

#### Test 3.11: `TestExtractAuthSettings_SSOConnectionUUID`
- **Verifies**: `extractAuthSettings` correctly extracts `sso_connection_uuid` from TF object
- **Key assertions**:
  - With `sso_connection_uuid: types.StringValue("uuid-123")`, result has `SSOConnectionUUID = &"uuid-123"`
  - With `sso_connection_uuid: types.StringNull()`, result has `SSOConnectionUUID = nil`
- **Layer**: 3

#### Test 3.12: `TestMapSettingsToTFWithFilter_SSOConnectionUUID`
- **Verifies**: `mapSettingsToTFWithFilter` maps API `SSOConnectionUUID` to TF state correctly
- **Key assertions**:
  - With `SSOConnectionUUID: &"uuid-abc"`, TF auth object has `sso_connection_uuid = "uuid-abc"`
  - With `SSOConnectionUUID: nil`, TF auth object has `sso_connection_uuid` as null
- **Layer**: 3

#### Test 3.13: `TestMapServiceToTFWithFilter_Description`
- **Verifies**: `mapServiceToTFWithFilter` maps API service `Description` map to TF state
- **Key assertions**:
  - With `Description: map[string]string{"en": "My API"}`, TF object has `description` map with `en -> "My API"`
  - With `Description: nil`, TF object has null `description` map
- **Layer**: 3

#### Test 3.14: `TestMapNestedServicesToTF_Description`
- **Verifies**: `mapNestedServicesToTF` maps API nested service `Description` map to TF state
- **Key assertions**:
  - With `Description: map[string]string{"en": "Nested svc"}`, TF object has `description` map with value
  - With `Description: nil`, TF object has null `description` map
- **Layer**: 3

#### Test 3.15: `TestMapTFToService_Description`
- **Verifies**: `mapTFToService` extracts `description` from TF and writes as plain string
- **Key assertions**:
  - With `description: {"en": "English desc"}`, result has `Description = &"English desc"`
  - With `description: {"fr": "French desc"}` (no `en`), result has `Description = &"French desc"` (fallback)
  - With null description map, result has `Description = nil`
- **Layer**: 3

#### Test 3.16: `TestMapTFToNestedServices_Description`
- **Verifies**: `mapTFToNestedServices` extracts `description` from TF nested services
- **Key assertions**:
  - With `description: {"en": "Nested desc"}`, nested service result has `Description = &"Nested desc"`
  - With null description map, result has `Description = nil`
- **Layer**: 3

---

## Layer 4: Resource Lifecycle Tests (Mock-server CRUD with new fields)

These are optional integration-level tests that validate the full resource lifecycle with the new fields. They use mock HTTP servers to simulate API behavior.

File: `internal/provider/statuspage_resource_test.go` or new file

#### Test 4.1: `TestStatusPageResource_SSOConnectionUUID_Lifecycle` (optional)
- **Verifies**: Full create-read-update-delete cycle with `sso_connection_uuid` in the authentication block
- **Key assertions**:
  - Create request body includes `sso_connection_uuid` when configured
  - Read response with `sso_connection_uuid` is correctly mapped to state
  - Update clears `sso_connection_uuid` when removed from config
- **Layer**: 4

#### Test 4.2: `TestStatusPageResource_ServiceDescription_Lifecycle` (optional)
- **Verifies**: Full create-read cycle with service `description`
- **Key assertions**:
  - Create request body includes `description` as plain string
  - Read response with localized `description` map is correctly mapped to state
- **Layer**: 4

---

## Summary

| Layer | New Tests | Updated Tests | Files |
|-------|-----------|---------------|-------|
| 1 - Client models | 7 | 0 | `internal/client/models_test.go` |
| 2 - AttrTypes counts | 2 new + 1 update | 1 | `internal/provider/statuspage_nested_service_test.go` |
| 3 - Mapping | 6 | 10 | `statuspage_mapping_test.go`, `statuspage_mapping_coverage_test.go`, `statuspage_mapping_sections_test.go`, `statuspage_nested_service_test.go` |
| 4 - Lifecycle (optional) | 2 | 0 | `statuspage_resource_test.go` |
| **Total** | **17 new** | **11 updated** | **5 files** |

## Verification Commands

```bash
# Layer 1: Client model tests
go test ./internal/client/... -v -run "TestCreate.*Updates|TestStatusPage.*SSO|TestStatusPage.*Description|TestCreateStatusPage.*Description"

# Layer 2: AttrTypes count tests
go test ./internal/provider/... -v -run "TestNestedServiceAttrTypes|TestServiceAttrTypes_Count|TestAuthenticationSettingsAttrTypes_Count"

# Layer 3: Mapping tests
go test ./internal/provider/... -v -run "TestExtractAuth|TestMapSettings|TestMapService|TestMapNestedServices|TestMapTFToService|TestMapTFToNestedServices|TestMapTFToSections|TestMapTFToServices"

# Full suite (must pass)
make test

# Lint (must pass)
make lint
```
