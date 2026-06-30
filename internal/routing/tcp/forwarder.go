package tcprouting

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/marcel-breuer/portnado/internal/domain"
)

type Forwarder struct {
	mu        sync.Mutex
	listeners map[string]net.Listener
	cancel    map[string]context.CancelFunc
	routes    map[string]domain.ConfirmedRoute
}

func NewForwarder() *Forwarder {
	return &Forwarder{
		listeners: make(map[string]net.Listener),
		cancel:    make(map[string]context.CancelFunc),
		routes:    make(map[string]domain.ConfirmedRoute),
	}
}

func (f *Forwarder) UpdateRoutes(ctx context.Context, routes []domain.ConfirmedRoute) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	desired := make(map[string]domain.ConfirmedRoute)
	for _, route := range routes {
		if route.Protocol != domain.ProtocolTCP || route.State != domain.RouteStateActive || route.FrontendPort == 0 {
			continue
		}
		desired[route.ID] = route
	}

	for id := range f.listeners {
		if _, ok := desired[id]; !ok {
			f.stopLocked(id)
		}
	}
	for id, route := range desired {
		if current, ok := f.routes[id]; ok && sameListener(current, route) {
			continue
		}
		f.stopLocked(id)
		if err := f.startLocked(ctx, route); err != nil {
			return err
		}
	}
	return nil
}

func (f *Forwarder) Close() {
	f.mu.Lock()
	defer f.mu.Unlock()
	for id := range f.listeners {
		f.stopLocked(id)
	}
}

func (f *Forwarder) startLocked(ctx context.Context, route domain.ConfirmedRoute) error {
	address := "127.0.0.1:" + strconv.Itoa(route.FrontendPort)
	listener, err := net.Listen("tcp4", address)
	if err != nil {
		return fmt.Errorf("listen tcp route %s on %s: %w", route.ID, address, err)
	}
	routeCtx, cancel := context.WithCancel(ctx)
	f.listeners[route.ID] = listener
	f.cancel[route.ID] = cancel
	f.routes[route.ID] = route
	go f.accept(routeCtx, listener, route)
	return nil
}

func (f *Forwarder) stopLocked(id string) {
	if cancel, ok := f.cancel[id]; ok {
		cancel()
		delete(f.cancel, id)
	}
	if listener, ok := f.listeners[id]; ok {
		_ = listener.Close()
		delete(f.listeners, id)
	}
	delete(f.routes, id)
}

func (f *Forwarder) accept(ctx context.Context, listener net.Listener, route domain.ConfirmedRoute) {
	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go handleConnection(ctx, conn, route)
	}
}

func handleConnection(ctx context.Context, frontend net.Conn, route domain.ConfirmedRoute) {
	defer frontend.Close()
	dialer := net.Dialer{Timeout: 2 * time.Second}
	backend, err := dialer.DialContext(ctx, "tcp4", fmt.Sprintf("%s:%d", route.BackendHost, route.BackendPort))
	if err != nil {
		return
	}
	defer backend.Close()

	done := make(chan struct{}, 2)
	go copyAndClose(frontend, backend, done)
	go copyAndClose(backend, frontend, done)
	<-done
}

func copyAndClose(dst net.Conn, src net.Conn, done chan<- struct{}) {
	_, _ = io.Copy(dst, src)
	if tcp, ok := dst.(*net.TCPConn); ok {
		_ = tcp.CloseWrite()
	} else {
		_ = dst.Close()
	}
	done <- struct{}{}
}

func sameListener(a, b domain.ConfirmedRoute) bool {
	return a.FrontendPort == b.FrontendPort && a.BackendHost == b.BackendHost && a.BackendPort == b.BackendPort
}
