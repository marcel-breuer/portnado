package system

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/marcel-breuer/portnado/internal/paths"
)

func TestBuildSetupPlan(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	plan, err := BuildSetupPlan(SetupOptions{LaunchAtLogin: true, PortlessHTTP: true})
	if err != nil {
		t.Fatalf("build setup plan: %v", err)
	}
	if len(plan.Changes) != 2 {
		t.Fatalf("changes = %d", len(plan.Changes))
	}
	if !plan.Changes[1].Privileged {
		t.Fatalf("expected PF change to be privileged")
	}
}

func TestApplySetupLaunchAgent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	plan, err := ApplySetup(SetupOptions{
		LaunchAtLogin: true,
		Apply:         true,
		DaemonPath:    "/tmp/portnado-daemon",
	})
	if err != nil {
		t.Fatalf("apply setup: %v", err)
	}
	if !plan.Changes[0].Applied {
		t.Fatalf("launch agent was not marked applied")
	}
}

func TestBuildUninstallPlanPreservesRepositoryFiles(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	plan, err := BuildUninstallPlan(true)
	if err != nil {
		t.Fatalf("build uninstall plan: %v", err)
	}
	foundState := false
	for _, change := range plan.Changes {
		if change.ID == "local-state" {
			foundState = true
		}
	}
	if !foundState {
		t.Fatalf("expected local state deletion change")
	}
}

func TestApplyUninstallDeletesOnlyManagedState(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	repositoryConfig := filepath.Join(t.TempDir(), ".portnado.yml")
	if err := os.WriteFile(repositoryConfig, []byte("version: 1\n"), 0o600); err != nil {
		t.Fatalf("write repo config: %v", err)
	}
	appSupport, err := paths.AppSupportDir()
	if err != nil {
		t.Fatalf("app support: %v", err)
	}
	if err := os.MkdirAll(appSupport, 0o700); err != nil {
		t.Fatalf("create app support: %v", err)
	}
	if err := os.WriteFile(filepath.Join(appSupport, "portnado.db"), []byte("state"), 0o600); err != nil {
		t.Fatalf("write state: %v", err)
	}

	if _, err := ApplyUninstall(true, true); err != nil {
		t.Fatalf("apply uninstall: %v", err)
	}
	if _, err := os.Stat(appSupport); !os.IsNotExist(err) {
		t.Fatalf("app support should be removed, stat err = %v", err)
	}
	if _, err := os.Stat(repositoryConfig); err != nil {
		t.Fatalf("repository config should be preserved: %v", err)
	}
}
