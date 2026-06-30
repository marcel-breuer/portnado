# Phase 1 Technical Spike

Date: 2026-06-30

## Environment

- Repository: `/Users/marcel/Developer/portnado`
- Repository state: empty Git repository, no commits yet.
- CPU architecture: `arm64`
- macOS: `26.5.1`, build `25F80`
- Docker: `29.6.1`
- Docker Compose: `v5.1.4`
- Go: `go1.26.4 darwin/arm64`
- Swift: Apple Swift `6.3.3`
- Xcode: `26.6`
- Homebrew: `6.0.6`

The target MVP remains macOS 14 or newer on Apple Silicon. The observed host is newer than that floor.

## `.localhost` Resolution

Read-only checks using Python `socket.getaddrinfo` resolved:

- `localhost` to `::1` and `127.0.0.1`
- `app.portnado-test.localhost` to `::1` and `127.0.0.1`
- `db.webguard.localhost` to `::1` and `127.0.0.1`

Conclusion: wildcard `.localhost` resolution works on this host. Portnado should still implement `portnado doctor` because resolver behavior can be affected by local configuration. Runtime listeners must bind to `127.0.0.1`; IPv6 routing remains out of MVP scope even if the resolver returns `::1`.

## Port 80 and High-port Fallback

Read-only connection checks:

- `127.0.0.1:80`: connection refused.
- `127.0.0.1:4780`: connection refused.

Conclusion: neither default HTTP frontend port was occupied during the spike. Phase 4 should implement high-port mode first on `127.0.0.1:4780`.

## PF Anchor Feasibility

Read-only inspection found:

- `/etc/pf.conf` exists and is root-owned.
- `/etc/pf.anchors` exists and is root-owned.
- The default PF config includes Apple anchor points and warns against flushing the main ruleset.
- `pfctl -sr` was not readable as the current user: `/dev/pf: Permission denied`.

Conclusion: direct non-privileged runtime inspection of active PF rules is limited. ADR 0004 selects a dedicated PF anchor model for portless mode, but implementation must be deferred until Phase 4 or Phase 5 and must use preview, explicit approval, rollback, and validation.

## App Bundle Layout

The proposed bundle layout is feasible:

```text
Portnado.app/
└── Contents/
    ├── MacOS/
    │   └── Portnado
    ├── Resources/
    │   └── bin/
    │       ├── portnado
    │       └── portnado-daemon
    └── Info.plist
```

SwiftUI owns the app executable. Go binaries are embedded in `Contents/Resources/bin` and exposed by the Homebrew Cask through a `binary` stanza.

## Signing and Gatekeeper

`codesign` is available. Phase 1 does not sign anything. Release engineering must not claim Developer ID signing or notarization. If ad-hoc signing is technically required for bundle integrity, it must be documented as ad-hoc only and not as Apple trust.

## Homebrew Cask Constraints

The custom Cask should install `Portnado.app`, expose the embedded `portnado` CLI, declare Apple Silicon and macOS constraints, include uninstall and `zap` behavior, and include caveats about unsigned/unnotarized distribution. It must not run Gatekeeper bypass commands.

## Commands Executed

```bash
uname -m
sw_vers
python3 - <<'PY'
import socket
for host in ['localhost', 'app.portnado-test.localhost', 'db.webguard.localhost']:
    print(host, socket.getaddrinfo(host, 80, proto=socket.IPPROTO_TCP))
PY
nc -vz -G 1 127.0.0.1 80
nc -vz -G 1 127.0.0.1 4780
ls -l /etc/pf.conf /etc/pf.anchors
sed -n '1,160p' /etc/pf.conf
pfctl -sr
swift --version
xcodebuild -version
brew --version
go version
docker --version
docker compose version
```
