package httprouting

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/marcel-breuer/portnado/internal/domain"
)

func TestProxyRoutesByHost(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hello" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("ok"))
	}))
	defer backend.Close()

	host, port := backendHostPort(t, backend.URL)
	proxy := NewProxy("127.0.0.1:0")
	proxy.UpdateRoutes([]domain.ConfirmedRoute{{
		ID:          "route_app",
		Protocol:    domain.ProtocolHTTP,
		RouteHost:   "app.webguard.localhost",
		BackendHost: host,
		BackendPort: port,
		State:       domain.RouteStateActive,
	}})

	request := httptest.NewRequest(http.MethodGet, "http://app.webguard.localhost/hello?x=1", nil)
	request.Host = "app.webguard.localhost"
	response := httptest.NewRecorder()
	proxy.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	if body := response.Body.String(); body != "ok" {
		t.Fatalf("body = %q", body)
	}
}

func TestProxyUnavailableBackend(t *testing.T) {
	proxy := NewProxy("127.0.0.1:0")
	proxy.UpdateRoutes([]domain.ConfirmedRoute{{
		ID:          "route_app",
		Protocol:    domain.ProtocolHTTP,
		RouteHost:   "app.webguard.localhost",
		BackendHost: "127.0.0.1",
		BackendPort: 1,
		State:       domain.RouteStateActive,
	}})

	request := httptest.NewRequest(http.MethodGet, "http://app.webguard.localhost/", nil)
	request.Host = "app.webguard.localhost:4780"
	response := httptest.NewRecorder()
	proxy.ServeHTTP(response, request)

	if response.Code != http.StatusBadGateway {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestProxyListenAndServeBindsLoopback(t *testing.T) {
	proxy := NewProxy("127.0.0.1:0")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() {
		errCh <- proxy.ListenAndServe(ctx)
	}()
	time.Sleep(20 * time.Millisecond)
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("proxy error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("proxy did not stop")
	}
}

func TestProxyServerHasResourceLimits(t *testing.T) {
	proxy := NewProxy("127.0.0.1:0")
	if proxy.server.ReadHeaderTimeout == 0 {
		t.Fatal("ReadHeaderTimeout must be set")
	}
	if proxy.server.ReadTimeout == 0 {
		t.Fatal("ReadTimeout must be set")
	}
	if proxy.server.IdleTimeout == 0 {
		t.Fatal("IdleTimeout must be set")
	}
	if proxy.server.MaxHeaderBytes == 0 {
		t.Fatal("MaxHeaderBytes must be set")
	}
}

func backendHostPort(t testing.TB, rawURL string) (string, int) {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		t.Fatalf("parse backend URL: %v", err)
	}
	host, portText, err := net.SplitHostPort(request.URL.Host)
	if err != nil {
		t.Fatalf("split host port: %v", err)
	}
	var port int
	_, err = fmt.Sscanf(portText, "%d", &port)
	if err != nil {
		t.Fatalf("parse port: %v", err)
	}
	return host, port
}

func TestProxySupportsStreaming(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		flusher, _ := w.(http.Flusher)
		_, _ = w.Write([]byte("first\n"))
		flusher.Flush()
		_, _ = w.Write([]byte("second\n"))
	}))
	defer backend.Close()
	host, port := backendHostPort(t, backend.URL)
	proxy := NewProxy("127.0.0.1:0")
	proxy.UpdateRoutes([]domain.ConfirmedRoute{{
		Protocol:    domain.ProtocolHTTP,
		RouteHost:   "app.webguard.localhost",
		BackendHost: host,
		BackendPort: port,
		State:       domain.RouteStateActive,
	}})
	request := httptest.NewRequest(http.MethodGet, "http://app.webguard.localhost/events", nil)
	request.Host = "app.webguard.localhost"
	response := httptest.NewRecorder()
	proxy.ServeHTTP(response, request)
	body, _ := io.ReadAll(response.Result().Body)
	if string(body) != "first\nsecond\n" {
		t.Fatalf("body = %q", string(body))
	}
}

func TestProxySupportsUpgrade(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.ToLower(r.Header.Get("Upgrade")) != "websocket" {
			t.Fatalf("missing upgrade header")
		}
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			t.Fatalf("backend does not support hijacking")
		}
		conn, _, err := hijacker.Hijack()
		if err != nil {
			t.Fatalf("hijack backend: %v", err)
		}
		defer conn.Close()
		_, _ = conn.Write([]byte("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n\r\n"))
	}))
	defer backend.Close()
	host, port := backendHostPort(t, backend.URL)
	proxy := NewProxy("127.0.0.1:0")
	proxy.UpdateRoutes([]domain.ConfirmedRoute{{
		Protocol:    domain.ProtocolHTTP,
		RouteHost:   "app.webguard.localhost",
		BackendHost: host,
		BackendPort: port,
		State:       domain.RouteStateActive,
	}})
	proxyServer := httptest.NewServer(proxy)
	defer proxyServer.Close()

	conn, err := net.Dial("tcp", strings.TrimPrefix(proxyServer.URL, "http://"))
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}
	defer conn.Close()
	_, _ = conn.Write([]byte("GET /socket HTTP/1.1\r\nHost: app.webguard.localhost\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\n"))
	buffer := make([]byte, 128)
	n, err := conn.Read(buffer)
	if err != nil {
		t.Fatalf("read upgrade response: %v", err)
	}
	if !strings.Contains(string(buffer[:n]), "101 Switching Protocols") {
		t.Fatalf("upgrade response = %q", string(buffer[:n]))
	}
}

func BenchmarkProxyServeHTTP(b *testing.B) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer backend.Close()
	host, port := backendHostPort(b, backend.URL)
	proxy := NewProxy("127.0.0.1:0")
	proxy.UpdateRoutes([]domain.ConfirmedRoute{{
		Protocol:    domain.ProtocolHTTP,
		RouteHost:   "app.webguard.localhost",
		BackendHost: host,
		BackendPort: port,
		State:       domain.RouteStateActive,
	}})

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		request := httptest.NewRequest(http.MethodGet, "http://app.webguard.localhost/", nil)
		request.Host = "app.webguard.localhost"
		response := httptest.NewRecorder()
		proxy.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			b.Fatalf("status = %d", response.Code)
		}
	}
}
