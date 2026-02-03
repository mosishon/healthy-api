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
	DefaultNetworkErrorTemplate    = "[🔌 Network Alert] {{.Metadata.ServiceName}}\n- Time: {{.Metadata.Timestamp}}\n{{.Metadata.Reason}}"
	DefaultHttpErrorTemplate       = "[❌ HTTP Alert] {{.Metadata.ServiceName}}\n- Time: {{.Metadata.Timestamp}}\n{{.Metadata.Reason}}"
	DefaultSlowResponseTemplate    = "[⏱️ Latency Alert] {{.Metadata.ServiceName}}\n- Time: {{.Metadata.Timestamp}}\n{{.Metadata.Reason}}"
	DefaultConditionFailedTemplate = "[🔍 Validation Alert] {{.Metadata.ServiceName}}\n- Time: {{.Metadata.Timestamp}}\n{{.Metadata.Reason}}"
	DefaultRecoveryTemplate        = "[✅ Recovery] {{.Metadata.ServiceName}}\n- Time: {{.Metadata.Timestamp}}\n- Status: Service is now Healthy"
	DefaultNotificationTemplate    = "[🔔 Alert] {{.Metadata.ServiceName}}\n- Time: {{.Metadata.Timestamp}}\n- Status: {{.Metadata.Status}}"
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
