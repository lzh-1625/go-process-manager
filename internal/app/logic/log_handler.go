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
	dirPath := path.Join(utils.UnwarpIgnore(os.UserHomeDir()), ".gpm", "log.diskqueue")
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
	for range max(1, config.CF.LogHandlerPoolSize) {
		go func() {
			var pl model.ProcessLog
			log.Logger.Infow("log handler started")
			for {
				select {
				case <-ctx.Done():
					return
				case msg, ok := <-queue.ReadChan():
					if !ok {
						return
					}
					_ = json.Unmarshal(msg, &pl)
					logLogic.Insert(pl.ID, pl.Log, pl.Name, pl.Using, pl.Time)
				}
			}
		}()
	}
	id := atomic.Int64{}
	id.Store(time.Now().UnixMilli() + queue.Depth())
	return &LogHandler{queue: queue, ILogLogic: logLogic, ctx: ctx, cancel: cancel, id: &id}
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
