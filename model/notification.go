package model

type Notification struct {
	ServiceName string
	Recipients  []string
	Reason      string 
	StatusCode   int    
	ResponseTime string 
}
