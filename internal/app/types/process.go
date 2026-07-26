package types

type ProcessState int32

const (
	ProcessStateStopped  ProcessState = iota // process is stopped
	ProcessStateStarting                     // process is starting
	ProcessStateWarning                      // process is in warning state
	ProcessStateRunning                      // process is running
	ProcessStateStopping                     // process is waiting stop
)

func (p ProcessState) String() string {
	switch p {
	case ProcessStateStopped:
		return "stopped"
	case ProcessStateStarting:
		return "starting"
	case ProcessStateWarning:
		return "warning"
	case ProcessStateRunning:
		return "running"
	case ProcessStateStopping:
		return "stopping"
	default:
		return "unknown"
	}
}
