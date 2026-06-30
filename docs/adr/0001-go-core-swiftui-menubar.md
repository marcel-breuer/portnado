# ADR 0001: Go Core and SwiftUI Menu Bar

## Status

Accepted

## Context

Portnado needs local networking, process discovery, SQLite persistence, a CLI, a daemon, and a native macOS menu bar app.

## Decision

Use Go for the CLI, daemon, discovery, routing, persistence, and platform integration. Use Swift and SwiftUI for the macOS menu bar application.

## Consequences

- Go keeps networking and CLI code simple and portable for future extension.
- SwiftUI provides native menu bar integration and accessibility behavior.
- Release packaging must embed Go binaries inside the app bundle.
- The UI and core communicate over a local versioned IPC protocol.

## Alternatives Considered

- All Swift: weaker fit for CLI, proxying, and future cross-platform core.
- All Go with non-native UI: weaker macOS menu bar experience.
- Electron: unnecessary runtime weight and broader attack surface for MVP.
