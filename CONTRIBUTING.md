# Contributing

Portnado is implemented in controlled phases. Keep changes focused on the approved phase and update traceability when requirements move from planned to implemented or verified.

## Development Rules

- Inspect existing code and Docker configuration before editing.
- Use Docker or Docker Compose for checks whenever technically possible.
- Use host macOS tooling only for SwiftUI and macOS-specific behavior that cannot run in Docker.
- Run targeted tests first, then the full relevant suite before completion.
- Use Conventional Commits.
- Do not include AI-tool attribution in commits, PRs, changelog entries, source comments, or docs.

## Checks

```bash
make phase7-hardening-check
make swift-build
make swift-test
```

## Security

Do not add listeners outside `127.0.0.1`, bypass Gatekeeper, log secrets, or introduce privileged system changes without an approved design and explicit confirmation.
