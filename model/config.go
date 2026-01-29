package model

type Service struct {
	Name          string   `yaml:"name"`
	URL           string   `yaml:"url"`
	Targets       []Target `yaml:"targets"`
	CheckPeriod   int      `yaml:"check_period"`
	SleepOnFail   int      `yaml:"sleep_on_fail"`
	ConditionName string   `yaml:"condition_id"`
	Threshold     int      `yaml:"threshold"` 
	UserAgent     string   `yaml:"user_agent"`
}

type Target struct {
	NotifierID string   `yaml:"notifier_id"`
	Recipients []string `yaml:"recipients"`
}

type Notifiers struct {
	IPPanels []IPPanel `yaml:"ippanel"`
	SMTPs    []SMTP    `yaml:"smtp"`
	Webhook  []Webhook `yaml:"webhook"`
	MeliPayamakPanels []MeliPayamakPanel `yaml:"meli_payamak_panel"`
}

type SMTP struct {
	ID       string `yaml:"id"`
	Sender   string `yaml:"sender"`
	Password string `yaml:"password"`
	Server   string `yaml:"server"`
	Port     string `yaml:"port"`
}
type Config struct {
	Services   []Service        `yaml:"services"`
	Notifiers  Notifiers        `yaml:"notifiers"`
	Conditions []NamedCondition `yaml:"conditions"`
}
