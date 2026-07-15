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
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/types"
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
		Controller         string
		ControlExpiredTime time.Time
	}
	writers map[string]io.WriteCloser
	wlock   sync.RWMutex
	Config  struct {
		AutoRestart       bool
		CompulsoryRestart bool // Restart automatically after reaching the restart limit when CompulsoryRestart is true.
		PushIDs           []int64
		LogReport         bool
		CgroupEnable      bool
		MemoryLimit       *float32
		CpuLimit          *float32
		logHandlerPipe    bool
		logHandlerFn      func(p *ProcessBase, log []byte)
	}
	State struct {
		StartTime      time.Time
		Info           string
		State          types.ProcessState //0 not running, 1 running, 2 warning state
		stateLock      sync.Mutex
		RestartTimes   int
		manualStopFlag bool
	}
	PerformanceStatus struct {
		Cpu  []float64
		Mem  []float64
		Time []time.Time
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

	logHandler    io.WriteCloser
	stateHook     func(p *ProcessBase, state types.ProcessState)
	addWriterHook func(p *ProcessBase, user string, c io.WriteCloser)
	delWriterHook func(p *ProcessBase, user string)
	pushHandle    func(p *ProcessBase, pushIDs []int64, messagePlaceholders map[string]string)
}

// SetOpertor sets the current process operator for a limited time.
func (p *ProcessBase) SetOpertor(operator string) {
	if p.operate.user.CompareAndSwap(nil, &operator) {
		p.operate.time = time.Now()
	}
}

// GetOpertor returns the current operator name and clears it.
func (p *ProcessBase) GetOpertor() string {
	s := p.operate.user.Swap(nil)
	if p.operate.time.Unix() < time.Now().Unix()-int64(config.CF.KillWaitTime) || s == nil {
		return ""
	}
	return *s
}

// fn function execution successfully, set state
// The process state cannot change while fn is running.
func (p *ProcessBase) SetState(state types.ProcessState, fn ...func() bool) bool {
	p.State.stateLock.Lock()
	defer p.State.stateLock.Unlock()
	if !p.checkStateChange(p.State.State, state) {
		return false
	}
	for _, v := range fn {
		if !v() {
			return false
		}
	}
	p.State.State = state
	if p.stateHook != nil {
		p.stateHook(p, state)
	}
	return true
}

func (p *ProcessBase) checkStateChange(old, new types.ProcessState) bool {
	switch old {
	case types.ProcessStateStarting:
		return new == types.ProcessStateRunning || new == types.ProcessStateWarning
	case types.ProcessStateRunning:
		return new == types.ProcessStateStopping || new == types.ProcessStateStopped
	case types.ProcessStateWarning, types.ProcessStateStopped:
		return new == types.ProcessStateStarting
	case types.ProcessStateStopping:
		return new == types.ProcessStateStopped
	default:
		return true
	}
}

// GetUserString returns the formatted list of terminal users for the current process.
func (p *ProcessBase) GetUserString() string {
	return strings.Join(p.GetUserList(), ";")
}

// GetUserList returns the terminal users for the current process.
func (p *ProcessBase) GetUserList() []string {
	p.wlock.RLock()
	defer p.wlock.RUnlock()
	userList := make([]string, 0, len(p.writers))
	for i := range p.writers {
		userList = append(userList, i)
	}
	return userList
}

// HasWriter reports whether the current terminal has the specified writer.
func (p *ProcessBase) HasWriter(userName string) bool {
	p.wlock.RLock()
	defer p.wlock.RUnlock()
	return p.writers[userName] != nil
}

// AddWriter adds a terminal writer.
func (p *ProcessBase) AddWriter(user string, c io.WriteCloser) {
	p.wlock.Lock()
	defer p.wlock.Unlock()

	if p.writers[user] != nil {
		log.Logger.Error("connection already exists")
		return
	}

	p.writers[user] = c
	if p.addWriterHook != nil {
		p.addWriterHook(p, user, c)
	}
}

// DeleteWriter removes a terminal writer.
func (p *ProcessBase) DeleteWriter(user string) {
	p.wlock.Lock()
	defer p.wlock.Unlock()
	delete(p.writers, user)
	if p.delWriterHook != nil {
		p.delWriterHook(p, user)
	}
}

func (p *ProcessBase) logReportHandler(log []byte) {
	if p.Config.LogReport && p.logHandler != nil {
		p.logHandler.Write(log)
	}
}

// ProcessControl disconnects all current users and makes the specified user the controller.
// Other users cannot operate the process terminal, and control is released automatically after a timeout.
func (p *ProcessBase) ProcessControl(name string) {
	p.Control.ControlExpiredTime = time.Now().Add(time.Second * time.Duration(config.CF.ProcessControlExpireTime))
	p.Control.Controller = name
	for _, ws := range p.writers {
		ws.Close()
	}
}

// not being controlled or control time expired
func (p *ProcessBase) VerifyControl() bool {
	return p.Control.Controller == "" || time.Now().After(p.Control.ControlExpiredTime)
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

// ResetRestartTimes resets the restart count.
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
		if p.pushHandle != nil {
			p.pushHandle(p, p.Config.PushIDs, messagePlaceholders)
		}
	}
}

func (p *ProcessBase) initPerformanceStatus() {
	p.PerformanceStatus.Cpu = make([]float64, config.CF.PerformanceInfoListLength)
	p.PerformanceStatus.Mem = make([]float64, config.CF.PerformanceInfoListLength)
	p.PerformanceStatus.Time = make([]time.Time, config.CF.PerformanceInfoListLength)
}

func (p *ProcessBase) addPerformanceRecord(cpu, mem float64) {
	p.PerformanceStatus.Cpu = append(p.PerformanceStatus.Cpu[1:], cpu)
	p.PerformanceStatus.Mem = append(p.PerformanceStatus.Mem[1:], mem)
	p.PerformanceStatus.Time = append(p.PerformanceStatus.Time[1:], time.Now())
}

// fetch performance information, return cpu usage and memory usage in KB
func (p *ProcessBase) GetPerformanceInfo() (float64, float64, error) {
	if p.monitor.pu == nil {
		return 0, 0, errors.New("process not running")
	}

	cpuPercent, err := p.monitor.pu.CPUPercent()
	if err != nil {
		return 0, 0, err
	}
	memInfo, err := p.monitor.pu.MemoryInfo()
	if err != nil {
		return 0, 0, err
	}
	return cpuPercent, float64(memInfo.RSS >> 10), nil
}

func (p *ProcessBase) monitorHandler() {
	if !p.monitor.enable {
		return
	}
	defer log.Logger.Infow("performance monitoring ended")
	ticker := time.NewTicker(time.Second * time.Duration(config.CF.PerformanceInfoInterval))
	defer ticker.Stop()
	for {
		if p.State.State != types.ProcessStateRunning {
			log.Logger.Debugw("process not running", "state", p.State.State)
			return
		}

		c, m, err := p.GetPerformanceInfo()
		if err != nil {
			log.Logger.Debugw("performance monitor exit", "err", err)
			return
		}
		p.addPerformanceRecord(c, m)
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

// Kill stops the process by sending SIGINT first, then forcibly kills it if it does not exit in time.
func (p *ProcessBase) Kill() error {
	if p.State.State != types.ProcessStateRunning {
		return errors.New("can't kill not running process")
	}
	p.State.manualStopFlag = true
	if err := p.op.Signal(syscall.SIGINT); err != nil {
		log.Logger.Errorw("send SIGINT signal failed", "err", err)
		return p.op.Kill()
	}
	p.SetState(types.ProcessStateStopping)
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

// Stop the process immediately.
func (p *ProcessBase) Kill9() error {
	return p.op.Kill()
}

func (p *ProcessBase) initLogHandler() {
	if p.Config.logHandlerFn == nil {
		return
	}
	if p.Config.logHandlerPipe {
		p.logHandler = newProcessLogHandlerByPipe(func(b []byte) {
			p.Config.logHandlerFn(p, b)
		})
	} else {
		p.logHandler = newProcessLogHandler(func(b []byte) {
			p.Config.logHandlerFn(p, b)
		})
	}
}

type ProcessOptions func(*ProcessBase)

// state change hook
func SetStateHook(fn func(p *ProcessBase, state types.ProcessState)) ProcessOptions {
	return func(p *ProcessBase) {
		p.stateHook = fn
	}
}

// ws connect hook
func SetAddWriterHook(fn func(p *ProcessBase, user string, c io.WriteCloser)) ProcessOptions {
	return func(p *ProcessBase) {
		p.addWriterHook = fn
	}
}

// ws disconnect hook
func SetDelWriterHook(fn func(p *ProcessBase, user string)) ProcessOptions {
	return func(p *ProcessBase) {
		p.delWriterHook = fn
	}
}

// log handle hook
func SetLogHandler(pipe bool, fn func(p *ProcessBase, log []byte)) ProcessOptions {
	return func(p *ProcessBase) {
		p.Config.logHandlerFn = fn
		p.Config.logHandlerPipe = pipe
	}
}

// push handle hook
func SetPushHandle(fn func(p *ProcessBase, pushIDs []int64, messagePlaceholders map[string]string)) ProcessOptions {
	return func(p *ProcessBase) {
		p.pushHandle = fn
	}
}

// NewProcessPty creates a process and configures its handlers.
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

func newProcessLogHandlerByPipe(fn func([]byte)) io.WriteCloser {
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

func newProcessLogHandler(fn func([]byte)) io.WriteCloser {
	return &processLogHandler{
		fn: fn,
	}
}
