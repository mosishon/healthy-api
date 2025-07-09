# Welcome to the Healthy API Wiki!

This wiki provides comprehensive documentation for understanding, configuring, and extending the Healthy API monitoring tool.

## Table of Contents

1.  [**Core Concepts**](#-core-concepts)
    -   [Services](#services)
    -   [Notifiers](#notifiers)
    -   [Health Checks](#health-checks)
2.  [**Configuration Deep Dive**](#-configuration-deep-dive)
    -   [The `services` Block](#the-services-block)
    -   [The `notifiers` Block](#the-notifiers-block)
        -   [SMTP Config](#smtp-config)
        -   [IPPanel (SMS) Config](#ippanel-sms-config)
        -   [Webhook Config & Templating](#webhook-config--templating)
3.  [**Extending the Application**](#-extending-the-application)
    -   [How to Add a New Notifier](#how-to-add-a-new-notifier)
4.  [**Running the Application**](#-running-the-application)

---

##  Core Concepts

### Services

A **Service** is the fundamental entity that you want to monitor. Each service is defined by a unique `name`, a `url` to check, and a set of conditions that determine its health status.

### Notifiers

A **Notifier** is a communication channel used to send alerts when a service fails. The application supports multiple types of notifiers out of the box (Email, SMS, Webhook). Each configured notifier is given a unique `id`, which is then referenced by services under the `targets` list. This design allows you to reuse a single notifier configuration (like a company-wide Slack channel) for multiple services.

### Health Checks

A **Health Check** is the process of sending an HTTP request to a service's `url` and validating the response. The two main parameters are:
-   `check_period`: The interval in seconds between each health check when the service is healthy.
-   `sleep_on_fail`: The interval in seconds to wait after a failure before checking again. This prevents alert spam while the service is down.

---

## ⚙️ Configuration Deep Dive

The application is configured using a single `config.yaml` file, which has two main sections: `services` and `notifiers`.

### The `services` Block

This is an array where you define every service to be monitored.

-   `name` (string): A human-readable name for the service (e.g., "Production API"). Used in alert messages.
-   `url` (string): The full URL of the health check endpoint (e.g., `https://api.example.com/health`).
-   `expected_status_code` (integer): The HTTP status code that indicates a healthy service. Typically `200`.
-   `check_period` (integer): The wait time in seconds between checks for a healthy service.
-   `sleep_on_fail` (integer): The wait time in seconds after a failure is detected.
-   `targets` (array): A list of notification targets to alert on failure.
    -   `notifier_id` (string): Must match the `id` of a configured notifier in the `notifiers` block.
    -   `recipients` (array of strings): A list of destinations (email addresses, phone numbers, or webhook URLs) for that notifier.

### The `notifiers` Block

This block contains the configuration for all your notification channels.

#### SMTP Config
-   `id` (string): A unique identifier (e.g., "admins-email-group").
-   `sender` (string): The "From" email address.
-   `password` (string): The SMTP server password.
-   `server` (string): The SMTP server address (e.g., "smtp.gmail.com").
-   `port` (integer): The SMTP server port (e.g., 587).

#### IPPanel (SMS) Config
-   `id` (string): A unique identifier (e.g., "on-call-sms-alert").
-   `url` (string): The API endpoint for the IPPanel service.
-   `user` (string): Your IPPanel username.
-   `pass` (string): Your IPPanel password.

#### Webhook Config & Templating
Webhooks are the most flexible notifier, allowing you to send custom HTTP requests to any endpoint (like Slack or a custom API).

-   `id` (string): A unique identifier (e.g., "slack-alerts").
-   `method` (string): The HTTP method to use (e.g., `POST`, `PUT`).
-   `headers` (map): A key-value map of HTTP headers.
-   `json` (map): The JSON payload to send in the request body.

**Templating:** The `headers` and `json` fields support dynamic values using Go's template engine. The following variables are available:
-   `{{ .ServiceName }}`: The `name` of the failed service.
-   `{{ .TimeStamp }}`: The timestamp of the failure in RFC3339 format (e.g., `2023-10-27T10:00:00Z`).
-   `{{ .URL }}`: The recipient URL that is currently being processed. This is useful if you have multiple webhook URLs for one notifier.

**Example:**
```yaml
json:
  text: "🔴 Alert: Service '{{ .ServiceName }}' is down!"
  fields:
    - title: "Timestamp"
      value: "{{ .TimeStamp }}"
      short: true
```

---

## 🔧 Extending the Application

### How to Add a New Notifier

The modular design makes it easy to add new notification channels (e.g., Telegram, Discord).

1.  **Implement the Interface:**
    -   Create a new file in the `notifier/` directory (e.g., `telegram.go`).
    -   Create a struct for your new notifier (e.g., `TelegramNotifier`).
    -   Implement the `notifier.Notifier` interface by adding the `Notify(n model.Notification) error` method. This method will contain the logic for sending the alert.

2.  **Update the Config Model:**
    -   In `model/config.go`, add a struct for your new notifier's configuration (e.g., `Telegram`).
    -   Add a new field to the `Notifiers` struct (e.g., `Telegrams []Telegram `yaml:"telegram"`).

3.  **Register the Notifier:**
    -   In `main.go`, create a new function `loadTelegramNotifiers` similar to the existing `load...` functions.
    -   This function should read the config, create an instance of your `TelegramNotifier`, and register it with the `notifierRegistry`.
    -   Call this new function from the `main` function.

---

## ▶️ Running the Application

To run the application, you must provide the path to your configuration file using the `-config` flag.

```bash
# Basic execution
go run main.go -config=config.yaml

# Build and run with verbose logging
go build -o healthy-api
./healthy-api -config=config.yaml -verbose
```

-   `-config` (required): Specifies the path to your YAML configuration file.
-   `-verbose` (optional): If present, the application will print detailed logs of every health check and notification attempt. This is highly recommended for debugging.
```