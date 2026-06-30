# ADR 0003: SQLite Persistence

## Status

Accepted

## Context

Portnado needs durable local state for projects, services, observations, route suggestions, confirmed routes, TCP port assignments, settings, and diagnostics.

## Decision

Use SQLite stored under:

```text
~/Library/Application Support/Portnado/portnado.db
```

The daemon is the single writer. Migrations are versioned, foreign keys are enabled, and WAL mode is used where appropriate.

## Consequences

- State remains local and inspectable.
- The daemon controls transaction boundaries.
- Migration tests become part of core test coverage.
- Phase 3 must choose a Go SQLite driver and document CGO implications.

## Alternatives Considered

- JSON files: simple but weak for relational state and transactional updates.
- BoltDB or similar embedded stores: workable but less natural for relational queries.
- External database: out of scope for a local developer tool.
