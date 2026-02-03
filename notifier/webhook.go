package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"healthy-api/model"
	"log/slog"
	"net/http"
	"text/template"
)

type WebhookNotifier struct {
	HookData model.Webhook
	Client   *http.Client
	Logger   *slog.Logger
}

func (w *WebhookNotifier) GetName() string {
	return fmt.Sprintf("WebhookNotifier(%s)", w.HookData.ID)
}

func executeTemplate(tmplStr string, ctx model.WebhookTemplate) (string, error) {
	tmpl, err := template.New("tmpl").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, ctx)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func FillTemplate(data map[string]interface{}, ctx model.WebhookTemplate) (map[string]interface{}, error) {

	result := make(map[string]interface{})
	for key, val := range data {
		switch v := val.(type) {
		case string:
			tmplRes, err := executeTemplate(v, ctx)
			if err != nil {
				return nil, err
			}
			result[key] = tmplRes

		case map[string]interface{}:
			nested, err := FillTemplate(v, ctx)
			if err != nil {
				return nil, err
			}
			result[key] = nested

		case []interface{}:
			newList := make([]interface{}, 0, len(v))
			for _, item := range v {
				switch itemVal := item.(type) {
				case string:
					tmplRes, err := executeTemplate(itemVal, ctx)
					if err != nil {
						return nil, err
					}
					newList = append(newList, tmplRes)
				default:
					newList = append(newList, item)
				}
			}
			result[key] = newList

		default:
			result[key] = val
		}
	}
	return result, nil
}

func (w *WebhookNotifier) sendRequest(url string, headers map[string]interface{}, body []byte) error {
	req, err := http.NewRequest(w.HookData.Method, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	for k, v := range headers {
		if valStr, ok := v.(string); ok {
			req.Header.Set(k, valStr)
		} else {
			w.Logger.Error("invalid_header_value", "key", k, "error", "not_a_string")
		}
	}
	resp, err := w.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		w.Logger.Error("bad_status_code", "status", resp.StatusCode)
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	return nil
}

func (w *WebhookNotifier) selectTemplate(n model.Notification) map[string]interface{} {
	// If no templates are defined in the group, use the legacy/direct JSON.
	// But the user wants intelligent selection.
	t := w.HookData.Templates
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
		return w.HookData.JSON
	}

	// If we have a template string, we assume it's a JSON string.
	// We'll try to unmarshal it.
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(tmplStr), &result); err != nil {
		// If it's not valid JSON, maybe it's just a message string?
		// To keep it simple and consistent with the legacy behavior:
		return w.HookData.JSON
	}
	return result
}

func (w *WebhookNotifier) Notify(n model.Notification) error {
	for _, recipient := range n.Recipients {

		ctx := model.WebhookTemplate{
			Metadata: n.Metadata,
			URL:      recipient,
		}

		jsonToUse := w.selectTemplate(n)

		filledHeaders, err := FillTemplate(w.HookData.Headers, ctx)
		if err != nil {
			return fmt.Errorf("failed to fill headers template: %w", err)
		}
		filledJSON, err := FillTemplate(jsonToUse, ctx)
		if err != nil {
			return fmt.Errorf("failed to fill JSON template: %w", err)
		}
		bodyBytes, err := json.Marshal(filledJSON)
		if err != nil {
			return fmt.Errorf("failed to marshal json body: %w", err)
		}

		go func(rec string, hdr map[string]interface{}, body []byte) {
			if err := w.sendRequest(rec, hdr, body); err != nil {
				w.Logger.Error("webhook_request_failed", "target", rec, "error", err)
			} else {
				w.Logger.Info("webhook_sent_success", "target", rec)
			}

		}(recipient, filledHeaders, bodyBytes)
	}
	return nil
}
