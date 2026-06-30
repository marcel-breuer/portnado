package routing

import (
	"context"
	"fmt"
	"sync"

	"github.com/marcel-breuer/portnado/internal/domain"
	httprouting "github.com/marcel-breuer/portnado/internal/routing/http"
	tcprouting "github.com/marcel-breuer/portnado/internal/routing/tcp"
)

type Manager struct {
	httpProxy  *httprouting.Proxy
	tcpForward *tcprouting.Forwarder
	once       sync.Once
	errCh      chan error
}

func NewManager(httpAddress string) *Manager {
	return &Manager{
		httpProxy:  httprouting.NewProxy(httpAddress),
		tcpForward: tcprouting.NewForwarder(),
		errCh:      make(chan error, 1),
	}
}

func (m *Manager) Start(ctx context.Context) {
	m.once.Do(func() {
		go func() {
			if err := m.httpProxy.ListenAndServe(ctx); err != nil {
				select {
				case m.errCh <- err:
				default:
				}
			}
		}()
		go func() {
			<-ctx.Done()
			m.tcpForward.Close()
		}()
	})
}

func (m *Manager) Reload(ctx context.Context, routes []domain.ConfirmedRoute) error {
	m.httpProxy.UpdateRoutes(routes)
	if err := m.tcpForward.UpdateRoutes(ctx, routes); err != nil {
		return fmt.Errorf("reload tcp routes: %w", err)
	}
	return nil
}

func (m *Manager) Errors() <-chan error {
	return m.errCh
}
