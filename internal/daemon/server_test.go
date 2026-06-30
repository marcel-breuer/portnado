package daemon

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/marcel-breuer/portnado/internal/paths"
)

func TestServerRespondsToStatus(t *testing.T) {
	home := filepath.Join("/tmp", "pn-"+strconv.FormatInt(time.Now().UnixNano(), 36))
	if err := os.MkdirAll(home, 0o700); err != nil {
		t.Fatalf("create short test home: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(home)
	})
	t.Setenv("HOME", home)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- NewServer().ListenAndServe(ctx)
	}()

	socketPath, err := paths.SocketPath()
	if err != nil {
		t.Fatalf("socket path: %v", err)
	}
	waitForSocket(t, socketPath, errCh)

	client, err := NewClient()
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	status, err := client.Status(context.Background())
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if status.DaemonState != "running" {
		t.Fatalf("daemon state = %q, want running", status.DaemonState)
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("server returned error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("server did not stop")
	}
}

func waitForSocket(t *testing.T, socketPath string, errCh <-chan error) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		select {
		case err := <-errCh:
			t.Fatalf("server stopped before socket was ready: %v", err)
		default:
		}
		if info, err := os.Stat(socketPath); err == nil && info.Mode()&os.ModeSocket != 0 {
			conn, err := net.DialTimeout("unix", socketPath, 50*time.Millisecond)
			if err == nil {
				_ = conn.Close()
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	parent := filepath.Dir(socketPath)
	entries, _ := os.ReadDir(parent)
	t.Fatalf("socket was not ready at %s; entries=%v", socketPath, entries)
}
