# Component Boundaries

## Go Daemon

Responsibilities:

- Own the single-writer SQLite connection.
- Run migrations.
- Run scans.
- Maintain route state.
- Own HTTP and TCP listeners.
- Serve local IPC.
- Enforce approval and target restrictions.
- Emit sanitized structured logs.
- Hold the single-instance lock.

The daemon must not start, stop, restart, or supervise development project processes.

## Go CLI

Responsibilities:

- Render human-readable and JSON output.
- Request daemon actions over IPC.
- Validate configuration.
- Perform setup and uninstall dry runs.
- Apply approved system setup in later phases through narrow platform helpers.
- Provide stable exit codes.

The CLI should not duplicate daemon routing state decisions.

## SwiftUI Menu Bar App

Responsibilities:

- Show daemon reachability.
- Show route health, project groups, warnings, and suggestions.
- Request scans, pause/resume, route approval, and settings changes through IPC.
- Configure launch at login through approved system integration.
- Avoid silent administrator prompts.

## Platform Adapters

Adapters isolate macOS-specific behavior:

- process inspection,
- LaunchAgent management,
- PF setup,
- app bundle checks,
- socket permissions,
- `.localhost` diagnostics.

## Shared Protocol Package

The public Go package `pkg/protocol` should contain versioned IPC envelopes and stable DTOs that are shared by the CLI, daemon, and Swift bridge generator if one is introduced.
