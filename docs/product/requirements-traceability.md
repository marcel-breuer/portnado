# Requirement Traceability

Status values:

- `planned`: designed for a later phase.
- `designed`: Phase 1 architecture exists.
- `implemented`: code exists.
- `verified`: automated or manual validation exists.
- `deferred`: explicitly out of current phase.

## Platform and Installation

| ID | Requirement | Phase | Status | Evidence |
| --- | --- | --- | --- | --- |
| FR-001 | MVP runs on Apple Silicon macOS only. | 1 | designed | PRD and architecture scope |
| FR-002 | Product installs from a custom Homebrew Cask. | 6 | implemented | `packaging/homebrew/Casks/portnado.rb` installs the app and exposes the CLI; `packaging/homebrew/publish-tap.sh` publishes the Cask to the tap |
| FR-003 | Release contains a macOS app bundle with embedded CLI and daemon. | 6 | implemented | `scripts/package-darwin-arm64.sh` creates `Portnado.app` with embedded Go binaries |
| FR-004 | Product operates without an Apple Developer Program certificate. | 6 | implemented | Release docs and Cask caveats describe unsigned/unnotarized status |
| FR-005 | Product never bypasses Gatekeeper automatically. | 6 | implemented | Cask and docs avoid Gatekeeper-bypass commands |

## Discovery

| ID | Requirement | Phase | Status | Evidence |
| --- | --- | --- | --- | --- |
| FR-010 | Detect running Docker Compose services. | 3 | verified | Docker CLI detector parses Compose projects and published ports; unit tests cover parser behavior |
| FR-011 | Docker discovery uses Docker-compatible CLI behavior. | 3 | implemented | Detector invokes fixed Docker CLI arguments |
| FR-012 | Detect native listening processes. | 3 | verified | Native detector parses fixed-argument `lsof` output; unit tests cover listener parsing |
| FR-013 | Classify Node.js, PHP, Python, Go, and Java processes. | 3 | verified | Runtime classifier covers requested families with unit tests |
| FR-014 | Associate detected services with project roots. | 3 | implemented | Native detector uses cwd and Git root when available |
| FR-015 | Include evidence and confidence for automatic associations. | 3 | implemented | Observations include evidence and confidence |
| FR-016 | Unknown runtimes remain routable after manual confirmation. | 3 | verified | Manual `.portnado.yml` targets create route suggestions with loopback backend validation |
| FR-017 | Discovery does not modify project files or processes. | 3 | designed | Threat model |

## Configuration

| ID | Requirement | Phase | Status | Evidence |
| --- | --- | --- | --- | --- |
| FR-020 | Repositories may define `.portnado.yml`. | 3 | verified | Strict YAML parser, `config validate`, unit tests, and Phase 7 fuzzing |
| FR-021 | Local overrides remain outside the repository. | 3 | implemented | Local override schema and merge helper |
| FR-022 | Configuration precedence is deterministic and documented. | 3 | implemented | Effective service merge helper follows ADR 0006 subset |
| FR-023 | Configuration validation produces actionable errors. | 3 | implemented | Parser returns field-specific errors |
| FR-024 | `portnado init` proposes config without overwriting existing files. | 3 | verified | `portnado init --dry-run` renders config and write mode refuses existing `.portnado.yml`; CLI and config tests cover behavior |

## HTTP Routing

| ID | Requirement | Phase | Status | Evidence |
| --- | --- | --- | --- | --- |
| FR-030 | HTTP services receive stable `.localhost` routes. | 4 | verified | Approved HTTP routes are served by the high-port proxy and proxy tests cover routing |
| FR-031 | Proxy routes according to Host header. | 4 | verified | HTTP proxy normalizes Host and strips optional port; route-by-host test covers behavior |
| FR-032 | Proxy supports WebSockets. | 4 | verified | Upgrade proxy test covers `101 Switching Protocols` |
| FR-033 | Proxy supports streaming and SSE. | 4 | verified | Streaming proxy test covers flushed response |
| FR-034 | Proxy returns safe local 502 diagnostics. | 4 | verified | Unavailable backend test covers sanitized 502 |
| FR-035 | Proxy binds only to IPv4 loopback. | 4 | verified | HTTP proxy listens on `tcp4` loopback address and test covers shutdown |
| FR-036 | Portless HTTP mode available after explicit setup. | 4 | planned | PF setup remains preview-only; apply/rollback is not implemented |
| FR-037 | High-port fallback works without privileged setup. | 4 | verified | Default proxy address is `127.0.0.1:4780`; live Phase 4 smoke and tests covered high-port mode |

## TCP Routing

| ID | Requirement | Phase | Status | Evidence |
| --- | --- | --- | --- | --- |
| FR-040 | Raw TCP services receive stable frontend ports. | 4 | verified | TCP forwarder tests cover approved TCP routes |
| FR-041 | TCP listeners bind only to IPv4 loopback. | 4 | verified | TCP forwarder listens on `tcp4` `127.0.0.1` |
| FR-042 | Frontend ports persist between scans and restarts. | 4 | verified | Confirmed route recovery after SQLite reopen is tested |
| FR-043 | Port conflicts detected before listener creation. | 4 | verified | TCP listener conflict behavior is tested before activation reload completes |
| FR-044 | Free frontend ports may be proposed automatically. | 4 | verified | TCP suggestions receive deterministic port from pool with service tests |
| FR-045 | Changing confirmed stable frontend port requires approval. | 4 | designed | PRD |

## State and Lifecycle

| ID | Requirement | Phase | Status | Evidence |
| --- | --- | --- | --- | --- |
| FR-050 | Persist projects, services, routes, and observations in SQLite. | 3 | verified | SQLite migrations, scan persistence, migration reopen, and route recovery tests |
| FR-051 | Routes remain configured when backend disappears. | 4 | verified | Confirmed routes remain persisted during scan reconciliation |
| FR-052 | Stale routes are marked but not deleted automatically. | 4 | verified | Store tests cover active route transition to `stale` and reactivation when observed again |
| FR-053 | Daemon supports graceful shutdown. | 2 | verified | Daemon closes listener on context cancellation and server test covers stop behavior |
| FR-054 | Only one daemon instance runs per user. | 2 | verified | Daemon rejects an already-active Unix socket before listening |
| FR-055 | Launch at login is configurable. | 5 | implemented | `setup --launch-at-login --yes` writes a user LaunchAgent plist; real login-session behavior remains unverified |
| FR-056 | Portnado does not restore development applications after login. | 5 | designed | PRD |

## User Interfaces

| ID | Requirement | Phase | Status | Evidence |
| --- | --- | --- | --- | --- |
| FR-060 | Native menu bar app displays daemon and route status. | 2, 5 | implemented | SwiftUI menu reads daemon status, suggestions, and confirmed routes over IPC |
| FR-061 | Menu bar app supports route approval and rejection. | 5 | implemented | Menu bar can approve suggested routes and enable or disable confirmed routes |
| FR-062 | Menu bar app allows route addresses to be copied. | 5 | implemented | Menu bar copy actions write route addresses to the pasteboard |
| FR-063 | Application does not open browsers automatically. | 5 | designed | PRD |
| FR-064 | CLI provides human-readable and structured output. | 2+ | verified | `status`, `scan`, `list`, `doctor`, `setup`, `uninstall`, and route list/actions support structured JSON where daemon-backed output is returned |
| FR-065 | CLI uses stable exit codes. | 2 | implemented | CLI returns `0`, `1`, or `2` for current commands |

## Diagnostics and Cleanup

| ID | Requirement | Phase | Status | Evidence |
| --- | --- | --- | --- | --- |
| FR-070 | `doctor` validates local installation. | 5 | verified | `portnado doctor` checks platform, localhost DNS, socket permissions, SQLite, proxy reachability, Docker, and LaunchAgent state; CLI tests cover JSON |
| FR-071 | Setup supports dry-run mode. | 5 | verified | `portnado setup --dry-run` and default setup preview output are covered by CLI and Docker checks |
| FR-072 | Setup previews privileged changes. | 5 | verified | Setup plan includes PF and hosts changes with privileged markers |
| FR-073 | System modifications are reversible. | 5 | implemented | `portnado uninstall --dry-run` previews managed cleanup; PF apply/rollback remains preview-only |
| FR-074 | Uninstall preserves repository configuration. | 5 | verified | Uninstall test verifies managed local state deletion preserves repository `.portnado.yml` files |
| FR-075 | Logs are sanitized and rotated. | 3+ | planned | Threat model |

## Security Requirements

| ID | Requirement | Phase | Status | Evidence |
| --- | --- | --- | --- | --- |
| SEC-001 | Local-only networking. | 4 | verified | HTTP and TCP listeners bind to IPv4 loopback; tests and security report cover listener scope |
| SEC-002 | Target restrictions. | 4 | designed | Threat model |
| SEC-003 | IPC permissions. | 2 | implemented | Daemon creates user socket with `0600` permissions |
| SEC-004 | Privilege separation. | 5 | implemented | Daemon remains unprivileged; privileged setup is previewed and root-gated |
| SEC-005 | Shell safety. | 3+ | designed | Threat model |
| SEC-006 | Path safety. | 3+ | designed | Threat model |
| SEC-007 | Secret handling. | 3+ | designed | Threat model |
| SEC-008 | Database permissions. | 3 | designed | SQLite schema |
| SEC-009 | Host validation. | 4 | verified | `.localhost` validation rejects unsafe characters, wildcards, IP-style labels, and is fuzzed |
| SEC-010 | Resource limits. | 2+ | verified | IPC frame limit, HTTP header/time limits, and Phase 7 tests are in place |
| SEC-011 | No telemetry. | 6 | implemented | Privacy docs state no telemetry, analytics, cloud processing, accounts, or crash uploads |
| SEC-012 | Gatekeeper integrity. | 6 | implemented | Release docs, README, Cask, and security policy avoid automatic Gatekeeper bypass |
| SEC-013 | Threat model maintained. | 1 | verified | Threat model and Phase 7 security report reflect current boundaries and residual risks |
