# ADR 0002: Coverage Threshold Strategy

## Status
Accepted

## Context
Go test coverage metrics don't capture code executed through the Terraform Plugin Testing framework because acceptance tests run in a subprocess. This creates a gap between actual test coverage and reported coverage.

## Decision
We set a 50% unit test coverage threshold with the understanding that:
- Client package: ~99% coverage (pure unit tests)
- Provider non-CRUD code: 100% coverage (factory, Configure, Metadata, Schema, mapping)
- Provider CRUD operations: 0% unit test coverage, but thoroughly tested via acceptance tests (TestAcc*)

## Consequences

### Positive
- Realistic threshold that reflects unit-testable code
- CRUD operations are still fully tested via acceptance tests
- CI doesn't fail due to unreachable coverage targets

### Negative
- Coverage percentage doesn't reflect true test coverage
- Requires documentation to explain the gap

## Alternatives Considered
1. **90% threshold with TF framework fixtures**: Too complex, diminishing returns
2. **No threshold**: Risks coverage regression
3. **Separate thresholds per package**: More complex CI

## Notes
Acceptance tests (`TestAcc*`) provide comprehensive coverage of CRUD operations including error handling, edge cases, and Terraform state management.
