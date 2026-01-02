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

// ProcessInfo 进程信息快照 - 用于安全地读取进程状态数据
type ProcessInfo struct {
	Name         string
	Pid          int
	Env          []string
	StartCommand []string
	WorkDir      string

	// 状态信息
	State        eum.ProcessState
	StateInfo    string
	StartTime    string
	RestartTimes int

	// 连接信息
	Users      []string
	Controller string

	// 性能数据
	CPU        []float64
	Mem        []float64
	RecordTime []string

	// 配置
	AutoRestart       bool
	CompulsoryRestart bool
	CgroupEnable      bool
	CPULimit          *float32
	MemoryLimit       *float32
	LogReport         bool
	PushIds           []int64
}

type Process interface {
	ReadCache(ConnectInstance) error
	Write(string) error
	WriteBytes([]byte) error
	Start() error
	Type() eum.TerminalType
	SetTerminalSize(int, int)
	SetOpertor(operator string)
	GetOpertor() string
	SetState(state eum.ProcessState, fn ...func() bool) bool
	GetUserString() string
	GetUserList() []string
	HasWsConn(userName string) bool
	AddConn(user string, c ConnectInstance)
	DeleteConn(user string)
	GetStartTimeFormat() string
	ProcessControl(name string)
	VerifyControl() bool
	ResetRestartTimes()
	InitPerformanceStatus()
	AddCpuUsage(usage float64)
	AddMemUsage(usage float64)
	AddRecordTime()
	Kill() error
	Info() ProcessInfo
	GetState() eum.ProcessState
	GetName() string
	StopSignal() <-chan struct{}
	IsRunning() bool
}

type ProcessBase struct {
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
		PushIds           []int64
		logReport         bool
		cgroupEnable      bool
		memoryLimit       *float32
		cpuLimit          *float32
	}
	State struct {
		startTime      time.Time
		Info           string
		State          eum.ProcessState //0 为未运行，1为运作中，2为异常状态
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

// fn 函数执行成功的情况下对state赋值
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
		log.Logger.Error("已存在连接")
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
	p.Config.PushIds = utils.JsonStrToStruct[[]int64](pconfig.PushIds)
	p.Config.compulsoryRestart = pconfig.CompulsoryRestart
	p.Config.cgroupEnable = pconfig.CgroupEnable
	p.Config.memoryLimit = pconfig.MemoryLimit
	p.Config.cpuLimit = pconfig.CpuLimit
}

func (p *ProcessBase) ResetRestartTimes() {
	p.State.restartTimes = 0
}

func (p *ProcessBase) push(message string) {
	if len(p.Config.PushIds) != 0 {
		messagePlaceholders := map[string]string{
			"{$name}":    p.Name,
			"{$user}":    p.GetUserString(),
			"{$message}": message,
			"{$status}":  strconv.Itoa(int(p.State.State)),
		}
		PushLogic.Push(p.Config.PushIds, messagePlaceholders)
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
	defer log.Logger.Infow("性能监控结束")
	ticker := time.NewTicker(time.Second * time.Duration(config.CF.PerformanceInfoInterval))
	defer ticker.Stop()
	for {
		if p.State.State != eum.ProcessStateRunning {
			log.Logger.Debugw("进程未在运行", "state", p.State.State)
			return
		}
		cpuPercent, err := p.monitor.pu.CPUPercent()
		if err != nil {
			log.Logger.Errorw("CPU使用率获取失败", "err", err)
			return
		}
		memInfo, err := p.monitor.pu.MemoryInfo()
		if err != nil {
			log.Logger.Errorw("内存使用率获取失败", "err", err)
			return
		}
		p.AddRecordTime()
		p.AddCpuUsage(cpuPercent)
		p.AddMemUsage(float64(memInfo.RSS >> 10))
		// log.Logger.Debugw("进程资源使用率获取成功", "cpu", cpuPercent, "mem", memInfo.RSS)
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
		log.Logger.Debug("pu进程获取失败")
	} else {
		p.monitor.enable = true
		log.Logger.Debug("pu进程获取成功")
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
		log.Logger.Errorw("发送SIGINT信号失败", "err", err)
		return p.op.Kill()
	}

	select {
	case <-p.StopChan:
		{
			return nil
		}
	case <-time.After(time.Second * time.Duration(config.CF.KillWaitTime)):
		{
			log.Logger.Debugw("进程kill超时,强制停止进程", "name", p.Name)
			return p.op.Kill()
		}
	}
}

func (p *ProcessBase) Info() ProcessInfo {
	p.State.stateLock.Lock()
	state := p.State.State
	stateInfo := p.State.Info
	startTime := p.State.startTime
	restartTimes := p.State.restartTimes
	p.State.stateLock.Unlock()

	return ProcessInfo{
		Name:              p.Name,
		Pid:               p.Pid,
		Env:               append([]string{}, p.Env...),
		StartCommand:      append([]string{}, p.StartCommand...),
		WorkDir:           p.WorkDir,
		State:             state,
		StateInfo:         stateInfo,
		StartTime:         startTime.Format(time.DateTime),
		RestartTimes:      restartTimes,
		Users:             p.GetUserList(),
		Controller:        p.Control.Controller,
		CPU:               append([]float64{}, p.performanceStatus.cpu...),
		Mem:               append([]float64{}, p.performanceStatus.mem...),
		RecordTime:        append([]string{}, p.performanceStatus.time...),
		AutoRestart:       p.Config.AutoRestart,
		CompulsoryRestart: p.Config.compulsoryRestart,
		CgroupEnable:      p.Config.cgroupEnable,
		CPULimit:          p.Config.cpuLimit,
		MemoryLimit:       p.Config.memoryLimit,
		LogReport:         p.Config.logReport,
		PushIds:           append([]int64{}, p.Config.PushIds...),
	}
}

func (p *ProcessBase) GetName() string {
	return p.Name
}

func (p *ProcessBase) StopSignal() <-chan struct{} {
	return p.StopChan
}

func (p *ProcessBase) IsRunning() bool {
	return p.State.State == eum.ProcessStateRunning
}

func (p *ProcessBase) GetState() eum.ProcessState {
	return p.State.State
}
