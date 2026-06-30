package daemon

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/marcel-breuer/portnado/internal/app"
	"github.com/marcel-breuer/portnado/internal/domain"
	"github.com/marcel-breuer/portnado/internal/paths"
	"github.com/marcel-breuer/portnado/internal/persistence"
	"github.com/marcel-breuer/portnado/internal/routing"
	"github.com/marcel-breuer/portnado/internal/version"
	"github.com/marcel-breuer/portnado/pkg/protocol"
)

type Server struct {
	startedAt time.Time
	app       *app.Service
	store     *persistence.Store
	routing   *routing.Manager
}

func NewServer() *Server {
	return &Server{startedAt: time.Now().UTC()}
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	if s.app == nil {
		store, err := persistence.Open(ctx)
		if err != nil {
			return err
		}
		defer store.Close()
		s.store = store
		s.app = app.NewService(store)
	}
	if s.routing == nil {
		s.routing = routing.NewManager("127.0.0.1:4780")
		s.routing.Start(ctx)
		if routes, err := s.app.ActiveRoutes(ctx); err == nil {
			_ = s.routing.Reload(ctx, routes)
		}
	}

	socketPath, err := paths.SocketPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(socketPath), 0o700); err != nil {
		return fmt.Errorf("create run directory: %w", err)
	}
	if err := removeInactiveSocket(socketPath); err != nil {
		return err
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("listen on control socket: %w", err)
	}
	defer listener.Close()
	defer os.Remove(socketPath)

	if err := os.Chmod(socketPath, 0o600); err != nil {
		return fmt.Errorf("restrict control socket permissions: %w", err)
	}

	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return nil
			}
			return fmt.Errorf("accept control connection: %w", err)
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	codec := protocol.NewCodec(conn)
	request, err := codec.ReadRequest()
	if err != nil {
		return
	}

	response := s.handleRequest(request)
	_ = codec.WriteResponse(response)
}

func (s *Server) handleRequest(request protocol.Request) protocol.Response {
	response := protocol.Response{
		ProtocolVersion: protocol.Version,
		RequestID:       request.RequestID,
	}

	if request.ProtocolVersion != protocol.Version {
		response.OK = false
		response.Error = &protocol.Error{
			Code:    "unsupported_protocol_version",
			Message: "Unsupported protocol version.",
		}
		return response
	}

	switch request.Method {
	case protocol.MethodDaemonStatus:
		socketPath, _ := paths.SocketPath()
		result, err := protocol.MarshalParams(protocol.Status{
			DaemonState:     "running",
			ProtocolVersion: protocol.Version,
			Version:         version.Version,
			SocketPath:      socketPath,
			StartedAt:       s.startedAt,
		})
		if err != nil {
			response.OK = false
			response.Error = &protocol.Error{Code: "internal_error", Message: "Could not encode status."}
			return response
		}
		response.OK = true
		response.Result = result
	case protocol.MethodDaemonVersion:
		result, err := protocol.MarshalParams(protocol.VersionInfo{Name: version.Name, Version: version.Version})
		if err != nil {
			response.OK = false
			response.Error = &protocol.Error{Code: "internal_error", Message: "Could not encode version."}
			return response
		}
		response.OK = true
		response.Result = result
	case protocol.MethodScanRun:
		var params protocol.ScanParams
		if len(request.Params) != 0 {
			if err := protocol.UnmarshalParams(request.Params, &params); err != nil {
				response.OK = false
				response.Error = &protocol.Error{Code: "invalid_params", Message: "Could not decode scan parameters."}
				return response
			}
		}
		result, err := s.app.Scan(context.Background(), params.Root)
		if err != nil {
			response.OK = false
			response.Error = &protocol.Error{Code: "scan_failed", Message: err.Error()}
			return response
		}
		raw, err := protocol.MarshalParams(result)
		if err != nil {
			response.OK = false
			response.Error = &protocol.Error{Code: "internal_error", Message: "Could not encode scan result."}
			return response
		}
		response.OK = true
		response.Result = raw
	case protocol.MethodRoutesList:
		result, err := s.app.List(context.Background())
		if err != nil {
			response.OK = false
			response.Error = &protocol.Error{Code: "list_failed", Message: err.Error()}
			return response
		}
		raw, err := protocol.MarshalParams(result)
		if err != nil {
			response.OK = false
			response.Error = &protocol.Error{Code: "internal_error", Message: "Could not encode route list."}
			return response
		}
		response.OK = true
		response.Result = raw
	case protocol.MethodRouteApprove, protocol.MethodRouteEnable, protocol.MethodRouteDisable:
		route, err := s.handleRouteAction(context.Background(), request)
		if err != nil {
			response.OK = false
			response.Error = &protocol.Error{Code: "route_action_failed", Message: err.Error()}
			return response
		}
		raw, err := protocol.MarshalParams(route)
		if err != nil {
			response.OK = false
			response.Error = &protocol.Error{Code: "internal_error", Message: "Could not encode route."}
			return response
		}
		response.OK = true
		response.Result = raw
	case protocol.MethodRouteList:
		result, err := s.app.Routes(context.Background())
		if err != nil {
			response.OK = false
			response.Error = &protocol.Error{Code: "route_list_failed", Message: err.Error()}
			return response
		}
		raw, err := protocol.MarshalParams(result)
		if err != nil {
			response.OK = false
			response.Error = &protocol.Error{Code: "internal_error", Message: "Could not encode route list."}
			return response
		}
		response.OK = true
		response.Result = raw
	default:
		response.OK = false
		response.Error = &protocol.Error{
			Code:    "unknown_method",
			Message: "Unknown IPC method.",
		}
	}

	return response
}

func (s *Server) handleRouteAction(ctx context.Context, request protocol.Request) (domain.ConfirmedRoute, error) {
	var params protocol.RouteIDParams
	if err := protocol.UnmarshalParams(request.Params, &params); err != nil {
		return domain.ConfirmedRoute{}, fmt.Errorf("decode route id: %w", err)
	}
	if params.ID == "" {
		return domain.ConfirmedRoute{}, fmt.Errorf("route id is required")
	}
	var route domain.ConfirmedRoute
	var err error
	switch request.Method {
	case protocol.MethodRouteApprove:
		route, err = s.app.ApproveRoute(ctx, params.ID)
	case protocol.MethodRouteEnable:
		route, err = s.app.EnableRoute(ctx, params.ID)
	case protocol.MethodRouteDisable:
		route, err = s.app.DisableRoute(ctx, params.ID)
	default:
		err = fmt.Errorf("unsupported route action")
	}
	if err != nil {
		return domain.ConfirmedRoute{}, err
	}
	if s.routing != nil {
		routes, err := s.app.ActiveRoutes(ctx)
		if err != nil {
			return domain.ConfirmedRoute{}, err
		}
		if err := s.routing.Reload(ctx, routes); err != nil {
			return domain.ConfirmedRoute{}, err
		}
	}
	return route, nil
}

func removeInactiveSocket(socketPath string) error {
	conn, err := net.DialTimeout("unix", socketPath, 100*time.Millisecond)
	if err == nil {
		_ = conn.Close()
		return fmt.Errorf("daemon already appears to be running at %s", socketPath)
	}
	if os.IsNotExist(err) {
		return nil
	}
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove stale control socket: %w", err)
	}
	return nil
}
