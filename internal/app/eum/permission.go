package eum

type OprPermission string

const (
	OperationStart         OprPermission = "Start"
	OperationStop          OprPermission = "Stop"
	OperationTerminal      OprPermission = "Terminal"
	OperationTerminalWrite OprPermission = "Write"
	OperationLog           OprPermission = "Log"
)
