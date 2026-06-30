package docker

import (
	"context"
	"testing"
	"time"

	"github.com/marcel-breuer/portnado/internal/domain"
)

type fakeRunner map[string][]byte

func (f fakeRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	key := name
	for _, arg := range args {
		key += " " + arg
	}
	return f[key], nil
}

func TestDetectorDiscoversComposePublishedPorts(t *testing.T) {
	detector := Detector{Runner: fakeRunner{
		"docker compose ls --format json": []byte(`[{"Name":"webguard"}]`),
		"docker compose -p webguard ps --format json": []byte(`[{
			"Name":"webguard-app-1",
			"Service":"app",
			"Project":"webguard",
			"State":"running",
			"Publishers":[{"URL":"127.0.0.1:4173","TargetPort":5173,"PublishedPort":4173,"Protocol":"tcp"}]
		}]`),
	}}
	observations, warnings := detector.Discover(context.Background())
	if len(warnings) != 0 {
		t.Fatalf("warnings = %v", warnings)
	}
	if len(observations) != 1 {
		t.Fatalf("observations = %d", len(observations))
	}
	if observations[0].Project.Name != "webguard" || observations[0].Service.Name != "app" || observations[0].BackendPort != 4173 {
		t.Fatalf("observation = %+v", observations[0])
	}
	if observations[0].Protocol != domain.ProtocolHTTP {
		t.Fatalf("protocol = %q", observations[0].Protocol)
	}
}

func TestDetectorClassifiesDatabaseAsTCP(t *testing.T) {
	observations := serviceObservations("webguard", composeService{
		Service: "database",
		Publishers: []publisher{{
			URL:           "0.0.0.0:5438",
			TargetPort:    5432,
			PublishedPort: 5438,
			Protocol:      "tcp",
		}},
	}, time.Unix(0, 0).UTC())
	if observations[0].Protocol != domain.ProtocolTCP {
		t.Fatalf("protocol = %q", observations[0].Protocol)
	}
	if observations[0].BackendHost != "127.0.0.1" {
		t.Fatalf("backend host = %q", observations[0].BackendHost)
	}
}
