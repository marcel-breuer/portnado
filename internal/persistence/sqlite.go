package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/marcel-breuer/portnado/internal/domain"
	"github.com/marcel-breuer/portnado/internal/paths"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(ctx context.Context) (*Store, error) {
	dbPath, err := paths.DatabasePath()
	if err != nil {
		return nil, err
	}
	return OpenPath(ctx, dbPath)
}

func OpenPath(ctx context.Context, dbPath string) (*Store, error) {
	if dbPath != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}
	store := &Store{db: db}
	if err := store.configure(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.Migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) configure(ctx context.Context) error {
	statements := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 5000",
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("configure sqlite %q: %w", statement, err)
		}
	}
	return nil
}

func (s *Store) Migrate(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migrations: %w", err)
	}
	defer tx.Rollback()

	for _, migration := range migrations {
		var exists int
		if migration.version > 1 {
			if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE version = ?", migration.version).Scan(&exists); err != nil {
				return fmt.Errorf("check migration %d: %w", migration.version, err)
			}
			if exists > 0 {
				continue
			}
		}
		if _, err := tx.ExecContext(ctx, migration.sql); err != nil {
			return fmt.Errorf("apply migration %d %s: %w", migration.version, migration.name, err)
		}
		if _, err := tx.ExecContext(ctx, "INSERT OR IGNORE INTO schema_migrations(version, name, applied_at) VALUES (?, ?, ?)", migration.version, migration.name, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
			return fmt.Errorf("record migration %d: %w", migration.version, err)
		}
	}

	return tx.Commit()
}

func (s *Store) SaveScanResult(ctx context.Context, result domain.ScanResult) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin scan save: %w", err)
	}
	defer tx.Rollback()

	run := result.Run
	if run.ID == "" {
		run.ID = domain.DeterministicID("scan", run.StartedAt.Format(time.RFC3339Nano))
	}
	summary, err := json.Marshal(map[string]int{"observations": len(result.Observations), "suggestions": len(result.Suggestions)})
	if err != nil {
		return fmt.Errorf("encode scan summary: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO scan_runs(id, started_at, finished_at, status, summary_json)
VALUES (?, ?, ?, ?, ?)`, run.ID, formatTime(run.StartedAt), formatTime(run.FinishedAt), run.Status, string(summary)); err != nil {
		return fmt.Errorf("insert scan run: %w", err)
	}

	for _, project := range result.Projects {
		if err := upsertProject(ctx, tx, project); err != nil {
			return err
		}
	}
	for _, service := range result.Services {
		if err := upsertService(ctx, tx, service); err != nil {
			return err
		}
	}
	for _, observation := range result.Observations {
		if err := insertObservation(ctx, tx, run.ID, observation); err != nil {
			return err
		}
	}
	for _, suggestion := range result.Suggestions {
		if err := upsertSuggestion(ctx, tx, suggestion); err != nil {
			return err
		}
	}
	if err := reconcileConfirmedRoutes(ctx, tx, result); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) ListServices(ctx context.Context) ([]domain.ServiceSummary, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT rs.id, p.name, svc.name, svc.protocol, rs.route_host, COALESCE(rs.frontend_port, 0),
       rs.backend_host, rs.backend_port, rs.state, obs.confidence, svc.source
FROM route_suggestions rs
JOIN services svc ON svc.id = rs.service_id
JOIN projects p ON p.id = svc.project_id
LEFT JOIN observations obs ON obs.id = rs.observation_id
ORDER BY p.name, svc.name, rs.created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	defer rows.Close()

	var summaries []domain.ServiceSummary
	for rows.Next() {
		var summary domain.ServiceSummary
		if err := rows.Scan(&summary.RouteID, &summary.ProjectName, &summary.ServiceName, &summary.Protocol, &summary.RouteHost, &summary.FrontendPort, &summary.BackendHost, &summary.BackendPort, &summary.State, &summary.Confidence, &summary.Source); err != nil {
			return nil, fmt.Errorf("scan service summary: %w", err)
		}
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate service summaries: %w", err)
	}
	return summaries, nil
}

func (s *Store) ApproveSuggestion(ctx context.Context, suggestionID string) (domain.ConfirmedRoute, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.ConfirmedRoute{}, fmt.Errorf("begin route approval: %w", err)
	}
	defer tx.Rollback()

	route, err := selectSuggestionRoute(ctx, tx, suggestionID)
	if err != nil {
		return domain.ConfirmedRoute{}, err
	}
	route.ID = domain.DeterministicID("route", route.ServiceID, route.RouteHost)
	route.State = domain.RouteStateActive
	now := time.Now().UTC()
	route.CreatedAt = now
	route.UpdatedAt = now
	if err := validateRouteConflict(ctx, tx, route); err != nil {
		return domain.ConfirmedRoute{}, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO confirmed_routes(id, service_id, route_host, frontend_port, backend_host, backend_port, state, approval_policy_json, last_observation_id, created_at, updated_at)
SELECT ?, service_id, route_host, frontend_port, backend_host, backend_port, ?, '{}', observation_id, ?, ?
FROM route_suggestions
WHERE id = ?
ON CONFLICT(id) DO UPDATE SET backend_host = excluded.backend_host, backend_port = excluded.backend_port, state = excluded.state, updated_at = excluded.updated_at`,
		route.ID, route.State, formatTime(now), formatTime(now), suggestionID); err != nil {
		return domain.ConfirmedRoute{}, fmt.Errorf("insert confirmed route: %w", err)
	}
	if route.Protocol == domain.ProtocolTCP && route.FrontendPort != 0 {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO tcp_port_assignments(id, route_id, port, source, created_at, updated_at)
VALUES (?, ?, ?, 'suggestion', ?, ?)
ON CONFLICT(route_id) DO UPDATE SET port = excluded.port, updated_at = excluded.updated_at`,
			domain.DeterministicID("tcpport", route.ID), route.ID, route.FrontendPort, formatTime(now), formatTime(now)); err != nil {
			return domain.ConfirmedRoute{}, fmt.Errorf("persist tcp frontend port: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, "UPDATE route_suggestions SET state = ?, updated_at = ? WHERE id = ?", domain.RouteStateActive, formatTime(now), suggestionID); err != nil {
		return domain.ConfirmedRoute{}, fmt.Errorf("mark suggestion active: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.ConfirmedRoute{}, fmt.Errorf("commit route approval: %w", err)
	}
	route.CreatedAt = now
	route.UpdatedAt = now
	return route, nil
}

func (s *Store) SetRouteState(ctx context.Context, routeID string, state domain.RouteState) (domain.ConfirmedRoute, error) {
	if state != domain.RouteStateActive && state != domain.RouteStateDisabled {
		return domain.ConfirmedRoute{}, fmt.Errorf("unsupported route state %q", state)
	}
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx, "UPDATE confirmed_routes SET state = ?, updated_at = ? WHERE id = ?", state, formatTime(now), routeID)
	if err != nil {
		return domain.ConfirmedRoute{}, fmt.Errorf("update route state: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return domain.ConfirmedRoute{}, sql.ErrNoRows
	}
	return s.GetRoute(ctx, routeID)
}

func (s *Store) GetRoute(ctx context.Context, routeID string) (domain.ConfirmedRoute, error) {
	rows, err := s.selectRoutes(ctx, "WHERE cr.id = ?", routeID)
	if err != nil {
		return domain.ConfirmedRoute{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return domain.ConfirmedRoute{}, sql.ErrNoRows
	}
	route, err := scanConfirmedRoute(rows)
	if err != nil {
		return domain.ConfirmedRoute{}, err
	}
	return route, rows.Err()
}

func (s *Store) ActiveRoutes(ctx context.Context) ([]domain.ConfirmedRoute, error) {
	return s.routes(ctx, "WHERE cr.state = ?", domain.RouteStateActive)
}

func (s *Store) Routes(ctx context.Context) ([]domain.ConfirmedRoute, error) {
	return s.routes(ctx, "")
}

func (s *Store) routes(ctx context.Context, where string, args ...any) ([]domain.ConfirmedRoute, error) {
	rows, err := s.selectRoutes(ctx, where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var routes []domain.ConfirmedRoute
	for rows.Next() {
		route, err := scanConfirmedRoute(rows)
		if err != nil {
			return nil, err
		}
		routes = append(routes, route)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return routes, nil
}

func (s *Store) selectRoutes(ctx context.Context, where string, args ...any) (*sql.Rows, error) {
	query := `
SELECT cr.id, cr.service_id, svc.name, p.name, svc.protocol, cr.route_host, COALESCE(cr.frontend_port, 0),
       COALESCE(cr.backend_host, ''), COALESCE(cr.backend_port, 0), cr.state, cr.created_at, cr.updated_at
FROM confirmed_routes cr
JOIN services svc ON svc.id = cr.service_id
JOIN projects p ON p.id = svc.project_id
` + where + `
ORDER BY p.name, svc.name`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select routes: %w", err)
	}
	return rows, nil
}

func scanConfirmedRoute(scanner interface {
	Scan(dest ...any) error
}) (domain.ConfirmedRoute, error) {
	var route domain.ConfirmedRoute
	var createdAt, updatedAt string
	if err := scanner.Scan(&route.ID, &route.ServiceID, &route.ServiceName, &route.ProjectName, &route.Protocol, &route.RouteHost, &route.FrontendPort, &route.BackendHost, &route.BackendPort, &route.State, &createdAt, &updatedAt); err != nil {
		return route, fmt.Errorf("scan confirmed route: %w", err)
	}
	route.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	route.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return route, nil
}

func selectSuggestionRoute(ctx context.Context, tx *sql.Tx, suggestionID string) (domain.ConfirmedRoute, error) {
	var route domain.ConfirmedRoute
	err := tx.QueryRowContext(ctx, `
SELECT rs.service_id, svc.name, p.name, svc.protocol, rs.route_host, COALESCE(rs.frontend_port, 0), rs.backend_host, rs.backend_port
FROM route_suggestions rs
JOIN services svc ON svc.id = rs.service_id
JOIN projects p ON p.id = svc.project_id
WHERE rs.id = ?`, suggestionID).Scan(&route.ServiceID, &route.ServiceName, &route.ProjectName, &route.Protocol, &route.RouteHost, &route.FrontendPort, &route.BackendHost, &route.BackendPort)
	if err != nil {
		if err == sql.ErrNoRows {
			return route, fmt.Errorf("route suggestion %s was not found", suggestionID)
		}
		return route, fmt.Errorf("load route suggestion: %w", err)
	}
	return route, nil
}

func validateRouteConflict(ctx context.Context, tx *sql.Tx, route domain.ConfirmedRoute) error {
	ip := net.ParseIP(route.BackendHost)
	if ip == nil || !ip.IsLoopback() || ip.To4() == nil {
		return fmt.Errorf("backend target %s is not an IPv4 loopback address", route.BackendHost)
	}
	var count int
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM confirmed_routes WHERE route_host = ? AND id != ?", route.RouteHost, route.ID).Scan(&count); err != nil {
		return fmt.Errorf("check route host conflict: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("route host %s is already confirmed", route.RouteHost)
	}
	if route.Protocol == domain.ProtocolTCP && route.FrontendPort != 0 {
		if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM confirmed_routes WHERE frontend_port = ? AND id != ?", route.FrontendPort, route.ID).Scan(&count); err != nil {
			return fmt.Errorf("check frontend port conflict: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("frontend port %d is already confirmed", route.FrontendPort)
		}
	}
	return nil
}

func upsertProject(ctx context.Context, tx *sql.Tx, project domain.Project) error {
	_, err := tx.ExecContext(ctx, `
INSERT INTO projects(id, name, root_path, source, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET name = excluded.name, root_path = excluded.root_path, source = excluded.source, updated_at = excluded.updated_at`,
		project.ID, project.Name, project.RootPath, project.Source, formatTime(project.CreatedAt), formatTime(project.UpdatedAt))
	if err != nil {
		return fmt.Errorf("upsert project %s: %w", project.Name, err)
	}
	return nil
}

func upsertService(ctx context.Context, tx *sql.Tx, service domain.Service) error {
	_, err := tx.ExecContext(ctx, `
INSERT INTO services(id, project_id, name, protocol, description, source, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(project_id, name) DO UPDATE SET protocol = excluded.protocol, description = excluded.description, source = excluded.source, updated_at = excluded.updated_at`,
		service.ID, service.ProjectID, service.Name, service.Protocol, service.Description, service.Source, formatTime(service.CreatedAt), formatTime(service.UpdatedAt))
	if err != nil {
		return fmt.Errorf("upsert service %s: %w", service.Name, err)
	}
	return nil
}

func insertObservation(ctx context.Context, tx *sql.Tx, scanRunID string, observation domain.Observation) error {
	evidence, err := json.Marshal(observation.Evidence)
	if err != nil {
		return fmt.Errorf("encode observation evidence: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
INSERT OR REPLACE INTO observations(id, scan_run_id, project_id, service_id, runtime, protocol, backend_host, backend_port, confidence, evidence_json, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		observation.ID, scanRunID, observation.Project.ID, observation.Service.ID, observation.Runtime, observation.Protocol, observation.BackendHost, observation.BackendPort, observation.Confidence, string(evidence), formatTime(observation.CreatedAt))
	if err != nil {
		return fmt.Errorf("insert observation %s: %w", observation.ID, err)
	}
	return nil
}

func upsertSuggestion(ctx context.Context, tx *sql.Tx, suggestion domain.RouteSuggestion) error {
	_, err := tx.ExecContext(ctx, `
INSERT INTO route_suggestions(id, service_id, observation_id, route_host, frontend_port, backend_host, backend_port, state, reason, created_at, updated_at)
VALUES (?, ?, ?, ?, NULLIF(?, 0), ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET observation_id = excluded.observation_id, backend_host = excluded.backend_host, backend_port = excluded.backend_port, state = excluded.state, reason = excluded.reason, updated_at = excluded.updated_at`,
		suggestion.ID, suggestion.ServiceID, suggestion.ObservationID, suggestion.RouteHost, suggestion.FrontendPort, suggestion.BackendHost, suggestion.BackendPort, suggestion.State, suggestion.Reason, formatTime(suggestion.CreatedAt), formatTime(suggestion.UpdatedAt))
	if err != nil {
		return fmt.Errorf("upsert suggestion %s: %w", suggestion.ID, err)
	}
	return nil
}

func reconcileConfirmedRoutes(ctx context.Context, tx *sql.Tx, result domain.ScanResult) error {
	now := formatTime(time.Now().UTC())
	currentRoutes := make(map[string]domain.RouteSuggestion)
	for _, suggestion := range result.Suggestions {
		routeID := domain.DeterministicID("route", suggestion.ServiceID, suggestion.RouteHost)
		currentRoutes[routeID] = suggestion
	}
	for routeID, suggestion := range currentRoutes {
		if _, err := tx.ExecContext(ctx, `
UPDATE confirmed_routes
SET backend_host = ?, backend_port = ?, last_observation_id = ?, state = CASE WHEN state = ? THEN ? ELSE state END, updated_at = ?
WHERE id = ? AND state IN (?, ?)`,
			suggestion.BackendHost, suggestion.BackendPort, suggestion.ObservationID,
			domain.RouteStateStale, domain.RouteStateActive, now, routeID,
			domain.RouteStateActive, domain.RouteStateStale); err != nil {
			return fmt.Errorf("refresh confirmed route %s: %w", routeID, err)
		}
	}

	for _, project := range result.Projects {
		rows, err := tx.QueryContext(ctx, `
SELECT cr.id
FROM confirmed_routes cr
JOIN services svc ON svc.id = cr.service_id
WHERE svc.project_id = ? AND cr.state = ?`, project.ID, domain.RouteStateActive)
		if err != nil {
			return fmt.Errorf("select active routes for project %s: %w", project.ID, err)
		}
		var stale []string
		for rows.Next() {
			var routeID string
			if err := rows.Scan(&routeID); err != nil {
				_ = rows.Close()
				return fmt.Errorf("scan active route for project %s: %w", project.ID, err)
			}
			if _, ok := currentRoutes[routeID]; !ok {
				stale = append(stale, routeID)
			}
		}
		if err := rows.Close(); err != nil {
			return fmt.Errorf("close active route rows for project %s: %w", project.ID, err)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterate active routes for project %s: %w", project.ID, err)
		}
		for _, routeID := range stale {
			if _, err := tx.ExecContext(ctx, "UPDATE confirmed_routes SET state = ?, updated_at = ? WHERE id = ? AND state = ?", domain.RouteStateStale, now, routeID, domain.RouteStateActive); err != nil {
				return fmt.Errorf("mark confirmed route %s stale: %w", routeID, err)
			}
		}
	}
	return nil
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		value = time.Now().UTC()
	}
	return value.UTC().Format(time.RFC3339Nano)
}
