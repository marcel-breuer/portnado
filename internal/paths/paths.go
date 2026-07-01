package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

const appSupportPath = "Library/Application Support/Portnado"
const maxDarwinUnixSocketPathLength = 104

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
	socketPath := filepath.Join(runDir, "portnado.sock")
	if len(socketPath) > maxDarwinUnixSocketPathLength {
		return "", fmt.Errorf("control socket path is %d bytes, exceeding the macOS limit of %d bytes; use a shorter HOME path", len(socketPath), maxDarwinUnixSocketPathLength)
	}
	return socketPath, nil
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
