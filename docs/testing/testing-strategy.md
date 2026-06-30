# Testing Strategy

## Coverage Target

Testable core logic targets at least 80% line coverage. Exclusions must be explicit and limited to generated files, trivial entry points, platform glue that cannot reasonably run in CI, and Xcode project metadata.

## Phase 1 Checks

Phase 1 provides a Dockerized validation harness:

```bash
docker compose run --rm phase1-checks
```

It verifies required documents, ADR structure, and JSON schema syntax.

## Phase 2 Checks

Phase 2 adds Dockerized Go checks:

```bash
docker compose run --rm phase2-go-checks
```

The check runs Phase 1 validation, `gofmt` verification, Go tests, and `go vet`. Swift checks run on the macOS host because SwiftUI and AppKit are not available in the Linux container.

## Phase 3 Checks

Phase 3 adds:

```bash
docker compose run --rm phase3-go-checks
```

The check runs Phase 1 and Phase 2 validation plus Go race tests for configuration, persistence, discovery, and route suggestions.

## Phase 4 Checks

Phase 4 adds:

```bash
docker compose run --rm phase4-go-checks
```

The check runs earlier validation plus routing package tests for HTTP proxying, WebSocket-style upgrades, streaming responses, and raw TCP forwarding. Host race tests should still be run separately because Docker uses CGO-disabled Alpine.

## Phase 5 Checks

Phase 5 adds:

```bash
docker compose run --rm phase5-go-checks
```

The check runs earlier validation plus setup, doctor, uninstall, and CLI tests. Swift menu bar route-action tests still run on the macOS host because SwiftUI and AppKit are not available in the Linux container.

## Phase 6 Checks

Phase 6 adds:

```bash
docker compose run --rm phase6-release-checks
```

The check validates packaging scripts, app metadata templates, Homebrew Cask
contents, and tap update behavior. Building `Portnado.app` and the release ZIP
requires macOS host tooling:

```bash
PORTNADO_VERSION=0.1.0 make package-darwin-arm64
```

## Phase 7 Checks

Phase 7 adds:

```bash
docker compose run --rm phase7-hardening-checks
```

The check runs earlier validation plus Go race tests, parser fuzzing, benchmarks,
module verification, and release-candidate document checks.

Run the vulnerability scanner on a network-enabled host or CI runner:

```bash
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

## Go Unit Tests

Later phases should cover:

- hostname normalization,
- project and service naming,
- configuration precedence,
- YAML parsing,
- environment interpolation,
- conflict detection,
- deterministic port allocation,
- route state transitions,
- runtime classification,
- command sanitization,
- secret redaction,
- IPC framing,
- database migrations,
- proxy header handling.

## Integration Tests

Use Docker for portable scenarios:

- Compose frontend, API, and PostgreSQL fixtures,
- dynamic published ports,
- unavailable backend,
- WebSocket echo,
- Server-Sent Events,
- large streaming response,
- raw TCP echo,
- overlapping container ports,
- malformed Compose metadata.

## macOS System Tests

Manual or host-specific automated tests are required for:

- menu bar app launch,
- LaunchAgent behavior,
- Unix socket permissions,
- `.localhost` resolution,
- PF setup and rollback,
- port 80 forwarding,
- Homebrew Cask installation,
- Gatekeeper documentation flow,
- uninstall cleanup.

Do not claim these are verified from Docker.

## Security Tests

Security test cases must include traversal, malicious YAML, shell metacharacters, symlink replacement, invalid socket permissions, remote target rejection, accidental `0.0.0.0` binds, secret redaction, malformed IPC frames, oversized IPC messages, and duplicate daemon startup.
