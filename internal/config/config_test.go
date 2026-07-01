package config

import (
	"strings"
	"testing"

	"github.com/marcel-breuer/portnado/internal/domain"
)

func TestParseRepositoryValidatesStrictSchema(t *testing.T) {
	_, err := ParseRepository([]byte(`
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
unexpected: true
`))
	if err == nil {
		t.Fatal("expected unknown field error")
	}
}

func TestParseRepositoryValid(t *testing.T) {
	cfg, err := ParseRepository([]byte(`
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
      preferredPort: 5173
`))
	if err != nil {
		t.Fatalf("parse repository: %v", err)
	}
	if cfg.Services["app"].Protocol != domain.ProtocolHTTP {
		t.Fatalf("protocol = %q", cfg.Services["app"].Protocol)
	}
}

func TestValidateLocalhostRejectsNonLocalhost(t *testing.T) {
	err := ValidateLocalhost("app.example.com")
	if err == nil || !strings.Contains(err.Error(), ".localhost") {
		t.Fatalf("error = %v, want .localhost rejection", err)
	}
}

func TestEffectiveServicesLocalOverrideWins(t *testing.T) {
	repo, err := ParseRepository([]byte(`
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
`))
	if err != nil {
		t.Fatalf("parse repository: %v", err)
	}
	disabled := false
	override := LocalOverride{
		Version: 1,
		Services: map[string]LocalServiceOverride{
			"app": {
				Enabled: &disabled,
				Route: &RouteConfig{
					Host: "frontend.webguard.localhost",
				},
			},
		},
	}
	services, err := EffectiveServices(repo, override)
	if err != nil {
		t.Fatalf("effective services: %v", err)
	}
	if services[0].Enabled {
		t.Fatal("expected local override to disable service")
	}
	if services[0].RouteHost != "frontend.webguard.localhost" {
		t.Fatalf("route host = %q", services[0].RouteHost)
	}
	if services[0].Sources["route"] != "local_override" {
		t.Fatalf("route source = %q", services[0].Sources["route"])
	}
}

func TestNewRepositoryConfigRendersValidDefault(t *testing.T) {
	cfg, err := NewRepositoryConfig(InitOptions{ProjectName: "WebGuard", ServiceName: "App", TargetPort: 5173})
	if err != nil {
		t.Fatalf("new repository config: %v", err)
	}
	data, err := RenderRepository(cfg)
	if err != nil {
		t.Fatalf("render repository config: %v", err)
	}
	parsed, err := ParseRepository(data)
	if err != nil {
		t.Fatalf("parse rendered config: %v\n%s", err, data)
	}
	if parsed.Project.Name != "webguard" {
		t.Fatalf("project name = %q", parsed.Project.Name)
	}
	service := parsed.Services["app"]
	if service.Route.Host != "app.webguard.localhost" || service.Target.PreferredPort != 5173 {
		t.Fatalf("service = %+v", service)
	}
}
