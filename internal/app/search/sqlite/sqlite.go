package sqlite

import (
	"strings"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/internal/app/search"
)

func init() {
	search.Register("sqlite", new(sqliteSearch))
}

type sqliteSearch struct{}

func (l *sqliteSearch) Search(req model.GetLogReq, filterProcessName ...string) model.LogResp {
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

func (l *sqliteSearch) Insert(log string, processName string, using string, ts int64) {
	repository.LogRepository.InsertLog(model.ProcessLog{
		Log:   log,
		Name:  processName,
		Using: using,
		Time:  ts,
	})
}

func (l *sqliteSearch) Init() error {
	return nil
}
