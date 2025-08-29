package eum

type TerminalType string

const (
	TerminalPty TerminalType = "pty"
	TerminalStd TerminalType = "std"
)

type ProcessState int32

const (
	ProcessStateStop ProcessState = iota
	ProcessStateStart
	ProcessStateWarnning
	ProcessStateRunning
)
