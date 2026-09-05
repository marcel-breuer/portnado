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
}

func NewManager(httpAddress string) *Manager {
	return &Manager{
		httpProxy:  httprouting.NewProxy(httpAddress),
		tcpForward: tcprouting.NewForwarder(),
	}
}

func (m *Manager) Start(ctx context.Context) {
	m.once.Do(func() {
		go func() {
			_ = m.httpProxy.ListenAndServe(ctx)
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
