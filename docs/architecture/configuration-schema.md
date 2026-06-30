# Configuration Schema Proposal

## Repository File

Path:

```text
.portnado.yml
```

The file is safe to commit and must not contain secrets, process IDs, or user-specific absolute paths.

## Local Override File

Path:

```text
~/Library/Application Support/Portnado/projects/<project-id>.yml
```

Overrides may contain machine-specific preferences such as enablement state, selected process, local project path, scan exclusions, and preferred TCP frontend ports.

## Precedence

Highest to lowest:

1. explicit CLI argument for the current command,
2. local project override,
3. repository `.portnado.yml`,
4. confirmed persisted user choice,
5. automatic discovery,
6. application defaults.

## Example

```yaml
version: 1

project:
  name: webguard

services:
  app:
    protocol: http
    description: Frontend application
    route:
      host: app.webguard.localhost
    target:
      discovery: auto
      preferredPort: 5173

  database:
    protocol: tcp
    route:
      host: db.webguard.localhost
      preferredPort: 15432
    target:
      discovery: docker-compose
      service: database
      containerPort: 5432
```

## Validation Rules

- `version` must be `1`.
- Project names must normalize to safe DNS labels.
- Service keys must be lowercase stable identifiers.
- Route hosts must end with `.localhost`.
- Route hosts must not contain control characters, slashes, empty labels, wildcard labels, or non-ASCII DNS labels in the MVP.
- TCP services must define or accept a stable frontend port.
- Unknown top-level keys are rejected.
- Secret-looking values in unsupported locations are rejected with an actionable error.

## Environment Interpolation

Portnado may read local environment files to resolve Compose-style port interpolation. It must not persist raw files or log resolved secret-looking values.
