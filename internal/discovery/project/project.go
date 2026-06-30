package project

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GitRoot(ctx context.Context, dir string) (string, bool) {
	if dir == "" {
		return "", false
	}
	command := exec.CommandContext(ctx, "git", "-C", dir, "rev-parse", "--show-toplevel")
	output, err := command.Output()
	if err != nil {
		return "", false
	}
	root := strings.TrimSpace(string(output))
	if root == "" {
		return "", false
	}
	return root, true
}

func ProjectNameFromRoot(root string) string {
	name := strings.ToLower(filepath.Base(root))
	name = strings.ReplaceAll(name, "_", "-")
	return name
}

func MarkerFiles(dir string) ([]string, error) {
	if dir == "" {
		return nil, nil
	}
	names := []string{"package.json", "composer.json", "pyproject.toml", "requirements.txt", "go.mod", "pom.xml", "build.gradle", "build.gradle.kts"}
	var found []string
	for _, name := range names {
		path := filepath.Join(dir, name)
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			found = append(found, path)
			continue
		}
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("inspect project marker %s: %w", path, err)
		}
	}
	return found, nil
}
