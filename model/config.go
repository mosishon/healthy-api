package model

type Service struct {
	Name               string   `yaml:"name"`
	URL                string   `yaml:"url"`
	Phones             []string `yaml:"phones"`
	CheckPeriod        int      `yaml:"check_period"`
	SleepOnFail        int      `yaml:"sleep_on_fail"`
	ExpectedStatusCode int      `yaml:"expected_status_code"`
}

type Config struct {
	Services []Service `yaml:"services"`
	User     string    `yaml:"user"`
	Pass     string    `yaml:"pass"`
}
