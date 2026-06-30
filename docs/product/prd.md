# Product Requirements Document

## Product

Portnado provides stable local routes for development services whose ports change over time.

Tagline: Stable local routes for changing development ports.

## MVP Platform

- macOS only.
- Apple Silicon only.
- `arm64` only.
- macOS 14 or newer unless implementation later proves a higher minimum is required.
- Distribution through the custom Homebrew tap `marcel-breuer/homebrew-portnado`.

## Problem

Local development services frequently move between ports such as `3000`, `4173`, `5173`, `8080`, or dynamically published Docker ports. Developers lose time checking which port is active and updating dependent tools.

## Goals

- Detect already-running local services.
- Suggest stable `.localhost` routes for HTTP and raw TCP services.
- Preserve confirmed stable addresses when backend ports change.
- Require approval before activating new routes or changing stable public route choices.
- Keep all processing local.
- Provide a native menu bar app, CLI, and user daemon.

## Non-goals

- Starting, stopping, supervising, or restoring project processes.
- HTTPS, local certificate authorities, or `mkcert`.
- Linux, Windows, Intel Mac, IPv6, LAN exposure, or remote access.
- Telemetry, analytics, accounts, crash uploads, or cloud sync.
- Kubernetes or production routing.
- Automatic browser opening.
- Gatekeeper bypass or notarization claims.

## Primary Scenarios

### Docker Compose Project

When a running Compose project publishes ports, Portnado proposes routes such as:

```text
app.webguard.localhost -> 127.0.0.1:4173
api.webguard.localhost -> 127.0.0.1:8082
db.webguard.localhost:15432 -> 127.0.0.1:5438
```

Only host-reachable published ports are routable in the MVP.

### Native Development Process

When a native development process listens on loopback and can be associated with a repository, Portnado proposes a service route such as:

```text
app.project-name.localhost -> 127.0.0.1:5173
```

### Backend Port Change

When the same confirmed project and service later listen on a different backend port, Portnado updates the backend target while preserving the stable route when the previous approval permits equivalent backend reassociation.

### Backend Unavailable

Portnado keeps confirmed routes when backends disappear. HTTP routes return a sanitized local `502` diagnostic response; TCP routes fail cleanly and record a sanitized diagnostic event.

## User Approval Model

Discovery is read-only and may run without approval. Approval is required before:

- Activating a newly suggested route.
- Changing a confirmed hostname.
- Changing a confirmed stable TCP frontend port.
- Adding launch-at-login configuration.
- Installing PF configuration.
- Modifying `/etc/hosts`.
- Deleting local state.
- Removing system integration.

## Success Criteria for v0.1.0

- CLI, daemon, and menu bar app build for Apple Silicon.
- Docker Compose and native services are detected.
- HTTP routes, WebSocket upgrades, streaming responses, and TCP forwarding work.
- Routes persist and require approval.
- High-port fallback works without privileged setup.
- Portless mode works after approved setup.
- Launch at login is configurable.
- `doctor` produces actionable diagnostics.
- Uninstall reverses managed system changes.
- Listeners bind only to `127.0.0.1`.
- No telemetry occurs.
- Logs and diagnostics do not expose secrets.
- Testable core logic reaches at least 80% line coverage.

## Assumptions

- `.localhost` wildcard names resolve to loopback on supported macOS systems, but `doctor` must verify this.
- Docker-compatible CLIs expose enough Compose labels and published port data for MVP discovery.
- High-port HTTP mode is always available unless another local process owns the chosen port.
- Portless mode can be implemented with a dedicated PF anchor after explicit approval.
