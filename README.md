# Portnado

Stable local routes for changing development ports.

Portnado is a macOS menu bar app and CLI for developers who run local services
whose ports keep changing. It discovers running development services, proposes
stable `.localhost` names, and routes approved names back to loopback backends.

No real product screenshot exists yet. Do not use mock screenshots in release
materials.

## Supported Platform

The MVP targets Apple Silicon Macs running macOS 14 or newer.

## Installation

Portnado is packaged as a single release archive:

```text
Portnado-vX.Y.Z-darwin-arm64.zip
```

The archive contains `Portnado.app`, with embedded `portnado` and
`portnado-daemon` binaries under `Contents/Resources/bin`.

Build a local archive:

```bash
PORTNADO_VERSION=0.1.0 make package-darwin-arm64
```

Install the release with Homebrew:

```bash
brew install --cask marcel-breuer/portnado/portnado
```

The tap repository is `marcel-breuer/homebrew-portnado`, exposed by Homebrew as
`marcel-breuer/portnado`. The Cask lives at
`packaging/homebrew/Casks/portnado.rb` in this repository and is published to
the tap during tagged releases when `HOMEBREW_TAP_TOKEN` is configured.

## Gatekeeper

Portnado currently ships without Developer ID notarization. macOS may report
that the developer cannot be verified, or that Apple cannot check the app for
malicious software, on first launch. Portnado does not include commands that
disable Gatekeeper, remove quarantine automatically, or modify global macOS
security policy.

## First Setup

Initialize optional repository config, start the daemon, scan the current
repository, and approve routes:

```bash
portnado init --dry-run
portnado init --project my-project --service app --target-port 5173
portnado-daemon
portnado scan --root "$PWD"
portnado list
portnado route approve <suggestion-id>
```

Preview setup changes:

```bash
portnado setup --dry-run
```

Install launch at login only after reviewing the plan:

```bash
portnado setup --launch-at-login --daemon-path /Applications/Portnado.app/Contents/Resources/bin/portnado-daemon --yes
```

## Current Commands

```bash
go run ./cmd/portnado --version
go run ./cmd/portnado status
go run ./cmd/portnado init --dry-run
go run ./cmd/portnado scan
go run ./cmd/portnado list
go run ./cmd/portnado doctor
go run ./cmd/portnado setup --dry-run
go run ./cmd/portnado uninstall --dry-run
go run ./cmd/portnado route approve <suggestion-id>
go run ./cmd/portnado route list --json
go run ./cmd/portnado route disable <route-id>
go run ./cmd/portnado route enable <route-id>
go run ./cmd/portnado config validate .portnado.yml
go run ./cmd/portnado-daemon
```

`portnado status`, `scan`, `list`, and `route` commands communicate with the daemon over the local Unix socket. `portnado doctor`, `setup`, `uninstall`, and `config validate` run local checks directly. HTTP routing currently listens on `127.0.0.1:4780`.

## Menu Bar Overview

The menu bar app displays daemon status, route suggestions, confirmed routes,
copy-address actions, route approval, and enable/disable controls. It does not
open browsers automatically.

## Runtime Compatibility

Docker Compose discovery reads Docker-compatible CLI output and detects
published loopback ports. Native runtime discovery inspects listening local
processes and classifies common development runtimes including Node.js, PHP,
Python, Go, and Java. Unknown or custom runtimes can still be routed by adding a
manual service target in `.portnado.yml`.

## Repository Configuration

Projects may generate `.portnado.yml` with `portnado init`. Existing config is
never overwritten. Manual targets can route services that automatic discovery
cannot classify:

```yaml
version: 1
project:
  name: webguard
services:
  app:
    protocol: http
    route:
      host: app.webguard.localhost
    target:
      discovery: manual
      preferredPort: 5173
```

Repository config is never removed by uninstall or Cask `zap`.

## Security And Privacy

Portnado binds local listeners to `127.0.0.1`, stores state locally in SQLite,
uses a user-scoped Unix socket, and has no telemetry, analytics, cloud service,
or account system. Optional privileged setup is previewed before use and the
daemon remains unprivileged.

## Troubleshooting

Run:

```bash
portnado doctor
```

Useful checks include `.localhost` resolution, control socket permissions,
SQLite readability, high-port proxy reachability, Docker CLI availability, and
LaunchAgent state.

If routes do not respond, confirm the backend is still running, the route is
approved and enabled, and the request uses the expected Host header or copied
menu bar address.

## Uninstall

Preview managed cleanup:

```bash
portnado uninstall --dry-run
```

Apply user-scope cleanup:

```bash
portnado uninstall --yes
```

Delete local Portnado state only when intended:

```bash
portnado uninstall --delete-state --yes
```

Repository `.portnado.yml` files are preserved.

## Development

Prefer Docker for checks that can run in containers:

```bash
make phase7-hardening-check
```

SwiftUI checks require macOS host tooling:

```bash
make swift-build
make swift-test
```

Full available local verification:

```bash
make verify
```

## Roadmap

Remaining release work is external verification: clean-machine installation,
installation from the public Homebrew tap, Developer ID signing/notarization if
a certificate becomes available, LaunchAgent behavior across a real login
session, PF apply/rollback with administrator approval, and sustained menu bar
resource profiling.

## Contributing

See `CONTRIBUTING.md`.

## License

MIT.
