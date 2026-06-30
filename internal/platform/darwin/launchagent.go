package darwin

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

const LaunchAgentLabel = "dev.portnado.daemon"

type LaunchAgent struct {
	Label      string
	DaemonPath string
	SocketPath string
	LogPath    string
}

func LaunchAgentPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, "Library", "LaunchAgents", LaunchAgentLabel+".plist"), nil
}

func RenderLaunchAgent(agent LaunchAgent) ([]byte, error) {
	if agent.Label == "" {
		agent.Label = LaunchAgentLabel
	}
	tpl := template.Must(template.New("launch-agent").Parse(launchAgentTemplate))
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, agent); err != nil {
		return nil, fmt.Errorf("render launch agent: %w", err)
	}
	return buf.Bytes(), nil
}

func InstallLaunchAgent(agent LaunchAgent) (string, error) {
	path, err := LaunchAgentPath()
	if err != nil {
		return "", err
	}
	data, err := RenderLaunchAgent(agent)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", fmt.Errorf("create LaunchAgents directory: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", fmt.Errorf("write launch agent: %w", err)
	}
	return path, nil
}

func RemoveLaunchAgent() (string, error) {
	path, err := LaunchAgentPath()
	if err != nil {
		return "", err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("remove launch agent: %w", err)
	}
	return path, nil
}

func LaunchAgentInstalled() (bool, string, error) {
	path, err := LaunchAgentPath()
	if err != nil {
		return false, "", err
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, path, nil
	}
	if os.IsNotExist(err) {
		return false, path, nil
	}
	return false, path, err
}

const launchAgentTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>{{.Label}}</string>
  <key>ProgramArguments</key>
  <array>
    <string>{{.DaemonPath}}</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <false/>
  <key>StandardOutPath</key>
  <string>{{.LogPath}}</string>
  <key>StandardErrorPath</key>
  <string>{{.LogPath}}</string>
</dict>
</plist>
`
