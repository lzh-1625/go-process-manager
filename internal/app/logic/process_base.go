package logic

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/log"
	"github.com/lzh-1625/go_process_manager/utils"

	pu "github.com/shirou/gopsutil/process"
)

type ProcessBase struct {
	UUID         int
	op           *os.Process
	Name         string
	Pid          int
	StartCommand []string
	WorkDir      string
	Env          []string
	Lock         sync.Mutex
	StopChan     chan struct{}
	Control      struct {
		Controller       string
		changControlTime time.Time
	}
	ws     map[string]ConnectInstance
	wsLock sync.RWMutex
	Config struct {
		AutoRestart       bool
		compulsoryRestart bool
		PushIDs           []int64
		logReport         bool
		cgroupEnable      bool
		memoryLimit       *float32
		cpuLimit          *float32
	}
	State struct {
		startTime      time.Time
		Info           string
		State          eum.ProcessState //0 not running, 1 running, 2 warning state
		stateLock      sync.Mutex
		restartTimes   int
		manualStopFlag bool
	}
	performanceStatus struct {
		cpu  []float64
		mem  []float64
		time []string
	}
	monitor struct {
		enable bool
		pu     *pu.Process
	}
	cgroup struct {
		enable bool
		delete func() error
	}
	operate struct {
		user atomic.Pointer[string]
		time time.Time
	}
}
type ConnectInstance interface {
	Write([]byte)
	WriteString(string)
	Cancel()
}

func (p *ProcessBase) SetOpertor(operator string) {
	if p.operate.user.CompareAndSwap(nil, &operator) {
		p.operate.time = time.Now()
	}
}

func (p *ProcessBase) GetOpertor() string {
	s := p.operate.user.Swap(nil)
	if p.operate.time.Unix() < time.Now().Unix()-int64(config.CF.KillWaitTime) || s == nil {
		return ""
	}
	return *s
}

// fn function execution successfully, set state
func (p *ProcessBase) SetState(state eum.ProcessState, fn ...func() bool) bool {
	p.State.stateLock.Lock()
	defer p.State.stateLock.Unlock()
	for _, v := range fn {
		if !v() {
			return false
		}
	}
	p.State.State = state
	ProcessWaitCond.Trigger()
	p.createEvent(state)
	go TaskLogic.RunTaskByTriggerEvent(p.Name, state)
	return true
}

func (p *ProcessBase) createEvent(state eum.ProcessState) {
	var eventType eum.EventType
	kv := []string{}
	switch state {
	case eum.ProcessStateRunning:
		eventType = eum.EventProcessStart
		kv = append(kv, "restartTimes", strconv.Itoa(p.State.restartTimes))
	case eum.ProcessStateStop:
		eventType = eum.EventProcessStop
		kv = append(kv, "startTime", p.State.startTime.Format(time.DateTime))
	case eum.ProcessStateWarnning:
		eventType = eum.EventProcessWarning
		kv = append(kv, "reason", p.State.Info, "startTime", p.State.startTime.Format(time.DateTime))
	default:
		return
	}
	kv = append(kv, "operator", p.GetOpertor())
	EventLogic.Create(p.Name, eventType, kv...)
}

func (p *ProcessBase) GetUserString() string {
	return strings.Join(p.GetUserList(), ";")
}

func (p *ProcessBase) GetUserList() []string {
	userList := make([]string, 0, len(p.ws))
	p.wsLock.RLock()
	defer p.wsLock.RUnlock()
	for i := range p.ws {
		userList = append(userList, i)
	}
	return userList
}

func (p *ProcessBase) HasWsConn(userName string) bool {
	p.wsLock.RLock()
	defer p.wsLock.RUnlock()
	return p.ws[userName] != nil
}

func (p *ProcessBase) AddConn(user string, c ConnectInstance) {
	p.wsLock.Lock()
	defer p.wsLock.Unlock()

	if p.ws[user] != nil {
		log.Logger.Error("connection already exists")
		return
	}

	p.ws[user] = c
	ProcessWaitCond.Trigger()
}

func (p *ProcessBase) DeleteConn(user string) {
	p.wsLock.Lock()
	defer p.wsLock.Unlock()
	delete(p.ws, user)
	ProcessWaitCond.Trigger()
}

func (p *ProcessBase) logReportHandler(log string) {
	if p.Config.logReport && len([]rune(log)) > config.CF.LogMinLenth {
		Loghandler.AddLog(model.ProcessLog{
			Log:   log,
			Using: p.GetUserString(),
			Name:  p.Name,
			Time:  time.Now().UnixMilli(),
		})
	}
}

func (p *ProcessBase) GetStartTimeFormat() string {
	return p.State.startTime.Format(time.DateTime)
}

func (p *ProcessBase) ProcessControl(name string) {
	p.Control.changControlTime = time.Now()
	p.Control.Controller = name
	for _, ws := range p.ws {
		ws.Cancel()
	}
}

// 没人在使用或控制时间过期
func (p *ProcessBase) VerifyControl() bool {
	return p.Control.Controller == "" || p.Control.changControlTime.Unix() < time.Now().Unix()-config.CF.ProcessExpireTime
}

func (p *ProcessBase) setProcessConfig(pconfig model.Process) {
	p.Config.AutoRestart = pconfig.AutoRestart
	p.Config.logReport = pconfig.LogReport
	p.Config.PushIDs = utils.JsonStrToStruct[[]int64](pconfig.PushIDs)
	p.Config.compulsoryRestart = pconfig.CompulsoryRestart
	p.Config.cgroupEnable = pconfig.CgroupEnable
	p.Config.memoryLimit = pconfig.MemoryLimit
	p.Config.cpuLimit = pconfig.CpuLimit
}

func (p *ProcessBase) ResetRestartTimes() {
	p.State.restartTimes = 0
}

func (p *ProcessBase) push(message string) {
	if len(p.Config.PushIDs) != 0 {
		messagePlaceholders := map[string]string{
			"{$name}":    p.Name,
			"{$user}":    p.GetUserString(),
			"{$message}": message,
			"{$status}":  strconv.Itoa(int(p.State.State)),
		}
		PushLogic.Push(p.Config.PushIDs, messagePlaceholders)
	}
}

func (p *ProcessBase) InitPerformanceStatus() {
	p.performanceStatus.cpu = make([]float64, config.CF.PerformanceInfoListLength)
	p.performanceStatus.mem = make([]float64, config.CF.PerformanceInfoListLength)
	p.performanceStatus.time = make([]string, config.CF.PerformanceInfoListLength)
}

func (p *ProcessBase) AddCpuUsage(usage float64) {
	p.performanceStatus.cpu = append(p.performanceStatus.cpu[1:], usage)
}

func (p *ProcessBase) AddMemUsage(usage float64) {
	p.performanceStatus.mem = append(p.performanceStatus.mem[1:], usage)
}

func (p *ProcessBase) AddRecordTime() {
	p.performanceStatus.time = append(p.performanceStatus.time[1:], time.Now().Format(time.DateTime))
}

func (p *ProcessBase) monitorHandler() {
	if !p.monitor.enable {
		return
	}
	defer log.Logger.Infow("performance monitoring ended")
	ticker := time.NewTicker(time.Second * time.Duration(config.CF.PerformanceInfoInterval))
	defer ticker.Stop()
	for {
		if p.State.State != eum.ProcessStateRunning {
			log.Logger.Debugw("process not running", "state", p.State.State)
			return
		}
		cpuPercent, err := p.monitor.pu.CPUPercent()
		if err != nil {
			log.Logger.Errorw("CPU usage get failed", "err", err)
			return
		}
		memInfo, err := p.monitor.pu.MemoryInfo()
		if err != nil {
			log.Logger.Errorw("memory usage get failed", "err", err)
			return
		}
		p.AddRecordTime()
		p.AddCpuUsage(cpuPercent)
		p.AddMemUsage(float64(memInfo.RSS >> 10))
		select {
		case <-ticker.C:
		case <-p.StopChan:
			return
		}
	}
}

func (p *ProcessBase) initPsutil() {
	pup, err := pu.NewProcess(int32(p.Pid))
	if err != nil {
		p.monitor.enable = false
		log.Logger.Debug("pu process get failed")
	} else {
		p.monitor.enable = true
		log.Logger.Debug("pu process get success")
		p.monitor.pu = pup
	}
}

func (p *ProcessBase) Kill() error {
	if p.State.State != eum.ProcessStateRunning {
		return errors.New("can't kill not running process")
	}
	p.State.stateLock.Lock()
	p.State.manualStopFlag = true
	p.State.stateLock.Unlock()

	if err := p.op.Signal(syscall.SIGINT); err != nil {
		log.Logger.Errorw("send SIGINT signal failed", "err", err)
		return p.op.Kill()
	}

	select {
	case <-p.StopChan:
		{
			return nil
		}
	case <-time.After(time.Second * time.Duration(config.CF.KillWaitTime)):
		{
			log.Logger.Debugw("process kill timeout, force stop process", "name", p.Name)
			return p.op.Kill()
		}
	}
}
