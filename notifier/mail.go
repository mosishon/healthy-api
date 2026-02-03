package notifier

import (
	"bytes"
	"fmt"
	"healthy-api/model"
	"log/slog"
	"net/smtp"
	"text/template"
)

type MailNotifier struct {
	Sender    string
	Server    string
	Port      string
	Password  string
	Templates model.TemplateGroup
	Logger    *slog.Logger
}

func (m *MailNotifier) selectTemplate(n model.Notification) string {
	t := m.Templates
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
		return model.GetDefaultTemplate(n.Type)
	}

	return tmplStr
}

func (m *MailNotifier) CreateMessage(n model.Notification, to string, subject string) string {
	tmplStr := m.selectTemplate(n)
	if tmplStr != "" {
		tmpl, err := template.New("mail").Parse(tmplStr)
		if err == nil {
			var tpl bytes.Buffer
			if err := tmpl.Execute(&tpl, n); err == nil {
				return fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n%s", m.Sender, to, subject, tpl.String())
			}
		}
	}

	// Fallback to legacy format
	metadata := n.Metadata
	return fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\nService **%s** (%s) is now %s.\nReason: %s\nStatus Code: %d\nResponse Time: %s\nTimestamp: %s\nFailure Count: %d\nThreshold: %d\nCheck it fast please.",
		m.Sender, to, subject, metadata.ServiceName, metadata.ServiceURL, metadata.Status, metadata.Reason, metadata.StatusCode, metadata.ResponseTime, metadata.Timestamp, metadata.FailureCount, metadata.Threshold)
}

func (m *MailNotifier) GetName() string {
	return fmt.Sprintf("MailNotifier(%s)", m.Server)
}

func (m *MailNotifier) Notify(n model.Notification) error {
	auth := smtp.PlainAuth("", m.Sender, m.Password, m.Server)
	addr := fmt.Sprintf("%s:%s", m.Server, m.Port)
	for _, mail := range n.Recipients {
		go func(target string) {
			msg := m.CreateMessage(n, target, "Alert")
			err := smtp.SendMail(addr, auth, m.Sender, []string{target}, bytes.NewBufferString(msg).Bytes())
			if err != nil {
				m.Logger.Error("email_send_failed", "target", target, "addr", addr, "error", err)
			} else {
				m.Logger.Info("alert_sent", "target", target, "service", n.Metadata.ServiceName)
			}
		}(mail)
	}
	return nil
}
