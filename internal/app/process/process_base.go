package process

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/google/shlex"
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
	ws     map[string]io.WriteCloser
	wsLock sync.RWMutex
	Config struct {
		AutoRestart       bool
		CompulsoryRestart bool
		PushIDs           []int64
		LogReport         bool
		CgroupEnable      bool
		MemoryLimit       *float32
		CpuLimit          *float32
	}
	State struct {
		StartTime      time.Time
		Info           string
		State          eum.ProcessState //0 not running, 1 running, 2 warning state
		stateLock      sync.Mutex
		RestartTimes   int
		manualStopFlag bool
	}
	PerformanceStatus struct {
		Cpu  []float64
		Mem  []float64
		Time []string
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

	LogHandler  IProcessLogHandler
	StateHook   func(p *ProcessBase, state eum.ProcessState)
	AddCoonHook func(p *ProcessBase, user string, c io.WriteCloser)
	DelCoonHook func(p *ProcessBase, user string)
	PushHandle  func(p *ProcessBase, pushIDs []int64, messagePlaceholders map[string]string)
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
	if p.StateHook != nil {
		p.StateHook(p, state)
	}
	return true
}

func (p *ProcessBase) GetUserString() string {
	return strings.Join(p.GetUserList(), ";")
}

func (p *ProcessBase) GetUserList() []string {
	p.wsLock.RLock()
	defer p.wsLock.RUnlock()
	userList := make([]string, 0, len(p.ws))
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

func (p *ProcessBase) AddConn(user string, c io.WriteCloser) {
	p.wsLock.Lock()
	defer p.wsLock.Unlock()

	if p.ws[user] != nil {
		log.Logger.Error("connection already exists")
		return
	}

	p.ws[user] = c
	if p.AddCoonHook != nil {
		p.AddCoonHook(p, user, c)
	}
}

func (p *ProcessBase) DeleteConn(user string) {
	p.wsLock.Lock()
	defer p.wsLock.Unlock()
	delete(p.ws, user)
	if p.DelCoonHook != nil {
		p.DelCoonHook(p, user)
	}
}

func (p *ProcessBase) logReportHandler(log []byte) {
	if p.LogHandler != nil {
		p.LogHandler.Write(log)
	}
}

func (p *ProcessBase) GetStartTimeFormat() string {
	return p.State.StartTime.Format(time.DateTime)
}

func (p *ProcessBase) ProcessControl(name string) {
	p.Control.changControlTime = time.Now()
	p.Control.Controller = name
	for _, ws := range p.ws {
		ws.Close()
	}
}

// not being controlled or control time expired
func (p *ProcessBase) VerifyControl() bool {
	return p.Control.Controller == "" || p.Control.changControlTime.Unix() < time.Now().Unix()-config.CF.ProcessControlExpireTime
}

func (p *ProcessBase) setProcessConfig(pconfig model.Process) {
	p.Config.AutoRestart = pconfig.AutoRestart
	p.Config.LogReport = pconfig.LogReport
	p.Config.PushIDs = utils.JsonStrToStruct[[]int64](pconfig.PushIDs)
	p.Config.CompulsoryRestart = pconfig.CompulsoryRestart
	p.Config.CgroupEnable = pconfig.CgroupEnable
	p.Config.MemoryLimit = pconfig.MemoryLimit
	p.Config.CpuLimit = pconfig.CpuLimit
}

func (p *ProcessBase) ResetRestartTimes() {
	p.State.RestartTimes = 0
}

func (p *ProcessBase) push(message string) {
	if len(p.Config.PushIDs) != 0 {
		messagePlaceholders := map[string]string{
			"{$name}":    p.Name,
			"{$user}":    p.GetUserString(),
			"{$message}": message,
			"{$status}":  strconv.Itoa(int(p.State.State)),
		}
		if p.PushHandle != nil {
			p.PushHandle(p, p.Config.PushIDs, messagePlaceholders)
		}
	}
}

func (p *ProcessBase) InitPerformanceStatus() {
	p.PerformanceStatus.Cpu = make([]float64, config.CF.PerformanceInfoListLength)
	p.PerformanceStatus.Mem = make([]float64, config.CF.PerformanceInfoListLength)
	p.PerformanceStatus.Time = make([]string, config.CF.PerformanceInfoListLength)
}

func (p *ProcessBase) AddCpuUsage(usage float64) {
	p.PerformanceStatus.Cpu = append(p.PerformanceStatus.Cpu[1:], usage)
}

func (p *ProcessBase) AddMemUsage(usage float64) {
	p.PerformanceStatus.Mem = append(p.PerformanceStatus.Mem[1:], usage)
}

func (p *ProcessBase) AddRecordTime() {
	p.PerformanceStatus.Time = append(p.PerformanceStatus.Time[1:], time.Now().Format(time.DateTime))
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

type ProcessOptions func(*ProcessBase)

// state change hook
func SetStateHook(fn func(p *ProcessBase, state eum.ProcessState)) ProcessOptions {
	return func(p *ProcessBase) {
		p.StateHook = fn
	}
}

// ws connect hook
func SetAddCoonHook(fn func(p *ProcessBase, user string, c io.WriteCloser)) ProcessOptions {
	return func(p *ProcessBase) {
		p.AddCoonHook = fn
	}
}

// ws disconnect hook
func SetDelCoonHook(fn func(p *ProcessBase, user string)) ProcessOptions {
	return func(p *ProcessBase) {
		p.DelCoonHook = fn
	}
}

// log handle hook
func SetLogHandler(fn func(p *ProcessBase) IProcessLogHandler) ProcessOptions {
	return func(p *ProcessBase) {
		p.LogHandler = fn(p)
	}
}

// push handle hook
func SetPushHandle(fn func(p *ProcessBase, pushIDs []int64, messagePlaceholders map[string]string)) ProcessOptions {
	return func(p *ProcessBase) {
		p.PushHandle = fn
	}
}

func NewProcessPty(pconfig model.Process, options ...ProcessOptions) *ProcessPty {
	p := &ProcessPty{
		ProcessBase: &ProcessBase{
			Name:         pconfig.Name,
			StartCommand: utils.UnwarpIgnore(shlex.Split(pconfig.Cmd)),
			WorkDir:      pconfig.Cwd,
			Env:          strings.Split(pconfig.Env, ";"),
		},
	}

	for _, option := range options {
		option(p.ProcessBase)
	}

	p.setProcessConfig(pconfig)
	return p
}

type IProcessLogHandler interface {
	io.WriteCloser
}

type processLogHandlerByPipe struct {
	pr *io.PipeReader
	pw *io.PipeWriter
	fn func([]byte)
}

func (p *processLogHandlerByPipe) Write(log []byte) (int, error) {
	return p.pw.Write(log)
}

func (p *processLogHandlerByPipe) Close() error {
	p.pr.Close()
	p.pw.Close()
	return nil
}

func NewProcessLogHandlerByPipe(fn func([]byte)) IProcessLogHandler {
	pr, pw := io.Pipe()
	pl := &processLogHandlerByPipe{
		pr: pr,
		pw: pw,
	}
	go func() {
		scanner := bufio.NewScanner(pr)
		if err := scanner.Err(); err != nil {
			log.Logger.Warn(err)
			return
		}
		for scanner.Scan() {
			if fn == nil {
				continue
			}
			fn(scanner.Bytes())
		}
		log.Logger.Debugw("process log handler by pipe closed")
	}()
	return pl
}

type processLogHandler struct {
	fn func([]byte)
}

func (p *processLogHandler) Write(log []byte) (int, error) {
	p.fn(log)
	return len(log), nil
}

func (p *processLogHandler) Close() error {
	return nil
}

func NewProcessLogHandler(fn func([]byte)) IProcessLogHandler {
	return &processLogHandler{
		fn: fn,
	}
}
