# Domain Model

## Project

A project represents a repository or Compose project that owns services. Identity should be derived from explicit configuration first, then stable local evidence such as Git root and Compose labels.

Fields:

- `id`
- `name`
- `root_path`
- `source`
- `created_at`
- `updated_at`

## Service

A service is a routable logical unit within a project.

Fields:

- `id`
- `project_id`
- `name`
- `protocol`: `http` or `tcp`
- `description`
- `source`

## Observation

An observation is read-only evidence from a scan.

Fields:

- `id`
- `scan_run_id`
- `project_id`
- `service_id`
- `runtime`
- `protocol`
- `backend_host`
- `backend_port`
- `evidence`
- `confidence`: `high`, `medium`, or `low`

## Route Suggestion

A suggestion is a proposed stable route that requires approval unless it matches an existing approved policy.

Fields:

- `id`
- `service_id`
- `observation_id`
- `route_host`
- `frontend_port`
- `backend_host`
- `backend_port`
- `state`
- `reason`

## Confirmed Route

A confirmed route is a user-approved stable frontend mapped to a service.

States:

- `suggested`
- `awaiting_approval`
- `active`
- `inactive`
- `backend_unavailable`
- `stale`
- `conflict`
- `invalid`
- `disabled`
- `error`

## TCP Port Assignment

TCP frontend ports are stable route resources. Assignment must be deterministic where practical and persisted before listener activation.

Rules:

- Preserve existing assignments.
- Prefer requested port if free.
- Otherwise choose a free port from the configured pool.
- Require approval before changing a confirmed stable frontend port.

## Diagnostic Event

Diagnostic events are sanitized local records for failures, doctor results, setup events, and routing errors. They must not contain raw environment values or credentials.
