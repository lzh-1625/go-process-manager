package logic

import (
	"bytes"
	"errors"
	"os"
	"strings"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/log"
	"github.com/lzh-1625/go_process_manager/utils"

	"github.com/runletapp/go-console"
)

type ProcessPty struct {
	*ProcessBase
	cacheBytesBuf *bytes.Buffer
	pty           console.Console
}

func (p *ProcessPty) doOnKilled() {
	p.pty.Close()
}

func (p *ProcessPty) Type() eum.TerminalType {
	return eum.TerminalPty
}

func (p *ProcessPty) Start() (err error) {
	defer func() {
		if err != nil {
			p.Config.AutoRestart = false
			p.SetState(eum.ProcessStateWarnning)
			p.State.Info = "进程启动失败:" + err.Error()
		}
	}()
	if ok := p.SetState(eum.ProcessStateStart, func() bool {
		return p.State.State != eum.ProcessStateRunning && p.State.State != eum.ProcessStateStart
	}); !ok {
		log.Logger.Warnw("进程已在运行，跳过启动")
		return nil
	}
	pty, err := console.New(100, 100)
	if err != nil {
		log.Logger.Errorw("进程启动失败", "err", err)
		return err
	}
	pty.SetCWD(p.WorkDir)
	err = pty.Start(p.StartCommand)
	if err != nil {
		log.Logger.Errorw("进程启动失败", "err", err)
		return err
	}
	p.pty = pty
	pid, err := pty.Pid()
	if err != nil {
		log.Logger.Errorw("进程启动失败", "err", err)
		return err
	}
	p.op, err = os.FindProcess(pid)
	if err != nil {
		log.Logger.Errorw("进程启动失败", "err", err)
		return err
	}
	log.Logger.Infow("进程启动成功", "进程名称", p.Name, "重启次数", p.State.restartTimes)
	p.pInit()
	if !p.SetState(eum.ProcessStateRunning, func() bool {
		return p.State.State == eum.ProcessStateStart
	}) {
		return errors.New("状态异常启动失败")
	}
	p.push("进程启动成功")
	return nil
}

func (p *ProcessPty) SetTerminalSize(cols, rows int) {
	if cols == 0 || rows == 0 || len(p.ws) != 0 {
		return
	}
	p.pty.SetSize(cols, rows)

}

func (p *ProcessPty) WriteBytes(input []byte) (err error) {
	_, err = p.pty.Write(input)
	return
}

func (p *ProcessPty) Write(input string) (err error) {
	_, err = p.pty.Write([]byte(input))
	return
}

func (p *ProcessPty) readInit() {
	log.Logger.Debugw("stdout读取线程已启动", "进程名", p.Name, "使用者", p.GetUserString())
	buf := make([]byte, 1024)
	for {
		select {
		case <-p.StopChan:
			{
				log.Logger.Debugw("stdout读取线程已退出", "进程名", p.Name, "使用者", p.GetUserString())
				return
			}
		default:
			{
				n, _ := p.pty.Read(buf)
				p.bufHanle(buf[:n])
				if len(p.ws) == 0 {
					continue
				}
				p.wsLock.Lock()
				for _, v := range p.ws {
					v.Write(buf[:n])
				}
				p.wsLock.Unlock()
			}
		}
	}
}

func (p *ProcessPty) ReadCache(ws ConnectInstance) {
	ws.Write(p.cacheBytesBuf.Bytes())
}

func (p *ProcessPty) bufHanle(b []byte) {
	log := strings.TrimSpace(string(b))
	if utils.RemoveANSI(log) != "" {
		p.logReportHandler(log)
	}
	p.cacheBytesBuf.Write(b)
	p.cacheBytesBuf.Next(len(b))
}

func (p *ProcessPty) doOnInit() {
	p.cacheBytesBuf = bytes.NewBuffer(make([]byte, config.CF.ProcessMsgCacheBufLimit))
}

func NewProcessPty(pconfig model.Process) *ProcessBase {
	p := ProcessBase{
		Name:         pconfig.Name,
		StartCommand: strings.Split(pconfig.Cmd, " "),
		WorkDir:      pconfig.Cwd,
	}
	p.Process = &ProcessPty{ProcessBase: &p}
	p.setProcessConfig(pconfig)
	return &p
}
