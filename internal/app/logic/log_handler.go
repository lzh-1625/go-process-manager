package logic

import (
	"time"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/search"
	"github.com/lzh-1625/go_process_manager/log"

	"github.com/panjf2000/ants/v2"
)

type LogHandler struct {
	antsPool  *ants.Pool
	ILogLogic search.ILogLogic
}

func NewLogHandler(ILogLogic search.ILogLogic) *LogHandler {
	antsPool, _ := ants.NewPool(config.CF.LogHandlerPoolSize, ants.WithNonblocking(true), ants.WithExpiryDuration(3*time.Second), ants.WithPanicHandler(func(i any) {
		log.Logger.Warnw("log storage failed", "err", i)
	}))
	return &LogHandler{
		antsPool:  antsPool,
		ILogLogic: ILogLogic,
	}
}

func (l *LogHandler) AddLog(data model.ProcessLog) {
	if err := l.antsPool.Submit(func() {
		l.ILogLogic.Insert(data.Log, data.Name, data.Using, data.Time)
	}); err != nil {
		log.Logger.Warnw("coroutine pool add task failed", "err", err, "current running number", l.antsPool.Running())
	}
}

func (l *LogHandler) GetRunning() int {
	return l.antsPool.Running()
}
