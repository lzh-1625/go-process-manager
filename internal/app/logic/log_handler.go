package logic

import (
	"encoding/json"
	"os"
	"path"
	"sync"
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
	wg        *sync.WaitGroup
}

func NewLogHandler(logLogic search.ILogLogic) *LogHandler {
	dirPath := path.Join(utils.UnwarpIgnore(os.UserHomeDir()), ".gpm", "log.diskqueue")
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		log.Logger.Panic(err)
	}
	queue := diskqueue.New(
		"log",
		dirPath,
		1024*1024*512*1,
		4,
		1024*1024*10,
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

	wg := &sync.WaitGroup{}
	for range max(1, config.CF.LogHandlerPoolSize) {
		wg.Go(func() {
			var pl model.ProcessLog
			for msg := range queue.ReadChan() {
				_ = json.Unmarshal(msg, &pl)
				logLogic.Insert(pl.Log, pl.Name, pl.Using, pl.Time)
			}
		})
	}
	return &LogHandler{queue: queue, ILogLogic: logLogic, wg: wg}
}

func (l *LogHandler) AddLog(data model.ProcessLog) {
	l.queue.Put(utils.Unwarp(json.Marshal(data)))
}

func (l *LogHandler) GetRunning() int {
	return int(l.queue.Depth())
}

func (l *LogHandler) Close() {
	l.queue.Close()
	l.wg.Wait()
}
