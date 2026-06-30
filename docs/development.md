# Development

## Tooling

- Go core: `go test ./...`, `go vet ./...`, and `gofmt`.
- Swift menu bar app: `swift build --package-path apps/Portnado` and `swift test --package-path apps/Portnado`.
- Dockerized checks: `make phase7-hardening-check`.

## Current Executables

Build Go binaries:

```bash
make go-build
```

Run the CLI:

```bash
go run ./cmd/portnado --version
go run ./cmd/portnado status
go run ./cmd/portnado scan
go run ./cmd/portnado list
go run ./cmd/portnado doctor
go run ./cmd/portnado setup --dry-run
go run ./cmd/portnado uninstall --dry-run
go run ./cmd/portnado route approve <suggestion-id>
go run ./cmd/portnado route list
go run ./cmd/portnado config validate configs/examples/portnado.yml
```

Run the daemon:

```bash
go run ./cmd/portnado-daemon
```

In another terminal:

```bash
go run ./cmd/portnado status
go run ./cmd/portnado scan --root "$PWD"
go run ./cmd/portnado list
```

Approved HTTP routes are served through the high-port proxy:

```bash
curl -H 'Host: app.webguard.localhost' http://127.0.0.1:4780/
```

## Swift Menu Bar

The Swift package shows daemon status, route suggestions, confirmed routes,
approval and enablement actions, copy-address commands, and basic settings. The
release packaging script builds `Portnado.app` and embeds the Go CLI and daemon.
