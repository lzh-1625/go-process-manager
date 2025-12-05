package eum

type EventType string

const (
	EventProcessStart   EventType = "ProcessStart"
	EventProcessStop    EventType = "ProcessStop"
	EventProcessWarning EventType = "ProcessWarning"
	EventTaskStart      EventType = "TaskStart"
	EventTaskStop       EventType = "TaskStop"
	EventApiRequest     EventType = "ApiRequest"
)
