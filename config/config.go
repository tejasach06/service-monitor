package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"service-monitor/models"

	"gopkg.in/yaml.v2"
)

const exampleConfigYAML = `# Example service-monitor config.yaml

services:
  AppDashboard:
    - port: 8080
      path: /
  AdminPanel:
    - port: 9090
      path: /login
  SecureLogsViewer:
    - port: 8443
      path: /esp

hosts:
  - ip: "192.168.1.10"
    services:
      - AppDashboard
      - AdminPanel
  - ip: "192.168.1.20"
    services:
      - SecureLogsViewer

webhook_url: "https://example.webhook.office.com/webhookb2/..."
mentions:
  - name: "Admin"
    email: "admin@example.com"

check_interval_seconds: 30
timeout_seconds: 5
`

// ✅ LoadConfig reads and parses the config YAML file.
func LoadConfig(path string) (*models.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	var cfg models.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("could not parse YAML: %w", err)
	}

	return &cfg, nil
}

// ✅ CheckAndCreateConfigIfMissing checks and creates the config file if it's missing.
func CheckAndCreateConfigIfMissing(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil // file exists
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("unable to access config path: %w", err)
	}

	fmt.Printf("⚠️  Config file not found at: %s\n", path)

	// Create parent directory
	dir := filepath.Dir(path)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write sample config content
	err = os.WriteFile(path, []byte(exampleConfigYAML), 0644)
	if err != nil {
		return fmt.Errorf("failed to create example config file: %w", err)
	}

	// Print success message and example
	fmt.Printf("\n✅ Created example config file at: %s\n", path)
	fmt.Println("📝 Please edit the file before starting the monitor.")
	fmt.Println("----------------------------------------------------")
	fmt.Println(exampleConfigYAML)
	fmt.Println("----------------------------------------------------\n")

	os.Exit(0)
	return nil
}
