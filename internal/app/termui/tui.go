package termui

import (
	"context"
	"os"
	"time"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/log"
	"github.com/lzh-1625/go_process_manager/utils"

	"github.com/rivo/tview"
)

type TermConnectInstance struct {
	CancelFunc context.CancelFunc
}

func (t *TermConnectInstance) Write(b []byte) {
	os.Stdout.Write(b)
}
func (t *TermConnectInstance) WriteString(s string) {
	os.Stdout.Write([]byte(s))

}
func (t *TermConnectInstance) Cancel() {
	t.CancelFunc()
}

type tui struct {
	app *tview.Application
}

var Tui = new(tui)

func (t *tui) TermuiInit() {
	if config.CF.Tui {
		t.drawProcessList()
	}
}

func (t *tui) drawProcessList() {
	t.app = tview.NewApplication()
	list := tview.NewList()
	for i, v := range logic.ProcessCtlLogic.GetProcessList() {
		if i >= 'r' {
			i++
		}
		list.AddItem(v.Name, utils.NewKVStr().Add("user_name", v.User).Add("start_time", v.StartTime).Add("state", v.State.State).Build(), 'a'+rune(i), func() {
			if v.State.State != 1 || v.TermType != eum.TerminalPty {
				return
			}
			t.teminal(v.UUID)
			t.app.Stop()
			t.drawProcessList()
		})
	}
	list.AddItem("Refresh", "refresh process list", 'r', func() {
		t.app.Stop()
		t.drawProcessList()
	})
	if err := t.app.SetRoot(list, true).SetFocus(list).Run(); err != nil {
		panic(err)
	}
}

func (t *tui) teminal(uuid int) {
	p, err := logic.ProcessCtlLogic.GetProcess(uuid)
	if err != nil {
		log.Logger.Error("不存在uuid", "uuid", uuid)
	}
	ctx, cancel := context.WithCancel(context.Background())
	tci := &TermConnectInstance{
		CancelFunc: cancel,
	}
	p.AddConn(eum.Console, tci)
	defer p.DeleteConn(eum.Console)
	os.Stdin.Write([]byte("\033[H\033[2J")) // 清空屏幕
	p.ReadCache(tci)
	go t.startConnect(p, ctx, cancel)
	log.Logger.Info("tui wait")
	select {
	case <-p.StopSignal():
	case <-time.After(time.Minute * 10):
	case <-ctx.Done():
	}
	log.Logger.Info("tui quit")
}

func (t *tui) startConnect(p logic.Process, ctx context.Context, cancel context.CancelFunc) {
	switch p.Type() {
	case eum.TerminalPty:
		{
			t.ptyConnect(p, ctx, cancel)
		}
	case eum.TerminalStd:
		{
			t.stdConnect(p, ctx, cancel)
		}
	}
}

func (t *tui) ptyConnect(p logic.Process, ctx context.Context, cancel context.CancelFunc) {
	buf := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			{
				n, err := os.Stdin.Read(buf)
				if err != nil {
					return
				}
				if buf[0] == 0x04 { // ctrl+d 信号
					cancel()
					continue
				}
				p.WriteBytes(buf[:n])
			}
		}
	}
}

func (t *tui) stdConnect(p logic.Process, ctx context.Context, cancel context.CancelFunc) {
	buf := make([]byte, 1024)
	var line string
	for {
		select {
		case <-ctx.Done():
			return
		default:
			{
				n, err := os.Stdin.Read(buf)
				if err != nil {
					return
				}
				if buf[0] == 0x04 { // ctrl+d 信号
					cancel()
					continue
				}
				if buf[0] == 0x13 { // enter 信号
					p.Write(line)
					line = ""
					continue
				}
				line += string(buf[:n])
			}
		}
	}
}
