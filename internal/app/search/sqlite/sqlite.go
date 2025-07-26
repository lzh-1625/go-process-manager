package sqlite

import (
	"slices"
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
	query := search.QueryStringAnalysis(req.Match.Log)
	data, total := repository.LogRepository.SearchLog(req, query)
	for _, v := range slices.DeleteFunc(query, func(q search.Query) bool {
		return q.Cond == search.NotMatch || q.Cond == search.NotWildCard
	}) {
		for i := range data {
			data[i].Log = strings.ReplaceAll(data[i].Log, v.Content, "\033[43m"+v.Content+"\033[0m")
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
