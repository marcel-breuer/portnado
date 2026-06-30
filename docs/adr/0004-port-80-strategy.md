# ADR 0004: Port 80 Strategy

## Status

Accepted for design, unimplemented

## Context

The desired HTTP route is:

```text
http://app.project.localhost
```

Port 80 requires privileges. The daemon must not run permanently as root.

## Decision

Implement high-port mode first on `127.0.0.1:4780`. For portless mode, use a dedicated PF redirect from loopback TCP port `80` to the unprivileged proxy port after explicit user approval.

## Consequences

- The daemon remains unprivileged.
- Portless setup is optional and reversible.
- `http://app.project.localhost:4780` remains the fallback.
- PF setup implementation must preview exact changes, require approval, avoid flushing the main ruleset, validate health, and support rollback.

## Alternatives Considered

- Root daemon binding port 80: rejected because it creates a long-lived privileged process.
- High-port only: retained as required fallback but not the ideal user experience.
- Local DNS server: out of MVP scope.
