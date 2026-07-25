//go:build windows

package process

import (
	"os"
	"os/exec"

	"github.com/lzh-1625/go_process_manager/log"
	"github.com/runletapp/go-console"
)

type ptyImpl struct {
	console.Console
}

func (p *ptyImpl) SetSize(cols, rows int) error {
	return p.Console.SetSize(cols, rows)
}

// Start starts the process.
func NewPTY(cmd *exec.Cmd) (ptyInterface, error) {
	pty, err := console.New(100, 100)
	if err != nil {
		log.Logger.Errorw("process start failed", "err", err)
		return nil, err
	}
	pty.SetCWD(cmd.Dir)
	pty.SetENV(append(os.Environ(), cmd.Env...))
	err = pty.Start(append([]string{cmd.Path}, cmd.Args...))
	if err != nil {
		log.Logger.Errorw("process start failed", "err", err)
		return nil, err
	}
	pid, err := pty.Pid()
	if err != nil {
		log.Logger.Errorw("process start failed", "err", err)
		return nil, err
	}
	op, err := os.FindProcess(pid)
	if err != nil {
		log.Logger.Errorw("process start failed", "err", err)
		return nil, err
	}
	cmd.Process = op
	return &ptyImpl{pty}, nil
}

func (p *ptyImpl) Wait() {

}
