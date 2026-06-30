# ADR 0002: Local IPC Protocol

## Status

Accepted

## Context

The CLI and menu bar app need to communicate with the daemon without exposing a network management endpoint.

## Decision

Use a Unix domain socket under:

```text
~/Library/Application Support/Portnado/run/portnado.sock
```

Use bounded newline-delimited JSON envelopes with `protocolVersion`, `requestId`, `method`, and `params`.

## Consequences

- No TCP control API is exposed.
- The protocol is easy to inspect and test.
- Message size limits and deadlines are required to avoid resource exhaustion.
- Versioning is mandatory from the first implementation.

## Alternatives Considered

- TCP localhost API: easier for debugging but unnecessary exposure.
- XPC: native macOS fit but adds complexity and is less natural for the Go CLI and daemon.
- gRPC: strong tooling but more dependency weight than needed for MVP.
