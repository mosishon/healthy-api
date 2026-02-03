package healthcheck

import (
	"context"
	"fmt"
	"healthy-api/model"
	"healthy-api/notifier"
	"healthy-api/registry"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type HealthChecker struct {
	Service           model.Service
	NotifierRegistry  *registry.Registry[notifier.Notifier]
	ConditionRegistry *registry.Registry[model.Condition]
	Client            *http.Client
	Logger            *slog.Logger
}

func (h *HealthChecker) Start(ctx context.Context) {
	h.Logger.Info("checker_started", "service", h.Service.Name)
	failureCount := 0

	for {
		waitDuration := time.Duration(h.Service.CheckPeriod) * time.Second

		h.performCheck(&failureCount, &waitDuration)

		select {
		case <-ctx.Done():
			return
		case <-time.After(waitDuration):
			// continue loop
		}
	}
}

func (h *HealthChecker) performCheck(failureCount *int, nextWait *time.Duration) {
	start := time.Now()
	request, err := http.NewRequest("GET", h.Service.URL, nil)

	var resp *http.Response
	var bodyData []byte

	evaluationRes := model.EvaluationResult{
		IsHealthy: false,
		Reason:    "Unknown error",
	}

	if h.Service.UserAgent != "" {
		request.Header.Set("User-Agent", h.Service.UserAgent)
	} else {
		request.Header.Set("User-Agent", "HealthyAPI(M.A)/1.0")
	}

	if err == nil {
		resp, err = h.Client.Do(request)
	}

	requestDuration := time.Since(start)
	sCode := 0

	if err != nil {
		evaluationRes.Reason = fmt.Sprintf("Network/Connection Error: %v", err)
	} else if resp != nil {
		sCode = resp.StatusCode

		bodyData, _ = io.ReadAll(resp.Body)
		resp.Body.Close()

		cond, ok := h.ConditionRegistry.Get(h.Service.ConditionName)
		if ok {
			evaluationRes = cond.Evaluate(resp, bodyData, requestDuration)
		} else {
			evaluationRes.Reason = "Condition registry not found"
		}
	}

	if !evaluationRes.IsHealthy {
		*failureCount++

		h.Logger.Warn("health_check_failed",
			"service", h.Service.Name,
			"attempt", *failureCount,
			"threshold", h.Service.Threshold,
			"status", sCode,
			"duration", requestDuration,
			"reason", evaluationRes.Reason)

		if *failureCount >= h.Service.Threshold {
			h.Logger.Error("threshold_reached", "service", h.Service.Name, "action", "sending_notifications")

			metadata := model.NotificationMetadata{
				ServiceName:  h.Service.Name,
				ServiceURL:   h.Service.URL,
				Reason:       evaluationRes.Reason,
				StatusCode:   sCode,
				ResponseTime: requestDuration.Round(time.Millisecond).String(),
				Timestamp:    time.Now().Format(time.RFC3339),
				FailureCount: *failureCount,
				Threshold:    h.Service.Threshold,
			}

			for _, target := range h.Service.Targets {
				if n, ok := h.NotifierRegistry.Get(target.NotifierID); ok {
					_ = n.Notify(model.Notification{
						Metadata:   metadata,
						Recipients: target.Recipients,
					})
				}
			}

			*nextWait = time.Duration(h.Service.SleepOnFail) * time.Second
			*failureCount = 0 // Reset after notification as per original logic
		}
	} else {
		if *failureCount > 0 {
			h.Logger.Info("service_recovery", "service", h.Service.Name, "after_failures", *failureCount)
		}
		*failureCount = 0
		h.Logger.Info("health_check_success", "service", h.Service.Name, "duration", requestDuration, "status_code", sCode)
	}
}

func (h *HealthChecker) StartInBackground(ctx context.Context) {
	go h.Start(ctx)
}
