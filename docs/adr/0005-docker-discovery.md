# ADR 0005: Docker Discovery Through CLI

## Status

Accepted

## Context

The MVP must detect Docker Compose services while remaining compatible with Docker Desktop, OrbStack, and Colima where possible.

## Decision

Use the installed Docker CLI with fixed argument arrays. Do not couple the core directly to one commercial Docker runtime.

## Consequences

- Discovery can work across Docker-compatible contexts.
- Absence of Docker is non-fatal.
- CLI output parsing must be defensive and tested with fixtures.
- Only host-published ports are routable in the MVP.

## Alternatives Considered

- Docker Engine API: more structured but couples to socket behavior and runtime details.
- Docker Desktop-specific integration: rejected because the product must not require Docker Desktop specifically.
- Compose file mutation: rejected because discovery is read-only.
