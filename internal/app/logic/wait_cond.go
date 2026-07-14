package logic

import (
	"context"
	"sync"
	"sync/atomic"
)

// WaitCond allows long-polling requests to respond immediately to state changes.
type WaitCond struct {
	Ch      chan struct{}
	Lock    sync.RWMutex
	Version *atomic.Int64
}

var (
	// ProcessWaitCond tracks process-related event changes.
	ProcessWaitCond = sync.OnceValue(func() *WaitCond {
		return newWaitCond()
	})

	// TaskWaitCond tracks task-related event changes.
	TaskWaitCond = sync.OnceValue(func() *WaitCond {
		return newWaitCond()
	})
)

func newWaitCond() *WaitCond {
	return &WaitCond{
		Ch:      make(chan struct{}),
		Version: &atomic.Int64{},
	}
}

// Trigger broadcasts an event.
func (w *WaitCond) Trigger() {
	w.Version.Add(1)
	w.Lock.Lock()
	oldCh := w.Ch
	w.Ch = make(chan struct{})
	w.Lock.Unlock()
	close(oldCh)
}

// Wait blocks until an event is broadcast or the context is canceled.
func (w *WaitCond) Wait(ctx context.Context, version int64) {
	if w.Version.Load() > version {
		return
	}
	w.Lock.RLock()
	ch := w.Ch
	w.Lock.RUnlock()

	select {
	case <-ctx.Done():
	case <-ch:
	}
}
