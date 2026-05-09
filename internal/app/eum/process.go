package eum

type ProcessState int32

const (
	ProcessStateStop     ProcessState = iota // process is stopped
	ProcessStateStart                        // process is starting
	ProcessStateWarnning                     // process is in warning state
	ProcessStateRunning                      // process is running
)
