package darwin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderLaunchAgent(t *testing.T) {
	data, err := RenderLaunchAgent(LaunchAgent{
		DaemonPath: "/Applications/Portnado.app/Contents/Resources/bin/portnado-daemon",
		LogPath:    "/tmp/portnado.log",
	})
	if err != nil {
		t.Fatalf("render launch agent: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, LaunchAgentLabel) {
		t.Fatalf("launch agent missing label: %s", text)
	}
	if !strings.Contains(text, "portnado-daemon") {
		t.Fatalf("launch agent missing daemon path: %s", text)
	}
}

func TestInstallAndRemoveLaunchAgent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path, err := InstallLaunchAgent(LaunchAgent{
		DaemonPath: "/tmp/portnado-daemon",
		LogPath:    "/tmp/portnado.log",
	})
	if err != nil {
		t.Fatalf("install launch agent: %v", err)
	}
	if want := filepath.Join(home, "Library", "LaunchAgents", LaunchAgentLabel+".plist"); path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat launch agent: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %s", info.Mode().Perm())
	}
	installed, installedPath, err := LaunchAgentInstalled()
	if err != nil {
		t.Fatalf("launch agent installed: %v", err)
	}
	if !installed || installedPath != path {
		t.Fatalf("installed=%v path=%q", installed, installedPath)
	}
	if _, err := RemoveLaunchAgent(); err != nil {
		t.Fatalf("remove launch agent: %v", err)
	}
	installed, _, err = LaunchAgentInstalled()
	if err != nil {
		t.Fatalf("launch agent installed after remove: %v", err)
	}
	if installed {
		t.Fatal("launch agent still installed")
	}
}
