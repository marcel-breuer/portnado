package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/marcel-breuer/portnado/internal/config"
	dockerdisc "github.com/marcel-breuer/portnado/internal/discovery/docker"
	nativedisc "github.com/marcel-breuer/portnado/internal/discovery/native"
	"github.com/marcel-breuer/portnado/internal/domain"
	"github.com/marcel-breuer/portnado/internal/paths"
	"github.com/marcel-breuer/portnado/internal/persistence"
)

type Detector interface {
	Discover(context.Context) ([]domain.Observation, []string)
}

type Service struct {
	store     *persistence.Store
	detectors []Detector
}

func NewService(store *persistence.Store, detectors ...Detector) *Service {
	if len(detectors) == 0 {
		detectors = []Detector{
			dockerdisc.Detector{},
			nativedisc.Detector{},
		}
	}
	return &Service{store: store, detectors: detectors}
}

func (s *Service) Scan(ctx context.Context, root string) (domain.ScanResult, error) {
	startedAt := time.Now().UTC()
	if root == "" {
		if cwd, err := os.Getwd(); err == nil {
			root = cwd
		}
	}
	repoConfig, hasRepoConfig, err := loadRepositoryConfig(root)
	if err != nil {
		return domain.ScanResult{}, err
	}
	localOverride, hasLocalOverride, err := loadLocalOverride(root, repoConfig, hasRepoConfig)
	if err != nil {
		return domain.ScanResult{}, err
	}

	var observations []domain.Observation
	var warnings []string
	for _, detector := range s.detectors {
		detected, detectorWarnings := detector.Discover(ctx)
		observations = append(observations, detected...)
		warnings = append(warnings, detectorWarnings...)
	}
	observations = applyRepositoryConfig(observations, repoConfig, hasRepoConfig, root)
	observations = append(observations, manualRepositoryObservations(repoConfig, hasRepoConfig, root, observations)...)
	suggestions := suggestionsForObservations(observations, repoConfig, hasRepoConfig, localOverride, hasLocalOverride)
	projects, services := collectEntities(observations)
	projects, services = mergeRepositoryEntities(projects, services, repoConfig, hasRepoConfig, root)
	finishedAt := time.Now().UTC()
	result := domain.ScanResult{
		Run: domain.ScanRun{
			ID:           domain.DeterministicID("scan", startedAt.Format(time.RFC3339Nano)),
			StartedAt:    startedAt,
			FinishedAt:   finishedAt,
			Status:       "completed",
			Summary:      fmt.Sprintf("%d observations, %d suggestions", len(observations), len(suggestions)),
			Observations: len(observations),
			Suggestions:  len(suggestions),
		},
		Projects:     projects,
		Services:     services,
		Observations: observations,
		Suggestions:  suggestions,
		Warnings:     warnings,
	}
	if err := s.store.SaveScanResult(ctx, result); err != nil {
		return domain.ScanResult{}, err
	}
	return result, nil
}

func mergeRepositoryEntities(projects []domain.Project, services []domain.Service, repoConfig config.RepositoryConfig, hasRepoConfig bool, root string) ([]domain.Project, []domain.Service) {
	if !hasRepoConfig {
		return projects, services
	}
	now := time.Now().UTC()
	projectID := domain.DeterministicID("project", "repository", root, repoConfig.Project.Name)
	project := domain.Project{
		ID:        projectID,
		Name:      repoConfig.Project.Name,
		RootPath:  root,
		Source:    "repository",
		CreatedAt: now,
		UpdatedAt: now,
	}
	projectByID := make(map[string]domain.Project)
	for _, existing := range projects {
		projectByID[existing.ID] = existing
	}
	projectByID[project.ID] = project
	projects = projects[:0]
	for _, existing := range projectByID {
		projects = append(projects, existing)
	}

	serviceByID := make(map[string]domain.Service)
	for _, existing := range services {
		serviceByID[existing.ID] = existing
	}
	for name, repoService := range repoConfig.Services {
		serviceID := domain.DeterministicID("service", projectID, name)
		if existing, ok := serviceByID[serviceID]; ok {
			existing.Protocol = repoService.Protocol
			existing.Description = repoService.Description
			existing.Source = mergeSource(existing.Source, "repository")
			existing.UpdatedAt = now
			serviceByID[serviceID] = existing
			continue
		}
		serviceByID[serviceID] = domain.Service{
			ID:          serviceID,
			ProjectID:   projectID,
			Name:        name,
			Protocol:    repoService.Protocol,
			Description: repoService.Description,
			Source:      "repository",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	}
	services = services[:0]
	for _, existing := range serviceByID {
		services = append(services, existing)
	}
	return projects, services
}

func manualRepositoryObservations(repoConfig config.RepositoryConfig, hasRepoConfig bool, root string, existing []domain.Observation) []domain.Observation {
	if !hasRepoConfig {
		return nil
	}
	now := time.Now().UTC()
	projectID := domain.DeterministicID("project", "repository", root, repoConfig.Project.Name)
	project := domain.Project{
		ID:        projectID,
		Name:      repoConfig.Project.Name,
		RootPath:  root,
		Source:    "repository",
		CreatedAt: now,
		UpdatedAt: now,
	}
	var observations []domain.Observation
	observedServices := make(map[string]bool)
	for _, observation := range existing {
		if observation.Project.ID == projectID {
			observedServices[observation.Service.Name] = true
		}
	}
	for name, repoService := range repoConfig.Services {
		if repoService.Target.Discovery != "manual" || repoService.Target.PreferredPort == 0 {
			continue
		}
		if observedServices[name] {
			continue
		}
		service := domain.Service{
			ID:          domain.DeterministicID("service", projectID, name),
			ProjectID:   projectID,
			Name:        name,
			Protocol:    repoService.Protocol,
			Description: repoService.Description,
			Source:      "repository",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		observations = append(observations, domain.Observation{
			ID:          domain.DeterministicID("observation", projectID, service.ID, "manual", fmt.Sprint(repoService.Target.PreferredPort)),
			Project:     project,
			Service:     service,
			Runtime:     "manual",
			Protocol:    repoService.Protocol,
			BackendHost: "127.0.0.1",
			BackendPort: repoService.Target.PreferredPort,
			Evidence: []domain.Evidence{{
				Source:  "repository",
				Summary: fmt.Sprintf(".portnado.yml manually maps %s to 127.0.0.1:%d", name, repoService.Target.PreferredPort),
			}},
			Confidence: domain.ConfidenceHigh,
			CreatedAt:  now,
		})
	}
	return observations
}

func (s *Service) List(ctx context.Context) ([]domain.ServiceSummary, error) {
	return s.store.ListServices(ctx)
}

func (s *Service) ApproveRoute(ctx context.Context, suggestionID string) (domain.ConfirmedRoute, error) {
	return s.store.ApproveSuggestion(ctx, suggestionID)
}

func (s *Service) EnableRoute(ctx context.Context, routeID string) (domain.ConfirmedRoute, error) {
	return s.store.SetRouteState(ctx, routeID, domain.RouteStateActive)
}

func (s *Service) DisableRoute(ctx context.Context, routeID string) (domain.ConfirmedRoute, error) {
	return s.store.SetRouteState(ctx, routeID, domain.RouteStateDisabled)
}

func (s *Service) ActiveRoutes(ctx context.Context) ([]domain.ConfirmedRoute, error) {
	return s.store.ActiveRoutes(ctx)
}

func (s *Service) Routes(ctx context.Context) ([]domain.ConfirmedRoute, error) {
	return s.store.Routes(ctx)
}

func loadRepositoryConfig(root string) (config.RepositoryConfig, bool, error) {
	if root == "" {
		return config.RepositoryConfig{}, false, nil
	}
	path := filepath.Join(root, config.RepositoryFileName)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return config.RepositoryConfig{}, false, nil
		}
		return config.RepositoryConfig{}, false, fmt.Errorf("inspect repository config: %w", err)
	}
	cfg, err := config.LoadRepositoryFile(path)
	if err != nil {
		return config.RepositoryConfig{}, false, err
	}
	return cfg, true, nil
}

func loadLocalOverride(root string, repoConfig config.RepositoryConfig, hasRepoConfig bool) (config.LocalOverride, bool, error) {
	if !hasRepoConfig {
		return config.LocalOverride{}, false, nil
	}
	projectID := domain.DeterministicID("project", "repository", root, repoConfig.Project.Name)
	path, err := paths.ProjectOverridePath(projectID)
	if err != nil {
		return config.LocalOverride{}, false, err
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return config.LocalOverride{}, false, nil
		}
		return config.LocalOverride{}, false, fmt.Errorf("inspect local override: %w", err)
	}
	override, err := config.LoadLocalOverrideFile(path)
	if err != nil {
		return config.LocalOverride{}, false, err
	}
	if _, err := config.EffectiveServices(repoConfig, override); err != nil {
		return config.LocalOverride{}, false, err
	}
	return override, true, nil
}

func applyRepositoryConfig(observations []domain.Observation, repoConfig config.RepositoryConfig, hasRepoConfig bool, root string) []domain.Observation {
	if !hasRepoConfig {
		return observations
	}
	now := time.Now().UTC()
	projectID := domain.DeterministicID("project", "repository", root, repoConfig.Project.Name)
	project := domain.Project{
		ID:        projectID,
		Name:      repoConfig.Project.Name,
		RootPath:  root,
		Source:    "repository",
		CreatedAt: now,
		UpdatedAt: now,
	}
	for i := range observations {
		if repoService, ok := repoConfig.Services[observations[i].Service.Name]; ok {
			observations[i].Project = project
			observations[i].Service.ProjectID = projectID
			observations[i].Service.Protocol = repoService.Protocol
			observations[i].Service.Source = mergeSource(observations[i].Service.Source, "repository")
			observations[i].Protocol = repoService.Protocol
			observations[i].Confidence = domain.ConfidenceHigh
			observations[i].Evidence = append(observations[i].Evidence, domain.Evidence{
				Source:  "repository",
				Summary: fmt.Sprintf(".portnado.yml declares service %s", observations[i].Service.Name),
			})
			observations[i].Service.ID = domain.DeterministicID("service", projectID, observations[i].Service.Name)
			observations[i].ID = domain.DeterministicID("observation", projectID, observations[i].Service.ID, observations[i].BackendHost, fmt.Sprint(observations[i].BackendPort))
		}
	}
	return observations
}

func suggestionsForObservations(observations []domain.Observation, repoConfig config.RepositoryConfig, hasRepoConfig bool, localOverride config.LocalOverride, hasLocalOverride bool) []domain.RouteSuggestion {
	now := time.Now().UTC()
	suggestions := make([]domain.RouteSuggestion, 0, len(observations))
	for _, observation := range observations {
		routeHost := fmt.Sprintf("%s.%s.localhost", observation.Service.Name, observation.Project.Name)
		frontendPort := 0
		enabled := true
		if hasRepoConfig {
			if repoService, ok := repoConfig.Services[observation.Service.Name]; ok {
				if repoService.Enabled != nil {
					enabled = *repoService.Enabled
				}
				if repoService.Route.Host != "" {
					routeHost = repoService.Route.Host
				}
				if repoService.Route.PreferredPort != 0 {
					frontendPort = repoService.Route.PreferredPort
				}
			}
		}
		if hasLocalOverride {
			if overrideService, ok := localOverride.Services[observation.Service.Name]; ok {
				if overrideService.ScanExcluded {
					continue
				}
				if overrideService.Enabled != nil {
					enabled = *overrideService.Enabled
				}
				if overrideService.Route != nil {
					if overrideService.Route.Host != "" {
						routeHost = overrideService.Route.Host
					}
					if overrideService.Route.PreferredPort != 0 {
						frontendPort = overrideService.Route.PreferredPort
					}
				}
			}
		}
		if !enabled {
			continue
		}
		if observation.Protocol == domain.ProtocolTCP && frontendPort == 0 {
			frontendPort = deterministicTCPPort(observation.Service.ID)
		}
		suggestionID := domain.DeterministicID("suggestion", observation.Service.ID, routeHost)
		suggestions = append(suggestions, domain.RouteSuggestion{
			ID:            suggestionID,
			ServiceID:     observation.Service.ID,
			ObservationID: observation.ID,
			RouteHost:     routeHost,
			FrontendPort:  frontendPort,
			BackendHost:   observation.BackendHost,
			BackendPort:   observation.BackendPort,
			State:         domain.RouteStateAwaitingApproval,
			Reason:        "New route suggestion from read-only discovery.",
			CreatedAt:     now,
			UpdatedAt:     now,
		})
	}
	return suggestions
}

func collectEntities(observations []domain.Observation) ([]domain.Project, []domain.Service) {
	projectByID := make(map[string]domain.Project)
	serviceByID := make(map[string]domain.Service)
	for _, observation := range observations {
		projectByID[observation.Project.ID] = observation.Project
		serviceByID[observation.Service.ID] = observation.Service
	}
	projects := make([]domain.Project, 0, len(projectByID))
	for _, project := range projectByID {
		projects = append(projects, project)
	}
	services := make([]domain.Service, 0, len(serviceByID))
	for _, service := range serviceByID {
		services = append(services, service)
	}
	return projects, services
}

func deterministicTCPPort(serviceID string) int {
	sum := 0
	for _, r := range serviceID {
		sum += int(r)
	}
	return 15400 + (sum % 600)
}

func mergeSource(existing, next string) string {
	if existing == "" {
		return next
	}
	if existing == next {
		return existing
	}
	for _, part := range strings.Split(existing, "+") {
		if part == next {
			return existing
		}
	}
	return existing + "+" + next
}
