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
