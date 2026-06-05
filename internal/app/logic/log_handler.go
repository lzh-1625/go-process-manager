package logic

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"sync/atomic"
	"time"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/search"
	"github.com/lzh-1625/go_process_manager/log"
	"github.com/lzh-1625/go_process_manager/utils"

	"github.com/nsqio/go-diskqueue"
)

type LogHandler struct {
	queue     diskqueue.Interface
	ILogLogic search.ILogLogic
	ctx       context.Context
	cancel    context.CancelFunc
	id        *atomic.Int64
}

func NewLogHandler(logLogic search.ILogLogic) *LogHandler {
	dirPath := path.Join(config.CF.ConfigDir, "log.diskqueue")
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		log.Logger.Panic(err)
	}
	queue := diskqueue.New(
		"log",
		dirPath,
		1024*1024*10,
		4,
		1024*1024,
		1000,
		5*time.Second,
		func(lvl diskqueue.LogLevel, f string, args ...any) {
			switch lvl {
			case diskqueue.DEBUG, diskqueue.INFO:
				log.Logger.Debugf(f, args...)
			case diskqueue.WARN:
				log.Logger.Warnf(f, args...)
			case diskqueue.ERROR:
				log.Logger.Errorf(f, args...)
			}
		})
	ctx, cancel := context.WithCancel(context.Background())
	id := atomic.Int64{}
	id.Store(time.Now().UnixMicro() + queue.Depth())
	lh := &LogHandler{queue: queue, ILogLogic: logLogic, ctx: ctx, cancel: cancel, id: &id}
	for range config.CF.LogHandlerPoolSize {
		go lh.worker(ctx)
	}
	return lh
}

func (l *LogHandler) AddLog(data model.ProcessLog) {
	data.ID = l.id.Add(1)
	l.queue.Put(utils.Unwarp(json.Marshal(data)))
}

func (l *LogHandler) GetRunning() int {
	return int(l.queue.Depth())
}

func (l *LogHandler) Close() {
	l.queue.Close()
	l.cancel()
}

func (l *LogHandler) worker(ctx context.Context) {
	logs := make([]model.ProcessLog, 0, 100)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-l.queue.ReadChan():
			if !ok {
				if len(logs) > 0 {
					l.ILogLogic.Insert(logs...)
				}
				return
			}
			var pl model.ProcessLog
			_ = json.Unmarshal(msg, &pl)
			logs = append(logs, pl)
			if len(logs) == cap(logs) {
				l.ILogLogic.Insert(logs...)
				logs = logs[:0]
				ticker.Reset(time.Second)
			}
		case <-ticker.C:
			if len(logs) > 0 {
				l.ILogLogic.Insert(logs...)
				logs = logs[:0]
			}
		}
	}
}
