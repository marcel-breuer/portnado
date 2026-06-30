#!/bin/sh
set -eu

required_files="
AGENTS.md
docs/product/prd.md
docs/product/requirements-traceability.md
docs/architecture/overview.md
docs/architecture/component-boundaries.md
docs/architecture/domain-model.md
docs/architecture/ipc-protocol.md
docs/architecture/configuration-schema.md
docs/architecture/sqlite-schema.md
docs/security/threat-model.md
docs/security/privileged-setup.md
docs/testing/testing-strategy.md
docs/release/release-strategy.md
docs/backlog/phased-backlog.md
docs/research/phase-1-technical-spike.md
docs/research/name-collision.md
configs/schema/portnado.schema.json
"

for file in $required_files; do
  test -s "$file"
done

python -m json.tool configs/schema/portnado.schema.json >/dev/null

for adr in docs/adr/*.md; do
  grep -q "^# ADR" "$adr"
  grep -q "^## Status" "$adr"
  grep -q "^## Decision" "$adr"
done

grep -q "FR-001" docs/product/requirements-traceability.md
grep -q "SEC-001" docs/security/threat-model.md

echo "Phase 1 checks passed"
