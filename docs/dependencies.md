# Dependencies

Portnado prefers the Go and Swift standard libraries. Runtime dependencies require clear value, active maintenance, and compatible licensing.

## Go

| Dependency | Purpose | License | Maintenance | Notes |
| --- | --- | --- | --- | --- |
| `modernc.org/sqlite` | Pure-Go SQLite driver for local persistence and Docker-friendly tests. | BSD-style | Actively maintained public module. | Avoids requiring CGO toolchains in Docker checks. Raises module floor to Go 1.25. |
| `gopkg.in/yaml.v3` | YAML parsing for `.portnado.yml` and local overrides. | Apache-2.0/MIT compatible | Mature public module. | Used with `KnownFields(true)` for strict schema handling. |

Transitive dependencies are tracked in `go.sum` and reviewed through Dependabot.
