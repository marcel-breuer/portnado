# Privileged Setup Analysis

## Principle

Portnado runs unprivileged by default. Privileged setup is optional, previewed, explicit, and reversible.

## Portless HTTP Strategy

Phase 1 selects PF redirect to an unprivileged proxy as the preferred design. The runtime daemon listens on `127.0.0.1:4780`. Approved system setup redirects loopback TCP port `80` to that high port.

Phase 5 implements dry-run planning for the managed PF redirect. Applying the PF anchor is still guarded behind administrator approval and is not performed automatically by the current implementation.

## Compared Options

| Option | Benefit | Risk | Decision |
| --- | --- | --- | --- |
| Root daemon binds port 80 | Simple runtime routing | Long-lived root process with broad attack surface | Reject |
| PF redirect to unprivileged proxy | Keeps daemon unprivileged, narrow system rule | Requires careful root-owned setup and rollback | Select |
| High-port only | No privileges required | URLs include `:4780` | Required fallback |

## Required Setup Behavior

- Support `--dry-run`.
- Show exact files and rules before applying.
- Require confirmation.
- Use static templates.
- Validate substituted ports and anchors.
- Back up replaced managed files.
- Validate after applying.
- Roll back on failure where possible.
- Never flush the main PF ruleset.
- Never disable or bypass Gatekeeper.

## Managed Files

- User LaunchAgent: `~/Library/LaunchAgents/dev.portnado.daemon.plist`.
- Managed PF anchor preview: `/etc/pf.anchors/dev.portnado`, based on
  `packaging/pf/dev.portnado.anchor.in`.
- Optional hosts fallback preview: `/etc/hosts`.

Repository `.portnado.yml` files must never be removed by uninstall.

## Host Entries

Phase 1 observed wildcard `.localhost` resolution working. `/etc/hosts` fallback remains optional and must require separate approval only if `doctor` proves resolution failure.
