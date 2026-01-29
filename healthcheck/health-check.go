package healthcheck

import (
	"healthy-api/model"
	"healthy-api/notifier"
	"healthy-api/registry"
	"io"
	"log"
	"net/http"
	"time"
	"fmt"
)

type HealthChecker struct {
	Service           model.Service
	NotifierRegistry  *registry.Registry[notifier.Notifier]
	ConditionRegistry *registry.Registry[model.Condition]
	Client            *http.Client
	Logger            *log.Logger
}

// func (h *HealthChecker) Start() {
// 	h.Logger.Printf("Health checker started for: %s[%s]", h.Service.Name, h.Service.URL)
// 	failureCount := 0
// 	for {
// 		start := time.Now()
// 		request, err := http.NewRequest("GET", h.Service.URL, nil)
// 		if err != nil {
// 			h.Logger.Printf("error while creating request in %v %v.\n", h.Service, err)
// 			break
// 		}
// 		resp, err := h.Client.Do(request)
// 		requestDuration := time.Since(start)
// 		if err != nil {
// 			h.Logger.Printf("error while sending request in %v %v.\n", h.Service, err)
// 			time.Sleep(time.Duration(h.Service.SleepOnFail) * time.Second)
// 			continue
// 		}
// 		h.Logger.Printf("Request [GET] sent to %s, status code:%d, time:%v\n", h.Service.URL, resp.StatusCode, requestDuration)
// 		bodyData, err := io.ReadAll(resp.Body)
// 		if err != nil {
// 			h.Logger.Printf("Cant read body for %v \n", resp)
// 			time.Sleep(time.Duration(h.Service.SleepOnFail) * time.Second)
// 			continue
// 		}
// 		resp.Body.Close()
// 		cond, ok := h.ConditionRegistry.Get(h.Service.ConditionName)
// 		if !ok {
// 			h.Logger.Printf("Condition with id %s not found\n", h.Service.ConditionName)
// 			return
// 		}
// 		h.Logger.Printf("Evaluating Condition : %s for service : %s\n", h.Service.ConditionName, h.Service.Name)
// 		if !cond.Evaluate(resp, bodyData,requestDuration) {
// 			failureCount++

// 			h.Logger.Printf("Evaluating Condition : %s for service : %s is DONE and failed. response code: %d, time: %v\n", h.Service.ConditionName, h.Service.Name, resp.StatusCode, requestDuration)
// 			if failureCount >= h.Service.Threshold {
// 				h.Logger.Printf("Threshold reached for %s! Sending notifications...\n", h.Service.Name)
// 				for _, target := range h.Service.Targets {
// 					notifierInst, ok := h.NotifierRegistry.Get(target.NotifierID)
// 					if ok == false {
// 						h.Logger.Printf("notifier with id %s not found\n", target.NotifierID)
// 						continue
// 					}
// 					err := notifierInst.Notify(model.Notification{
// 						ServiceName: h.Service.Name,
// 						Recipients:  target.Recipients,
// 					})
// 					if err != nil {
// 						h.Logger.Printf("Failed to Notify using %v,%v\n", notifierInst.GetName(), err)
// 					}
// 				}
// 				h.Logger.Printf("[SLEEP] sleeping for %d.\n", h.Service.SleepOnFail)
// 				time.Sleep(time.Duration(h.Service.SleepOnFail) * time.Second)
// 			}else {
// 				time.Sleep(time.Duration(h.Service.CheckPeriod) * time.Second)
// 			}
// 		} else {
// 			if failureCount > 0 {
// 				h.Logger.Printf("Service %s is healthy again. Resetting failure count.\n", h.Service.Name)
// 			}
// 			failureCount = 0
// 			h.Logger.Printf("Evaluating Condition : %s for service : %s is DONE and successfull.response code is : %d\n", h.Service.ConditionName, h.Service.Name, resp.StatusCode)

// 			h.Logger.Printf("[SLEEP] sleeping for %d.\n", h.Service.CheckPeriod)

// 			time.Sleep(time.Duration(h.Service.CheckPeriod) * time.Second)
// 		}

// 	}
// }
func (h *HealthChecker) Start() {
	h.Logger.Printf("Started for: %s", h.Service.Name)
	failureCount := 0

	for {
		start := time.Now()
		request, err := http.NewRequest("GET", h.Service.URL, nil)
		
		var resp *http.Response
		var bodyData []byte
		
		// مقدار پیش‌فرض در صورت خطای شبکه یا نیل بودن ریسپانس
		evaluationRes := model.EvaluationResult{
			IsHealthy: false,
			Reason:    "Unknown error",
		}

		// ارسال درخواست
		if err == nil {
			resp, err = h.Client.Do(request)
		}
		
		requestDuration := time.Since(start)

		if err != nil {
			// در صورت خطای شبکه (مثل خاموش بودن سرور مقصد)
			evaluationRes.Reason = fmt.Sprintf("Network/Connection Error: %v", err)
		} else if resp != nil {
			// خواندن بادی و بستن فوری آن برای جلوگیری از Memory Leak
			bodyData, _ = io.ReadAll(resp.Body)
			resp.Body.Close() 
			
			cond, ok := h.ConditionRegistry.Get(h.Service.ConditionName)
			if ok {
				// ارزیابی تمام شرط‌ها و گرفتن دلیل (Reason)
				evaluationRes = cond.Evaluate(resp, bodyData, requestDuration)
			} else {
				evaluationRes.Reason = "Condition registry not found"
			}
		}

		if !evaluationRes.IsHealthy {
			failureCount++
			
			sCode := 0
			if resp != nil {
				sCode = resp.StatusCode
			}
			
			// چاپ در کنسول با ذکر دلیل دقیق
			h.Logger.Printf("FAIL %s [%d/%d] - Status: %d, Time: %v, Reason: %s", 
				h.Service.Name, failureCount, h.Service.Threshold, sCode, requestDuration, evaluationRes.Reason)

			if failureCount >= h.Service.Threshold {
				h.Logger.Printf("Threshold reached for %s. Sending notifications...", h.Service.Name)
				
				for _, target := range h.Service.Targets {
					if n, ok := h.NotifierRegistry.Get(target.NotifierID); ok {
						_ = n.Notify(model.Notification{
							ServiceName: h.Service.Name,
							Recipients:  target.Recipients,
							Reason:      evaluationRes.Reason, 
							StatusCode:   sCode,                    
                    		ResponseTime: requestDuration.Round(time.Millisecond).String(),
							
						})
					}
				}
				
				time.Sleep(time.Duration(h.Service.SleepOnFail) * time.Second)
				failureCount = 0 // ریست کردن پس از اطلاع‌رسانی برای شروع سیکل جدید
			} else {
				time.Sleep(time.Duration(h.Service.CheckPeriod) * time.Second)
			}
		} else {
			// اگر سرویس سالم بود
			if failureCount > 0 {
				h.Logger.Printf("Service %s is back to NORMAL after %d failures", h.Service.Name, failureCount)
			}
			failureCount = 0
			
			h.Logger.Printf("SUCCESS %s - Time: %v", h.Service.Name, requestDuration)
			
			time.Sleep(time.Duration(h.Service.CheckPeriod) * time.Second)
		}
	}
}
func (h *HealthChecker) StartInBackground() {
	go h.Start()
}
