package tcprouting

import (
	"bufio"
	"context"
	"io"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/marcel-breuer/portnado/internal/domain"
)

func TestForwarderProxiesTCP(t *testing.T) {
	backend := startTCPEcho(t)
	defer backend.Close()
	_, backendPortText, err := net.SplitHostPort(backend.Addr().String())
	if err != nil {
		t.Fatalf("split backend address: %v", err)
	}
	backendPort, _ := strconv.Atoi(backendPortText)
	frontend := freePort(t)

	forwarder := NewForwarder()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer forwarder.Close()
	err = forwarder.UpdateRoutes(ctx, []domain.ConfirmedRoute{{
		ID:           "route_db",
		Protocol:     domain.ProtocolTCP,
		RouteHost:    "db.webguard.localhost",
		FrontendPort: frontend,
		BackendHost:  "127.0.0.1",
		BackendPort:  backendPort,
		State:        domain.RouteStateActive,
	}})
	if err != nil {
		t.Fatalf("update routes: %v", err)
	}

	conn, err := net.DialTimeout("tcp4", "127.0.0.1:"+strconv.Itoa(frontend), time.Second)
	if err != nil {
		t.Fatalf("dial frontend: %v", err)
	}
	defer conn.Close()
	_, _ = conn.Write([]byte("ping\n"))
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		t.Fatalf("read echo: %v", err)
	}
	if line != "ping\n" {
		t.Fatalf("line = %q", line)
	}
}

func startTCPEcho(t *testing.T) net.Listener {
	t.Helper()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen backend: %v", err)
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				_, _ = io.Copy(conn, conn)
			}()
		}
	}()
	return listener
}

func freePort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen free port: %v", err)
	}
	defer listener.Close()
	_, portText, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatalf("split free port: %v", err)
	}
	port, _ := strconv.Atoi(portText)
	return port
}
