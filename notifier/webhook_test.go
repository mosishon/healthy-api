package notifier_test

import (
	"encoding/json"
	"healthy-api/model"
	"healthy-api/notifier"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"os"
)

func TestWebhookNotifier_Notify(t *testing.T) {
	var receivedBody map[string]interface{}
	var receivedHeader http.Header
	var method string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header
		method = r.Method
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&receivedBody); err != nil {
			t.Errorf("failed to decode request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	hookData := model.Webhook{
		Method: "POST",
		Headers: map[string]interface{}{
			"Content-Type": "application/json",
			"X-Test":       "{{ .ServiceName }}",
		},
		JSON: map[string]interface{}{
			"message":   "Service {{ .ServiceName }} is down",
			"timestamp": "{{ .TimeStamp }}",
			"url":       "{{ .URL }}",
		},
	}
	wh := &notifier.WebhookNotifier{
		HookData: hookData,
		Client:   &http.Client{Timeout: 3 * time.Second},
		Logger:   slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	notif := model.Notification{
		ServiceName: "user-service",
		Recipients:  []string{server.URL},
	}

	err := wh.Notify(notif)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	if receivedBody == nil {
		t.Fatal("no request received by test server")
	}

	if receivedBody["message"] != "Service user-service is down" {
		t.Errorf("unexpected message: %v", receivedBody["message"])
	}

	if receivedHeader.Get("X-Test") != "user-service" {
		t.Errorf("unexpected header value for X-Test: %v", receivedHeader.Get("X-Test"))
	}
	if method != "POST" {
		t.Errorf("unexpected method %s should be POST", method)
	}

	if !strings.Contains(receivedBody["url"].(string), server.URL) {
		t.Errorf("unexpected url: %v", receivedBody["url"])
	}
}
