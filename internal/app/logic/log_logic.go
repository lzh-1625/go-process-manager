package logic

import (
	"strings"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/log"
)

type LogLogic interface {
	Search(req model.GetLogReq, filterProcessName ...string) model.LogResp
	Insert(log string, processName string, using string, ts int64)
}

var LogLogicImpl LogLogic

func InitLog() {
	switch config.CF.StorgeType {
	case "es":
		LogLogicImpl = LogEs
		EsLogic.InitEs()
		log.Logger.Infow("使用es作为日志存储")
	case "bleve":
		LogLogicImpl = BleveLogic
		BleveLogic.InitBleve()
		log.Logger.Infow("使用bleve作为日志存储")
	default:
		LogLogicImpl = LogSqlite
		log.Logger.Infow("使用sqlite作为日志存储")
	}
}

type logSqlite struct{}

var LogSqlite = new(logSqlite)

func (l *logSqlite) Search(req model.GetLogReq, filterProcessName ...string) model.LogResp {
	req.FilterName = filterProcessName
	data, total := repository.LogRepository.SearchLog(req)
	if req.Match.Log != "" {
		for i := range data {
			data[i].Log = strings.ReplaceAll(data[i].Log, req.Match.Log, "\033[43m"+req.Match.Log+"\033[0m")
		}
	}
	return model.LogResp{
		Data:  data,
		Total: total,
	}
}

func (l *logSqlite) Insert(log string, processName string, using string, ts int64) {
	repository.LogRepository.InsertLog(model.ProcessLog{
		Log:   log,
		Name:  processName,
		Using: using,
		Time:  ts,
	})
}

type logEs struct{}

var LogEs = new(logEs)

func (l *logEs) Search(req model.GetLogReq, filterProcessName ...string) model.LogResp {
	return EsLogic.Search(req, filterProcessName...)
}

func (l *logEs) Insert(log string, processName string, using string, ts int64) {
	EsLogic.Insert(log, processName, using, ts)
}
