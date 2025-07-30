
# Service Monitor

**Service Monitor** is a lightweight, extensible Go application that periodically checks your network services—including specific HTTP/HTTPS paths—to ensure they are up and running. It sends intelligent notifications to Microsoft Teams on service status changes, with support for user mentions, and remembers service states across restarts to avoid duplicate alerts.

---

## 🚀 Features

- Define services with multiple ports and paths (e.g., `/login`, `/api/health`)
- Track service state across restarts with persistent storage
- Send rich Microsoft Teams Adaptive Card notifications with mentions
- Auto-generate an example YAML config file on first run
- Optional `systemd` service file included for production deployments

---

## 📁 Project Structure

```

checker/         \# Service health check logic (HTTP/HTTPS with paths and TLS skip)
cmd/monitor/     \# Main application entry point
config/          \# Configuration loading and example file creation
models/          \# Data models and YAML mapping
notifier/        \# Microsoft Teams notification logic

```

---

## 🏁 Getting Started

### 1. Build the binary

```

go build -o service-monitor ./cmd/monitor

```

### 2. First run: generate example config

Run once to create the example config file at `/etc/service-monitor/config.yaml`:

```

sudo ./service-monitor

```

Edit the config file generated at `/etc/service-monitor/config.yaml` to fit your environment and services.

### 3. Run the monitor

```

sudo ./service-monitor

```

Or specify a custom config path:

```

./service-monitor -config /path/to/your/config.yaml

```

---

## 📝 Example Configuration (`config.yaml`)

```

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

webhook_url: "https://example.webhook.office.com/<your-teams-webhook>"
mentions:

- name: "Admin"
email: "admin@example.com"

check_interval_seconds: 30
timeout_seconds: 5

```

**Config Explanation:**

- `services`: Map service names to a list of port/path endpoints to check.
- `hosts`: List of IP addresses with associated service names.
- `webhook_url`: Your Teams incoming webhook URL.
- `mentions`: Optional list of people to mention in alerts.
- `check_interval_seconds`: Interval in seconds between health checks.
- `timeout_seconds`: Timeout in seconds per request.

---

## 🔧 Systemd Service Setup (Optional)

Create `/etc/systemd/system/service-monitor.service`:

```

[Unit]
Description=Service Monitor - Go-based web service health checker and notifier
After=network.target

[Service]
Type=simple
User=service-monitor
Group=service-monitor
WorkingDirectory=/usr/local/bin
ExecStart=/usr/local/bin/service-monitor -config /etc/service-monitor/config.yaml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target

```

### Set up and enable the service

```

sudo systemctl daemon-reload
sudo systemctl enable service-monitor.service
sudo systemctl start service-monitor.service

```

### View status and logs

```

sudo systemctl status service-monitor.service
sudo journalctl -u service-monitor.service -f

```

---

## ⚙️ How It Works

- Loads YAML config specifying hosts and service endpoints.
- Periodically checks the full URL `http(s)://IP:PORT/PATH`, supporting HTTPS with self-signed certs.
- Uses a persistent JSON file `/etc/service-monitor/last_state.json` to track service statuses across restarts.
- Sends Microsoft Teams Adaptive Card notifications **only** on changes (DOWN or UP after DOWN).
- Supports Teams user mentions in alerts.

---

## 🔧 Customization

- Notification logic: `notifier/teams.go`
- Checker behavior: `checker/checker.go`
- Config schema: `models/models.go`

Customization is straightforward—fork the repo and adjust components as needed.

---

## 📄 License

MIT License

---

## 👥 Authors

- Tejas Acharya (TJ)
---

**Questions or feature requests?** Feel free to open an issue or reach out!

---

Ready to get started?  
Run the binary, edit the auto-generated config, and monitor your critical services with confidence! 🚀
```
