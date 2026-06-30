package paths

import (
	"path/filepath"
	"testing"
)

func TestPathsUseApplicationSupport(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	appSupport, err := AppSupportDir()
	if err != nil {
		t.Fatalf("app support: %v", err)
	}
	if want := filepath.Join(home, "Library", "Application Support", "Portnado"); appSupport != want {
		t.Fatalf("app support = %q, want %q", appSupport, want)
	}
	socket, err := SocketPath()
	if err != nil {
		t.Fatalf("socket path: %v", err)
	}
	if want := filepath.Join(appSupport, "run", "portnado.sock"); socket != want {
		t.Fatalf("socket = %q, want %q", socket, want)
	}
	database, err := DatabasePath()
	if err != nil {
		t.Fatalf("database path: %v", err)
	}
	if want := filepath.Join(appSupport, "portnado.db"); database != want {
		t.Fatalf("database = %q, want %q", database, want)
	}
	override, err := ProjectOverridePath("project_webguard")
	if err != nil {
		t.Fatalf("override path: %v", err)
	}
	if want := filepath.Join(appSupport, "projects", "project_webguard.yml"); override != want {
		t.Fatalf("override = %q, want %q", override, want)
	}
}
