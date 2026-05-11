package repository

import (
	"context"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/query"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/search"
)

func NewLogRepository(query *query.Query) *LogRepository {
	return &LogRepository{
		query: query,
	}
}

type LogRepository struct {
	query *query.Query
}

func (l *LogRepository) InsertLog(data model.ProcessLog) error {
	return l.query.ProcessLog.Create(&data)
}

func (l *LogRepository) SearchLog(req model.GetLogReq, logQuery []search.Query) (result []*model.ProcessLog, total int64) {
	q := l.query.ProcessLog.WithContext(context.TODO())
	if req.Match.Name != "" {
		q = q.Where(l.query.ProcessLog.Name.Eq(req.Match.Name))
	}
	if req.Match.Using != "" {
		q = q.Where(l.query.ProcessLog.Using.Eq(req.Match.Using))
	}

	for _, v := range logQuery {
		switch v.Cond {
		case search.Match, search.WildCard:
			q = q.Where(l.query.ProcessLog.Log.Like("%" + v.Content + "%"))
		case search.NotMatch, search.NotWildCard:
			q = q.Where(l.query.ProcessLog.Log.NotLike("%" + v.Content + "%"))
		}
	}

	if req.Sort == "desc" {
		q = q.Order(l.query.ProcessLog.Time.Desc())
	}
	if req.TimeRange.StartTime != 0 {
		q = q.Where(l.query.ProcessLog.Time.Gte(req.TimeRange.StartTime))
	}
	if req.TimeRange.EndTime != 0 {
		q = q.Where(l.query.ProcessLog.Time.Lt(req.TimeRange.EndTime))
	}
	if len(req.FilterName) != 0 {
		q = q.Where(l.query.ProcessLog.Name.In(req.FilterName...))
	}
	result, total, _ = q.FindByPage(req.Page.From, req.Page.Size)
	return
}
