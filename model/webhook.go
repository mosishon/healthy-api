package model

type Webhook struct {
	ID      string                 `yaml:"id"`
	Method  string                 `yaml:"method"`
	Headers map[string]interface{} `yaml:"headers"`
	JSON    map[string]interface{} `yaml:"json"`
}

type WebhookTemplate struct {
	ServiceName string
	TimeStamp   string
	URL         string
}
