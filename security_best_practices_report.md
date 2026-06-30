# Security Best Practices Report

## Executive Summary

No critical or high severity findings were identified in the Phase 7 Go security
review. The review found local-only listener controls, bounded IPC frames,
parameterized SQL, fixed-argument subprocess execution, and CI vulnerability
scanning. One low-severity residual risk remains around long-lived streaming
HTTP responses, documented below.

## Critical Findings

None.

## High Findings

None.

## Medium Findings

None.

## Low Findings

### SBP-001: HTTP proxy allows unbounded write duration for streaming compatibility

- Rule ID: GO-HTTP-001
- Severity: Low
- Location: `internal/routing/http/proxy.go:34`
- Evidence: `http.Server` sets `ReadHeaderTimeout`, `ReadTimeout`,
  `IdleTimeout`, and `MaxHeaderBytes`; `WriteTimeout` is intentionally `0` so
  WebSocket upgrades and server-sent events are not cut off.
- Impact: A local client could keep a proxied streaming response open for a long
  time. The exposure is limited because the listener is loopback-only.
- Fix: Add route-aware write deadlines only if future profiling shows local
  resource exhaustion.
- Mitigation: Keep loopback-only binding, bounded headers, read timeout, and
  idle timeout. Add connection limits in a later hardening pass if needed.
- False positive notes: This is compatible with the product requirement for
  WebSocket and streaming support.

## Positive Controls Reviewed

- Local-only HTTP proxy defaults to `127.0.0.1:4780` and explicit server limits:
  `internal/routing/http/proxy.go:28-42`.
- Route hosts are constrained to `.localhost`, reject wildcards, separators,
  whitespace, and IP-style labels: `internal/config/config.go:232-257`.
- IPC frame size is bounded by `MaxFrameSize`: `pkg/protocol/codec.go:49-56`.
- Docker, native, and Git discovery use `exec.CommandContext` with fixed command
  names and argument arrays, not shell strings:
  `internal/discovery/docker/docker.go:24-26`,
  `internal/discovery/native/native.go:23-25`,
  `internal/discovery/project/project.go:16-17`.
- SQLite uses parameterized statements for route writes and state changes:
  `internal/persistence/sqlite.go:189-207` and
  `internal/persistence/sqlite.go:218-231`.
- Local database setup creates the parent directory with user-only permissions:
  `internal/persistence/sqlite.go:30-36`.
- Uninstall preserves repository files and only deletes managed local state when
  explicitly requested: `internal/system/setup.go:123-170`.
- CI now runs Dockerized Phase 7 checks and govulncheck:
  `.github/workflows/ci.yml:18-22`.

## Review Evidence

- `go test ./...`: passed.
- `go test -race ./...`: passed.
- Parser fuzzing: passed for repository YAML, localhost validation, and IPC
  request frames.
- `go mod verify`: passed.
- `govulncheck ./...`: no vulnerabilities found.
