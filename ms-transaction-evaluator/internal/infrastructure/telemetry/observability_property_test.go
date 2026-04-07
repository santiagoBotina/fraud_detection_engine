package telemetry

// Feature: grafana-observability, Property 1: All observability containers are defined in docker-compose.yml

import (
	"encoding/json"
	"os"
	"path/filepath"
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
	Networks []string `yaml:"networks"`
}

// **Validates: Requirements 1.1, 2.1, 2.5, 3.1, 3.2, 5.1, 6.1, 8.1**
func TestProperty_AllObservabilityContainersDefined(t *testing.T) {
	requiredContainers := []string{
		"grafana",
		"loki",
		"promtail",
		"tempo",
		"otel-collector",
		"prometheus",
		"kafka-exporter",
	}

	dc := loadDockerCompose(t)

	rapid.Check(t, func(t *rapid.T) {
		idx := rapid.IntRange(0, len(requiredContainers)-1).Draw(t, "containerIndex")
		container := requiredContainers[idx]

		if _, exists := dc.Services[container]; !exists {
			t.Fatalf("observability container %q is not defined as a service in docker-compose.yml", container)
		}
	})
}

// Feature: grafana-observability, Property 2: All observability containers are on the local-network

// **Validates: Requirements 8.2**
func TestProperty_AllObservabilityContainersOnLocalNetwork(t *testing.T) {
	observabilityContainers := []string{
		"grafana",
		"loki",
		"promtail",
		"tempo",
		"otel-collector",
		"prometheus",
		"kafka-exporter",
	}

	dc := loadDockerCompose(t)

	rapid.Check(t, func(t *rapid.T) {
		idx := rapid.IntRange(0, len(observabilityContainers)-1).Draw(t, "containerIndex")
		container := observabilityContainers[idx]

		svcRaw, exists := dc.Services[container]
		if !exists {
			t.Fatalf("observability container %q is not defined in docker-compose.yml", container)
		}

		// Re-marshal the service value and unmarshal into serviceConfig to extract networks.
		rawBytes, err := yaml.Marshal(svcRaw)
		if err != nil {
			t.Fatalf("failed to marshal service %q: %v", container, err)
		}

		var svc serviceConfig
		if err := yaml.Unmarshal(rawBytes, &svc); err != nil {
			t.Fatalf("failed to parse service config for %q: %v", container, err)
		}

		found := false
		for _, net := range svc.Networks {
			if net == "local-network" {
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("observability container %q does not include local-network in its networks (got %v)", container, svc.Networks)
		}
	})
}

// Feature: grafana-observability, Property 4: All provisioned dashboards reference a valid datasource

// dashboardJSON represents the top-level structure of a Grafana dashboard JSON file.
type dashboardJSON struct {
	Panels []panelJSON `json:"panels"`
}

// panelJSON represents a panel within a Grafana dashboard.
type panelJSON struct {
	Datasource *datasourceRef `json:"datasource"`
	Targets    []targetJSON   `json:"targets"`
	Title      string         `json:"title"`
}

// targetJSON represents a query target within a panel.
type targetJSON struct {
	Datasource *datasourceRef `json:"datasource"`
}

// datasourceRef represents a datasource reference in a panel or target.
type datasourceRef struct {
	Type string `json:"type"`
	UID  string `json:"uid"`
}

// validDatasourceTypes are the datasource types provisioned in datasources.yml
// plus "grafana" for built-in annotations.
var validDatasourceTypes = map[string]bool{
	"loki":       true,
	"tempo":      true,
	"prometheus": true,
	"grafana":    true,
}

func loadDashboardFiles(t *testing.T) []string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to determine test file path")
	}

	dashboardDir := filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "observability", "grafana", "dashboards")

	entries, err := os.ReadDir(dashboardDir)
	if err != nil {
		t.Fatalf("failed to read dashboards directory: %v", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			files = append(files, filepath.Join(dashboardDir, entry.Name()))
		}
	}

	if len(files) == 0 {
		t.Fatal("no dashboard JSON files found in grafana/dashboards/")
	}

	return files
}

func parseDashboard(t testing.TB, path string) dashboardJSON {
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read dashboard file %s: %v", path, err)
	}

	var db dashboardJSON
	if err := json.Unmarshal(data, &db); err != nil {
		t.Fatalf("failed to parse dashboard JSON %s: %v", path, err)
	}

	return db
}

// **Validates: Requirements 9.3, 9.4, 9.5, 9.6, 9.7, 9.8**
func TestProperty_AllDashboardsReferenceValidDatasources(t *testing.T) {
	files := loadDashboardFiles(t)

	// Collect all panels across all dashboards for random selection.
	type panelRef struct {
		file  string
		panel panelJSON
	}

	var allPanels []panelRef
	for _, f := range files {
		db := parseDashboard(t, f)
		for _, p := range db.Panels {
			allPanels = append(allPanels, panelRef{file: filepath.Base(f), panel: p})
		}
	}

	if len(allPanels) == 0 {
		t.Fatal("no panels found across any dashboard files")
	}

	rapid.Check(t, func(t *rapid.T) {
		idx := rapid.IntRange(0, len(allPanels)-1).Draw(t, "panelIndex")
		ref := allPanels[idx]

		// Check panel-level datasource.
		if ref.panel.Datasource != nil && ref.panel.Datasource.Type != "" {
			if !validDatasourceTypes[ref.panel.Datasource.Type] {
				t.Fatalf("panel %q in %s has invalid datasource type %q", ref.panel.Title, ref.file, ref.panel.Datasource.Type)
			}
		}

		// Check target-level datasources.
		for i, target := range ref.panel.Targets {
			if target.Datasource != nil && target.Datasource.Type != "" {
				if !validDatasourceTypes[target.Datasource.Type] {
					t.Fatalf("panel %q target[%d] in %s has invalid datasource type %q", ref.panel.Title, i, ref.file, target.Datasource.Type)
				}
			}
		}
	})
}
