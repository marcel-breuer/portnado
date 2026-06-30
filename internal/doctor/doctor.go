package doctor

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"time"

	_ "modernc.org/sqlite"

	"github.com/marcel-breuer/portnado/internal/paths"
	darwinplatform "github.com/marcel-breuer/portnado/internal/platform/darwin"
)

type Status string

const (
	StatusPass Status = "pass"
	StatusWarn Status = "warn"
	StatusFail Status = "fail"
)

type Check struct {
	ID      string `json:"id"`
	Status  Status `json:"status"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

type Report struct {
	GeneratedAt time.Time `json:"generatedAt"`
	Checks      []Check   `json:"checks"`
}

type Options struct {
	ProxyAddress string
	LookupHost   string
	Now          func() time.Time
}

func Run(ctx context.Context, options Options) Report {
	if options.ProxyAddress == "" {
		options.ProxyAddress = "127.0.0.1:4780"
	}
	if options.LookupHost == "" {
		options.LookupHost = "app.portnado-test.localhost"
	}
	now := time.Now
	if options.Now != nil {
		now = options.Now
	}

	checks := []Check{
		checkPlatform(),
		checkLocalhostResolution(options.LookupHost),
		checkSocketPermissions(),
		checkSQLite(),
		checkProxy(ctx, options.ProxyAddress),
		checkDocker(),
		checkLaunchAgent(),
	}
	return Report{GeneratedAt: now().UTC(), Checks: checks}
}

func (r Report) HasFailures() bool {
	for _, check := range r.Checks {
		if check.Status == StatusFail {
			return true
		}
	}
	return false
}

func checkPlatform() Check {
	status := StatusPass
	message := "Supported local development platform"
	if runtime.GOOS != "darwin" {
		status = StatusWarn
		message = "Portnado is designed for macOS first"
	}
	return Check{
		ID:      "platform",
		Status:  status,
		Message: message,
		Detail:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

func checkLocalhostResolution(host string) Check {
	addresses, err := net.LookupHost(host)
	if err != nil {
		return Check{ID: "localhost-resolution", Status: StatusFail, Message: "Wildcard .localhost resolution failed", Detail: err.Error()}
	}
	return Check{ID: "localhost-resolution", Status: StatusPass, Message: ".localhost names resolve locally", Detail: fmt.Sprintf("%s -> %v", host, addresses)}
}

func checkSocketPermissions() Check {
	socketPath, err := paths.SocketPath()
	if err != nil {
		return Check{ID: "control-socket", Status: StatusFail, Message: "Could not resolve control socket path", Detail: err.Error()}
	}
	info, err := os.Stat(socketPath)
	if os.IsNotExist(err) {
		return Check{ID: "control-socket", Status: StatusWarn, Message: "Control socket is not present", Detail: socketPath}
	}
	if err != nil {
		return Check{ID: "control-socket", Status: StatusFail, Message: "Control socket cannot be inspected", Detail: err.Error()}
	}
	if info.Mode()&0o077 != 0 {
		return Check{ID: "control-socket", Status: StatusFail, Message: "Control socket is too permissive", Detail: fmt.Sprintf("%s mode %s", socketPath, info.Mode().Perm())}
	}
	return Check{ID: "control-socket", Status: StatusPass, Message: "Control socket permissions are user-scoped", Detail: fmt.Sprintf("%s mode %s", socketPath, info.Mode().Perm())}
}

func checkSQLite() Check {
	databasePath, err := paths.DatabasePath()
	if err != nil {
		return Check{ID: "sqlite", Status: StatusFail, Message: "Could not resolve database path", Detail: err.Error()}
	}
	if _, err := os.Stat(databasePath); os.IsNotExist(err) {
		return Check{ID: "sqlite", Status: StatusWarn, Message: "SQLite database has not been initialized", Detail: databasePath}
	} else if err != nil {
		return Check{ID: "sqlite", Status: StatusFail, Message: "SQLite database cannot be inspected", Detail: err.Error()}
	}

	db, err := sql.Open("sqlite", databasePath+"?mode=ro")
	if err != nil {
		return Check{ID: "sqlite", Status: StatusFail, Message: "SQLite database could not be opened", Detail: err.Error()}
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return Check{ID: "sqlite", Status: StatusFail, Message: "SQLite health check failed", Detail: err.Error()}
	}
	return Check{ID: "sqlite", Status: StatusPass, Message: "SQLite database is readable", Detail: databasePath}
}

func checkProxy(ctx context.Context, address string) Check {
	dialer := net.Dialer{Timeout: 250 * time.Millisecond}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return Check{ID: "http-proxy", Status: StatusWarn, Message: "High-port HTTP proxy is not accepting connections", Detail: err.Error()}
	}
	_ = conn.Close()
	return Check{ID: "http-proxy", Status: StatusPass, Message: "High-port HTTP proxy is reachable", Detail: address}
}

func checkDocker() Check {
	path, err := exec.LookPath("docker")
	if err != nil {
		return Check{ID: "docker", Status: StatusWarn, Message: "Docker CLI was not found", Detail: err.Error()}
	}
	return Check{ID: "docker", Status: StatusPass, Message: "Docker CLI is available", Detail: path}
}

func checkLaunchAgent() Check {
	installed, path, err := darwinplatform.LaunchAgentInstalled()
	if err != nil {
		return Check{ID: "launch-agent", Status: StatusFail, Message: "LaunchAgent state cannot be inspected", Detail: err.Error()}
	}
	if installed {
		return Check{ID: "launch-agent", Status: StatusPass, Message: "Launch at login is installed", Detail: path}
	}
	return Check{ID: "launch-agent", Status: StatusWarn, Message: "Launch at login is not installed", Detail: path}
}
