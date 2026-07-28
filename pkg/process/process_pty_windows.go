//go:build windows

package process

import (
	"os"
	"os/exec"

	"github.com/runletapp/go-console"
)

type ptyImpl struct {
	console.Console
}

func (p *ptyImpl) SetSize(cols, rows int) error {
	return p.Console.SetSize(cols, rows)
}
func (p *ptyImpl) Wait() {}

// Start starts the process.
func startWithPty(cmd *exec.Cmd) (ptyInterface, error) {
	pty, err := console.New(100, 100)
	if err != nil {
		logger.Error("process start failed", "err", err)
		return nil, err
	}
	pty.SetCWD(cmd.Dir)
	pty.SetENV(append(os.Environ(), cmd.Env...))
	err = pty.Start(append([]string{cmd.Path}, cmd.Args...))
	if err != nil {
		logger.Error("process start failed", "err", err)
		return nil, err
	}
	pid, err := pty.Pid()
	if err != nil {
		logger.Error("process start failed", "err", err)
		return nil, err
	}
	op, err := os.FindProcess(pid)
	if err != nil {
		logger.Error("process start failed", "err", err)
		return nil, err
	}
	cmd.Process = op
	return &ptyImpl{pty}, nil
}
