package notifier

import (
	"bytes"
	"healthy-api/model"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"
)

type PayamakNotifier struct {
	Username  string
	Password  string
	Sender    string
	Template  string
	Templates model.TemplateGroup
	Logger    *slog.Logger
}

func (p *PayamakNotifier) selectTemplate(n model.Notification) string {
	t := p.Templates
	var tmplStr string

	switch n.Type {
	case model.NotificationNetworkError:
		tmplStr = t.NetworkError
	case model.NotificationHttpError:
		tmplStr = t.HttpError
	case model.NotificationSlowResponse:
		tmplStr = t.SlowResponse
	case model.NotificationConditionFailed:
		tmplStr = t.ConditionFailed
	case model.NotificationRecovery:
		tmplStr = t.Recovery
	default:
		tmplStr = t.Default
	}

	if tmplStr == "" {
		if p.Template != "" {
			return p.Template
		}
		return model.GetDefaultTemplate(n.Type)
	}
	return tmplStr
}

func (p *PayamakNotifier) Notify(notification model.Notification) error {
	baseURL := "https://rest.payamak-panel.com/api/SendSMS/SendSMS"

	// رندر کردن تمپلیت
	tmplStr := p.selectTemplate(notification)
	tmpl, err := template.New("sms").Parse(tmplStr)
	if err != nil {
		// اگر تمپلیت مشکل داشت، یک متن پیش‌فرض استفاده کن
		tmplStr = "Service {{.Metadata.ServiceName}} is {{.Metadata.Status}}!"
		tmpl, _ = template.New("sms").Parse(tmplStr)
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
			p.Logger.Error("network_request_failed",
				"provider", "payamak_panel",
				"target",   number,
				"error",    err,
			)
			continue
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		responseString := string(bodyBytes)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK || len(responseString) < 5 {
			p.Logger.Error("sms_delivery_failed",
				"provider", "payamak_panel",
				"target",   number,
				"status",   resp.Status,
				"response", responseString,
			)
		} else {
			p.Logger.Info("sms_delivery_success",
				"provider", "payamak_panel",
				"target",   number,
				"result_id", responseString,
			)
		}
	}
	return nil
}

func (p *PayamakNotifier) GetName() string { return "PayamakPanel" }
