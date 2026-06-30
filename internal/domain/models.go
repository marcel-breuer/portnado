package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

type Protocol string

const (
	ProtocolHTTP Protocol = "http"
	ProtocolTCP  Protocol = "tcp"
)

type Confidence string

const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
)

type RouteState string

const (
	RouteStateSuggested        RouteState = "suggested"
	RouteStateAwaitingApproval RouteState = "awaiting_approval"
	RouteStateActive           RouteState = "active"
	RouteStateInactive         RouteState = "inactive"
	RouteStateUnavailable      RouteState = "backend_unavailable"
	RouteStateStale            RouteState = "stale"
	RouteStateConflict         RouteState = "conflict"
	RouteStateInvalid          RouteState = "invalid"
	RouteStateDisabled         RouteState = "disabled"
	RouteStateError            RouteState = "error"
)

type Project struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	RootPath  string    `json:"rootPath,omitempty"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Service struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"projectId"`
	Name        string    `json:"name"`
	Protocol    Protocol  `json:"protocol"`
	Description string    `json:"description,omitempty"`
	Source      string    `json:"source"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Evidence struct {
	Source  string `json:"source"`
	Summary string `json:"summary"`
}

type Observation struct {
	ID          string     `json:"id"`
	ScanRunID   string     `json:"scanRunId,omitempty"`
	Project     Project    `json:"project"`
	Service     Service    `json:"service"`
	Runtime     string     `json:"runtime,omitempty"`
	Protocol    Protocol   `json:"protocol"`
	BackendHost string     `json:"backendHost"`
	BackendPort int        `json:"backendPort"`
	Evidence    []Evidence `json:"evidence"`
	Confidence  Confidence `json:"confidence"`
	CreatedAt   time.Time  `json:"createdAt"`
}

type RouteSuggestion struct {
	ID            string     `json:"id"`
	ServiceID     string     `json:"serviceId"`
	ObservationID string     `json:"observationId"`
	RouteHost     string     `json:"routeHost"`
	FrontendPort  int        `json:"frontendPort,omitempty"`
	BackendHost   string     `json:"backendHost"`
	BackendPort   int        `json:"backendPort"`
	State         RouteState `json:"state"`
	Reason        string     `json:"reason"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

type ConfirmedRoute struct {
	ID           string     `json:"id"`
	ServiceID    string     `json:"serviceId"`
	ServiceName  string     `json:"serviceName,omitempty"`
	ProjectName  string     `json:"projectName,omitempty"`
	Protocol     Protocol   `json:"protocol"`
	RouteHost    string     `json:"routeHost"`
	FrontendPort int        `json:"frontendPort,omitempty"`
	BackendHost  string     `json:"backendHost"`
	BackendPort  int        `json:"backendPort"`
	State        RouteState `json:"state"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type ScanRun struct {
	ID           string    `json:"id"`
	StartedAt    time.Time `json:"startedAt"`
	FinishedAt   time.Time `json:"finishedAt"`
	Status       string    `json:"status"`
	Summary      string    `json:"summary"`
	Observations int       `json:"observations"`
	Suggestions  int       `json:"suggestions"`
}

type ScanResult struct {
	Run          ScanRun           `json:"run"`
	Projects     []Project         `json:"projects"`
	Services     []Service         `json:"services"`
	Observations []Observation     `json:"observations"`
	Suggestions  []RouteSuggestion `json:"suggestions"`
	Warnings     []string          `json:"warnings,omitempty"`
}

type ServiceSummary struct {
	RouteID      string     `json:"routeId"`
	ProjectName  string     `json:"projectName"`
	ServiceName  string     `json:"serviceName"`
	Protocol     Protocol   `json:"protocol"`
	RouteHost    string     `json:"routeHost,omitempty"`
	FrontendPort int        `json:"frontendPort,omitempty"`
	BackendHost  string     `json:"backendHost,omitempty"`
	BackendPort  int        `json:"backendPort,omitempty"`
	State        RouteState `json:"state"`
	Confidence   Confidence `json:"confidence"`
	Source       string     `json:"source"`
}

func DeterministicID(kind string, parts ...string) string {
	hash := sha256.Sum256([]byte(kind + ":" + strings.Join(parts, "\x00")))
	return kind + "_" + hex.EncodeToString(hash[:8])
}
