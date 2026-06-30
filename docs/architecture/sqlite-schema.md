# SQLite Schema Proposal

## Database Location

```text
~/Library/Application Support/Portnado/portnado.db
```

The database file and parent directory must be current-user only. The daemon enables foreign keys and uses WAL mode when compatible with local filesystem semantics.

## Migration Rules

- Migrations are versioned and monotonic.
- Migrations run inside transactions where SQLite permits it.
- Destructive migrations require a backup first.
- Migration tests must cover clean install, incremental upgrade, rollback on failure, and foreign key behavior.

## Proposed Tables

### schema_migrations

Tracks applied migrations.

```sql
version INTEGER PRIMARY KEY,
name TEXT NOT NULL,
applied_at TEXT NOT NULL
```

### projects

```sql
id TEXT PRIMARY KEY,
name TEXT NOT NULL,
root_path TEXT,
source TEXT NOT NULL,
created_at TEXT NOT NULL,
updated_at TEXT NOT NULL
```

### services

```sql
id TEXT PRIMARY KEY,
project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
name TEXT NOT NULL,
protocol TEXT NOT NULL CHECK (protocol IN ('http', 'tcp')),
description TEXT,
source TEXT NOT NULL,
created_at TEXT NOT NULL,
updated_at TEXT NOT NULL,
UNIQUE(project_id, name)
```

### scan_runs

```sql
id TEXT PRIMARY KEY,
started_at TEXT NOT NULL,
finished_at TEXT,
status TEXT NOT NULL,
summary_json TEXT NOT NULL
```

### observations

```sql
id TEXT PRIMARY KEY,
scan_run_id TEXT NOT NULL REFERENCES scan_runs(id) ON DELETE CASCADE,
project_id TEXT REFERENCES projects(id) ON DELETE SET NULL,
service_id TEXT REFERENCES services(id) ON DELETE SET NULL,
runtime TEXT,
protocol TEXT NOT NULL,
backend_host TEXT NOT NULL,
backend_port INTEGER NOT NULL,
confidence TEXT NOT NULL CHECK (confidence IN ('high', 'medium', 'low')),
evidence_json TEXT NOT NULL,
created_at TEXT NOT NULL
```

### route_suggestions

```sql
id TEXT PRIMARY KEY,
service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
observation_id TEXT REFERENCES observations(id) ON DELETE SET NULL,
route_host TEXT NOT NULL,
frontend_port INTEGER,
backend_host TEXT NOT NULL,
backend_port INTEGER NOT NULL,
state TEXT NOT NULL,
reason TEXT NOT NULL,
created_at TEXT NOT NULL,
updated_at TEXT NOT NULL
```

### confirmed_routes

```sql
id TEXT PRIMARY KEY,
service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
route_host TEXT NOT NULL,
frontend_port INTEGER,
backend_host TEXT,
backend_port INTEGER,
state TEXT NOT NULL,
approval_policy_json TEXT NOT NULL,
last_observation_id TEXT REFERENCES observations(id) ON DELETE SET NULL,
created_at TEXT NOT NULL,
updated_at TEXT NOT NULL,
UNIQUE(route_host),
UNIQUE(frontend_port)
```

### tcp_port_assignments

```sql
id TEXT PRIMARY KEY,
route_id TEXT NOT NULL REFERENCES confirmed_routes(id) ON DELETE CASCADE,
port INTEGER NOT NULL UNIQUE,
source TEXT NOT NULL,
created_at TEXT NOT NULL,
updated_at TEXT NOT NULL
```

### settings

```sql
key TEXT PRIMARY KEY,
value_json TEXT NOT NULL,
updated_at TEXT NOT NULL
```

### diagnostic_events

```sql
id TEXT PRIMARY KEY,
level TEXT NOT NULL,
component TEXT NOT NULL,
event_name TEXT NOT NULL,
message TEXT NOT NULL,
context_json TEXT NOT NULL,
created_at TEXT NOT NULL
```

## Indexes

- `observations(scan_run_id)`
- `observations(project_id)`
- `services(project_id, name)`
- `confirmed_routes(state)`
- `diagnostic_events(created_at)`

## Open Decisions for Phase 3

- Final SQLite driver choice and CGO policy.
- JSON storage shape for evidence and approval policies.
- Retention policy for observations and diagnostic events.
