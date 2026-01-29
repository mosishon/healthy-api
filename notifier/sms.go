package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"healthy-api/model"
	"io"
	"net/http"
	"time"
	"log/slog"

)

type SMSNotifier struct {
	User   string
	Pass   string
	URL    string
	Logger *slog.Logger
}

func newSendSMSHeader() http.Header {
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	return header
}
func (s *SMSNotifier) GetCodePattern() string {
	return "1103e3zul9izcls"
}
func (s *SMSNotifier) GetSender() string {
	return "3000505"
}
func (s *SMSNotifier) GetUser() string {
	return s.User
}
func (s *SMSNotifier) GetPass() string {
	return s.Pass
}

func (s *SMSNotifier) GetDataKey() string {
	return "app-name"
}
func (s *SMSNotifier) GetURL() string {
	return s.URL
}
func (s *SMSNotifier) GetName() string {
	return fmt.Sprintf("SMSNotifier(%s)", s.URL)
}
func (s SMSNotifier) Notify(n model.Notification) error {

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	for _, target := range n.Recipients {

		data, err := json.Marshal(model.SendSMSRequest{
			Op:          model.SMSTypes.PATTERN(),
			User:        s.GetUser(),
			Pass:        s.GetPass(),
			Sender:      s.GetSender(),
			Recipient:   target,
			PatternCode: s.GetCodePattern(),
			InputData: []map[string]string{
				{s.GetDataKey(): n.ServiceName},
			},
		})
		if err != nil {
			return fmt.Errorf("Error while marshaling SendSMSRequest %w", err)
		}
		request, err := http.NewRequest("POST", s.GetURL(), bytes.NewBuffer(data))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		request.Header = newSendSMSHeader()
		resp, err := client.Do(request)
		if err != nil {
			return fmt.Errorf("failed to send SMS to %s: %w", target, err)
		}
		bodyData, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("Error response code is %d. Body: %s", resp.StatusCode, string(bodyData))
		}
s.Logger.Info("sms_sent", "target", target, "status", resp.StatusCode, "body", string(bodyData))	}

	return nil
}
