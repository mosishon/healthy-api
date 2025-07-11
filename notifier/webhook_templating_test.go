package notifier_test

import (
	"healthy-api/model"
	"healthy-api/notifier"
	"testing"
	"time"
)

func TestFillTemplate(t *testing.T) {
	templateData := map[string]interface{}{
		"message": "Service '{{ .ServiceName }}' is down!",
		"details": map[string]interface{}{
			"url":       "Checked URL was {{ .URL }}",
			"timestamp": "{{ .TimeStamp }}",
		},
		"static_value": 123,
	}

	testTime := time.Now()
	context := model.WebhookTemplate{
		ServiceName: "Login-API",
		TimeStamp:   testTime.Format(time.RFC3339),
		URL:         "https://api.example.com/login",
	}

	result, err := notifier.FillTemplate(templateData, context)

	if err != nil {
		t.Fatalf("FillTemplate returned an unexpected error: %v", err)
	}

	// Check top-level message
	expectedMessage := "Service 'Login-API' is down!"
	if result["message"] != expectedMessage {
		t.Errorf("Expected message to be '%s', got '%s'", expectedMessage, result["message"])
	}

	// Check nested details
	detailsMap, ok := result["details"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'details' to be a map")
	}

	expectedURL := "Checked URL was https://api.example.com/login"
	if detailsMap["url"] != expectedURL {
		t.Errorf("Expected nested url to be '%s', got '%s'", expectedURL, detailsMap["url"])
	}

	if detailsMap["timestamp"] != context.TimeStamp {
		t.Errorf("Expected nested timestamp to be '%s', got '%s'", context.TimeStamp, detailsMap["timestamp"])
	}

	// Check that static values are preserved
	if result["static_value"] != 123 {
		t.Errorf("Expected static_value to be 123, got %v", result["static_value"])
	}
}
