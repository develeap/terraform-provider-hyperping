# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) for the terraform-provider-hyperping project.

## What is an ADR?

An ADR is a document that captures an important architectural decision made along with its context and consequences.

## ADRs

| ADR | Title | Status |
|-----|-------|--------|
| [0001](0001-client-interface-for-testability.md) | Client Interface for Testability | Accepted |
| [0002](0002-coverage-threshold-strategy.md) | Coverage Threshold Strategy | Accepted |
| [0003](0003-single-source-of-truth-for-constants.md) | Single Source of Truth for Constants | Accepted |

## Creating a New ADR

1. Copy the template below
2. Name the file `NNNN-title-with-dashes.md` where NNNN is the next number
3. Fill in the sections
4. Add to the table above

### Template

```markdown
# ADR NNNN: Title

## Status
[Proposed | Accepted | Deprecated | Superseded]

## Context
[Describe the context and problem]

## Decision
[Describe the decision]

## Consequences
[Describe the consequences, both positive and negative]
```
