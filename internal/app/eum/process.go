package eum

type ProcessState int32

const (
	ProcessStateStop ProcessState = iota
	ProcessStateStart
	ProcessStateWarnning
	ProcessStateRunning
)
