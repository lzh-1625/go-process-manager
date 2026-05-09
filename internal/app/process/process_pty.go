//go:build !windows

package process

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/log"
	"github.com/lzh-1625/go_process_manager/utils"

	"github.com/creack/pty"
)

type ProcessPty struct {
	*ProcessBase
	cacheBytesBuf *bytes.Buffer
	pty           *os.File
}

func (p *ProcessPty) Start() (err error) {
	defer func() {
		if err != nil {
			p.Config.AutoRestart = false
			p.SetState(eum.ProcessStateWarnning)
			p.State.Info = "process start failed: " + err.Error()
		}
	}()
	if ok := p.SetState(eum.ProcessStateStart, func() bool {
		return p.State.State != eum.ProcessStateStart && p.State.State != eum.ProcessStateRunning
	}); !ok {
		log.Logger.Warnw("process is running, skip start")
		return nil
	}
	cmd := exec.Command(p.StartCommand[0], p.StartCommand[1:]...)
	cmd.Dir = p.WorkDir
	cmd.Env = append(cmd.Env, p.Env...)
	pf, err := pty.Start(cmd)
	if err != nil || cmd.Process == nil {
		log.Logger.Errorw("process start failed", "err", err)
		return err
	}
	pty.Setsize(pf, &pty.Winsize{
		Rows: 100,
		Cols: 100,
	})
	p.pty = pf
	log.Logger.Infow("process start success", "process name", p.Name, "restart times", p.State.RestartTimes)
	p.op = cmd.Process
	p.pInit()
	if !p.SetState(eum.ProcessStateRunning, func() bool {
		return p.State.State == eum.ProcessStateStart
	}) {
		return errors.New("state abnormal start failed")
	}
	p.push("process start success")
	return nil
}

func (p *ProcessPty) pInit() {
	log.Logger.Infow("create process success")
	p.StopChan = make(chan struct{})
	p.State.manualStopFlag = false
	p.State.StartTime = time.Now()
	p.ws = make(map[string]ConnectInstance)
	p.Pid = p.op.Pid
	p.cacheBytesBuf = bytes.NewBuffer(make([]byte, config.CF.ProcessMsgCacheBufLimit))
	p.InitPerformanceStatus()
	p.initPsutil()
	p.initCgroup()
	go p.watchDog()
	go p.readInit()
	go p.monitorHandler()
}

func (p *ProcessPty) SetTerminalSize(cols, rows int) {
	if cols == 0 || rows == 0 || len(p.ws) != 0 {
		return
	}
	if err := pty.Setsize(p.pty, &pty.Winsize{
		Cols: uint16(cols),
		Rows: uint16(rows),
	}); err != nil {
		log.Logger.Error("set terminal size failed", "err", err)
	}

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
	log.Logger.Debugw("stdout read thread started", "process name", p.Name, "user", p.GetUserString())
	buf := make([]byte, 1024)
	for {
		select {
		case <-p.StopChan:
			{
				log.Logger.Debugw("stdout read thread exited", "process name", p.Name, "user", p.GetUserString())
				return
			}
		default:
			{
				n, err := p.pty.Read(buf)
				if err != nil {
					log.Logger.Debugw("stdout read failed", "err", err)
					return
				}
				p.bufHandle(buf[:n])
				if len(p.ws) == 0 {
					continue
				}
				p.wsLock.RLock()
				for _, v := range p.ws {
					v.Write(buf[:n])
				}
				p.wsLock.RUnlock()
			}
		}
	}
}

func (p *ProcessPty) ReadCache(ws ConnectInstance) error {
	if p.cacheBytesBuf == nil {
		return errors.New("cache is null")
	}
	ws.Write(p.cacheBytesBuf.Bytes())
	return nil
}

func (p *ProcessPty) bufHandle(b []byte) {
	log := strings.TrimSpace(string(b))
	if utils.RemoveANSI(log) != "" {
		p.logReportHandler(log)
	}
	p.cacheBytesBuf.Write(b)
	p.cacheBytesBuf.Next(len(b))
}

func (p *ProcessPty) watchDog() {
	state, _ := p.op.Wait()
	if p.cgroup.enable && p.cgroup.delete != nil {
		err := p.cgroup.delete()
		if err != nil {
			log.Logger.Errorw("cgroup delete failed", "err", err, "process name", p.Name)
		}
	}
	if !p.SetState(eum.ProcessStateStop, func() bool {
		// process is already stopped or warning state, no need to repeat set state
		if eum.ProcessStateStop == p.State.State || eum.ProcessStateWarnning == p.State.State {
			return false
		}
		close(p.StopChan)
		p.pty.Close()
		return true
	}) {
		return
	}
	if state.ExitCode() != 0 {
		log.Logger.Infow("process stopped", "process name", p.Name, "exitCode", state.ExitCode())
		p.push(fmt.Sprintf("process stopped, exit code %d", state.ExitCode()))
	} else {
		log.Logger.Infow("process normal exit", "process name", p.Name)
		p.push("process normal exit")
	}
	if !p.Config.AutoRestart || p.State.manualStopFlag { // not restart or manual close
		return
	}
	if p.Config.CompulsoryRestart { // compulsory restart
		p.Start()
		return
	}
	if state.ExitCode() == 0 { // normal exit
		return
	}
	if p.State.RestartTimes < config.CF.ProcessRestartsLimit { // restart times not reached limit
		p.Start()
		p.State.RestartTimes++
		return
	}
	log.Logger.Warnw("restart times reached limit", "name", p.Name, "limit", config.CF.ProcessRestartsLimit)
	p.SetState(eum.ProcessStateWarnning)
	p.State.Info = "restart times abnormal"
	p.push("restart times reached limit")
}
