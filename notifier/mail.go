package notifier

import (
	"bytes"
	"fmt"
	"healthy-api/model"
	"log"
	"net/smtp"
)

type MailNotifier struct {
	Sender   string
	Server   string
	Port     string
	Password string
	Logger   *log.Logger
}

func (m *MailNotifier) CreateMessage(serviceName string, to string, subject string) string {
	return fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\nService **%s** is not working good check it fast please.", m.Sender, to, subject, serviceName)
}
func (m *MailNotifier) GetName() string {
	return fmt.Sprintf("MailNotifier(%s)", m.Server)
}
func (m *MailNotifier) Notify(n model.Notification) error {
	auth := smtp.PlainAuth("", m.Sender, m.Password, m.Server)
	addr := fmt.Sprintf("%s:%s", m.Server, m.Port)
	for _, mail := range n.Recipients {
		go func(target string) {
			msg := m.CreateMessage(n.ServiceName, target, "Alert")
			err := smtp.SendMail(addr, auth, m.Sender, []string{mail}, bytes.NewBufferString(msg).Bytes())
			if err != nil {
				m.Logger.Printf("error while sending mail to :%s,addr=%v\n", target, addr)
				// return fmt.Errorf("error while sending mail to %s:%w", target, err)
			}
			m.Logger.Printf("alert sent to %s app-name:%s\n", target, n.ServiceName)
		}(mail)
	}
	return nil
}
