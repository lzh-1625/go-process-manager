package types

type ProcessState int32

const (
	ProcessStateStopped  ProcessState = iota // process is stopped
	ProcessStateStarting                     // process is starting
	ProcessStateWarning                      // process is in warning state
	ProcessStateRunning                      // process is running
	ProcessStateStopping                     // process is waiting stop
)
