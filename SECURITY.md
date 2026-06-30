# Security Policy

Portnado is pre-release software. Do not rely on it for production traffic.

## Reporting

Use private maintainer contact channels when available. Until a dedicated disclosure address exists, avoid posting exploitable details publicly.

## Scope

Security-sensitive areas include:

- local routing listeners,
- IPC socket permissions,
- target validation,
- privileged setup,
- logs and diagnostic redaction,
- repository configuration parsing.

## Current Status

Phase 6 contains local discovery, SQLite state, high-port loopback routing,
route lifecycle commands, setup previews, doctor diagnostics, uninstall
planning, a Swift menu bar app, and release packaging templates.

The project is not Developer ID signed or notarized. Portnado must not disable
Gatekeeper, remove quarantine automatically, or modify global macOS security
policy.
