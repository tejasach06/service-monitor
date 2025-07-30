package models

type MentionUser struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}

type ServiceEndpoint struct {
	Port int    `yaml:"port"`
	Path string `yaml:"path"`
}

type Config struct {
	Services      map[string][]ServiceEndpoint `yaml:"services"`
	Hosts         []Host                       `yaml:"hosts"`
	WebhookURL    string                       `yaml:"webhook_url"`
	Mentions      []MentionUser                `yaml:"mentions"`
	CheckInterval int                          `yaml:"check_interval_seconds"`
	Timeout       int                          `yaml:"timeout_seconds"`
}

type Host struct {
	IP       string   `yaml:"ip"`
	Services []string `yaml:"services"`
}
