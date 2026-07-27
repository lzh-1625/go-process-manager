//go:build !windows && (!epoll || !linux)

package process

import (
	"os"
	"os/exec"

	"github.com/creack/pty"
)

type ptyImpl struct {
	*os.File
}

func (p *ptyImpl) SetSize(cols, rows int) error {
	return pty.Setsize(p.File, &pty.Winsize{
		Cols: uint16(cols),
		Rows: uint16(rows),
	})
}

// Start starts the process.
func startWithPty(cmd *exec.Cmd) (ptyInterface, error) {
	pf, err := pty.Start(cmd)
	if err != nil || cmd.Process == nil {
		logger.Error("process start failed", "err", err)
		return nil, err
	}
	return &ptyImpl{pf}, nil
}

func (p *ptyImpl) Wait() {

}
