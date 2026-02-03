# 🩺 Healthy API - Advanced Service Monitoring

## 🌐 Languages

- [English](README.md)
- [فارسی](README.fa.md)

## WIKI

[WIKI](https://github.com/mosishon/healthy-api/wiki)

**Healthy API** is a powerful and extensible tool for real-time health checking of your web services. Written in **Go**, this project helps you ensure the availability and proper functioning of your services through periodic checks. If any service fails, it instantly alerts you through multiple channels like Email, SMS, and custom Webhooks.

---

## ✨ Key Features

- **Multi-Service Monitoring:** Define and monitor an unlimited number of services simultaneously.
- **Multi-Channel Alerting System:** Get notified via **SMTP (Email)**, **SMS (IPPanel)**, and **Webhooks**. The architecture is extensible for adding new channels.
- **Intelligent Periodic Checks:** Set custom intervals (`check_period`) for monitoring each service.
- **Spam Prevention:** Define a cooldown period (`sleep_on_fail`) after a failure is detected to avoid repetitive alerts.
- **Customizable Health Conditions:** Define complex health rules using **AND**, **OR**, **NOT**, **Regex**, **Header**, and **Response Time** checks.
- **Recovery Notifications:** Get notified when a service comes back online after a failure.
- **Smart Templating:** Use custom message templates for different types of failures (Network, HTTP, Latency, etc.).
- **Concurrent by Design:** Utilizes Goroutines to monitor all services concurrently without blocking.
- **Easy Configuration:** All settings are managed through a single, human-readable `YAML` file.

---

## 🛠️ Advanced Features

### Smart Metadata & Templating
You can use variables in your notification templates using Go's `text/template` syntax.

| Variable | Description |
| :--- | :--- |
| `{{.Metadata.ServiceName}}` | The name of the service |
| `{{.Metadata.ServiceURL}}` | The URL being checked |
| `{{.Metadata.Status}}` | Current status (UP or DOWN) |
| `{{.Metadata.Reason}}` | Detailed reason for failure (supports recursive reporting) |
| `{{.Metadata.StatusCode}}` | HTTP status code received |
| `{{.Metadata.ResponseTime}}`| Duration of the request |
| `{{.Metadata.Timestamp}}` | When the check occurred |
| `{{.Metadata.FailureCount}}`| Consecutive failures detected |

### Template Groups
Notifiers support `TemplateGroups`, allowing different messages for different failure types. This allows for intelligent selection:
- `network_error`: Network/Connection issues.
- `http_error`: Unexpected status codes.
- `slow_response`: Latency threshold exceeded.
- `condition_failed`: Logic/Regex/Header mismatch.
- `recovery`: Triggered when a service returns to healthy state.
- `default`: Fallback template.

### Recursive Error Reporting
When using nested `AND`/`OR` conditions, Healthy-API provides a detailed failure tree in the `Reason` field. This helps identify exactly which part of a complex condition caused the failure.

### Recovery Notifications
Set `notify_on_recovery: true` in your service configuration to receive alerts when services come back online.

---

## 🚀 Getting Started

### Prerequisites

- **Go 1.21+**
- Access to an SMTP server, an IPPanel SMS gateway, or a webhook endpoint.

### Installation & Usage

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/mosishon/healthy-api.git
    cd healthy-api
    ```

2.  **Create a configuration file** (e.g., `config.yaml`) by copying and modifying the `sample.yaml` file.

3.  **Run the application:**
    ```bash
    go run main.go -config=config.yaml
    ```

    Alternatively, you can build the executable first:
    ```bash
    go build -o healthy-api
    ./healthy-api -config=config.yaml -verbose
    ```
    > Use the `-verbose` flag to see detailed application logs.

---

## ⚙️ Configuration

All settings are managed in a single YAML file. The structure is as follows:

```yaml
#===========================================
#        Services to Monitor
#===========================================
services:
  - name: "production-api-service" # A descriptive name for display in alerts
    url: "https://api.my-domain.com/health"
    
    expected_status_code: 200 # The expected HTTP status code for a successful check
    check_period: 60 # Check every 60 seconds
    sleep_on_fail: 300 # If the service fails, wait 5 minutes before the next check to prevent spam
    # On failure, send alerts to these targets
    targets:
      - notifier_id: "admins-email-group"
        recipients:
          - "admin1@example.com"
          - "cto@example.com"
      - notifier_id: "on-call-sms-alert"
        recipients:
          - "+989120000001"
      - notifier_id: "slack-notification-hook"
        recipients:
          # You can define multiple webhook URLs for a single notifier ID
          - "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
          - "https://your-custom-api-endpoint.com/notify"

#===========================================
#        Notification Channel Configuration
#===========================================
notifiers:
  # ------ Email Servers (SMTP) ------
  smtp:
    - id: "admins-email-group" # This ID must match the 'notifier_id' in services
      sender: "notifier@your-domain.com"
      password: "your-smtp-password"
      server: "smtp.your-domain.com"
      port: 587

  # ------ SMS Gateways (e.g., IPPanel) ------
  ippanel: 
    - id: "on-call-sms-alert"
      url: <YOUR_IPPANEL_URL>
      user: <YOUR_IPPANEL_USERNAME>
      pass: <YOUR_IPPANEL_PASSWORD>

  # ------ Webhooks (For sending custom POST requests) ------
  webhook:
    - id: "slack-notification-hook"
      # The HTTP method to use for the webhook (e.g., POST, PUT)
      method: POST
      # Custom headers for the request
      headers:
        Content-Type: "application/json"
        Authorization: "Bearer your-secret-token" # Example for an auth header
      # The JSON body of the request.
      # You can use template variables for dynamic values.
      json:
        # {{ .ServiceName }} is replaced with the service name
        message: "🔴 Alert: Service '{{ .ServiceName }}' is down!"
        # {{ .TimeStamp }} is replaced with the failure timestamp
        timestamp: "{{ .TimeStamp }}"
        details: "Request to {{ .URL }} failed."
```

---

## 🏗️ Project Structure

The project is designed with a modular architecture to easily accommodate new features.

```
.
├── config/         # Logic for loading and parsing the YAML config file
├── healthcheck/    # The core engine for running service check loops
├── model/          # Struct definitions (Service, Notifier, Config, etc.)
├── notifier/       # The alert notification system (Email, SMS, etc.)
│   ├── notifier.go # The main interface for all notifiers
│   ├── mail.go     # SMTP email implementation
│   ├── sms.go      # IPPanel SMS implementation
│   └── webhook.go  # Webhook implementation
├── registry/registry.go        # Manages and registers different notifiers and conditions
├── main.go         # The entry point that coordinates all modules
└── sample.yaml     # An example configuration file
```

---

## 🗺️ Roadmap

- [x] Implement **Graceful Shutdown** using `context` for better Goroutine management.
- [x] Add **Unit Tests** for the `healthcheck` and `notifier` modules.
- [x] Support **Response Body Validation** using regular expressions (Regex).
- [x] Add more notifiers (e.g., **Slack**, **Discord** via Webhooks).
- [x] Persist logs to a file or database for historical analysis.
- [ ] Develop a simple **Web UI** to display the real-time status of services.
- [ ] Add cronjob instead of check_period.
- [x] Enhance logging with `slog`.
- [x] Add response time condition.
- [ ] Add json path condition.
- [x] Add retry policy (Threshold).
- [x] Add Recovery notifications.
- [x] Add Smart Template Groups.

---

## 🤝 Contributing

Contributions (PRs and issues) are highly welcome! If you have an idea for improving the project, we would love to hear from you.

Please follow these principles when contributing:
- Adhere to Go's **Naming Conventions**.
- Use **Interface-based Design** for greater flexibility.
- Use a controllable **Logger** for better debugging.

---

## 📄 License

**Mostafa Arshadi** (Proudly built for learning, growth, and teamwork ❤️)
