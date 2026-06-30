package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/marcel-breuer/portnado/internal/paths"
	darwinplatform "github.com/marcel-breuer/portnado/internal/platform/darwin"
)

type Change struct {
	ID         string `json:"id"`
	Scope      string `json:"scope"`
	Action     string `json:"action"`
	Path       string `json:"path,omitempty"`
	Privileged bool   `json:"privileged"`
	Applied    bool   `json:"applied"`
	Message    string `json:"message"`
}

type Plan struct {
	Mode    string   `json:"mode"`
	Changes []Change `json:"changes"`
}

type SetupOptions struct {
	LaunchAtLogin bool
	PortlessHTTP  bool
	HostsFallback bool
	Apply         bool
	DaemonPath    string
}

func BuildSetupPlan(options SetupOptions) (Plan, error) {
	var changes []Change
	if options.LaunchAtLogin {
		path, err := darwinplatform.LaunchAgentPath()
		if err != nil {
			return Plan{}, err
		}
		changes = append(changes, Change{
			ID:      "launch-agent",
			Scope:   "user",
			Action:  "write LaunchAgent plist",
			Path:    path,
			Message: "Starts the Portnado daemon at user login. It does not start development projects.",
		})
	}
	if options.PortlessHTTP {
		changes = append(changes, Change{
			ID:         "pf-portless-http",
			Scope:      "system",
			Action:     "install managed PF redirect from 127.0.0.1:80 to 127.0.0.1:4780",
			Path:       "/etc/pf.anchors/dev.portnado",
			Privileged: true,
			Message:    "Requires administrator approval. The daemon remains unprivileged.",
		})
	}
	if options.HostsFallback {
		changes = append(changes, Change{
			ID:         "hosts-localhost-fallback",
			Scope:      "system",
			Action:     "add clearly marked managed fallback host entries",
			Path:       "/etc/hosts",
			Privileged: true,
			Message:    "Only needed if doctor proves wildcard .localhost resolution is broken.",
		})
	}
	return Plan{Mode: "setup", Changes: changes}, nil
}

func ApplySetup(options SetupOptions) (Plan, error) {
	plan, err := BuildSetupPlan(options)
	if err != nil {
		return Plan{}, err
	}
	if !options.Apply {
		return plan, nil
	}
	for i := range plan.Changes {
		change := &plan.Changes[i]
		if change.Privileged && os.Geteuid() != 0 {
			return plan, fmt.Errorf("%s requires root privileges; rerun with explicit administrator approval", change.ID)
		}
		switch change.ID {
		case "launch-agent":
			daemonPath := options.DaemonPath
			if daemonPath == "" {
				if path, err := exec.LookPath("portnado-daemon"); err == nil {
					daemonPath = path
				}
			}
			if daemonPath == "" {
				return plan, fmt.Errorf("daemon path is required to install launch at login")
			}
			logPath, err := defaultLogPath()
			if err != nil {
				return plan, err
			}
			path, err := darwinplatform.InstallLaunchAgent(darwinplatform.LaunchAgent{
				DaemonPath: daemonPath,
				LogPath:    logPath,
			})
			if err != nil {
				return plan, err
			}
			change.Path = path
			change.Applied = true
		case "pf-portless-http":
			change.Applied = false
			return plan, fmt.Errorf("PF setup is not applied by this phase implementation; use setup --dry-run to inspect the planned managed anchor")
		case "hosts-localhost-fallback":
			change.Applied = false
			return plan, fmt.Errorf("hosts fallback is not applied automatically; use doctor output before considering this fallback")
		}
	}
	return plan, nil
}

func BuildUninstallPlan(deleteState bool) (Plan, error) {
	launchPath, err := darwinplatform.LaunchAgentPath()
	if err != nil {
		return Plan{}, err
	}
	changes := []Change{
		{ID: "launch-agent", Scope: "user", Action: "remove LaunchAgent plist", Path: launchPath, Message: "Stops future login startup after the daemon is manually stopped."},
		{ID: "pf-portless-http", Scope: "system", Action: "remove managed PF redirect", Path: "/etc/pf.anchors/dev.portnado", Privileged: true, Message: "Only removes Portnado-managed PF state."},
		{ID: "hosts-localhost-fallback", Scope: "system", Action: "remove managed host entries", Path: "/etc/hosts", Privileged: true, Message: "Only removes clearly marked Portnado entries."},
	}
	if deleteState {
		appSupport, err := paths.AppSupportDir()
		if err != nil {
			return Plan{}, err
		}
		changes = append(changes, Change{ID: "local-state", Scope: "user", Action: "delete local Portnado state", Path: appSupport, Message: "Repository .portnado.yml files are preserved."})
	}
	return Plan{Mode: "uninstall", Changes: changes}, nil
}

func ApplyUninstall(deleteState bool, apply bool) (Plan, error) {
	plan, err := BuildUninstallPlan(deleteState)
	if err != nil {
		return Plan{}, err
	}
	if !apply {
		return plan, nil
	}
	for i := range plan.Changes {
		change := &plan.Changes[i]
		if change.Privileged {
			if os.Geteuid() != 0 {
				continue
			}
			continue
		}
		switch change.ID {
		case "launch-agent":
			if _, err := darwinplatform.RemoveLaunchAgent(); err != nil {
				return plan, err
			}
			change.Applied = true
		case "local-state":
			if err := os.RemoveAll(change.Path); err != nil {
				return plan, fmt.Errorf("delete local state: %w", err)
			}
			change.Applied = true
		}
	}
	return plan, nil
}

func defaultLogPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	logDir := filepath.Join(home, "Library", "Logs", "Portnado")
	if runtime.GOOS != "darwin" {
		logDir = filepath.Join(home, ".portnado", "logs")
	}
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(logDir, "portnado.log"), nil
}
