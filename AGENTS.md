# Repository Instructions

These instructions apply to all development assistants and contributors working in this repository.

## Project Scope

Portnado is a macOS Apple Silicon developer tool. The MVP is limited to macOS 14 or newer on `arm64`; do not add Linux, Windows, Intel Mac, HTTPS, telemetry, process supervision, or automatic project startup unless a later approved phase explicitly changes scope.

## Architecture

- Go owns the CLI, daemon, discovery, routing, persistence, local IPC, and platform integration.
- Swift and SwiftUI own the macOS menu bar application.
- SQLite is the local state store.
- YAML is used for repository configuration and local overrides.
- Keep platform-specific code isolated under Darwin-specific packages or app targets.
- Keep security boundaries explicit: the daemon runs as the current user and privileged setup is minimal, reversible, and approval-gated.

## Workflow

- Inspect existing code and Docker configuration before editing.
- Work inside the provided Docker environment whenever technically possible.
- Prefer project-defined Docker Compose, Make, or script commands over ad-hoc host commands.
- Run targeted checks first, then the full relevant suite before completion.
- Do not perform unrelated refactors.
- Do not delete, rewrite, or revert user work without explicit approval.
- Use Conventional Commits for commits.

## UI Work

- For new UI views, UI redesigns, and focused UI reviews, use the project skill in `.agents/skills/unslop/` (`$unslop`) before changing or reviewing the interface. Follow its audit-first approach and fix only issues supported by the audit; preserve existing behavior and avoid an unrequested redesign.
- Keep Portnado's product identity, existing SwiftUI patterns, and reusable components. The upstream skill includes web-specific guidance; apply its visual principles to native SwiftUI rather than introducing web technologies or copying web-only implementation steps.
- For UI changes, check the supported macOS layouts, accessibility and contrast, and visual consistency. Run relevant existing UI checks when available and include before/after evidence in the change review when the rendered interface changes.

## Testing

- Target at least 80% line coverage for testable core logic.
- Run Go tests, race tests, vet, formatting checks, schema checks, and Swift tests when those components exist.
- Docker-based integration tests should cover portable scenarios.
- macOS PF, LaunchAgent, Gatekeeper, app bundle, and Homebrew Cask behavior require explicit macOS host procedures when they cannot run in Docker.

## Security

- Bind runtime listeners only to `127.0.0.1`.
- Do not bind to `0.0.0.0`; IPv6 routing is out of MVP scope.
- Do not construct shell command strings from project configuration.
- Do not log secrets, raw environment files, credentials, or sensitive command arguments.
- Reject non-local proxy targets by default.
- Do not bypass Gatekeeper automatically.
- Do not run `xattr` to remove quarantine automatically.
- Do not modify `/etc/hosts`, PF, launchd, or other system files without preview and explicit user approval.
- Do not include AI-tool attribution in code comments, docs, commits, changelog entries, PR titles, or PR descriptions.

## Dependencies

- Prefer the standard library.
- Add runtime dependencies only when they substantially reduce complexity or security risk.
- Document dependency purpose, license, maintenance status, and pinning.
- Avoid redundant libraries and unmaintained packages.

## Documentation

- Keep ADRs current when architecture decisions change.
- Update requirement traceability when implementing or deferring requirements.
- Document unverified macOS behavior clearly.
- Keep change summaries concise and concrete.
