package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/marcel-breuer/portnado/internal/domain"
)

type fakeStatusClient struct {
	status statusResult
	scan   domain.ScanResult
	list   []domain.ServiceSummary
	routes []domain.ConfirmedRoute
	route  domain.ConfirmedRoute
	err    error
}

func (f fakeStatusClient) Status(context.Context) (statusResult, error) {
	return f.status, f.err
}

func (f fakeStatusClient) Scan(context.Context, string) (domain.ScanResult, error) {
	return f.scan, f.err
}

func (f fakeStatusClient) List(context.Context) ([]domain.ServiceSummary, error) {
	return f.list, f.err
}

func (f fakeStatusClient) Routes(context.Context) ([]domain.ConfirmedRoute, error) {
	return f.routes, f.err
}

func (f fakeStatusClient) ApproveRoute(context.Context, string) (domain.ConfirmedRoute, error) {
	return f.route, f.err
}

func (f fakeStatusClient) EnableRoute(context.Context, string) (domain.ConfirmedRoute, error) {
	return f.route, f.err
}

func (f fakeStatusClient) DisableRoute(context.Context, string) (domain.ConfirmedRoute, error) {
	return f.route, f.err
}

func TestVersionCommand(t *testing.T) {
	var out bytes.Buffer
	app := App{Out: &out, Err: &bytes.Buffer{}}

	code := app.Run(context.Background(), []string{"--version"})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "portnado 0.1.0-dev") {
		t.Fatalf("version output = %q", out.String())
	}
}

func TestStatusCommandUnavailable(t *testing.T) {
	var out bytes.Buffer
	app := App{
		Out: &out,
		Err: &bytes.Buffer{},
		Client: fakeStatusClient{
			err: errors.New("dial unix /tmp/missing.sock: connect: no such file or directory"),
		},
	}

	code := app.Run(context.Background(), []string{"status"})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	output := out.String()
	if !strings.Contains(output, "Daemon: unavailable") {
		t.Fatalf("status output = %q", output)
	}
	if strings.Contains(output, "/tmp/missing.sock") {
		t.Fatalf("status output leaked raw socket detail: %q", output)
	}
}

func TestStatusCommandJSON(t *testing.T) {
	var out bytes.Buffer
	startedAt := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	app := App{
		Out: &out,
		Err: &bytes.Buffer{},
		Client: fakeStatusClient{
			status: statusResult{
				DaemonState:     "running",
				ProtocolVersion: 1,
				Version:         "0.1.0-dev",
				SocketPath:      "/tmp/portnado.sock",
				StartedAt:       startedAt,
			},
		},
	}

	code := app.Run(context.Background(), []string{"status", "--json"})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(out.String(), `"daemonState": "running"`) {
		t.Fatalf("json output = %q", out.String())
	}
}

func TestScanCommand(t *testing.T) {
	var out bytes.Buffer
	app := App{
		Out: &out,
		Err: &bytes.Buffer{},
		Client: fakeStatusClient{scan: domain.ScanResult{
			Observations: []domain.Observation{{ID: "obs"}},
			Suggestions: []domain.RouteSuggestion{{
				RouteHost:   "app.webguard.localhost",
				BackendHost: "127.0.0.1",
				BackendPort: 5173,
				State:       domain.RouteStateAwaitingApproval,
			}},
		}},
	}

	code := app.Run(context.Background(), []string{"scan"})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "Suggestions: 1") {
		t.Fatalf("scan output = %q", out.String())
	}
}

func TestListCommand(t *testing.T) {
	var out bytes.Buffer
	app := App{
		Out: &out,
		Err: &bytes.Buffer{},
		Client: fakeStatusClient{list: []domain.ServiceSummary{{
			ProjectName: "webguard",
			ServiceName: "app",
			RouteHost:   "app.webguard.localhost",
			BackendHost: "127.0.0.1",
			BackendPort: 5173,
			State:       domain.RouteStateAwaitingApproval,
			Confidence:  domain.ConfidenceHigh,
		}}},
	}

	code := app.Run(context.Background(), []string{"list"})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "webguard/app") {
		t.Fatalf("list output = %q", out.String())
	}
}

func TestRouteApproveCommand(t *testing.T) {
	var out bytes.Buffer
	app := App{
		Out: &out,
		Err: &bytes.Buffer{},
		Client: fakeStatusClient{route: domain.ConfirmedRoute{
			RouteHost: "app.webguard.localhost",
			State:     domain.RouteStateActive,
		}},
	}
	code := app.Run(context.Background(), []string{"route", "approve", "suggestion_app"})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "app.webguard.localhost") {
		t.Fatalf("route output = %q", out.String())
	}
}

func TestRouteListCommand(t *testing.T) {
	var out bytes.Buffer
	app := App{
		Out: &out,
		Err: &bytes.Buffer{},
		Client: fakeStatusClient{routes: []domain.ConfirmedRoute{{
			ID:          "route_app",
			ProjectName: "webguard",
			ServiceName: "app",
			RouteHost:   "app.webguard.localhost",
			BackendHost: "127.0.0.1",
			BackendPort: 5173,
			State:       domain.RouteStateActive,
		}}},
	}
	code := app.Run(context.Background(), []string{"route", "list"})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "route_app") {
		t.Fatalf("route list output = %q", out.String())
	}
}

func TestSetupDryRunCommand(t *testing.T) {
	var out bytes.Buffer
	app := App{Out: &out, Err: &bytes.Buffer{}}

	code := app.Run(context.Background(), []string{"setup", "--dry-run", "--launch-at-login"})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	output := out.String()
	if !strings.Contains(output, "launch-agent") || !strings.Contains(output, "Dry run only") {
		t.Fatalf("setup output = %q", output)
	}
}

func TestUninstallDryRunPreservesRepositoryConfig(t *testing.T) {
	var out bytes.Buffer
	app := App{Out: &out, Err: &bytes.Buffer{}}

	code := app.Run(context.Background(), []string{"uninstall", "--dry-run", "--delete-state"})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	output := out.String()
	if !strings.Contains(output, "local-state") || !strings.Contains(output, "Repository .portnado.yml files are preserved") {
		t.Fatalf("uninstall output = %q", output)
	}
}

func TestDoctorJSONCommand(t *testing.T) {
	var out bytes.Buffer
	app := App{Out: &out, Err: &bytes.Buffer{}}

	code := app.Run(context.Background(), []string{"doctor", "--json", "--proxy-address", "127.0.0.1:1"})
	if code != 0 && code != 1 {
		t.Fatalf("exit code = %d, want 0 or 1", code)
	}
	output := out.String()
	if !strings.Contains(output, `"checks"`) || !strings.Contains(output, `"platform"`) {
		t.Fatalf("doctor output = %q", output)
	}
}
