# ADR 0003: Single Source of Truth for Constants

## Status
Accepted

## Context
The provider needs to validate regions, incident statuses, and other enumerated values. These values need to be consistent across schema validation, documentation, and tests.

## Decision
All enumerated constants are defined once in `internal/client/models.go`:
- `AllowedRegions` - Valid monitor check regions
- `AllowedIncidentStatuses` - Valid incident status values
- `AllowedIncidentSeverities` - Valid incident severity values
- `AllowedMaintenanceStatuses` - Valid maintenance status values

Schema validators and documentation reference these constants rather than duplicating the values.

## Consequences

### Positive
- Single source of truth prevents inconsistencies
- Easy to update when API changes
- Tests can verify schema matches allowed values
- Documentation stays in sync automatically

### Negative
- Constants are in client package, creating a dependency from provider to client
- Changes require updating one file only (but that's a feature)

## Implementation
```go
// In schema definition
Validators: []validator.List{
    listvalidator.ValueStringsAre(stringvalidator.OneOf(client.AllowedRegions...)),
}
```
