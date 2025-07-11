package model

type SMSType string

const (
	pattern SMSType = "pattern"
)

type IPPanel struct {
	ID   string `yaml:"id"`
	Url  string `yaml:"url"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}

type SendSMSRequest struct {
	Op          SMSType             `json:"op"`
	User        string              `json:"user"`
	Pass        string              `json:"pass"`
	Sender      string              `json:"fromNum"`
	Recipient   string              `json:"toNum"`
	PatternCode string              `json:"patternCode"`
	InputData   []map[string]string `json:"inputData"`
}

type SendSMSVariables struct {
	AppName string `json:"app-name"`
}
type sendSMSHeader struct {
	ApiKey      string `json:"api-key"`
	ContentType string `json:"content-type"`
}

type smsTypes struct{}

func (s *smsTypes) PATTERN() SMSType {
	return pattern
}

var SMSTypes smsTypes

func init() {
	SMSTypes = smsTypes{}
}
