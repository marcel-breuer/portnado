package persistence

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/marcel-breuer/portnado/internal/domain"
)

func TestMigrateIsIdempotent(t *testing.T) {
	ctx := context.Background()
	store, err := OpenPath(ctx, ":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
}

func TestMigrationsPersistAcrossReopen(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "portnado.db")
	store, err := OpenPath(ctx, dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("database was not created: %v", err)
	}
	reopened, err := OpenPath(ctx, dbPath)
	if err != nil {
		t.Fatalf("reopen migrated store: %v", err)
	}
	defer reopened.Close()
	if err := reopened.Migrate(ctx); err != nil {
		t.Fatalf("migrate reopened store: %v", err)
	}
}

func TestSaveScanResultAndListServices(t *testing.T) {
	ctx := context.Background()
	store, err := OpenPath(ctx, ":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	now := time.Unix(0, 0).UTC()
	project := domain.Project{ID: "project_webguard", Name: "webguard", Source: "test", CreatedAt: now, UpdatedAt: now}
	service := domain.Service{ID: "service_app", ProjectID: project.ID, Name: "app", Protocol: domain.ProtocolHTTP, Source: "test", CreatedAt: now, UpdatedAt: now}
	observation := domain.Observation{
		ID:          "observation_app",
		Project:     project,
		Service:     service,
		Runtime:     "node",
		Protocol:    domain.ProtocolHTTP,
		BackendHost: "127.0.0.1",
		BackendPort: 5173,
		Confidence:  domain.ConfidenceHigh,
		Evidence:    []domain.Evidence{{Source: "test", Summary: "fixture"}},
		CreatedAt:   now,
	}
	suggestion := domain.RouteSuggestion{
		ID:            "suggestion_app",
		ServiceID:     service.ID,
		ObservationID: observation.ID,
		RouteHost:     "app.webguard.localhost",
		BackendHost:   "127.0.0.1",
		BackendPort:   5173,
		State:         domain.RouteStateAwaitingApproval,
		Reason:        "test",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	result := domain.ScanResult{
		Run: domain.ScanRun{
			ID:         "scan_fixture",
			StartedAt:  now,
			FinishedAt: now,
			Status:     "completed",
		},
		Projects:     []domain.Project{project},
		Services:     []domain.Service{service},
		Observations: []domain.Observation{observation},
		Suggestions:  []domain.RouteSuggestion{suggestion},
	}

	if err := store.SaveScanResult(ctx, result); err != nil {
		t.Fatalf("save scan result: %v", err)
	}
	summaries, err := store.ListServices(ctx)
	if err != nil {
		t.Fatalf("list services: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("summaries = %d", len(summaries))
	}
	if summaries[0].RouteHost != "app.webguard.localhost" || summaries[0].BackendPort != 5173 {
		t.Fatalf("summary = %+v", summaries[0])
	}
}

func TestApproveSuggestionCreatesActiveRoute(t *testing.T) {
	ctx := context.Background()
	store, err := OpenPath(ctx, ":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	fixture := saveRouteFixture(t, ctx, store, "127.0.0.1")
	route, err := store.ApproveSuggestion(ctx, fixture.suggestionID)
	if err != nil {
		t.Fatalf("approve suggestion: %v", err)
	}
	if route.State != domain.RouteStateActive {
		t.Fatalf("state = %q", route.State)
	}
	active, err := store.ActiveRoutes(ctx)
	if err != nil {
		t.Fatalf("active routes: %v", err)
	}
	if len(active) != 1 || active[0].RouteHost != "app.webguard.localhost" {
		t.Fatalf("active routes = %+v", active)
	}
	disabled, err := store.SetRouteState(ctx, route.ID, domain.RouteStateDisabled)
	if err != nil {
		t.Fatalf("disable route: %v", err)
	}
	if disabled.State != domain.RouteStateDisabled {
		t.Fatalf("disabled state = %q", disabled.State)
	}
}

func TestRouteRecoveryAfterStoreReopen(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "portnado.db")
	store, err := OpenPath(ctx, dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	fixture := saveRouteFixture(t, ctx, store, "127.0.0.1")
	route, err := store.ApproveSuggestion(ctx, fixture.suggestionID)
	if err != nil {
		t.Fatalf("approve suggestion: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}

	reopened, err := OpenPath(ctx, dbPath)
	if err != nil {
		t.Fatalf("reopen store: %v", err)
	}
	defer reopened.Close()
	active, err := reopened.ActiveRoutes(ctx)
	if err != nil {
		t.Fatalf("active routes after reopen: %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("active routes = %d", len(active))
	}
	if active[0].ID != route.ID || active[0].RouteHost != "app.webguard.localhost" {
		t.Fatalf("recovered route = %+v", active[0])
	}
}

func TestApproveSuggestionRejectsRemoteBackend(t *testing.T) {
	ctx := context.Background()
	store, err := OpenPath(ctx, ":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	fixture := saveRouteFixture(t, ctx, store, "192.168.1.10")
	if _, err := store.ApproveSuggestion(ctx, fixture.suggestionID); err == nil {
		t.Fatal("expected remote backend rejection")
	}
}

type routeFixture struct {
	suggestionID string
}

func saveRouteFixture(t *testing.T, ctx context.Context, store *Store, backendHost string) routeFixture {
	t.Helper()
	now := time.Unix(0, 0).UTC()
	project := domain.Project{ID: "project_webguard", Name: "webguard", Source: "test", CreatedAt: now, UpdatedAt: now}
	service := domain.Service{ID: "service_app", ProjectID: project.ID, Name: "app", Protocol: domain.ProtocolHTTP, Source: "test", CreatedAt: now, UpdatedAt: now}
	observation := domain.Observation{
		ID:          "observation_app_" + backendHost,
		Project:     project,
		Service:     service,
		Runtime:     "node",
		Protocol:    domain.ProtocolHTTP,
		BackendHost: backendHost,
		BackendPort: 5173,
		Confidence:  domain.ConfidenceHigh,
		Evidence:    []domain.Evidence{{Source: "test", Summary: "fixture"}},
		CreatedAt:   now,
	}
	suggestion := domain.RouteSuggestion{
		ID:            "suggestion_app_" + backendHost,
		ServiceID:     service.ID,
		ObservationID: observation.ID,
		RouteHost:     "app.webguard.localhost",
		BackendHost:   backendHost,
		BackendPort:   5173,
		State:         domain.RouteStateAwaitingApproval,
		Reason:        "test",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := store.SaveScanResult(ctx, domain.ScanResult{
		Run:          domain.ScanRun{ID: "scan_" + backendHost, StartedAt: now, FinishedAt: now, Status: "completed"},
		Projects:     []domain.Project{project},
		Services:     []domain.Service{service},
		Observations: []domain.Observation{observation},
		Suggestions:  []domain.RouteSuggestion{suggestion},
	}); err != nil {
		t.Fatalf("save route fixture: %v", err)
	}
	return routeFixture{suggestionID: suggestion.ID}
}
