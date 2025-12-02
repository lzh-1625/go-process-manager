package repository

import (
	"context"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/query"
	"github.com/lzh-1625/go_process_manager/internal/app/search"
	"github.com/lzh-1625/go_process_manager/log"
)

type logRepository struct{}

var LogRepository = new(logRepository)

func (l *logRepository) InsertLog(data model.ProcessLog) {
	if err := db.Create(&data).Error; err != nil {
		log.Logger.Errorw("日志插入失败", "err", err)
	}
}

func (l *logRepository) SearchLog(req model.GetLogReq, logQuery []search.Query) (result []*model.ProcessLog, total int64) {
	q := query.ProcessLog.WithContext(context.TODO())
	if req.Match.Name != "" {
		q = q.Where(query.ProcessLog.Name.Eq(req.Match.Name))
	}
	if req.Match.Using != "" {
		q = q.Where(query.ProcessLog.Using.Eq(req.Match.Using))
	}

	for _, v := range logQuery {
		switch v.Cond {
		case search.Match, search.WildCard:
			q = q.Where(query.ProcessLog.Log.Like("%" + v.Content + "%"))
		case search.NotMatch, search.NotWildCard:
			q = q.Where(query.ProcessLog.Log.NotLike("%" + v.Content + "%"))
		}
	}

	if req.Sort == "desc" {
		q = q.Order(query.ProcessLog.Time.Desc())
	}
	if req.TimeRange.StartTime != 0 {
		q = q.Where(query.ProcessLog.Time.Gt(req.TimeRange.StartTime))
	}
	if req.TimeRange.EndTime != 0 {
		q = q.Where(query.ProcessLog.Time.Lt(req.TimeRange.EndTime))
	}
	if len(req.FilterName) != 0 {
		q = q.Where(query.ProcessLog.Name.In(req.FilterName...))
	}
	result, total, _ = q.FindByPage(req.Page.From, req.Page.Size)
	return
}
