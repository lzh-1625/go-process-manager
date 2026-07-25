package process

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
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

type ptyInterface interface {
	io.ReadWriteCloser
	SetSize(cols, rows int) error
	Wait()
}

type Process struct {
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
		logHandlerFn      func(p *Process, log []byte)
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
	cacheBytesBuf *bytes.Buffer
	pty           ptyInterface

	logHandler    io.WriteCloser
	stateHook     func(p *Process, state types.ProcessState)
	addWriterHook func(p *Process, user string, c io.WriteCloser)
	delWriterHook func(p *Process, user string)
	pushHandle    func(p *Process, pushIDs []int64, messagePlaceholders map[string]string)
}

// SetOpertor sets the current process operator for a limited time.
func (p *Process) SetOpertor(operator string) {
	if p.operate.user.CompareAndSwap(nil, &operator) {
		p.operate.time = time.Now()
	}
}

func (p *Process) bufHandle(b []byte) {
	p.logReportHandler(b)
	p.cacheBytesBuf.Write(b)
	p.cacheBytesBuf.Next(len(b))
}

// ReadCache reads the cached terminal data.
// The process caches some recent output so that terminal clients can view a portion of its output history.
func (p *Process) ReadCache(ws io.WriteCloser) error {
	if p.cacheBytesBuf == nil {
		return errors.New("cache is null")
	}
	_, err := ws.Write(p.cacheBytesBuf.Bytes())
	return err
}

// GetOpertor returns the current operator name and clears it.
func (p *Process) GetOpertor() string {
	s := p.operate.user.Swap(nil)
	if p.operate.time.Unix() < time.Now().Unix()-int64(config.CF.KillWaitTime) || s == nil {
		return ""
	}
	return *s
}

// fn function execution successfully, set state
// The process state cannot change while fn is running.
func (p *Process) SetState(state types.ProcessState, fn ...func() bool) bool {
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

func (p *Process) checkStateChange(old, new types.ProcessState) bool {
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
func (p *Process) GetUserString() string {
	return strings.Join(p.GetUserList(), ";")
}

// GetUserList returns the terminal users for the current process.
func (p *Process) GetUserList() []string {
	p.wlock.RLock()
	defer p.wlock.RUnlock()
	userList := make([]string, 0, len(p.writers))
	for i := range p.writers {
		userList = append(userList, i)
	}
	return userList
}

// HasWriter reports whether the current terminal has the specified writer.
func (p *Process) HasWriter(userName string) bool {
	p.wlock.RLock()
	defer p.wlock.RUnlock()
	return p.writers[userName] != nil
}

// AddWriter adds a terminal writer.
func (p *Process) AddWriter(user string, c io.WriteCloser) {
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
func (p *Process) DeleteWriter(user string) {
	p.wlock.Lock()
	defer p.wlock.Unlock()
	delete(p.writers, user)
	if p.delWriterHook != nil {
		p.delWriterHook(p, user)
	}
}

func (p *Process) logReportHandler(log []byte) {
	if p.Config.LogReport && p.logHandler != nil {
		p.logHandler.Write(log)
	}
}

// ProcessControl disconnects all current users and makes the specified user the controller.
// Other users cannot operate the process terminal, and control is released automatically after a timeout.
func (p *Process) ProcessControl(name string) {
	p.Control.ControlExpiredTime = time.Now().Add(time.Second * time.Duration(config.CF.ProcessControlExpireTime))
	p.Control.Controller = name
	for _, ws := range p.writers {
		ws.Close()
	}
}

// not being controlled or control time expired
func (p *Process) VerifyControl() bool {
	return p.Control.Controller == "" || time.Now().After(p.Control.ControlExpiredTime)
}

func (p *Process) setProcessConfig(pconfig model.Process) {
	p.Config.AutoRestart = pconfig.AutoRestart
	p.Config.LogReport = pconfig.LogReport
	p.Config.PushIDs = utils.JsonStrToStruct[[]int64](pconfig.PushIDs)
	p.Config.CompulsoryRestart = pconfig.CompulsoryRestart
	p.Config.CgroupEnable = pconfig.CgroupEnable
	p.Config.MemoryLimit = pconfig.MemoryLimit
	p.Config.CpuLimit = pconfig.CpuLimit
}

// ResetRestartTimes resets the restart count.
func (p *Process) ResetRestartTimes() {
	p.State.RestartTimes = 0
}

func (p *Process) push(message string) {
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

func (p *Process) initPerformanceStatus() {
	p.PerformanceStatus.Cpu = make([]float64, config.CF.PerformanceInfoListLength)
	p.PerformanceStatus.Mem = make([]float64, config.CF.PerformanceInfoListLength)
	p.PerformanceStatus.Time = make([]time.Time, config.CF.PerformanceInfoListLength)
}

func (p *Process) addPerformanceRecord(cpu, mem float64) {
	p.PerformanceStatus.Cpu = append(p.PerformanceStatus.Cpu[1:], cpu)
	p.PerformanceStatus.Mem = append(p.PerformanceStatus.Mem[1:], mem)
	p.PerformanceStatus.Time = append(p.PerformanceStatus.Time[1:], time.Now())
}

// fetch performance information, return cpu usage and memory usage in KB
func (p *Process) GetPerformanceInfo() (float64, float64, error) {
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

func (p *Process) monitorHandler() {
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

func (p *Process) initPsutil() {
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
func (p *Process) Kill() error {
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
func (p *Process) Kill9() error {
	return p.op.Kill()
}

func (p *Process) initLogHandler() {
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

func (p *Process) readInit() {
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
				if len(p.writers) == 0 {
					continue
				}
				p.wlock.RLock()
				for _, v := range p.writers {
					v.Write(buf[:n])
				}
				p.wlock.RUnlock()
			}
		}
	}
}

// WriteBytes writes data to the process terminal.
func (p *Process) WriteBytes(input []byte) (err error) {
	_, err = p.pty.Write(input)
	return
}

// SetTerminalSize sets the process terminal size.
func (p *Process) SetTerminalSize(cols, rows int) {
	if cols == 0 || rows == 0 || len(p.writers) != 0 {
		return
	}
	p.pty.SetSize(cols, rows)
}

func (p *Process) pInit() {
	log.Logger.Infow("create process success")
	p.StopChan = make(chan struct{})
	p.State.manualStopFlag = false
	p.State.StartTime = time.Now()
	p.writers = make(map[string]io.WriteCloser)
	p.Pid = p.op.Pid
	p.cacheBytesBuf = bytes.NewBuffer(make([]byte, config.CF.ProcessMsgCacheBufLimit))
	p.initPerformanceStatus()
	p.initPsutil()
	p.initCgroup()
	p.initLogHandler()
	go p.watchDog()
	go p.readInit()
	go p.monitorHandler()
}

// Start starts the process.
func (p *Process) Start() (err error) {
	defer func() {
		if err != nil {
			p.Config.AutoRestart = false
			p.SetState(types.ProcessStateWarning)
			p.State.Info = "process start failed: " + err.Error()
		}
	}()
	if ok := p.SetState(types.ProcessStateStarting); !ok {
		log.Logger.Warnw("process is running, skip start")
		return nil
	}
	cmd := exec.Command(p.StartCommand[0], p.StartCommand[1:]...)
	cmd.Dir = p.WorkDir
	cmd.Env = append(os.Environ(), p.Env...)
	pty, err := startWithPty(cmd)
	if err != nil {
		log.Logger.Warnw("process pty init failed")
		return err
	}
	p.pty = pty
	log.Logger.Infow("process start success", "process name", p.Name, "restart times", p.State.RestartTimes)
	p.op = cmd.Process
	p.pInit()
	if !p.SetState(types.ProcessStateRunning) {
		return errors.New("state abnormal start failed")
	}
	p.push("process start success")
	return nil
}

func (p *Process) watchDog() {
	p.pty.Wait()
	state, _ := p.op.Wait()
	if p.cgroup.enable && p.cgroup.delete != nil {
		err := p.cgroup.delete()
		if err != nil {
			log.Logger.Errorw("cgroup delete failed", "err", err, "process name", p.Name)
		}
	}
	if p.logHandler != nil {
		p.logHandler.Close()
	}
	if !p.SetState(types.ProcessStateStopped, func() bool {
		// process is already stopped or warning state, no need to repeat set state
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
	p.SetState(types.ProcessStateWarning)
	p.State.Info = "restart times abnormal"
	p.push("restart times reached limit")
}

type ProcessOptions func(*Process)

// state change hook
func SetStateHook(fn func(p *Process, state types.ProcessState)) ProcessOptions {
	return func(p *Process) {
		p.stateHook = fn
	}
}

// ws connect hook
func SetAddWriterHook(fn func(p *Process, user string, c io.WriteCloser)) ProcessOptions {
	return func(p *Process) {
		p.addWriterHook = fn
	}
}

// ws disconnect hook
func SetDelWriterHook(fn func(p *Process, user string)) ProcessOptions {
	return func(p *Process) {
		p.delWriterHook = fn
	}
}

// log handle hook
func SetLogHandler(pipe bool, fn func(p *Process, log []byte)) ProcessOptions {
	return func(p *Process) {
		p.Config.logHandlerFn = fn
		p.Config.logHandlerPipe = pipe
	}
}

// push handle hook
func SetPushHandle(fn func(p *Process, pushIDs []int64, messagePlaceholders map[string]string)) ProcessOptions {
	return func(p *Process) {
		p.pushHandle = fn
	}
}

// NewProcessPty creates a process and configures its handlers.
func NewProcess(pconfig model.Process, options ...ProcessOptions) *Process {
	p := &Process{
		UUID:         pconfig.UUID,
		Name:         pconfig.Name,
		StartCommand: utils.UnwarpIgnore(shlex.Split(pconfig.Cmd)),
		WorkDir:      pconfig.Cwd,
		Env:          strings.Split(pconfig.Env, ";"),
	}

	for _, option := range options {
		option(p)
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
