# Threat Model

## Scope

Portnado is a local-only routing layer for development machines. The MVP has no cloud service, no account system, and no telemetry.

## Assets

- Local route configuration.
- SQLite state.
- Local override files.
- Repository `.portnado.yml` files.
- Logs and diagnostic events.
- Optional PF and launchd system integration.
- Developer workstation network boundaries.

## Trust Boundaries

- Repository configuration is untrusted input.
- Running local processes are untrusted.
- HTTP requests from browsers and local tools are untrusted.
- Docker CLI output is external input and must be parsed defensively.
- IPC clients are local but still must be authenticated by socket permissions and validated protocol envelopes.
- Privileged setup crosses into root-owned system configuration.

## Requirements Coverage

### SEC-001 Local-only Networking

All listeners bind to `127.0.0.1`. Binding to `0.0.0.0` is a security regression. IPv6 routing is out of MVP scope.

### SEC-002 Target Restrictions

Proxy targets are loopback or recognized local published container ports. Public IPs, LAN targets, arbitrary hostnames, and metadata endpoints are rejected by default.

### SEC-003 IPC Permissions

The IPC socket and parent directory are current-user only. The daemon validates protocol version, frame size, method names, and request parameters.

### SEC-004 Privilege Separation

The daemon does not run permanently as root. Privileged changes use static templates, absolute executable paths, safe permissions, and rollback.

### SEC-005 Shell Safety

Project configuration must never be interpolated into shell command strings. Use direct process execution with fixed executable paths and argument arrays.

### SEC-006 Path Safety

File writes must defend against symlinks, traversal, world-writable parent directories, unsafe temporary files, and race conditions.

### SEC-007 Secret Handling

Environment files may be read for interpolation but raw values are not logged or persisted. Credential-like values are redacted from command lines and diagnostics.

### SEC-008 Database Permissions

The application support directory, database, and run directory are restricted to the current user.

### SEC-009 Host Validation

Route hosts are normalized and limited to `.localhost`. Invalid labels, control characters, path separators, wildcard labels, and duplicate names are rejected.

### SEC-010 Resource Limits

IPC frame size, scan frequency, command length, log retention, diagnostic retention, and proxy concurrency must have explicit limits.

### SEC-011 No Telemetry

The MVP performs no analytics or telemetry requests.

### SEC-012 Gatekeeper Integrity

Portnado never removes quarantine automatically and never modifies global macOS security policies.

### SEC-013 Threat Model Maintenance

This document is updated when architecture decisions or trust boundaries change.

## Key Threats

| Threat | Risk | Mitigation |
| --- | --- | --- |
| Malicious repository config defines unsafe host or target | Open proxy or confusing route | Strict schema, `.localhost` only, local targets only |
| Compromised local process receives proxied traffic | Data exposure between local apps | Explicit route approval and clear backend display |
| Hostile browser request abuses proxy | SSRF or header spoofing | Host allowlist, forwarding header normalization, local targets only |
| DNS rebinding assumptions fail | Unexpected route resolution | `doctor` resolver checks and no LAN binding |
| Privileged setup injection | Root compromise | Static templates, validation, previews, rollback |
| Symlink replacement during setup | Unsafe file write | Open with safe flags, verify ownership and path ancestry |
| IPC impersonation | Unauthorized route changes | User-only socket path and bounded versioned protocol |
| Log leakage | Secret disclosure | Redaction and no raw environment persistence |
| Dependency compromise | Supply-chain risk | Minimal dependencies, pinned versions, dependency review |
