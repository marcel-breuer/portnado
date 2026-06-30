#!/bin/sh
set -eu

check-phase3

test -s internal/routing/http/proxy.go
test -s internal/routing/tcp/forwarder.go
test -s internal/routing/manager.go

go test ./internal/routing/...

echo "Phase 4 checks passed"
