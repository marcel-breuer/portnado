package project

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestProjectNameFromRoot(t *testing.T) {
	got := ProjectNameFromRoot("/Users/example/Shop_API")
	if got != "shop-api" {
		t.Fatalf("name = %q", got)
	}
}

func TestMarkerFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte("{}"), 0o600); err != nil {
		t.Fatalf("write marker: %v", err)
	}
	if err := os.Mkdir(filepath.Join(root, "go.mod"), 0o700); err != nil {
		t.Fatalf("write marker dir: %v", err)
	}
	markers, err := MarkerFiles(root)
	if err != nil {
		t.Fatalf("marker files: %v", err)
	}
	if len(markers) != 1 || filepath.Base(markers[0]) != "package.json" {
		t.Fatalf("markers = %v", markers)
	}
}

func TestMarkerFilesEmptyDir(t *testing.T) {
	markers, err := MarkerFiles("")
	if err != nil {
		t.Fatalf("marker files: %v", err)
	}
	if markers != nil {
		t.Fatalf("markers = %v, want nil", markers)
	}
}

func TestGitRoot(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	root := t.TempDir()
	if err := exec.Command("git", "-C", root, "init").Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}
	child := filepath.Join(root, "app")
	if err := os.Mkdir(child, 0o700); err != nil {
		t.Fatalf("mkdir child: %v", err)
	}
	got, ok := GitRoot(context.Background(), child)
	if !ok {
		t.Fatal("expected git root")
	}
	want, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("eval root symlink: %v", err)
	}
	got, err = filepath.EvalSymlinks(got)
	if err != nil {
		t.Fatalf("eval git root symlink: %v", err)
	}
	if got != want {
		t.Fatalf("git root = %q, want %q", got, want)
	}
}
