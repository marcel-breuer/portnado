package persistence

type migration struct {
	version int
	name    string
	sql     string
}

var migrations = []migration{
	{
		version: 1,
		name:    "initial_schema",
		sql: `
CREATE TABLE IF NOT EXISTS schema_migrations (
  version INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  applied_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS projects (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  root_path TEXT,
  source TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS services (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  protocol TEXT NOT NULL CHECK (protocol IN ('http', 'tcp')),
  description TEXT,
  source TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(project_id, name)
);

CREATE TABLE IF NOT EXISTS scan_runs (
  id TEXT PRIMARY KEY,
  started_at TEXT NOT NULL,
  finished_at TEXT,
  status TEXT NOT NULL,
  summary_json TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS observations (
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
);

CREATE TABLE IF NOT EXISTS route_suggestions (
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
);

CREATE TABLE IF NOT EXISTS confirmed_routes (
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
);

CREATE TABLE IF NOT EXISTS tcp_port_assignments (
  id TEXT PRIMARY KEY,
  route_id TEXT NOT NULL REFERENCES confirmed_routes(id) ON DELETE CASCADE,
  port INTEGER NOT NULL UNIQUE,
  source TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(route_id)
);

CREATE TABLE IF NOT EXISTS project_roots (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  root_path TEXT NOT NULL,
  source TEXT NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(project_id, root_path)
);

CREATE TABLE IF NOT EXISTS settings (
  key TEXT PRIMARY KEY,
  value_json TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS diagnostic_events (
  id TEXT PRIMARY KEY,
  level TEXT NOT NULL,
  component TEXT NOT NULL,
  event_name TEXT NOT NULL,
  message TEXT NOT NULL,
  context_json TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_observations_scan_run_id ON observations(scan_run_id);
CREATE INDEX IF NOT EXISTS idx_observations_project_id ON observations(project_id);
CREATE INDEX IF NOT EXISTS idx_services_project_name ON services(project_id, name);
CREATE INDEX IF NOT EXISTS idx_route_suggestions_service_id ON route_suggestions(service_id);
CREATE INDEX IF NOT EXISTS idx_confirmed_routes_state ON confirmed_routes(state);
CREATE INDEX IF NOT EXISTS idx_diagnostic_events_created_at ON diagnostic_events(created_at);
`,
	},
}
