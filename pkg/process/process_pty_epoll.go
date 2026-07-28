//go:build epoll && linux

package process

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sync"

	"github.com/containerd/console"
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
		panic(err)
	}
	ew, err := newProcessExitWatcher(context.TODO())
	if err != nil {
		panic(err)
	}
	return &epoller{
		exitEpoller: ew,
		readEpoller: ep,
	}
})

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

type ptyImpl struct {
	*console.EpollConsole
	exit <-chan int
}

func (p *ptyImpl) SetSize(cols, rows int) error {
	return p.Console.Resize(console.WinSize{
		Width:  uint16(cols),
		Height: uint16(rows),
	})
}

func (p *ptyImpl) Wait() { <-p.exit }

// Start starts the process.
func startWithPty(cmd *exec.Cmd) (ptyInterface, error) {
	pf, err := pty.Start(cmd)
	if err != nil || cmd.Process == nil {
		logger.Error("process start failed", "err", err)
		return nil, err
	}
	pty.Setsize(pf, &pty.Winsize{
		Rows: 100,
		Cols: 100,
	})
	cs, _, err := console.NewPtyFromFile(pf)
	if err != nil {
		logger.Error("console new pty from file failed", "err", err)
		return nil, err
	}
	ep, err := globalEpoller().readEpoller.Add(cs)
	if err != nil {
		logger.Error("read epoller add failed", "err", err)
		return nil, err
	}

	exitCh, err := globalEpoller().exitEpoller.Add(cmd.Process.Pid)
	if err != nil {
		logger.Error("exit epoller add failed", "err", err)
		return nil, err
	}

	return &ptyImpl{ep, exitCh}, nil
}
