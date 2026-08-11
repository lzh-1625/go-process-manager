package process

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
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
	lock         sync.Mutex
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
		lock     sync.Mutex
		cond     *sync.Cond
		operator string
		active   int
	}
	cacheBytesBuf *bytes.Buffer
	pty           ptyInterface

	logHandler    io.WriteCloser
	stateHook     func(p *Process, state types.ProcessState)
	addWriterHook func(p *Process, user string, c io.WriteCloser)
	delWriterHook func(p *Process, user string)
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

// GetOperator returns the current operator name.
func (p *Process) GetOperator() string {
	p.operate.lock.Lock()
	defer p.operate.lock.Unlock()
	return p.operate.operator
}

// Perform process modification operations through the operator.
func (p *Process) Operate(operator string, fn func() error) error {
	p.operate.lock.Lock()
	for p.operate.active > 0 && p.operate.operator != operator {
		p.operate.cond.Wait()
	}
	if p.operate.active == 0 {
		p.operate.operator = operator
	}
	p.operate.active++
	p.operate.lock.Unlock()

	defer func() {
		p.operate.lock.Lock()
		defer p.operate.lock.Unlock()
		p.operate.active--
		if p.operate.active == 0 {
			p.operate.operator = ""
			p.operate.cond.Broadcast()
		}
	}()
	return fn()
}

// fn function execution successfully, set state
func (p *Process) setState(state types.ProcessState, afterHookFns ...func()) bool {
	if !p.checkStateChange(p.State.State, state) {
		return false
	}
	p.State.State = state
	if p.stateHook != nil {
		p.stateHook(p, state)
	}
	for _, fn := range afterHookFns {
		fn()
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
	p.Config.CompulsoryRestart = pconfig.CompulsoryRestart
	p.Config.CgroupEnable = pconfig.CgroupEnable
	p.Config.MemoryLimit = pconfig.MemoryLimit
	p.Config.CpuLimit = pconfig.CpuLimit
}

// ResetRestartTimes resets the restart count.
func (p *Process) ResetRestartTimes() {
	p.State.RestartTimes = 0
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
		return p.Kill9()
	}
	p.setState(types.ProcessStateStopping)
	select {
	case <-p.StopChan:
		{
			return nil
		}
	case <-time.After(time.Second * time.Duration(config.CF.KillWaitTime)):
		{
			log.Logger.Debugw("process kill timeout, force stop process", "name", p.Name)
			return p.Kill9()
		}
	}
}

// Stop the process immediately.
func (p *Process) Kill9() error {
	if err := p.op.Kill(); err != nil {
		return err
	}
	<-p.StopChan
	return nil
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
	if p.State.State == types.ProcessStateRunning {
		log.Logger.Warnw("process is running, skip start")
		return nil
	}
	p.lock.Lock()
	defer p.lock.Unlock()
	defer func() {
		if err != nil {
			p.Config.AutoRestart = false
			p.setState(types.ProcessStateWarning)
			p.State.Info = "process start failed: " + err.Error()
		}
	}()
	if ok := p.setState(types.ProcessStateStarting); !ok {
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
	if !p.setState(types.ProcessStateRunning) {
		return errors.New("state abnormal start failed")
	}
	return nil
}

func (p *Process) watchDog() {
	p.pty.Wait()
	state, _ := p.op.Wait()
	p.lock.Lock()
	if p.cgroup.enable && p.cgroup.delete != nil {
		err := p.cgroup.delete()
		if err != nil {
			log.Logger.Errorw("cgroup delete failed", "err", err, "process name", p.Name)
		}
	}
	if p.logHandler != nil {
		p.logHandler.Close()
	}
	if !p.setState(types.ProcessStateStopped) {
		log.Logger.Errorw("", "name", p.Name, "state", p.State.State.String())
		p.lock.Unlock()
		return
	}
	p.pty.Close()
	close(p.StopChan)
	p.lock.Unlock()
	if state.ExitCode() != 0 {
		log.Logger.Infow("process stopped", "process name", p.Name, "exitCode", state.ExitCode())
	} else {
		log.Logger.Infow("process normal exit", "process name", p.Name)
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
	p.setState(types.ProcessStateWarning)
	p.State.Info = "restart times abnormal"
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

// NewProcessPty creates a process and configures its handlers.
func NewProcess(pconfig model.Process, options ...ProcessOptions) *Process {
	p := &Process{
		UUID:         pconfig.UUID,
		Name:         pconfig.Name,
		StartCommand: utils.UnwarpIgnore(shlex.Split(pconfig.Cmd)),
		WorkDir:      pconfig.Cwd,
		Env:          strings.Split(pconfig.Env, ";"),
	}
	p.operate.cond = sync.NewCond(&p.operate.lock)
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

// Write logs to a pipe, then collect them in another goroutine after a newline is read.
// This avoids truncating logs before the line terminator whenever possible.
// Buffered logs are not collected until the next newline is received.
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
		// Read through the next newline, or truncate early to prevent unbounded lines from blocking the terminal buffer.
		scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			for i, b := range data {
				if b == '\n' {
					return i + 1, data[:i], nil
				}
				if i+1 >= 4096 {
					return i + 1, data[:i+1], nil
				}
			}
			if atEOF && len(data) > 0 {
				return len(data), data, nil
			}
			return 0, nil, nil
		})
		for scanner.Scan() {
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

// Read all logs from the buffer immediately.
// Fast log writes may cause logs to be truncated.
func newProcessLogHandler(fn func([]byte)) io.WriteCloser {
	return &processLogHandler{
		fn: fn,
	}
}
