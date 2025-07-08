package healthcheck

import (
	"healthy-api/model"
	"healthy-api/notifier"
	"log"
	"net/http"
	"time"
)

type HealthChecker struct {
	Service          model.Service
	NotifierRegistry *notifier.Registry
	Client           *http.Client
	Logger           *log.Logger
}

func (h *HealthChecker) Start() {
	h.Logger.Printf("Health checker started for :%v", h.Service)
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
		resp.Body.Close()

		if resp.StatusCode != h.Service.ExpectedStatusCode {
			for _, target := range h.Service.Targets {
				notifierInst := h.NotifierRegistry.Get(target.NotifierID)
				if notifierInst == nil {
					h.Logger.Printf("notifier with id %s not found\n", target.NotifierID)
					continue
				}
				err := notifierInst.Notify(model.Notification{
					ServiceName: h.Service.Name,
					Recipients:  target.Recipients,
				})
				if err != nil {
					h.Logger.Printf("Failed to Notify using %v,%v\n", notifierInst, err)
				} else {
					h.Logger.Printf("Notified using %v\n", notifierInst)

				}
			}
			h.Logger.Printf("[SLEEP] sleeping for %d.\n", h.Service.SleepOnFail)
			time.Sleep(time.Duration(h.Service.SleepOnFail) * time.Second)

		} else {
			h.Logger.Printf("[SLEEP] sleeping for %d.\n", h.Service.CheckPeriod)

			time.Sleep(time.Duration(h.Service.CheckPeriod) * time.Second)
		}

	}
}
func (h *HealthChecker) StartInBackground() {
	go h.Start()
}
