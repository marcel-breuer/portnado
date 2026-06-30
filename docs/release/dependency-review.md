# Dependency Review

## Summary

Phase 7 dependency review passed for the current module set.

## Commands

```bash
go mod verify
go list -m all
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

## Results

- `go mod verify`: all modules verified.
- `govulncheck`: no vulnerabilities found.
- CI runs govulncheck on every push and pull request.
- Homebrew Cask style and strict audit pass when the Cask is copied into a
  temporary local tap.

## Direct Runtime Dependencies

- `gopkg.in/yaml.v3 v3.0.1`
- `modernc.org/sqlite v1.53.0`

## Review Notes

The dependency graph includes transitive packages from `modernc.org/sqlite` and
the Go vulnerability scanner toolchain. No module authenticity bypass settings
are configured in repository scripts, Dockerfiles, or workflows.
