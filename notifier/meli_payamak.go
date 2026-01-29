package notifier

import (
	"bytes"
	"fmt"
	"healthy-api/model"
	"log"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"
)

type PayamakNotifier struct {
	Username string
	Password string
	Sender   string
	Template string
	Logger   *log.Logger
}

func (p *PayamakNotifier) Notify(notification model.Notification) error {
	baseURL := "https://rest.payamak-panel.com/api/SendSMS/SendSMS"
	
	// رندر کردن تمپلیت
	tmpl, err := template.New("sms").Parse(p.Template)
	if err != nil {
		// اگر تمپلیت مشکل داشت، یک متن پیش‌فرض استفاده کن
		p.Template = "Service {{.ServiceName}} is DOWN!"
		tmpl, _ = template.New("sms").Parse(p.Template)
	}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, notification); err != nil {
		return err
	}
	message := tpl.String()

	client := &http.Client{Timeout: 10 * time.Second}

	for _, number := range notification.Recipients {
		data := url.Values{}
		data.Set("username", p.Username)
		data.Set("password", p.Password)
		data.Set("to", number)
		data.Set("from", p.Sender)
		data.Set("text", message)
		data.Set("isFlash", "false")

		resp, err := client.Post(baseURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
		if err != nil {
			p.Logger.Printf("[PayamakPanel Error] %s: %v\n", number, err)
			continue
		}
		resp.Body.Close()
	}
	return nil
}

func (p *PayamakNotifier) GetName() string { return "PayamakPanel" }