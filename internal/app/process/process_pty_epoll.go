//go:build epoll && linux

package process

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/containerd/console"
	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/types"
	"github.com/lzh-1625/go_process_manager/log"
	"golang.org/x/sys/unix"

	"github.com/creack/pty"
)

type epoller struct {
	readEpoller *console.Epoller
	exitEpoller *ProcessExitWatcher
}

var globalEpoller = sync.OnceValue(func() *epoller {
	ep, err := console.NewEpoller()
	if err != nil {
		log.Logger.Panic(err)
	}
	ew, err := newProcessExitWatcher(context.TODO())
	if err != nil {
		log.Logger.Panic(err)
	}
	return &epoller{
		exitEpoller: ew,
		readEpoller: ep,
	}
})

type ProcessPty struct {
	*ProcessBase
	cacheBytesBuf *bytes.Buffer
	pty           *console.EpollConsole
	exit          <-chan int
}

// Start starts the process.
func (p *ProcessPty) Start() (err error) {
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
	pf, err := pty.Start(cmd)
	if err != nil || cmd.Process == nil {
		log.Logger.Errorw("process start failed", "err", err)
		return err
	}
	pty.Setsize(pf, &pty.Winsize{
		Rows: 100,
		Cols: 100,
	})
	cs, _, err := console.NewPtyFromFile(pf)
	if err != nil {
		log.Logger.Errorw("console new pty from file failed", "err", err)
		return err
	}
	ep, err := globalEpoller().readEpoller.Add(cs)
	if err != nil {
		log.Logger.Errorw("read epoller add failed", "err", err)
		return err
	}
	p.pty = ep

	exitCh, err := globalEpoller().exitEpoller.Add(cmd.Process.Pid)
	if err != nil {
		log.Logger.Errorw("exit epoller add failed", "err", err)
		return err
	}
	p.exit = exitCh

	log.Logger.Infow("process start success", "process name", p.Name, "restart times", p.State.RestartTimes)
	p.op = cmd.Process
	p.pInit()
	if !p.SetState(types.ProcessStateRunning) {
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

// SetTerminalSize sets the process terminal size.
func (p *ProcessPty) SetTerminalSize(cols, rows int) {
	if cols == 0 || rows == 0 || len(p.writers) != 0 {
		return
	}
	p.pty.Resize(console.WinSize{
		Width:  uint16(cols),
		Height: uint16(rows),
	})

}

// WriteBytes writes data to the process terminal.
func (p *ProcessPty) WriteBytes(input []byte) (err error) {
	_, err = p.pty.Write(input)
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

// ReadCache reads the cached terminal data.
// The process caches some recent output so that terminal clients can view a portion of its output history.
func (p *ProcessPty) ReadCache(ws io.WriteCloser) error {
	if p.cacheBytesBuf == nil {
		return errors.New("cache is null")
	}
	_, err := ws.Write(p.cacheBytesBuf.Bytes())
	return err
}

func (p *ProcessPty) bufHandle(b []byte) {
	p.logReportHandler(b)
	p.cacheBytesBuf.Write(b)
	p.cacheBytesBuf.Next(len(b))
}

func (p *ProcessPty) watchDog() {
	<-p.exit
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

var ErrProcessExitWatcherClosed = errors.New("process exit watcher is closed")

// ProcessExitWatcher reports process exits for PIDs registered with Add.
// It is safe to call Add concurrently with Events and Close.
type ProcessExitWatcher struct {
	epollFD   int
	wakeFD    int
	events    chan int
	stop      chan struct{}
	done      chan struct{}
	closeOnce sync.Once

	mu      sync.Mutex
	closed  bool
	pidByFD map[int]*processExitRegistration
	byPID   map[int]*processExitRegistration
}

type processExitRegistration struct {
	pid   int
	pidfd int
	exits chan int
}

// NewProcessExitWatcher creates a Linux pidfd watcher backed by one epoll fd.
func newProcessExitWatcher(ctx context.Context) (*ProcessExitWatcher, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	epollFD, err := unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if err != nil {
		return nil, fmt.Errorf("create epoll instance: %w", err)
	}
	wakeFD, err := unix.Eventfd(0, unix.EFD_CLOEXEC|unix.EFD_NONBLOCK)
	if err != nil {
		_ = unix.Close(epollFD)
		return nil, fmt.Errorf("create epoll wake event: %w", err)
	}
	if err := unix.EpollCtl(epollFD, unix.EPOLL_CTL_ADD, wakeFD, &unix.EpollEvent{
		Events: unix.EPOLLIN,
		Fd:     int32(wakeFD),
	}); err != nil {
		_ = unix.Close(wakeFD)
		_ = unix.Close(epollFD)
		return nil, fmt.Errorf("add wake event to epoll: %w", err)
	}

	w := &ProcessExitWatcher{
		epollFD: epollFD,
		wakeFD:  wakeFD,
		events:  make(chan int, 64),
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
		pidByFD: make(map[int]*processExitRegistration),
		byPID:   make(map[int]*processExitRegistration),
	}
	go w.run(ctx)
	return w, nil
}

// Add begins listening for pid and returns a channel for that PID's exit.
// Adding an already watched PID returns its existing channel.
func (w *ProcessExitWatcher) Add(pid int) (<-chan int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil, ErrProcessExitWatcherClosed
	}
	if pid <= 0 {
		return nil, fmt.Errorf("invalid pid %d", pid)
	}
	if registration, exists := w.byPID[pid]; exists {
		return registration.exits, nil
	}

	pidfd, err := pidfdOpen(pid)
	if err != nil {
		return nil, fmt.Errorf("open pidfd for %d: %w", pid, err)
	}
	if err := unix.EpollCtl(w.epollFD, unix.EPOLL_CTL_ADD, pidfd, &unix.EpollEvent{
		Events: unix.EPOLLIN,
		Fd:     int32(pidfd),
	}); err != nil {
		_ = unix.Close(pidfd)
		return nil, fmt.Errorf("add pidfd for %d to epoll: %w", pid, err)
	}
	registration := &processExitRegistration{pid: pid, pidfd: pidfd, exits: make(chan int, 1)}
	w.pidByFD[pidfd] = registration
	w.byPID[pid] = registration
	return registration.exits, nil
}

// Close stops the watcher and waits until every kernel fd has been released.
func (w *ProcessExitWatcher) Close() {
	w.requestClose()
	<-w.done
}

func (w *ProcessExitWatcher) requestClose() {
	w.closeOnce.Do(func() {
		w.mu.Lock()
		w.closed = true
		w.mu.Unlock()
		close(w.stop)
		_, _ = unix.Write(w.wakeFD, []byte{1, 0, 0, 0, 0, 0, 0, 0})
	})
}

func (w *ProcessExitWatcher) run(ctx context.Context) {
	defer w.cleanup()
	events := make([]unix.EpollEvent, 64)

	for {
		if ctx.Err() != nil {
			w.requestClose()
		}
		n, err := unix.EpollWait(w.epollFD, events, 100)
		if err == unix.EINTR {
			continue
		}
		if err != nil {
			return
		}
		for _, event := range events[:n] {
			fd := int(event.Fd)
			if fd == w.wakeFD {
				var value [8]byte
				_, _ = unix.Read(w.wakeFD, value[:])
				select {
				case <-w.stop:
					return
				default:
				}
				continue
			}

			w.mu.Lock()
			registration, ok := w.pidByFD[fd]
			if ok {
				delete(w.pidByFD, fd)
				delete(w.byPID, registration.pid)
				_ = unix.EpollCtl(w.epollFD, unix.EPOLL_CTL_DEL, fd, nil)
				_ = unix.Close(fd)
			}
			w.mu.Unlock()
			if !ok {
				continue
			}
			registration.exits <- registration.pid
			close(registration.exits)
			select {
			case w.events <- registration.pid:
			case <-w.stop:
				return
			}
		}
	}
}

func (w *ProcessExitWatcher) cleanup() {
	w.mu.Lock()
	for fd, registration := range w.pidByFD {
		_ = unix.Close(fd)
		close(registration.exits)
	}
	w.pidByFD = nil
	w.byPID = nil
	_ = unix.Close(w.wakeFD)
	_ = unix.Close(w.epollFD)
	w.mu.Unlock()
	close(w.events)
	close(w.done)
}

func pidfdOpen(pid int) (int, error) {
	fd, _, errno := unix.Syscall(unix.SYS_PIDFD_OPEN, uintptr(pid), 0, 0)
	if errno != 0 {
		return 0, errno
	}
	return int(fd), nil
}
