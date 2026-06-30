package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/marcel-breuer/portnado/internal/domain"
	"github.com/marcel-breuer/portnado/internal/persistence"
)

type fakeDetector struct {
	observations []domain.Observation
	warnings     []string
}

func (f fakeDetector) Discover(context.Context) ([]domain.Observation, []string) {
	return f.observations, f.warnings
}

func TestScanPersistsSuggestionsWithRepositoryRoute(t *testing.T) {
	ctx := context.Background()
	t.Setenv("HOME", t.TempDir())
	store, err := persistence.OpenPath(ctx, ":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".portnado.yml"), `
version: 1
project:
  name: webguard
services:
  app:
    protocol: http
    route:
      host: app.webguard.localhost
    target:
      discovery: auto
`)
	now := time.Unix(0, 0).UTC()
	project := domain.Project{ID: "project_native", Name: "webguard", Source: "native", CreatedAt: now, UpdatedAt: now}
	service := domain.Service{ID: "service_app", ProjectID: project.ID, Name: "app", Protocol: domain.ProtocolHTTP, Source: "native", CreatedAt: now, UpdatedAt: now}
	appService := NewService(store, fakeDetector{observations: []domain.Observation{{
		ID:          "observation_app",
		Project:     project,
		Service:     service,
		Runtime:     "node",
		Protocol:    domain.ProtocolHTTP,
		BackendHost: "127.0.0.1",
		BackendPort: 5173,
		Confidence:  domain.ConfidenceMedium,
		CreatedAt:   now,
	}}})

	result, err := appService.Scan(ctx, root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(result.Suggestions) != 1 {
		t.Fatalf("suggestions = %d", len(result.Suggestions))
	}
	if result.Suggestions[0].RouteHost != "app.webguard.localhost" {
		t.Fatalf("route host = %q", result.Suggestions[0].RouteHost)
	}

	summaries, err := appService.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(summaries) != 1 || summaries[0].BackendPort != 5173 {
		t.Fatalf("summaries = %+v", summaries)
	}
}

func TestScanAppliesLocalOverride(t *testing.T) {
	ctx := context.Background()
	home := t.TempDir()
	t.Setenv("HOME", home)
	store, err := persistence.OpenPath(ctx, ":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".portnado.yml"), `
version: 1
project:
  name: webguard
services:
  app:
    protocol: http
    route:
      host: app.webguard.localhost
    target:
      discovery: auto
`)
	projectID := domain.DeterministicID("project", "repository", root, "webguard")
	overridePath := filepath.Join(home, "Library", "Application Support", "Portnado", "projects", projectID+".yml")
	if err := os.MkdirAll(filepath.Dir(overridePath), 0o700); err != nil {
		t.Fatalf("create override dir: %v", err)
	}
	writeFile(t, overridePath, `
version: 1
services:
  app:
    route:
      host: frontend.webguard.localhost
`)

	now := time.Unix(0, 0).UTC()
	project := domain.Project{ID: "project_native", Name: "webguard", Source: "native", CreatedAt: now, UpdatedAt: now}
	service := domain.Service{ID: "service_app", ProjectID: project.ID, Name: "app", Protocol: domain.ProtocolHTTP, Source: "native", CreatedAt: now, UpdatedAt: now}
	appService := NewService(store, fakeDetector{observations: []domain.Observation{{
		ID:          "observation_app",
		Project:     project,
		Service:     service,
		Runtime:     "node",
		Protocol:    domain.ProtocolHTTP,
		BackendHost: "127.0.0.1",
		BackendPort: 5173,
		Confidence:  domain.ConfidenceMedium,
		CreatedAt:   now,
	}}})

	result, err := appService.Scan(ctx, root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if result.Suggestions[0].RouteHost != "frontend.webguard.localhost" {
		t.Fatalf("route host = %q", result.Suggestions[0].RouteHost)
	}
}

func writeFile(t *testing.T, path string, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
