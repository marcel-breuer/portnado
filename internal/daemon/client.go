package daemon

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/marcel-breuer/portnado/internal/domain"
	"github.com/marcel-breuer/portnado/internal/paths"
	"github.com/marcel-breuer/portnado/pkg/protocol"
)

type Client struct {
	SocketPath string
	Timeout    time.Duration
}

func NewClient() (*Client, error) {
	socketPath, err := paths.SocketPath()
	if err != nil {
		return nil, err
	}
	return &Client{SocketPath: socketPath, Timeout: time.Second}, nil
}

func (c *Client) Status(ctx context.Context) (protocol.Status, error) {
	response, err := c.call(ctx, protocol.MethodDaemonStatus, nil)
	if err != nil {
		return protocol.Status{}, err
	}
	if !response.OK {
		return protocol.Status{}, fmt.Errorf("%s: %s", response.Error.Code, response.Error.Message)
	}
	return protocol.DecodeResult[protocol.Status](response)
}

func (c *Client) Scan(ctx context.Context, root string) (domain.ScanResult, error) {
	response, err := c.call(ctx, protocol.MethodScanRun, protocol.ScanParams{Root: root})
	if err != nil {
		return domain.ScanResult{}, err
	}
	if !response.OK {
		return domain.ScanResult{}, fmt.Errorf("%s: %s", response.Error.Code, response.Error.Message)
	}
	return protocol.DecodeResult[domain.ScanResult](response)
}

func (c *Client) List(ctx context.Context) ([]domain.ServiceSummary, error) {
	response, err := c.call(ctx, protocol.MethodRoutesList, nil)
	if err != nil {
		return nil, err
	}
	if !response.OK {
		return nil, fmt.Errorf("%s: %s", response.Error.Code, response.Error.Message)
	}
	return protocol.DecodeResult[[]domain.ServiceSummary](response)
}

func (c *Client) Routes(ctx context.Context) ([]domain.ConfirmedRoute, error) {
	response, err := c.call(ctx, protocol.MethodRouteList, nil)
	if err != nil {
		return nil, err
	}
	if !response.OK {
		return nil, fmt.Errorf("%s: %s", response.Error.Code, response.Error.Message)
	}
	return protocol.DecodeResult[[]domain.ConfirmedRoute](response)
}

func (c *Client) ApproveRoute(ctx context.Context, suggestionID string) (domain.ConfirmedRoute, error) {
	return c.routeAction(ctx, protocol.MethodRouteApprove, suggestionID)
}

func (c *Client) EnableRoute(ctx context.Context, routeID string) (domain.ConfirmedRoute, error) {
	return c.routeAction(ctx, protocol.MethodRouteEnable, routeID)
}

func (c *Client) DisableRoute(ctx context.Context, routeID string) (domain.ConfirmedRoute, error) {
	return c.routeAction(ctx, protocol.MethodRouteDisable, routeID)
}

func (c *Client) routeAction(ctx context.Context, method string, id string) (domain.ConfirmedRoute, error) {
	response, err := c.call(ctx, method, protocol.RouteIDParams{ID: id})
	if err != nil {
		return domain.ConfirmedRoute{}, err
	}
	if !response.OK {
		return domain.ConfirmedRoute{}, fmt.Errorf("%s: %s", response.Error.Code, response.Error.Message)
	}
	return protocol.DecodeResult[domain.ConfirmedRoute](response)
}

func (c *Client) call(ctx context.Context, method string, params any) (protocol.Response, error) {
	timeout := c.Timeout
	if timeout == 0 {
		timeout = time.Second
	}
	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "unix", c.SocketPath)
	if err != nil {
		return protocol.Response{}, err
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	} else {
		_ = conn.SetDeadline(time.Now().Add(timeout))
	}

	rawParams, err := protocol.MarshalParams(params)
	if err != nil {
		return protocol.Response{}, err
	}

	codec := protocol.NewCodec(conn)
	if err := codec.WriteRequest(protocol.Request{
		ProtocolVersion: protocol.Version,
		RequestID:       "cli",
		Method:          method,
		Params:          rawParams,
	}); err != nil {
		return protocol.Response{}, err
	}
	return codec.ReadResponse()
}
