package healthcheck

import (
	"bytes"
	"healthy-api/model"
	"healthy-api/notifier"
	"log"
	"net/http"
	"time"
)

type HealthChecker struct {
	Service  model.Service
	Notifier notifier.Notifier
	Client   *http.Client
	Logger   *log.Logger
}

func (h *HealthChecker) Start() {
	h.Logger.Printf("Health checker started for :%v", h.Service)
	for {
		request, err := http.NewRequest("GET", h.Service.URL, bytes.NewBuffer(nil))
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
			h.Logger.Println("Notifing using notifier")
			err := h.Notifier.Notif(model.Notification{
				ServiceName: h.Service.Name,
				Recipients:  h.Service.Phones,
			})
			if err != nil {
				h.Logger.Printf("failed to notify: %v", err)
			}

			time.Sleep(time.Duration(h.Service.SleepOnFail) * time.Second)

		} else {
			time.Sleep(time.Duration(h.Service.CheckPeriod) * time.Second)
		}

	}
}
func (h *HealthChecker) StartInBackground() {
	go h.Start()
}
