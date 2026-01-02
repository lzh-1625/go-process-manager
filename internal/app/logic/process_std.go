package logic

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/google/shlex"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/log"
	"github.com/lzh-1625/go_process_manager/utils"
)

type ProcessStd struct {
	*ProcessBase
	cacheLine []string
	stdin     io.WriteCloser
	stdout    *bufio.Scanner
}

func (p *ProcessStd) Type() eum.TerminalType {
	return eum.TerminalStd
}

func (p *ProcessStd) WriteBytes(input []byte) (err error) {
	p.logReportHandler(config.CF.ProcessInputPrefix + string(input))
	_, err = p.stdin.Write(append(input, '\n'))
	return
}

func (p *ProcessStd) Write(input string) (err error) {
	p.logReportHandler(config.CF.ProcessInputPrefix + input)
	_, err = p.stdin.Write([]byte(input + "\n"))
	return
}

func (p *ProcessStd) Start() (err error) {
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
	cmd := exec.Command(p.StartCommand[0], p.StartCommand[1:]...)
	cmd.Dir = p.WorkDir
	cmd.Env = append(cmd.Env, p.Env...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		log.Logger.Errorw("启动失败，输出管道获取失败", "err", err)
		return err
	}
	p.stdout = bufio.NewScanner(out)
	p.stdin, err = cmd.StdinPipe()
	if err != nil {
		log.Logger.Errorw("启动失败，输入管道获取失败", "err", err)
		return err
	}
	err = cmd.Start()
	if err != nil {
		log.Logger.Errorw("启动失败，进程启动出错:", "err", err)
		return err
	}
	log.Logger.Infow("进程启动成功", "重启次数", p.State.restartTimes)
	p.op = cmd.Process
	p.pInit()
	if !p.SetState(eum.ProcessStateRunning, func() bool {
		return p.State.State == eum.ProcessStateStart
	}) {
		return errors.New("状态异常启动失败")
	}
	p.push("进程启动成功")
	return nil
}

func (p *ProcessStd) ReadCache(ws ConnectInstance) error {
	if len(p.cacheLine) == 0 {
		return errors.New("cache is null")
	}
	for _, line := range p.cacheLine {
		ws.WriteString(line)
	}
	return nil
}

func (p *ProcessStd) SetTerminalSize(cols, rows int) {
	log.Logger.Debug("当前终端不支持修改尺寸")
}

func (p *ProcessStd) readInit() {
	var output string
	log.Logger.Debugw("stdout读取线程已启动", "进程名", p.Name, "使用者", p.GetUserString())
	for {
		select {
		case <-p.StopChan:
			{
				log.Logger.Debugw("stdout读取线程已退出", "进程名", p.Name, "使用者", p.GetUserString())
				return
			}
		default:
			{
				output = p.Read()
				if len(p.ws) == 0 {
					continue
				}
				p.wsLock.Lock()
				for _, v := range p.ws {
					v.WriteString(output)
				}
				p.wsLock.Unlock()
			}
		}
	}
}
func (p *ProcessStd) Read() string {
	if p.stdout.Scan() {
		output := utils.RemoveNotValidUtf8InString(p.stdout.Text())
		p.logReportHandler(output + "\n")
		p.cacheLine = p.cacheLine[1:]
		p.cacheLine = append(p.cacheLine, output)
		return output
	}
	return ""
}

func (p *ProcessStd) watchDog() {
	state, _ := p.op.Wait()
	if p.cgroup.enable && p.cgroup.delete != nil {
		err := p.cgroup.delete()
		if err != nil {
			log.Logger.Errorw("cgroup删除失败", "err", err, "进程名称", p.Name)
		}
	}
	close(p.StopChan)
	p.stdin.Close()
	p.cacheLine = make([]string, config.CF.ProcessMsgCacheLinesLimit)
	p.SetState(eum.ProcessStateStop)
	if state.ExitCode() != 0 {
		log.Logger.Infow("进程停止", "进程名称", p.Name, "exitCode", state.ExitCode(), "进程类型", p.Type())
		p.push(fmt.Sprintf("进程停止,退出码 %d", state.ExitCode()))
	} else {
		log.Logger.Infow("进程正常退出", "进程名称", p.Name)
		p.push("进程正常退出")
	}
	if !p.Config.AutoRestart || p.State.manualStopFlag { // 不重启或手动关闭
		return
	}
	if p.Config.compulsoryRestart { // 强制重启
		p.Start()
		return
	}
	if state.ExitCode() == 0 { // 正常退出
		return
	}
	if p.State.restartTimes < config.CF.ProcessRestartsLimit { // 重启次数未达限制
		p.Start()
		p.State.restartTimes++
		return
	}
	log.Logger.Warnw("重启次数达到上限", "name", p.Name, "limit", config.CF.ProcessRestartsLimit)
	p.SetState(eum.ProcessStateWarnning)
	p.State.Info = "重启次数异常"
	p.push("进程重启次数达到上限")
}

func (p *ProcessStd) pInit() {
	log.Logger.Infow("创建进程成功")
	p.StopChan = make(chan struct{})
	p.State.manualStopFlag = false
	p.State.startTime = time.Now()
	p.ws = make(map[string]ConnectInstance)
	p.cacheLine = make([]string, config.CF.ProcessMsgCacheLinesLimit)
	p.Pid = p.op.Pid
	p.InitPerformanceStatus()
	p.initPsutil()
	p.initCgroup()
	go p.watchDog()
	go p.readInit()
	go p.monitorHandler()
}

func NewProcessStd(pconfig model.Process) *ProcessStd {
	p := &ProcessStd{
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
