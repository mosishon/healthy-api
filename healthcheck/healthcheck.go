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
	"strings"
	"time"
)

type HealthChecker struct {
	Service           model.Service
	NotifierRegistry  *registry.Registry[notifier.Notifier]
	ConditionRegistry *registry.Registry[model.Condition]
	Client            *http.Client
	Logger            *slog.Logger
	isDown            bool
	failureCount      int
}

func (h *HealthChecker) Start(ctx context.Context) {
	h.Logger.Info("checker_started", "service", h.Service.Name)

	for {
		waitDuration := time.Duration(h.Service.CheckPeriod) * time.Second

		h.performCheck(&waitDuration)

		select {
		case <-ctx.Done():
			return
		case <-time.After(waitDuration):
			// continue loop
		}
	}
}

func (h *HealthChecker) performCheck(nextWait *time.Duration) {
	start := time.Now()

	// تنظیم پیش‌فرض زمان انتظار برای دور بعدی
	*nextWait = time.Duration(h.Service.CheckPeriod) * time.Second

	request, err := http.NewRequest("GET", h.Service.URL, nil)
	if h.Service.UserAgent != "" {
		request.Header.Set("User-Agent", h.Service.UserAgent)
	} else {
		request.Header.Set("User-Agent", "HealthyAPI(M.A)/1.0")
	}

	var resp *http.Response
	var bodyData []byte
	evaluationRes := model.EvaluationResult{
		IsHealthy: false,
		Reason:    []string{"Unknown error"},
	}

	// اجرای درخواست
	if err == nil {
		resp, err = h.Client.Do(request)
	}

	requestDuration := time.Since(start)
	sCode := 0

	// ارزیابی نتیجه
	if err != nil {
		evaluationRes.Reason = []string{fmt.Sprintf("Network Error: %v", err)}
		evaluationRes.Type = model.NotificationNetworkError
	} else if resp != nil {
		sCode = resp.StatusCode
		bodyData, _ = io.ReadAll(resp.Body)
		resp.Body.Close()

		cond, ok := h.ConditionRegistry.Get(h.Service.ConditionName)
		if ok {
			evaluationRes = cond.Evaluate(resp, bodyData, requestDuration)
		} else {
			evaluationRes.Reason = []string{"Condition configuration not found"}
		}
	}

	joinedReason := strings.Join(evaluationRes.Reason, "\n")

	// --- منطق اصلی سلامت و Sleep ---
	if !evaluationRes.IsHealthy {
		h.failureCount++

		h.Logger.Warn("health_check_failed",
			"service", h.Service.Name,
			"attempt", h.failureCount,
			"threshold", h.Service.Threshold,
			"reason", joinedReason)

		// رسیدن به حد نصاب خطا
		if h.failureCount >= h.Service.Threshold {
			// تنظیم زمان استراحت طولانی چون سرویس Down است
			*nextWait = time.Duration(h.Service.SleepOnFail) * time.Second

			// ارسال نوتیفیکیشن فقط در صورتی که قبلاً DOWN نشده بود (برای جلوگیری از اسپم)
			if !h.isDown {
				h.isDown = true
				h.Logger.Error("threshold_reached", "service", h.Service.Name, "action", "sending_notifications")

				h.sendNotification("DOWN", joinedReason, sCode, requestDuration, evaluationRes.Type)
			}
		}
	} else {
		// سرویس سالم است
		if h.isDown {
			h.Logger.Info("service_recovery", "service", h.Service.Name)
			if h.Service.NotifyOnRecovery {
				h.sendNotification("UP", "Service recovered", sCode, requestDuration, model.NotificationRecovery)
			}
			h.isDown = false
		} else if h.failureCount > 0 {
			h.Logger.Info("service_recovered_before_threshold", "service", h.Service.Name, "after_failures", h.failureCount)
		}

		// ریست کردن کانتر فقط در صورت موفقیت
		h.failureCount = 0
		h.Logger.Info("health_check_success", "service", h.Service.Name, "duration", requestDuration)
	}
}

// متد کمکی برای ارسال نوتیفیکیشن (تمیزتر شدن کد)
func (h *HealthChecker) sendNotification(status, reason string, sCode int, duration time.Duration, nType model.NotificationType) {
	metadata := model.NotificationMetadata{
		ServiceName:  h.Service.Name,
		ServiceURL:   h.Service.URL,
		Reason:       reason,
		StatusCode:   sCode,
		ResponseTime: duration.Round(time.Millisecond).String(),
		Timestamp:    time.Now().Format(time.RFC3339),
		FailureCount: h.failureCount,
		Threshold:    h.Service.Threshold,
		Status:       status,
	}

	for _, target := range h.Service.Targets {
		if n, ok := h.NotifierRegistry.Get(target.NotifierID); ok {
			_ = n.Notify(model.Notification{
				Metadata:   metadata,
				Recipients: target.Recipients,
				Type:       nType,
			})
		}
	}
}

func (h *HealthChecker) StartInBackground(ctx context.Context) {
	go h.Start(ctx)
}
