package notifier

import "healthy-api/model"

type Notifier interface {
	Notif(n model.Notification) error
}
