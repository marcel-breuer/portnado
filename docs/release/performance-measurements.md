# Performance Measurements

Measured on Apple M1, Darwin arm64, during Phase 7.

## Commands

```bash
go test -bench=. -benchmem ./internal/config ./internal/routing/http
```

## Results

| Benchmark | Time | Memory | Allocations |
| --- | ---: | ---: | ---: |
| `BenchmarkParseRepository-8` | `31572 ns/op` | `21982 B/op` | `338 allocs/op` |
| `BenchmarkProxyServeHTTP-8` | `51561 ns/op` | `45857 B/op` | `103 allocs/op` |

## Interpretation

Repository parsing is comfortably small for interactive CLI scans. The proxy
benchmark includes an in-process backend and request recorder overhead, so it is
best used as a regression baseline rather than an absolute production throughput
claim.

## Memory And CPU Notes

Go benchmark memory allocation data is recorded above.

A short daemon profile was run from the packaged daemon with a temporary HOME:

```text
PID    RSS  %CPU ELAPSED COMMAND
91423  20128   0.0   00:10 dist/Portnado.app/Contents/Resources/bin/portnado-daemon
```

The daemon responded to `portnado status`, and `portnado doctor --json` reported
passing platform, `.localhost`, control socket, SQLite, HTTP proxy, and Docker
checks. LaunchAgent was expected to warn because launch at login was not
installed in the temporary HOME.

This is a short smoke/profile only. Sustained multi-hour menu bar and daemon
resource usage remains a release-candidate follow-up.
