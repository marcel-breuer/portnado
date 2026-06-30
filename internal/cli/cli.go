package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/marcel-breuer/portnado/internal/config"
	"github.com/marcel-breuer/portnado/internal/daemon"
	"github.com/marcel-breuer/portnado/internal/doctor"
	"github.com/marcel-breuer/portnado/internal/domain"
	"github.com/marcel-breuer/portnado/internal/paths"
	"github.com/marcel-breuer/portnado/internal/system"
	"github.com/marcel-breuer/portnado/internal/version"
)

type App struct {
	Out    io.Writer
	Err    io.Writer
	Client client
}

type client interface {
	Status(context.Context) (statusResult, error)
	Scan(context.Context, string) (domain.ScanResult, error)
	List(context.Context) ([]domain.ServiceSummary, error)
	Routes(context.Context) ([]domain.ConfirmedRoute, error)
	ApproveRoute(context.Context, string) (domain.ConfirmedRoute, error)
	EnableRoute(context.Context, string) (domain.ConfirmedRoute, error)
	DisableRoute(context.Context, string) (domain.ConfirmedRoute, error)
}

type statusResult struct {
	DaemonState     string    `json:"daemonState"`
	ProtocolVersion int       `json:"protocolVersion"`
	Version         string    `json:"version"`
	SocketPath      string    `json:"socketPath"`
	StartedAt       time.Time `json:"startedAt"`
}

type daemonClientAdapter struct {
	client *daemon.Client
}

func (a daemonClientAdapter) Status(ctx context.Context) (statusResult, error) {
	status, err := a.client.Status(ctx)
	if err != nil {
		return statusResult{}, err
	}
	return statusResult{
		DaemonState:     status.DaemonState,
		ProtocolVersion: status.ProtocolVersion,
		Version:         status.Version,
		SocketPath:      status.SocketPath,
		StartedAt:       status.StartedAt,
	}, nil
}

func (a daemonClientAdapter) Scan(ctx context.Context, root string) (domain.ScanResult, error) {
	return a.client.Scan(ctx, root)
}

func (a daemonClientAdapter) List(ctx context.Context) ([]domain.ServiceSummary, error) {
	return a.client.List(ctx)
}

func (a daemonClientAdapter) Routes(ctx context.Context) ([]domain.ConfirmedRoute, error) {
	return a.client.Routes(ctx)
}

func (a daemonClientAdapter) ApproveRoute(ctx context.Context, id string) (domain.ConfirmedRoute, error) {
	return a.client.ApproveRoute(ctx, id)
}

func (a daemonClientAdapter) EnableRoute(ctx context.Context, id string) (domain.ConfirmedRoute, error) {
	return a.client.EnableRoute(ctx, id)
}

func (a daemonClientAdapter) DisableRoute(ctx context.Context, id string) (domain.ConfirmedRoute, error) {
	return a.client.DisableRoute(ctx, id)
}

func New(out, errOut io.Writer) (*App, error) {
	client, err := daemon.NewClient()
	if err != nil {
		return nil, err
	}
	return &App{Out: out, Err: errOut, Client: daemonClientAdapter{client: client}}, nil
}

func (a *App) Run(ctx context.Context, args []string) int {
	if len(args) == 0 {
		printHelp(a.Out)
		return 0
	}

	switch args[0] {
	case "--help", "-h", "help":
		printHelp(a.Out)
		return 0
	case "--version", "version":
		fmt.Fprintf(a.Out, "portnado %s\n", version.Version)
		return 0
	case "status":
		return a.runStatus(ctx, args[1:])
	case "scan":
		return a.runScan(ctx, args[1:])
	case "list":
		return a.runList(ctx, args[1:])
	case "config":
		return a.runConfig(args[1:])
	case "route":
		return a.runRoute(ctx, args[1:])
	case "doctor":
		return a.runDoctor(ctx, args[1:])
	case "setup":
		return a.runSetup(args[1:])
	case "uninstall":
		return a.runUninstall(args[1:])
	default:
		fmt.Fprintf(a.Err, "unknown command: %s\n\n", args[0])
		printHelp(a.Err)
		return 2
	}
}

func (a *App) runRoute(ctx context.Context, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(a.Err, "usage: portnado route list|approve|enable|disable <route-id>")
		return 2
	}
	if args[0] == "list" {
		return a.runRouteList(ctx)
	}
	if len(args) < 2 {
		fmt.Fprintln(a.Err, "usage: portnado route approve|enable|disable <route-id>")
		return 2
	}
	action, id := args[0], args[1]
	var route domain.ConfirmedRoute
	var err error
	switch action {
	case "approve":
		route, err = a.Client.ApproveRoute(ctx, id)
	case "enable":
		route, err = a.Client.EnableRoute(ctx, id)
	case "disable":
		route, err = a.Client.DisableRoute(ctx, id)
	default:
		fmt.Fprintln(a.Err, "usage: portnado route approve|enable|disable <route-id>")
		return 2
	}
	if err != nil {
		fmt.Fprintf(a.Err, "route %s failed: %s\n", action, sanitizeStatusError(err))
		return 1
	}
	fmt.Fprintf(a.Out, "Route %s: %s %s (%s)\n", action, route.ID, displayConfirmedRoute(route), route.State)
	return 0
}

func (a *App) runRouteList(ctx context.Context) int {
	routes, err := a.Client.Routes(ctx)
	if err != nil {
		fmt.Fprintf(a.Err, "route list failed: %s\n", sanitizeStatusError(err))
		return 1
	}
	fmt.Fprintln(a.Out, "Portnado confirmed routes")
	if len(routes) == 0 {
		fmt.Fprintln(a.Out, "No confirmed routes.")
		return 0
	}
	for _, route := range routes {
		fmt.Fprintf(a.Out, "- %s %s/%s %s -> %s:%d [%s]\n", route.ID, route.ProjectName, route.ServiceName, displayConfirmedRoute(route), route.BackendHost, route.BackendPort, route.State)
	}
	return 0
}

func (a *App) runStatus(ctx context.Context, args []string) int {
	flags := flag.NewFlagSet("status", flag.ContinueOnError)
	flags.SetOutput(a.Err)
	jsonOutput := flags.Bool("json", false, "write machine-readable JSON")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	status, err := a.Client.Status(ctx)
	if err != nil {
		return a.writeUnavailable(*jsonOutput, err)
	}

	if *jsonOutput {
		return writeJSON(a.Out, status)
	}

	fmt.Fprintln(a.Out, "Portnado status")
	fmt.Fprintf(a.Out, "Daemon: %s\n", status.DaemonState)
	fmt.Fprintf(a.Out, "Version: %s\n", status.Version)
	fmt.Fprintf(a.Out, "Protocol: %d\n", status.ProtocolVersion)
	fmt.Fprintf(a.Out, "Socket: %s\n", status.SocketPath)
	fmt.Fprintf(a.Out, "Started: %s\n", status.StartedAt.Format(time.RFC3339))
	return 0
}

func (a *App) runScan(ctx context.Context, args []string) int {
	flags := flag.NewFlagSet("scan", flag.ContinueOnError)
	flags.SetOutput(a.Err)
	jsonOutput := flags.Bool("json", false, "write machine-readable JSON")
	root := flags.String("root", "", "project root to scan for .portnado.yml")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	result, err := a.Client.Scan(ctx, *root)
	if err != nil {
		fmt.Fprintf(a.Err, "scan failed: %s\n", sanitizeStatusError(err))
		return 1
	}
	if *jsonOutput {
		return writeJSON(a.Out, result)
	}
	fmt.Fprintln(a.Out, "Portnado scan")
	fmt.Fprintf(a.Out, "Observations: %d\n", len(result.Observations))
	fmt.Fprintf(a.Out, "Suggestions: %d\n", len(result.Suggestions))
	for _, warning := range result.Warnings {
		fmt.Fprintf(a.Out, "Warning: %s\n", warning)
	}
	for _, suggestion := range result.Suggestions {
		fmt.Fprintf(a.Out, "- %s -> %s:%d (%s)\n", displayRoute(suggestion), suggestion.BackendHost, suggestion.BackendPort, suggestion.State)
	}
	return 0
}

func (a *App) runList(ctx context.Context, args []string) int {
	flags := flag.NewFlagSet("list", flag.ContinueOnError)
	flags.SetOutput(a.Err)
	jsonOutput := flags.Bool("json", false, "write machine-readable JSON")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	summaries, err := a.Client.List(ctx)
	if err != nil {
		fmt.Fprintf(a.Err, "list failed: %s\n", sanitizeStatusError(err))
		return 1
	}
	if *jsonOutput {
		return writeJSON(a.Out, summaries)
	}
	fmt.Fprintln(a.Out, "Portnado routes")
	if len(summaries) == 0 {
		fmt.Fprintln(a.Out, "No route suggestions. Run portnado scan after starting the daemon.")
		return 0
	}
	for _, summary := range summaries {
		route := summary.RouteHost
		if summary.FrontendPort != 0 {
			route = fmt.Sprintf("%s:%d", route, summary.FrontendPort)
		}
		fmt.Fprintf(a.Out, "- %s %s/%s %s -> %s:%d [%s, %s]\n", summary.RouteID, summary.ProjectName, summary.ServiceName, route, summary.BackendHost, summary.BackendPort, summary.State, summary.Confidence)
	}
	return 0
}

func (a *App) runConfig(args []string) int {
	if len(args) == 0 || args[0] != "validate" {
		fmt.Fprintln(a.Err, "usage: portnado config validate [path]")
		return 2
	}
	path := config.RepositoryFileName
	if len(args) > 1 {
		path = args[1]
	}
	if _, err := config.LoadRepositoryFile(path); err != nil {
		fmt.Fprintf(a.Err, "config invalid: %s\n", err)
		return 1
	}
	fmt.Fprintf(a.Out, "Config valid: %s\n", path)
	return 0
}

func (a *App) runDoctor(ctx context.Context, args []string) int {
	flags := flag.NewFlagSet("doctor", flag.ContinueOnError)
	flags.SetOutput(a.Err)
	jsonOutput := flags.Bool("json", false, "write machine-readable JSON")
	proxyAddress := flags.String("proxy-address", "127.0.0.1:4780", "HTTP proxy address to check")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	report := doctor.Run(ctx, doctor.Options{ProxyAddress: *proxyAddress})
	if *jsonOutput {
		return writeJSON(a.Out, report)
	}

	fmt.Fprintln(a.Out, "Portnado doctor")
	fmt.Fprintf(a.Out, "Generated: %s\n", report.GeneratedAt.Format(time.RFC3339))
	for _, check := range report.Checks {
		fmt.Fprintf(a.Out, "- [%s] %s: %s", strings.ToUpper(string(check.Status)), check.ID, check.Message)
		if check.Detail != "" {
			fmt.Fprintf(a.Out, " (%s)", check.Detail)
		}
		fmt.Fprintln(a.Out)
	}
	if report.HasFailures() {
		return 1
	}
	return 0
}

func (a *App) runSetup(args []string) int {
	flags := flag.NewFlagSet("setup", flag.ContinueOnError)
	flags.SetOutput(a.Err)
	jsonOutput := flags.Bool("json", false, "write machine-readable JSON")
	dryRun := flags.Bool("dry-run", false, "preview planned changes")
	yes := flags.Bool("yes", false, "apply approved non-privileged changes")
	launchAtLogin := flags.Bool("launch-at-login", false, "install user LaunchAgent")
	portlessHTTP := flags.Bool("portless-http", false, "plan managed local port 80 redirect")
	hostsFallback := flags.Bool("hosts-fallback", false, "plan managed /etc/hosts fallback entries")
	daemonPath := flags.String("daemon-path", "", "portnado-daemon path for launch at login")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if !*launchAtLogin && !*portlessHTTP && !*hostsFallback {
		*launchAtLogin = true
		*portlessHTTP = true
	}

	options := system.SetupOptions{
		LaunchAtLogin: *launchAtLogin,
		PortlessHTTP:  *portlessHTTP,
		HostsFallback: *hostsFallback,
		Apply:         *yes && !*dryRun,
		DaemonPath:    *daemonPath,
	}
	plan, err := system.ApplySetup(options)
	if *jsonOutput {
		return writePlanJSON(a.Out, plan, err)
	}
	printPlan(a.Out, plan)
	if err != nil {
		fmt.Fprintf(a.Err, "setup failed: %s\n", err)
		return 1
	}
	if !options.Apply {
		fmt.Fprintln(a.Out, "Dry run only. Re-run with --yes to apply user-scope changes.")
	}
	return 0
}

func (a *App) runUninstall(args []string) int {
	flags := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	flags.SetOutput(a.Err)
	jsonOutput := flags.Bool("json", false, "write machine-readable JSON")
	dryRun := flags.Bool("dry-run", false, "preview planned changes")
	yes := flags.Bool("yes", false, "apply approved non-privileged removals")
	deleteState := flags.Bool("delete-state", false, "delete local Portnado state; repository config files are preserved")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	apply := *yes && !*dryRun
	plan, err := system.ApplyUninstall(*deleteState, apply)
	if *jsonOutput {
		return writePlanJSON(a.Out, plan, err)
	}
	printPlan(a.Out, plan)
	if err != nil {
		fmt.Fprintf(a.Err, "uninstall failed: %s\n", err)
		return 1
	}
	if !apply {
		fmt.Fprintln(a.Out, "Dry run only. Re-run with --yes to apply user-scope removals.")
	}
	return 0
}

func (a *App) writeUnavailable(jsonOutput bool, statusErr error) int {
	socketPath, err := paths.SocketPath()
	if err != nil {
		socketPath = "unavailable"
	}
	if jsonOutput {
		return writeJSON(a.Out, map[string]any{
			"daemonState": "unavailable",
			"socketPath":  socketPath,
			"error":       sanitizeStatusError(statusErr),
		})
	}
	fmt.Fprintln(a.Out, "Portnado status")
	fmt.Fprintln(a.Out, "Daemon: unavailable")
	fmt.Fprintf(a.Out, "Socket: %s\n", socketPath)
	fmt.Fprintf(a.Out, "Detail: %s\n", sanitizeStatusError(statusErr))
	return 0
}

func sanitizeStatusError(err error) string {
	if err == nil {
		return "none"
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "daemon did not respond before timeout"
	}
	message := err.Error()
	if strings.Contains(message, "no such file") || strings.Contains(message, "connect: connection refused") {
		return "daemon control socket is not reachable"
	}
	return message
}

func writeJSON(out io.Writer, value any) int {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		fmt.Fprintf(os.Stderr, "encode JSON output: %v\n", err)
		return 1
	}
	return 0
}

func writePlanJSON(out io.Writer, plan system.Plan, planErr error) int {
	payload := map[string]any{
		"plan": plan,
	}
	if planErr != nil {
		payload["error"] = planErr.Error()
	}
	return writeJSON(out, payload)
}

func printPlan(out io.Writer, plan system.Plan) {
	title := "Portnado setup plan"
	if plan.Mode == "uninstall" {
		title = "Portnado uninstall plan"
	}
	fmt.Fprintln(out, title)
	if len(plan.Changes) == 0 {
		fmt.Fprintln(out, "No changes planned.")
		return
	}
	for _, change := range plan.Changes {
		scope := change.Scope
		if change.Privileged {
			scope += ", privileged"
		}
		applied := "planned"
		if change.Applied {
			applied = "applied"
		}
		if change.Path != "" {
			fmt.Fprintf(out, "- %s [%s, %s] %s at %s\n", change.ID, scope, applied, change.Action, change.Path)
		} else {
			fmt.Fprintf(out, "- %s [%s, %s] %s\n", change.ID, scope, applied, change.Action)
		}
		if change.Message != "" {
			fmt.Fprintf(out, "  %s\n", change.Message)
		}
	}
}

func printHelp(out io.Writer) {
	fmt.Fprint(out, `Portnado routes stable local names to changing development ports.

Usage:
  portnado --version
  portnado status [--json]
  portnado scan [--json] [--root PATH]
  portnado list [--json]
  portnado config validate [path]
  portnado doctor [--json]
  portnado setup [--dry-run] [--yes] [--launch-at-login] [--portless-http] [--hosts-fallback]
  portnado uninstall [--dry-run] [--yes] [--delete-state]
  portnado route list
  portnado route approve <suggestion-id>
  portnado route enable <route-id>
  portnado route disable <route-id>

Phase 5 implements setup previews, doctor diagnostics, uninstall planning, and menu bar route actions.
`)
}

func displayRoute(suggestion domain.RouteSuggestion) string {
	if suggestion.FrontendPort != 0 {
		return fmt.Sprintf("%s:%d", suggestion.RouteHost, suggestion.FrontendPort)
	}
	return suggestion.RouteHost
}

func displayConfirmedRoute(route domain.ConfirmedRoute) string {
	if route.FrontendPort != 0 {
		return fmt.Sprintf("%s:%d", route.RouteHost, route.FrontendPort)
	}
	return route.RouteHost
}
