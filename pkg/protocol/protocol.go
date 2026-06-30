package protocol

import "time"

const (
	Version      = 1
	MaxFrameSize = 1 << 20

	MethodDaemonStatus  = "daemon.status"
	MethodDaemonVersion = "daemon.version"
	MethodScanRun       = "scan.run"
	MethodRoutesList    = "routes.list"
	MethodRouteApprove  = "route.approve"
	MethodRouteEnable   = "route.enable"
	MethodRouteDisable  = "route.disable"
	MethodRouteList     = "route.list"
)

type Request struct {
	ProtocolVersion int             `json:"protocolVersion"`
	RequestID       string          `json:"requestId"`
	Method          string          `json:"method"`
	Params          jsonRawEnvelope `json:"params,omitempty"`
}

type Response struct {
	ProtocolVersion int             `json:"protocolVersion"`
	RequestID       string          `json:"requestId"`
	OK              bool            `json:"ok"`
	Result          jsonRawEnvelope `json:"result,omitempty"`
	Error           *Error          `json:"error,omitempty"`
}

type Error struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type Status struct {
	DaemonState     string    `json:"daemonState"`
	ProtocolVersion int       `json:"protocolVersion"`
	Version         string    `json:"version"`
	SocketPath      string    `json:"socketPath"`
	StartedAt       time.Time `json:"startedAt"`
}

type VersionInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ScanParams struct {
	Root string `json:"root,omitempty"`
}

type RouteIDParams struct {
	ID string `json:"id"`
}
