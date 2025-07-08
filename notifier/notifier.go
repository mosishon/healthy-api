package notifier

import "healthy-api/model"

type Notifier interface {
	Notify(n model.Notification) error
}
