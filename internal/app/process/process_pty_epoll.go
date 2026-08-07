//go:build epoll && linux

package process

import (
	"os/exec"
	"sync"

	"github.com/containerd/console"
	"github.com/lzh-1625/go_process_manager/log"

	"github.com/creack/pty"
)

type epoller struct {
	readEpoller *console.Epoller
}

var globalEpoller = sync.OnceValue(func() *epoller {
	ep, err := console.NewEpoller()
	if err != nil {
		log.Logger.Panic(err)
	}
	return &epoller{
		readEpoller: ep,
	}
})

type ptyImpl struct {
	*console.EpollConsole
}

func (p *ptyImpl) SetSize(cols, rows int) error {
	return p.Console.Resize(console.WinSize{
		Width:  uint16(cols),
		Height: uint16(rows),
	})
}

// Start starts the process.
func startWithPty(cmd *exec.Cmd) (ptyInterface, error) {
	pf, err := pty.Start(cmd)
	if err != nil || cmd.Process == nil {
		log.Logger.Errorw("process start failed", "err", err)
		return nil, err
	}
	pty.Setsize(pf, &pty.Winsize{
		Rows: 100,
		Cols: 100,
	})
	cs, _, err := console.NewPtyFromFile(pf)
	if err != nil {
		log.Logger.Errorw("console new pty from file failed", "err", err)
		return nil, err
	}
	ep, err := globalEpoller().readEpoller.Add(cs)
	if err != nil {
		log.Logger.Errorw("read epoller add failed", "err", err)
		return nil, err
	}
	return &ptyImpl{ep}, nil
}

func (p *ptyImpl) Wait() {

}
