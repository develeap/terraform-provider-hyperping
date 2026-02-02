# ADR 0001: Client Interface for Testability

## Status
Accepted

## Context
The Terraform provider resources need to interact with the Hyperping API through a client. For proper unit testing, we need the ability to mock API responses without making actual HTTP calls.

## Decision
We created interfaces (`MonitorAPI`, `IncidentAPI`, `MaintenanceAPI`, `HyperpingAPI`) that define the contract for API operations. Resources use these interfaces instead of the concrete `*Client` type.

## Consequences

### Positive
- Resources can be unit tested with mock implementations
- Enables testing error handling paths without real API calls
- Follows dependency inversion principle
- Makes the code more flexible for future changes

### Negative
- Slightly more complex type structure
- Interface must be kept in sync with Client methods

## Implementation
- `internal/client/interface.go` - Interface definitions
- Resources use interface types (e.g., `client client.MonitorAPI`)
- Mock implementations can be created in tests
