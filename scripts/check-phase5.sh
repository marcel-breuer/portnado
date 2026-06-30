#!/bin/sh
set -eu

check-phase4

test -s internal/system/setup.go
test -s internal/doctor/doctor.go
test -s internal/platform/darwin/launchagent.go

go test ./internal/system ./internal/doctor ./internal/cli
go run ./cmd/portnado setup --dry-run --json >/tmp/portnado-setup.json
go run ./cmd/portnado uninstall --dry-run --json >/tmp/portnado-uninstall.json

grep -q '"launch-agent"' /tmp/portnado-setup.json
grep -q '"pf-portless-http"' /tmp/portnado-setup.json
grep -q '"local-state"' /tmp/portnado-uninstall.json || go run ./cmd/portnado uninstall --dry-run --delete-state --json | grep -q '"local-state"'

echo "Phase 5 checks passed"
