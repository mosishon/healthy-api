package model



type MeliPayamakPanel struct {
    ID       string `yaml:"id"`
    Username string `yaml:"username"`
    Password string `yaml:"password"`
    Sender   string `yaml:"sender"`
	Template string `yaml:"template"`
}