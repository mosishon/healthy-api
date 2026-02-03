package model

type NotificationMetadata struct {
	ServiceName  string
	ServiceURL   string
	Reason       string
	StatusCode   int
	ResponseTime string
	Timestamp    string
}

type Notification struct {
	Metadata   NotificationMetadata
	Recipients []string
}
