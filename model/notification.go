package model

type NotificationMetadata struct {
	ServiceName  string
	ServiceURL   string
	Reason       string
	StatusCode   int
	ResponseTime string
	Timestamp    string
	FailureCount int
	Threshold    int
	Status       string
}

type Notification struct {
	Metadata   NotificationMetadata
	Recipients []string
	Type       NotificationType
}

type TemplateGroup struct {
	NetworkError    string `yaml:"network_error"`
	HttpError       string `yaml:"http_error"`
	SlowResponse    string `yaml:"slow_response"`
	ConditionFailed string `yaml:"condition_failed"`
	Recovery       string `yaml:"recovery"`
	Default        string `yaml:"default"`
}

const (
	DefaultNetworkErrorTemplate    = "[🔌 Network Alert] {{.Metadata.ServiceName}} - Connection failed at {{.Metadata.Timestamp}}. Error: {{.Metadata.Reason}}"
	DefaultHttpErrorTemplate       = "[❌ HTTP Alert] {{.Metadata.ServiceName}} returned {{.Metadata.StatusCode}} at {{.Metadata.Timestamp}}. URL: {{.Metadata.ServiceURL}}"
	DefaultSlowResponseTemplate    = "[⏱️ Latency Alert] {{.Metadata.ServiceName}} is slow! Response time: {{.Metadata.ResponseTime}} (Threshold exceeded) at {{.Metadata.Timestamp}}."
	DefaultConditionFailedTemplate = "[🔍 Validation Alert] {{.Metadata.ServiceName}} failed health criteria at {{.Metadata.Timestamp}}. Detail: {{.Metadata.Reason}}"
	DefaultRecoveryTemplate        = "[✅ Recovery] {{.Metadata.ServiceName}} is back online! Status: Healthy. Restored at: {{.Metadata.Timestamp}}."
	DefaultNotificationTemplate    = "[🔔 Alert] {{.Metadata.ServiceName}} status is {{.Metadata.Status}} at {{.Metadata.Timestamp}}."
)

func GetDefaultTemplate(t NotificationType) string {
	switch t {
	case NotificationNetworkError:
		return DefaultNetworkErrorTemplate
	case NotificationHttpError:
		return DefaultHttpErrorTemplate
	case NotificationSlowResponse:
		return DefaultSlowResponseTemplate
	case NotificationConditionFailed:
		return DefaultConditionFailedTemplate
	case NotificationRecovery:
		return DefaultRecoveryTemplate
	default:
		return DefaultNotificationTemplate
	}
}
