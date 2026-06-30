#!/bin/sh
set -eu

check-phase1

test -s go.mod
test -s cmd/portnado/main.go
test -s cmd/portnado-daemon/main.go
test -s apps/Portnado/Package.swift
test -s Makefile

unformatted="$(gofmt -l cmd internal pkg)"
if [ -n "$unformatted" ]; then
  echo "Go files need gofmt:" >&2
  echo "$unformatted" >&2
  exit 1
fi

go test ./...
go vet ./...

echo "Phase 2 Go checks passed"
