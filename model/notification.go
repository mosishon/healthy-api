package model

type NotificationMetadata struct {
	ServiceName  string
	ServiceURL   string
	Reason       string
	StatusCode   int
	ResponseTime string
	Timestamp    string
	FailureCount int
	Threshold    int
}

type Notification struct {
	Metadata   NotificationMetadata
	Recipients []string
}
