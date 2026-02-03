package notifier

import (
	"bytes"
	"fmt"
	"healthy-api/model"
	"log/slog"
	"net/smtp"
)

type MailNotifier struct {
	Sender   string
	Server   string
	Port     string
	Password string
	Logger   *slog.Logger
}

func (m *MailNotifier) CreateMessage(metadata model.NotificationMetadata, to string, subject string) string {
	return fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\nService **%s** (%s) is not working good.\nReason: %s\nStatus Code: %d\nResponse Time: %s\nTimestamp: %s\nCheck it fast please.",
		m.Sender, to, subject, metadata.ServiceName, metadata.ServiceURL, metadata.Reason, metadata.StatusCode, metadata.ResponseTime, metadata.Timestamp)
}

func (m *MailNotifier) GetName() string {
	return fmt.Sprintf("MailNotifier(%s)", m.Server)
}

func (m *MailNotifier) Notify(n model.Notification) error {
	auth := smtp.PlainAuth("", m.Sender, m.Password, m.Server)
	addr := fmt.Sprintf("%s:%s", m.Server, m.Port)
	for _, mail := range n.Recipients {
		go func(target string) {
			msg := m.CreateMessage(n.Metadata, target, "Alert")
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
