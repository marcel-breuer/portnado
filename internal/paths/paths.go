package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

const appSupportPath = "Library/Application Support/Portnado"

func AppSupportDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home: %w", err)
	}
	return filepath.Join(home, appSupportPath), nil
}

func RunDir() (string, error) {
	base, err := AppSupportDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "run"), nil
}

func SocketPath() (string, error) {
	runDir, err := RunDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(runDir, "portnado.sock"), nil
}

func DatabasePath() (string, error) {
	base, err := AppSupportDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "portnado.db"), nil
}

func ProjectOverridePath(projectID string) (string, error) {
	base, err := AppSupportDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "projects", projectID+".yml"), nil
}
