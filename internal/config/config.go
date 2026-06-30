package config

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/marcel-breuer/portnado/internal/domain"
	"gopkg.in/yaml.v3"
)

const RepositoryFileName = ".portnado.yml"

var dnsLabelPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

type RepositoryConfig struct {
	Version  int                      `yaml:"version"`
	Project  ProjectConfig            `yaml:"project"`
	Services map[string]ServiceConfig `yaml:"services"`
}

type ProjectConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
}

type ServiceConfig struct {
	Protocol    domain.Protocol `yaml:"protocol"`
	Description string          `yaml:"description,omitempty"`
	Enabled     *bool           `yaml:"enabled,omitempty"`
	Route       RouteConfig     `yaml:"route"`
	Target      TargetConfig    `yaml:"target"`
}

type RouteConfig struct {
	Host          string `yaml:"host"`
	PreferredPort int    `yaml:"preferredPort,omitempty"`
}

type TargetConfig struct {
	Discovery     string `yaml:"discovery"`
	PreferredPort int    `yaml:"preferredPort,omitempty"`
	Service       string `yaml:"service,omitempty"`
	ContainerPort int    `yaml:"containerPort,omitempty"`
}

type LocalOverride struct {
	Version  int                             `yaml:"version"`
	Project  LocalOverrideProject            `yaml:"project,omitempty"`
	Services map[string]LocalServiceOverride `yaml:"services,omitempty"`
}

type LocalOverrideProject struct {
	RootPath string `yaml:"rootPath,omitempty"`
}

type LocalServiceOverride struct {
	Enabled      *bool        `yaml:"enabled,omitempty"`
	Route        *RouteConfig `yaml:"route,omitempty"`
	SelectedPID  int          `yaml:"selectedPid,omitempty"`
	ScanExcluded bool         `yaml:"scanExcluded,omitempty"`
}

type EffectiveService struct {
	Name         string
	Protocol     domain.Protocol
	RouteHost    string
	FrontendPort int
	Target       TargetConfig
	Enabled      bool
	Sources      map[string]string
}

func LoadRepositoryFile(path string) (RepositoryConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return RepositoryConfig{}, fmt.Errorf("read repository config: %w", err)
	}
	return ParseRepository(data)
}

func ParseRepository(data []byte) (RepositoryConfig, error) {
	var cfg RepositoryConfig
	if err := decodeStrict(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, cfg.Validate()
}

func ParseLocalOverride(data []byte) (LocalOverride, error) {
	var override LocalOverride
	if err := decodeStrict(data, &override); err != nil {
		return override, err
	}
	if override.Version != 1 {
		return override, fmt.Errorf("local override version must be 1")
	}
	return override, nil
}

func LoadLocalOverrideFile(path string) (LocalOverride, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return LocalOverride{}, fmt.Errorf("read local override: %w", err)
	}
	return ParseLocalOverride(data)
}

func (c RepositoryConfig) Validate() error {
	if c.Version != 1 {
		return fmt.Errorf("config version must be 1")
	}
	projectName := NormalizeName(c.Project.Name)
	if projectName == "" || !dnsLabelPattern.MatchString(projectName) {
		return fmt.Errorf("project.name must be a DNS-safe label")
	}
	if len(c.Services) == 0 {
		return fmt.Errorf("services must contain at least one service")
	}

	names := make([]string, 0, len(c.Services))
	for name := range c.Services {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		service := c.Services[name]
		if name != NormalizeName(name) || !dnsLabelPattern.MatchString(name) {
			return fmt.Errorf("service %q must be a lowercase DNS-safe label", name)
		}
		if service.Protocol != domain.ProtocolHTTP && service.Protocol != domain.ProtocolTCP {
			return fmt.Errorf("service %q protocol must be http or tcp", name)
		}
		if err := ValidateLocalhost(service.Route.Host); err != nil {
			return fmt.Errorf("service %q route.host: %w", name, err)
		}
		if service.Protocol == domain.ProtocolTCP && service.Route.PreferredPort != 0 {
			if err := validatePort(service.Route.PreferredPort); err != nil {
				return fmt.Errorf("service %q route.preferredPort: %w", name, err)
			}
		}
		if service.Target.PreferredPort != 0 {
			if err := validatePort(service.Target.PreferredPort); err != nil {
				return fmt.Errorf("service %q target.preferredPort: %w", name, err)
			}
		}
		if service.Target.ContainerPort != 0 {
			if err := validatePort(service.Target.ContainerPort); err != nil {
				return fmt.Errorf("service %q target.containerPort: %w", name, err)
			}
		}
		switch service.Target.Discovery {
		case "auto", "docker-compose", "native", "manual":
		default:
			return fmt.Errorf("service %q target.discovery must be auto, docker-compose, native, or manual", name)
		}
	}
	return nil
}

func EffectiveServices(repo RepositoryConfig, override LocalOverride) ([]EffectiveService, error) {
	if err := repo.Validate(); err != nil {
		return nil, err
	}
	services := make([]EffectiveService, 0, len(repo.Services))
	names := make([]string, 0, len(repo.Services))
	for name := range repo.Services {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		repoService := repo.Services[name]
		enabled := true
		if repoService.Enabled != nil {
			enabled = *repoService.Enabled
		}
		routeHost := repoService.Route.Host
		frontendPort := repoService.Route.PreferredPort
		sources := map[string]string{
			"protocol": "repository",
			"route":    "repository",
			"enabled":  "repository",
			"target":   "repository",
		}

		if overrideService, ok := override.Services[name]; ok {
			if overrideService.Enabled != nil {
				enabled = *overrideService.Enabled
				sources["enabled"] = "local_override"
			}
			if overrideService.Route != nil {
				if overrideService.Route.Host != "" {
					if err := ValidateLocalhost(overrideService.Route.Host); err != nil {
						return nil, fmt.Errorf("override service %q route.host: %w", name, err)
					}
					routeHost = overrideService.Route.Host
					sources["route"] = "local_override"
				}
				if overrideService.Route.PreferredPort != 0 {
					if err := validatePort(overrideService.Route.PreferredPort); err != nil {
						return nil, fmt.Errorf("override service %q route.preferredPort: %w", name, err)
					}
					frontendPort = overrideService.Route.PreferredPort
					sources["route"] = "local_override"
				}
			}
		}

		services = append(services, EffectiveService{
			Name:         name,
			Protocol:     repoService.Protocol,
			RouteHost:    routeHost,
			FrontendPort: frontendPort,
			Target:       repoService.Target,
			Enabled:      enabled,
			Sources:      sources,
		})
	}
	return services, nil
}

func NormalizeName(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, "_", "-")
	return value
}

func ValidateLocalhost(host string) error {
	host = strings.TrimSpace(strings.ToLower(host))
	if host == "" {
		return fmt.Errorf("host is required")
	}
	if strings.ContainsAny(host, "/\\\x00\r\n\t ") {
		return fmt.Errorf("host contains invalid characters")
	}
	if strings.Contains(host, "*") {
		return fmt.Errorf("wildcard hosts are not allowed")
	}
	if !strings.HasSuffix(host, ".localhost") {
		return fmt.Errorf("host must end with .localhost")
	}
	if ip := net.ParseIP(strings.TrimSuffix(host, ".localhost")); ip != nil {
		return fmt.Errorf("host labels must not be IP addresses")
	}
	for _, label := range strings.Split(host, ".") {
		if label == "localhost" {
			continue
		}
		if !dnsLabelPattern.MatchString(label) {
			return fmt.Errorf("invalid DNS label %q", label)
		}
	}
	return nil
}

func decodeStrict(data []byte, target any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode YAML: %w", err)
	}
	return nil
}

func validatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	return nil
}
