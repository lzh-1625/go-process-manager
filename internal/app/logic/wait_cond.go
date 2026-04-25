package logic

import (
	"context"
	"sync"
	"sync/atomic"
)

type WaitCond struct {
	Ch      chan struct{}
	Lock    sync.RWMutex
	Version *atomic.Int64
}

var (
	ProcessWaitCond *WaitCond = newWaitCond()
	TaskWaitCond    *WaitCond = newWaitCond()
)

func newWaitCond() *WaitCond {
	wc := &WaitCond{
		Ch:      make(chan struct{}),
		Version: &atomic.Int64{},
	}
	return wc
}

func (w *WaitCond) Trigger() {
	w.Version.Add(1)
	w.Lock.Lock()
	oldCh := w.Ch
	w.Ch = make(chan struct{})
	w.Lock.Unlock()
	close(oldCh)
}

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
