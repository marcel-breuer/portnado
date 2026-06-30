# ADR 0006: Configuration Precedence

## Status

Accepted

## Context

Portnado combines CLI arguments, local overrides, repository configuration, persisted choices, discovery, and defaults. Ambiguity would make route behavior hard to explain.

## Decision

Use this precedence from highest to lowest:

1. explicit CLI argument for the current command,
2. local project override,
3. repository `.portnado.yml`,
4. confirmed persisted user choice,
5. automatic discovery,
6. application defaults.

Reject ambiguous or invalid combinations with actionable errors.

## Consequences

- Behavior is deterministic.
- Local machine preferences stay outside the repository.
- Confirmed user choices remain durable but can be overridden intentionally.
- Validation must report the source of each effective value.

## Alternatives Considered

- Repository config above local override: rejected because local machine routing may need private adjustments.
- Discovery above persisted choice: rejected because it would undermine stable approved routes.
