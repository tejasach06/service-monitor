package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	"service-monitor/checker"
	"service-monitor/config"
	"service-monitor/notifier"
)

const (
	configPath        = "/etc/service-monitor/config.yaml"
	stateFilePath     = "/etc/service-monitor/last_state.json"
	stateFilePermsDir = 0755
	stateFilePerms    = 0644
)

func loadLastState() map[string]bool {
	data, err := os.ReadFile(stateFilePath)
	if err != nil {
		return map[string]bool{}
	}
	var state map[string]bool
	if err := json.Unmarshal(data, &state); err != nil {
		return map[string]bool{}
	}
	return state
}

func saveLastState(state map[string]bool) {
	_ = os.MkdirAll(filepath.Dir(stateFilePath), stateFilePermsDir)
	data, err := json.MarshalIndent(state, "", "  ")
	if err == nil {
		_ = os.WriteFile(stateFilePath, data, stateFilePerms)
	}
}

func main() {
	// Optional CLI override
	userConfigPath := flag.String("config", configPath, "Path to config file")
	flag.Parse()

	// Ensure config exists
	if err := config.CheckAndCreateConfigIfMissing(*userConfigPath); err != nil {
		log.Fatalf("❌ %v", err)
	}

	cfg, err := config.LoadConfig(*userConfigPath)
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	interval := time.Duration(cfg.CheckInterval) * time.Second
	timeout := time.Duration(cfg.Timeout) * time.Second

	lastStatus := loadLastState()

	log.Printf("✅ Monitoring started using config: %s\n", *userConfigPath)

	for {
		currentStatus := make(map[string]bool)

		for _, host := range cfg.Hosts {
			for _, serviceName := range host.Services {
				endpoints, ok := cfg.Services[serviceName]
				if !ok {
					log.Printf("⚠ Service %s not defined in config", serviceName)
					continue
				}
				for _, ep := range endpoints {
					up, proto := checker.CheckService(host.IP, ep.Port, ep.Path, timeout)
					keyBytes, err := json.Marshal([]any{host.IP, serviceName, ep.Port, ep.Path})
					if err != nil {
						log.Printf("❌ Failed to marshal key for %s:%d%s - %v", host.IP, ep.Port, ep.Path, err)
						continue
					}
					key := string(keyBytes)
					prevUp := lastStatus[key]
					currentStatus[key] = up

					status := "UP"
					if !up {
						status = "DOWN"
					}
					protoMsg := "unreachable"
					if up {
						protoMsg = proto
					}

					if !up || (up && !prevUp) {
						err := notifier.SendTeamsNotification(cfg.WebhookURL, serviceName, host.IP, ep.Port, status, protoMsg, cfg.Mentions)
						if err != nil {
							log.Printf("❌ Notification failed for %s:%d%s - %v", host.IP, ep.Port, ep.Path, err)
						} else {
							log.Printf("📣 Notified %s:%d%s [%s]", host.IP, ep.Port, ep.Path, status)
						}
					}
				}
			}
		}

		saveLastState(currentStatus)
		lastStatus = currentStatus
		time.Sleep(interval)
	}
}
