package sqlite

import (
	"errors"
	"slices"
	"strings"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/search"
	logger "github.com/lzh-1625/go_process_manager/log"
)

type sqliteSearch struct {
	logRepository *repository.LogRepository
}

func NewSqliteSearch(logRepository *repository.LogRepository) search.ILogLogic {
	return &sqliteSearch{
		logRepository: logRepository,
	}
}

func (l *sqliteSearch) Search(req model.GetLogReq) model.LogResp {
	query := search.QueryStringAnalysis(req.Match.Log)
	data, total := l.logRepository.SearchLog(req, query)

	if req.Match.HighLight {
		for _, v := range slices.DeleteFunc(query, func(q search.Query) bool {
			return q.Cond == search.NotMatch || q.Cond == search.NotWildCard
		}) {
			for i := range data {
				data[i].Log = strings.ReplaceAll(data[i].Log, v.Content, "\033[43m"+v.Content+"\033[0m")
			}
		}
	}

	return model.LogResp{
		Data:  data,
		Total: total,
	}
}

func (l *sqliteSearch) Insert(logs ...model.ProcessLog) {
	if err := l.logRepository.InsertLog(logs...); err != nil {
		logger.Logger.Errorw("Log insert failed", "err", err)
	}
}

func (l *sqliteSearch) Reload() error {
	return errors.New("sqlite not support reload")
}
