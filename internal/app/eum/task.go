package eum

type Condition int

const (
	TaskCondRunning Condition = iota
	TaskCondNotRunning
	TaskCondException
	TaskCondPass
)

type TaskOperation int

const (
	TaskStart TaskOperation = iota
	TaskStop
	TaskStartWaitDone
	TaskStopWaitDone
)

type CtxTaskTraceID struct{}
