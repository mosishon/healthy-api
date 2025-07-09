package model

type Service struct {
	Name               string   `yaml:"name"`
	URL                string   `yaml:"url"`
	Targets            []Target `yaml:"targets"`
	CheckPeriod        int      `yaml:"check_period"`
	SleepOnFail        int      `yaml:"sleep_on_fail"`
	ExpectedStatusCode int      `yaml:"expected_status_code"`
	ExpectedRegex      string   `yaml:"expected_regex"`
}

type Target struct {
	NotifierID string   `yaml:"notifier_id"`
	Recipients []string `yaml:"recipients"`
}

type Webhook struct {
	ID      string                 `yaml:"id"`
	Method  string                 `yaml:"method"`
	Headers map[string]interface{} `yaml:"headers"`
	JSON    map[string]interface{} `yaml:"json"`
}
type Notifiers struct {
	IPPanels []IPPanel `yaml:"ippanel"`
	SMTPs    []SMTP    `yaml:"smtp"`
	Webhook  []Webhook `yaml:"webhook"`
}
type IPPanel struct {
	ID   string `yaml:"id"`
	Url  string `yaml:"url"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}
type SMTP struct {
	ID       string `yaml:"id"`
	Sender   string `yaml:"sender"`
	Password string `yaml:"password"`
	Server   string `yaml:"server"`
	Port     string `yaml:"port"`
}
type Config struct {
	Services  []Service `yaml:"services"`
	Notifiers Notifiers `yaml:"notifiers"`
}
