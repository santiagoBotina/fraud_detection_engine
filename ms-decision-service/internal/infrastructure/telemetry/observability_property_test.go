package telemetry

// Feature: grafana-observability, Property 3: All configurable ports use environment variables with defaults

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"

	"gopkg.in/yaml.v3"
	"pgregory.net/rapid"
)

// dockerCompose represents the top-level structure of a docker-compose.yml file.
type dockerCompose struct {
	Services map[string]any `yaml:"services"`
}

func loadDockerCompose(t *testing.T) dockerCompose {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to determine test file path")
	}

	composePath := filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "docker-compose.yml")

	data, err := os.ReadFile(composePath)
	if err != nil {
		t.Fatalf("failed to read docker-compose.yml: %v", err)
	}

	var dc dockerCompose
	if err := yaml.Unmarshal(data, &dc); err != nil {
		t.Fatalf("failed to parse docker-compose.yml: %v", err)
	}

	return dc
}

// serviceConfig represents the relevant fields of a docker-compose service definition.
type serviceConfig struct {
	Ports []string `yaml:"ports"`
}

// envVarWithDefaultPattern matches the ${VAR:-default} syntax in port mappings.
var envVarWithDefaultPattern = regexp.MustCompile(`\$\{[A-Z_]+:-[^}]+\}`)

// configurablePortContainers maps observability containers that expose configurable
// host ports to their expected port mapping pattern using ${VAR:-default} syntax.
var configurablePortContainers = map[string]string{
	"grafana":  "${GRAFANA_PORT:-3003}:3000",
	"cadvisor": "${CADVISOR_PORT:-8081}:8080",
}

// **Validates: Requirements 8.4**
func TestProperty_ConfigurablePortsUseEnvVarsWithDefaults(t *testing.T) {
	dc := loadDockerCompose(t)

	containerNames := make([]string, 0, len(configurablePortContainers))
	for name := range configurablePortContainers {
		containerNames = append(containerNames, name)
	}

	rapid.Check(t, func(t *rapid.T) {
		idx := rapid.IntRange(0, len(containerNames)-1).Draw(t, "containerIndex")
		container := containerNames[idx]
		expectedPort := configurablePortContainers[container]

		svcRaw, exists := dc.Services[container]
		if !exists {
			t.Fatalf("container %q is not defined in docker-compose.yml", container)
		}

		// Re-marshal and unmarshal to extract ports.
		rawBytes, err := yaml.Marshal(svcRaw)
		if err != nil {
			t.Fatalf("failed to marshal service %q: %v", container, err)
		}

		var svc serviceConfig
		if err := yaml.Unmarshal(rawBytes, &svc); err != nil {
			t.Fatalf("failed to parse service config for %q: %v", container, err)
		}

		if len(svc.Ports) == 0 {
			t.Fatalf("container %q has no port mappings defined", container)
		}

		// Verify at least one port mapping uses ${VAR:-default} syntax.
		foundEnvVar := false
		for _, port := range svc.Ports {
			if envVarWithDefaultPattern.MatchString(port) {
				foundEnvVar = true
				break
			}
		}
		if !foundEnvVar {
			t.Fatalf("container %q port mappings %v do not use ${VAR:-default} syntax", container, svc.Ports)
		}

		// Verify the exact expected port mapping is present.
		foundExpected := false
		for _, port := range svc.Ports {
			if port == expectedPort {
				foundExpected = true
				break
			}
		}
		if !foundExpected {
			t.Fatalf("container %q expected port mapping %q not found in %v", container, expectedPort, svc.Ports)
		}
	})
}
