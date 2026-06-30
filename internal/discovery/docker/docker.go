package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

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

type composeProject struct {
	Name string `json:"Name"`
}

type composeService struct {
	ID         string      `json:"ID"`
	Name       string      `json:"Name"`
	Service    string      `json:"Service"`
	Project    string      `json:"Project"`
	State      string      `json:"State"`
	Publishers []publisher `json:"Publishers"`
}

type publisher struct {
	URL           string `json:"URL"`
	TargetPort    int    `json:"TargetPort"`
	PublishedPort int    `json:"PublishedPort"`
	Protocol      string `json:"Protocol"`
}

func (d Detector) Discover(ctx context.Context) ([]domain.Observation, []string) {
	runner := d.Runner
	if runner == nil {
		runner = ExecRunner{}
	}
	projectsRaw, err := runner.Run(ctx, "docker", "compose", "ls", "--format", "json")
	if err != nil {
		return nil, []string{"Docker discovery unavailable: docker compose ls failed"}
	}
	projects, err := parseProjects(projectsRaw)
	if err != nil {
		return nil, []string{fmt.Sprintf("Docker discovery unavailable: %v", err)}
	}
	var observations []domain.Observation
	var warnings []string
	for _, project := range projects {
		if project.Name == "" {
			continue
		}
		raw, err := runner.Run(ctx, "docker", "compose", "-p", project.Name, "ps", "--format", "json")
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Docker Compose project %s could not be inspected", project.Name))
			continue
		}
		services, err := parseServices(raw)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Docker Compose project %s returned unreadable service data", project.Name))
			continue
		}
		now := time.Now().UTC()
		for _, service := range services {
			observations = append(observations, serviceObservations(project.Name, service, now)...)
		}
	}
	return observations, warnings
}

func parseProjects(data []byte) ([]composeProject, error) {
	var projects []composeProject
	if err := json.Unmarshal(bytes.TrimSpace(data), &projects); err != nil {
		return nil, fmt.Errorf("parse docker compose project list: %w", err)
	}
	return projects, nil
}

func parseServices(data []byte) ([]composeService, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, nil
	}
	var services []composeService
	if data[0] == '[' {
		if err := json.Unmarshal(data, &services); err != nil {
			return nil, fmt.Errorf("parse docker compose services: %w", err)
		}
		return services, nil
	}
	for _, line := range bytes.Split(data, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var service composeService
		if err := json.Unmarshal(line, &service); err != nil {
			return nil, fmt.Errorf("parse docker compose service line: %w", err)
		}
		services = append(services, service)
	}
	return services, nil
}

func serviceObservations(projectName string, service composeService, now time.Time) []domain.Observation {
	serviceName := normalizeServiceName(firstNonEmpty(service.Service, service.Name))
	if serviceName == "" {
		return nil
	}
	var observations []domain.Observation
	for _, published := range service.Publishers {
		if strings.ToLower(published.Protocol) != "tcp" {
			continue
		}
		host, port, ok := hostPort(published.URL, published.PublishedPort)
		if !ok || port == 0 {
			continue
		}
		protocol := domain.ProtocolHTTP
		if likelyTCPService(serviceName, published.TargetPort) {
			protocol = domain.ProtocolTCP
		}
		project := domain.Project{
			ID:        domain.DeterministicID("project", "docker", projectName),
			Name:      normalizeServiceName(projectName),
			Source:    "docker-compose",
			CreatedAt: now,
			UpdatedAt: now,
		}
		svc := domain.Service{
			ID:        domain.DeterministicID("service", project.ID, serviceName),
			ProjectID: project.ID,
			Name:      serviceName,
			Protocol:  protocol,
			Source:    "docker-compose",
			CreatedAt: now,
			UpdatedAt: now,
		}
		observations = append(observations, domain.Observation{
			ID:          domain.DeterministicID("observation", project.ID, svc.ID, strconv.Itoa(port)),
			Project:     project,
			Service:     svc,
			Runtime:     "docker-compose",
			Protocol:    protocol,
			BackendHost: host,
			BackendPort: port,
			Confidence:  domain.ConfidenceHigh,
			Evidence: []domain.Evidence{
				{Source: "docker", Summary: fmt.Sprintf("Compose service %s published host port %d", serviceName, port)},
			},
			CreatedAt: now,
		})
	}
	return observations
}

func hostPort(url string, fallbackPort int) (string, int, bool) {
	url = strings.TrimSpace(url)
	if url == "" {
		if fallbackPort == 0 {
			return "", 0, false
		}
		return "127.0.0.1", fallbackPort, true
	}
	host := "127.0.0.1"
	if strings.Contains(url, ":") {
		parts := strings.Split(url, ":")
		host = strings.Trim(parts[0], "[]")
		port, err := strconv.Atoi(parts[len(parts)-1])
		if err == nil {
			return normalizeLoopback(host), port, true
		}
	}
	portPattern := regexp.MustCompile(`(\d+)$`)
	match := portPattern.FindStringSubmatch(url)
	if len(match) == 2 {
		port, err := strconv.Atoi(match[1])
		if err == nil {
			return normalizeLoopback(host), port, true
		}
	}
	if fallbackPort > 0 {
		return normalizeLoopback(host), fallbackPort, true
	}
	return "", 0, false
}

func normalizeLoopback(host string) string {
	switch host {
	case "", "0.0.0.0", "::", "::1", "localhost":
		return "127.0.0.1"
	default:
		return host
	}
}

func likelyTCPService(name string, targetPort int) bool {
	switch name {
	case "db", "database", "postgres", "postgresql", "mysql", "redis":
		return true
	}
	return targetPort == 5432 || targetPort == 3306 || targetPort == 6379
}

func normalizeServiceName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "_", "-")
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

var ErrDockerUnavailable = errors.New("docker unavailable")
