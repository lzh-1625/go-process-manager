package logic

import (
	"time"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/log"

	"github.com/panjf2000/ants/v2"
)

type loghandler struct {
	antsPool *ants.Pool
}

var (
	Loghandler = new(loghandler)
)

func InitLogHandle() {
	Loghandler.antsPool, _ = ants.NewPool(config.CF.LogHandlerPoolSize, ants.WithNonblocking(true), ants.WithExpiryDuration(3*time.Second), ants.WithPanicHandler(func(i any) {
		log.Logger.Warnw("log storage failed", "err", i)
	}))
}

func (l *loghandler) AddLog(data model.ProcessLog) {
	if err := l.antsPool.Submit(func() {
		LogLogicImpl.Insert(data.Log, data.Name, data.Using, data.Time)
	}); err != nil {
		log.Logger.Warnw("coroutine pool add task failed", "err", err, "current running number", l.antsPool.Running())
	}
}

func (l *loghandler) GetRunning() int {
	return l.antsPool.Running()
}
