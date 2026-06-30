#!/bin/sh
set -eu

check-phase6

if [ "$(go env CGO_ENABLED)" = "1" ]; then
  go test -race ./...
else
  echo "Skipping race tests in Docker because CGO is disabled"
fi
go test -run=Fuzz -fuzz=FuzzParseRepository -fuzztime=3s ./internal/config
go test -run=Fuzz -fuzz=FuzzValidateLocalhost -fuzztime=3s ./internal/config
go test -run=Fuzz -fuzz=Fuzz -fuzztime=3s ./pkg/protocol
go test -bench=. -benchmem ./internal/config ./internal/routing/http
go mod verify
go list -m all >/tmp/portnado-modules.txt

test -s docs/release/v0.1.0-checklist.md
test -s docs/release/performance-measurements.md
test -s docs/release/dependency-review.md
test -s docs/release/known-limitations.md
test -s security_best_practices_report.md

grep -q 'clean Apple Silicon Mac' docs/release/known-limitations.md
grep -q 'No critical or high severity findings' security_best_practices_report.md
grep -q 'v0.1.0 Release Candidate Checklist' docs/release/v0.1.0-checklist.md

echo "Phase 7 checks passed"
