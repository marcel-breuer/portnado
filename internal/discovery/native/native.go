package native

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/marcel-breuer/portnado/internal/discovery/project"
	"github.com/marcel-breuer/portnado/internal/discovery/runtime"
	"github.com/marcel-breuer/portnado/internal/domain"
)

type Runner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, name, args...)
	return command.Output()
}

type Detector struct {
	Runner Runner
}

type processListener struct {
	PID     int
	Command string
	Address string
	Port    int
	Cwd     string
}

func (d Detector) Discover(ctx context.Context) ([]domain.Observation, []string) {
	runner := d.Runner
	if runner == nil {
		runner = ExecRunner{}
	}
	raw, err := runner.Run(ctx, "lsof", "-nP", "-iTCP", "-sTCP:LISTEN", "-FpPnac")
	if err != nil {
		return nil, []string{"Native process discovery unavailable: lsof failed"}
	}
	listeners, err := parseListeners(raw)
	if err != nil {
		return nil, []string{fmt.Sprintf("Native process discovery unavailable: %v", err)}
	}
	now := time.Now().UTC()
	var observations []domain.Observation
	for _, listener := range listeners {
		if !isLoopback(listener.Address) {
			continue
		}
		listener.Cwd = d.cwd(ctx, runner, listener.PID)
		markers, _ := project.MarkerFiles(listener.Cwd)
		classification := runtime.Classify(listener.Command, nil, markers)
		projectRoot := listener.Cwd
		if gitRoot, ok := project.GitRoot(ctx, listener.Cwd); ok {
			projectRoot = gitRoot
		}
		projectName := "local"
		if projectRoot != "" {
			projectName = project.ProjectNameFromRoot(projectRoot)
		}
		proj := domain.Project{
			ID:        domain.DeterministicID("project", "native", projectRoot, projectName),
			Name:      projectName,
			RootPath:  projectRoot,
			Source:    "native",
			CreatedAt: now,
			UpdatedAt: now,
		}
		service := domain.Service{
			ID:        domain.DeterministicID("service", proj.ID, classification.Service),
			ProjectID: proj.ID,
			Name:      classification.Service,
			Protocol:  domain.ProtocolHTTP,
			Source:    "native",
			CreatedAt: now,
			UpdatedAt: now,
		}
		confidence := domain.ConfidenceLow
		if classification.Confidence == "high" {
			confidence = domain.ConfidenceHigh
		}
		if classification.Confidence == "medium" {
			confidence = domain.ConfidenceMedium
		}
		evidence := []domain.Evidence{
			{Source: "native", Summary: fmt.Sprintf("Process %d listens on %s:%d", listener.PID, listener.Address, listener.Port)},
		}
		for _, item := range classification.Evidence {
			evidence = append(evidence, domain.Evidence{Source: "runtime", Summary: item})
		}
		observations = append(observations, domain.Observation{
			ID:          domain.DeterministicID("observation", proj.ID, service.ID, strconv.Itoa(listener.PID), strconv.Itoa(listener.Port)),
			Project:     proj,
			Service:     service,
			Runtime:     classification.Runtime,
			Protocol:    domain.ProtocolHTTP,
			BackendHost: listener.Address,
			BackendPort: listener.Port,
			Evidence:    evidence,
			Confidence:  confidence,
			CreatedAt:   now,
		})
	}
	return observations, nil
}

func (d Detector) cwd(ctx context.Context, runner Runner, pid int) string {
	raw, err := runner.Run(ctx, "lsof", "-a", "-p", strconv.Itoa(pid), "-d", "cwd", "-Fn")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(line, "n") {
			return strings.TrimSpace(strings.TrimPrefix(line, "n"))
		}
	}
	return ""
}

func parseListeners(data []byte) ([]processListener, error) {
	var listeners []processListener
	var current processListener
	for _, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}
		switch line[0] {
		case 'p':
			pid, err := strconv.Atoi(strings.TrimPrefix(line, "p"))
			if err != nil {
				return nil, fmt.Errorf("parse pid %q: %w", line, err)
			}
			current = processListener{PID: pid}
		case 'c':
			current.Command = strings.TrimPrefix(line, "c")
		case 'n':
			address, port, ok := parseEndpoint(strings.TrimPrefix(line, "n"))
			if ok && current.PID != 0 {
				listener := current
				listener.Address = address
				listener.Port = port
				listeners = append(listeners, listener)
			}
		}
	}
	return listeners, nil
}

func parseEndpoint(value string) (string, int, bool) {
	value = strings.TrimSpace(value)
	if strings.Contains(value, "->") {
		return "", 0, false
	}
	host, portText, err := net.SplitHostPort(value)
	if err != nil {
		index := strings.LastIndex(value, ":")
		if index < 0 {
			return "", 0, false
		}
		host = value[:index]
		portText = value[index+1:]
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port <= 0 || port > 65535 {
		return "", 0, false
	}
	host = strings.Trim(host, "[]")
	if host == "localhost" || host == "" {
		host = "127.0.0.1"
	}
	return host, port, true
}

func isLoopback(host string) bool {
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback() && ip.To4() != nil
}
