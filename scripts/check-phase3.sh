#!/bin/sh
set -eu

check-phase2

test -s internal/app/service.go
test -s internal/config/config.go
test -s internal/persistence/sqlite.go
test -s internal/discovery/docker/docker.go
test -s internal/discovery/native/native.go
test -s internal/discovery/runtime/classifier.go

if [ "$(go env CGO_ENABLED)" = "1" ]; then
  go test -race ./...
else
  echo "Skipping race tests in Docker because CGO is disabled"
fi

echo "Phase 3 checks passed"
