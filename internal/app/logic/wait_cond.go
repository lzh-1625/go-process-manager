package logic

import (
	"sync"
	"time"

	"github.com/lzh-1625/go_process_manager/config"
)

type WaitCond struct {
	Cond        sync.Cond
	Ts          int64
	TimeMap     sync.Map
	TriggerChan chan struct{}
}

var (
	ProcessWaitCond *WaitCond
	TaskWaitCond    *WaitCond
)

func InitWaitCond() {
	ProcessWaitCond = newWaitCond()
	TaskWaitCond = newWaitCond()
}

func newWaitCond() *WaitCond {
	wc := &WaitCond{
		Cond:        *sync.NewCond(&sync.Mutex{}),
		Ts:          time.Now().UnixNano(),
		TimeMap:     sync.Map{},
		TriggerChan: make(chan struct{}),
	}
	go wc.timing()
	return wc
}

func (p *WaitCond) Trigger() {
	p.TriggerChan <- struct{}{}
	p.Ts = time.Now().UnixMicro()
}

func (p *WaitCond) timing() { // 添加定时信号清理阻塞协程
	ticker := time.NewTicker(time.Second * time.Duration(config.CF.CondWaitTime))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
		case <-p.TriggerChan:
		}
		ticker.Reset(time.Second * time.Duration(config.CF.CondWaitTime))
		p.Cond.Broadcast()
	}
}
