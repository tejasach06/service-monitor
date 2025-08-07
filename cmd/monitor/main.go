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

	retryCount      = 3
	retryDelay      = 2 * time.Second
	alertDelay2     = 10 * time.Minute // 2nd alert delay
	alertDelay3Plus = 30 * time.Minute // 3rd and subsequent alerts delay
)

// AlertTracker holds alert count and last notification time for a monitored endpoint
type AlertTracker struct {
	TimesSent int
	LastSent  time.Time
}

// loadLastState loads the previous monitoring states (UP/DOWN) from JSON file.
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

// saveLastState saves the current monitoring states to JSON file.
func saveLastState(state map[string]bool) {
	_ = os.MkdirAll(filepath.Dir(stateFilePath), stateFilePermsDir)
	data, err := json.MarshalIndent(state, "", "  ")
	if err == nil {
		_ = os.WriteFile(stateFilePath, data, stateFilePerms)
	}
}

func main() {
	userConfigPath := flag.String("config", configPath, "Path to config file")
	flag.Parse()

	if err := config.CheckAndCreateConfigIfMissing(*userConfigPath); err != nil {
		log.Fatalf("❌ %v", err)
	}

	cfg, err := config.LoadConfig(*userConfigPath)
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	interval := time.Duration(cfg.CheckInterval) * time.Second
	timeout := time.Duration(cfg.Timeout) * time.Second

	// Load last known states (site UP/DOWN)
	lastStatus := loadLastState()

	// Map to track alert info per service key
	alertTimes := make(map[string]*AlertTracker)

	log.Printf("✅ Monitoring started using config: %s\n", *userConfigPath)

	for {
		currentStatus := make(map[string]bool)
		now := time.Now()

		for _, host := range cfg.Hosts {
			for _, serviceName := range host.Services {
				endpoints, ok := cfg.Services[serviceName]
				if !ok {
					log.Printf("⚠ Service %s not defined in config", serviceName)
					continue
				}

				for _, ep := range endpoints {
					up := checker.CheckService(host.IP, ep.Port, retryCount, retryDelay, timeout)

					// Compose key for tracking using hostIP, serviceName, and port (drop path)
					keyBytes, err := json.Marshal([]any{host.IP, serviceName, ep.Port})
					if err != nil {
						log.Printf("❌ Failed to marshal key for %s:%d - %v", host.IP, ep.Port, err)
						continue
					}
					key := string(keyBytes)

					prevUp := lastStatus[key]
					currentStatus[key] = up

					status := "UP"
					if !up {
						status = "DOWN"
					}

					protoMsg := "TCP port check"
					if !up {
						protoMsg = "unreachable"
					}

					// Alert decision logic
					sendNotification := false

					alertTracker, exists := alertTimes[key]
					if !up {
						if !exists {
							// First DOWN alert
							sendNotification = true
							alertTimes[key] = &AlertTracker{TimesSent: 1, LastSent: now}
						} else {
							elapsed := now.Sub(alertTracker.LastSent)
							switch alertTracker.TimesSent {
							case 1:
								if elapsed >= alertDelay2 {
									sendNotification = true
									alertTracker.TimesSent++
									alertTracker.LastSent = now
								}
							default: // 2 or more sent
								if elapsed >= alertDelay3Plus {
									sendNotification = true
									alertTracker.TimesSent++
									alertTracker.LastSent = now
								}
							}
						}
					} else {
						// Service recovered (UP)
						if exists {
							delete(alertTimes, key)
						}
						// Send recovery notification if status changed from DOWN to UP
						if up && !prevUp {
							sendNotification = true
						}
					}

					if sendNotification {
						alertCount := 1
						if exists {
							alertCount = alertTracker.TimesSent
						}

						err := notifier.SendTeamsNotification(
							cfg.WebhookURL,
							serviceName,
							host.IP,
							ep.Port,
							status,
							protoMsg,
							cfg.Mentions,
							alertCount,
						)
						if err != nil {
							log.Printf("❌ Notification failed for %s:%d - %v", host.IP, ep.Port, err)
						} else {
							log.Printf("📣 Notified %s:%d [%s] (Alert count: %d)", host.IP, ep.Port, status, alertCount)
						}
					}
				}
			}
		}

		// Persist monitoring states for next iteration
		saveLastState(currentStatus)
		lastStatus = currentStatus

		time.Sleep(interval)
	}
}
