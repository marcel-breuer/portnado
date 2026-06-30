package httprouting

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/marcel-breuer/portnado/internal/config"
	"github.com/marcel-breuer/portnado/internal/domain"
)

type Proxy struct {
	address string
	server  *http.Server
	mu      sync.RWMutex
	routes  map[string]domain.ConfirmedRoute
}

func NewProxy(address string) *Proxy {
	if address == "" {
		address = "127.0.0.1:4780"
	}
	proxy := &Proxy{
		address: address,
		routes:  make(map[string]domain.ConfirmedRoute),
	}
	proxy.server = &http.Server{
		Addr:              address,
		Handler:           proxy,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      0,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    32 << 10,
	}
	return proxy
}

func (p *Proxy) Address() string {
	return p.address
}

func (p *Proxy) UpdateRoutes(routes []domain.ConfirmedRoute) {
	next := make(map[string]domain.ConfirmedRoute)
	for _, route := range routes {
		if route.Protocol != domain.ProtocolHTTP || route.State != domain.RouteStateActive {
			continue
		}
		host := normalizeHost(route.RouteHost)
		if config.ValidateLocalhost(host) != nil {
			continue
		}
		next[host] = route
	}
	p.mu.Lock()
	p.routes = next
	p.mu.Unlock()
}

func (p *Proxy) ListenAndServe(ctx context.Context) error {
	listener, err := net.Listen("tcp4", p.address)
	if err != nil {
		return fmt.Errorf("listen http proxy: %w", err)
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = p.server.Shutdown(shutdownCtx)
	}()
	err = p.server.Serve(listener)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve http proxy: %w", err)
	}
	return nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := normalizeHost(r.Host)
	route, ok := p.route(host)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if route.BackendHost == "" || route.BackendPort == 0 {
		writeUnavailable(w, route)
		return
	}
	target := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", route.BackendHost, route.BackendPort),
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
		req.URL.Scheme = "http"
		req.URL.Host = target.Host
		req.Header.Del("X-Forwarded-Host")
		req.Header.Del("X-Forwarded-Proto")
		req.Header.Del("X-Forwarded-For")
		req.Header.Set("X-Forwarded-Host", host)
		req.Header.Set("X-Forwarded-Proto", "http")
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, _ error) {
		writeUnavailable(w, route)
	}
	proxy.ServeHTTP(w, r)
}

func (p *Proxy) route(host string) (domain.ConfirmedRoute, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	route, ok := p.routes[host]
	return route, ok
}

func normalizeHost(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	if strings.HasPrefix(host, "[") {
		return host
	}
	if splitHost, _, err := net.SplitHostPort(host); err == nil {
		return strings.ToLower(splitHost)
	}
	if index := strings.LastIndex(host, ":"); index > -1 {
		maybePort := host[index+1:]
		if maybePort != "" && allDigits(maybePort) {
			return host[:index]
		}
	}
	return host
}

func allDigits(value string) bool {
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func writeUnavailable(w http.ResponseWriter, route domain.ConfirmedRoute) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusBadGateway)
	_, _ = fmt.Fprintf(w, "Portnado route %s is unavailable.\n", route.RouteHost)
}
