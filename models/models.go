package models

type MentionUser struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}

type ServiceEndpoint struct {
	Port int `yaml:"port"`
	// Removed Path, no longer needed
}

type Config struct {
	Services      map[string][]ServiceEndpoint `yaml:"services"`
	Hosts         []Host                       `yaml:"hosts"`
	WebhookURL    string                       `yaml:"webhook_url"`
	Mentions      []MentionUser                `yaml:"mentions"`
	CheckInterval int                          `yaml:"check_interval_seconds"`
	Timeout       int                          `yaml:"timeout_seconds"`

	RetryCount         int `yaml:"retry_count"`
	RetryDelaySec      int `yaml:"retry_delay_seconds"`
	AlertDelay2Min     int `yaml:"alert_delay_2_minutes"`
	AlertDelay3PlusMin int `yaml:"alert_delay_3_plus_minutes"`
}

type Host struct {
	IP       string   `yaml:"ip"`
	Services []string `yaml:"services"`
}
