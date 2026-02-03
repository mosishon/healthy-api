package model

type Webhook struct {
	ID        string                 `yaml:"id"`
	Method    string                 `yaml:"method"`
	Headers   map[string]interface{} `yaml:"headers"`
	JSON      map[string]interface{} `yaml:"json"`
	Templates TemplateGroup          `yaml:"templates"`
}

type WebhookTemplate struct {
	Metadata NotificationMetadata
	URL      string // This is the recipient URL
}
