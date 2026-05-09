//go:build windows

package logic

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/shlex"
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

func (p *ProcessPty) Start() (err error) {
	defer func() {
		if err != nil {
			p.Config.AutoRestart = false
			p.SetState(eum.ProcessStateWarnning)
			p.State.Info = "process start failed: " + err.Error()
		}
	}()
	if ok := p.SetState(eum.ProcessStateStart, func() bool {
		return p.State.State != eum.ProcessStateRunning && p.State.State != eum.ProcessStateStart
	}); !ok {
		log.Logger.Warnw("process is running, skip start")
		return nil
	}
	pty, err := console.New(100, 100)
	if err != nil {
		log.Logger.Errorw("process start failed", "err", err)
		return err
	}
	pty.SetCWD(p.WorkDir)
	pty.SetENV(p.Env)
	err = pty.Start(p.StartCommand)
	if err != nil {
		log.Logger.Errorw("process start failed", "err", err)
		return err
	}
	p.pty = pty
	pid, err := pty.Pid()
	if err != nil {
		log.Logger.Errorw("process start failed", "err", err)
		return err
	}
	p.op, err = os.FindProcess(pid)
	if err != nil {
		log.Logger.Errorw("process start failed", "err", err)
		return err
	}
	log.Logger.Infow("process start success", "process name", p.Name, "restart times", p.State.restartTimes)
	p.pInit()
	if !p.SetState(eum.ProcessStateRunning, func() bool {
		return p.State.State == eum.ProcessStateStart
	}) {
		return errors.New("state abnormal start failed")
	}
	p.push("process start success")
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
				p.wsLock.Lock()
				for _, v := range p.ws {
					v.Write(buf[:n])
				}
				p.wsLock.Unlock()
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

func (p *ProcessPty) pInit() {
	log.Logger.Infow("create process success")
	p.StopChan = make(chan struct{})
	p.State.manualStopFlag = false
	p.State.startTime = time.Now()
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
	if p.Config.compulsoryRestart { // compulsory restart
		p.Start()
		return
	}
	if state.ExitCode() == 0 { // normal exit
		return
	}
	if p.State.restartTimes < config.CF.ProcessRestartsLimit { // restart times not reached limit
		p.Start()
		p.State.restartTimes++
		return
	}
	log.Logger.Warnw("restart times reached limit", "name", p.Name, "limit", config.CF.ProcessRestartsLimit)
	p.SetState(eum.ProcessStateWarnning)
	p.State.Info = "restart times abnormal"
	p.push("restart times reached limit")
}

func NewProcessPty(pconfig model.Process) *ProcessPty {
	p := &ProcessPty{
		ProcessBase: &ProcessBase{
			Name:         pconfig.Name,
			StartCommand: utils.UnwarpIgnore(shlex.Split(pconfig.Cmd)),
			WorkDir:      pconfig.Cwd,
			Env:          strings.Split(pconfig.Env, ";"),
		},
	}
	p.setProcessConfig(pconfig)
	return p
}
