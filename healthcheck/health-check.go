package healthcheck

import (
	"healthy-api/model"
	"healthy-api/notifier"
	"healthy-api/registry"
	"io"
	"log"
	"net/http"
	"time"
)

type HealthChecker struct {
	Service           model.Service
	NotifierRegistry  *registry.Registry[notifier.Notifier]
	ConditionRegistry *registry.Registry[model.Condition]
	Client            *http.Client
	Logger            *log.Logger
}

func (h *HealthChecker) Start() {
	h.Logger.Printf("Health checker started for: %s[%s]", h.Service.Name, h.Service.URL)
	for {
		request, err := http.NewRequest("GET", h.Service.URL, nil)
		if err != nil {
			h.Logger.Printf("error while creating request in %v %v.\n", h.Service, err)
			break
		}
		resp, err := h.Client.Do(request)
		if err != nil {
			h.Logger.Printf("error while sending request in %v %v.\n", h.Service, err)
			time.Sleep(time.Duration(h.Service.SleepOnFail) * time.Second)
			continue
		}
		h.Logger.Printf("Request [GET] sent to %s, status code:%d\n", h.Service.URL, resp.StatusCode)
		bodyData, err := io.ReadAll(resp.Body)
		if err != nil {
			h.Logger.Printf("Cant read body for %v \n", resp)
			time.Sleep(time.Duration(h.Service.SleepOnFail) * time.Second)
			continue
		}
		resp.Body.Close()
		cond, ok := h.ConditionRegistry.Get(h.Service.ConditionName)
		if !ok {
			h.Logger.Printf("Condition with id %s not found\n", h.Service.ConditionName)
			return
		}
		h.Logger.Printf("Evaluating Condition : %s for service : %s\n", h.Service.ConditionName, h.Service.Name)
		if !cond.Evaluate(resp, bodyData) {
			h.Logger.Printf("Evaluating Condition : %s for service : %s is DONE and failed.response code is : %d\n", h.Service.ConditionName, h.Service.Name, resp.StatusCode)

			for _, target := range h.Service.Targets {
				notifierInst, ok := h.NotifierRegistry.Get(target.NotifierID)
				if ok == false {
					h.Logger.Printf("notifier with id %s not found\n", target.NotifierID)
					continue
				}
				err := notifierInst.Notify(model.Notification{
					ServiceName: h.Service.Name,
					Recipients:  target.Recipients,
				})
				if err != nil {
					h.Logger.Printf("Failed to Notify using %v,%v\n", notifierInst.GetName(), err)
				}
			}
			h.Logger.Printf("[SLEEP] sleeping for %d.\n", h.Service.SleepOnFail)
			time.Sleep(time.Duration(h.Service.SleepOnFail) * time.Second)

		} else {
			h.Logger.Printf("Evaluating Condition : %s for service : %s is DONE and successfull.response code is : %d\n", h.Service.ConditionName, h.Service.Name, resp.StatusCode)

			h.Logger.Printf("[SLEEP] sleeping for %d.\n", h.Service.CheckPeriod)

			time.Sleep(time.Duration(h.Service.CheckPeriod) * time.Second)
		}

	}
}
func (h *HealthChecker) StartInBackground() {
	go h.Start()
}
