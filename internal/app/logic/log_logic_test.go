package logic

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/search"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/search/bleve"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/search/es"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/search/sqlite"
	"github.com/lzh-1625/go_process_manager/log"
)

var testDatas []model.ProcessLog = []model.ProcessLog{
	{
		ID:    1,
		Log:   "test content debug",
		Name:  "test",
		Using: "test",
		Time:  1000,
	},
	{
		ID:    2,
		Log:   "test content error",
		Name:  "test1",
		Using: "test",
		Time:  2000,
	},
	{
		ID:    3,
		Log:   "test content warn",
		Name:  "test",
		Using: "test",
		Time:  3000,
	},
	{
		ID:    4,
		Log:   "test content filter info",
		Name:  "test",
		Using: "test",
		Time:  4000,
	},
	{
		ID:    5,
		Log:   "test content filter",
		Name:  "test",
		Using: "test",
		Time:  5000,
	},
	{
		ID:    6,
		Log:   "test content filter",
		Name:  "test2",
		Using: "user",
		Time:  6000,
	},
}

func req() model.GetLogReq {
	return model.GetLogReq{
		Page: struct {
			From int `json:"from"`
			Size int `json:"size"`
		}{
			From: 0,
			Size: 100,
		},
	}
}

func TestSearchLog(t *testing.T) {
	config.CF.ConfigDir = t.TempDir()
	var bleve = bleve.NewBleveSearch()
	var sqlite = sqlite.NewSqliteSearch(repository.NewLogRepository(repository.NewQuery(repository.NewDB())))
	var es = es.NewEsSearch()

	test(t, bleve)
	test(t, sqlite)

	if config.CF.EsIndex != "" && config.CF.EsUrl != "" {
		config.CF.EsIndex = "server_log_test_" + time.Now().Format("20060102150405")
		test(t, es)
	}

}

func test(t *testing.T, logic search.ILogLogic) {
	logic.Insert(testDatas...)

	time.Sleep(1 * time.Second)

	reqs := []model.GetLogReq{}
	reqs = append(reqs, req())

	r := req()
	r.CursorID = 3
	r.Sort = "desc"
	reqs = append(reqs, r)

	r = req()
	r.CursorID = 3
	r.Sort = "asc"
	reqs = append(reqs, r)

	r = req()
	r.Match.Log = "test content"
	r.Match.Using = "test"
	reqs = append(reqs, r)

	r = req()
	r.Match.Log = "!warn !filter"
	reqs = append(reqs, r)

	r = req()
	r.FilterName = []string{"test1", "test2"}
	reqs = append(reqs, r)

	r = req()
	r.Match.Using = "user"
	reqs = append(reqs, r)

	for _, v := range reqs {
		if !check(logic, v) {
			t.Errorf("data is not correct")
		}
	}
}

func check(logic search.ILogLogic, req model.GetLogReq) bool {
	if req.Page.Size == 0 {
		req.Page.Size = 100
	}
	result1 := []model.ProcessLog{}
	for _, v := range testDatas {
		if req.Match.Name != "" && v.Name != req.Match.Name {
			continue
		}
		if req.Match.Using != "" && v.Using != req.Match.Using {
			continue
		}
		if req.TimeRange.StartTime != 0 && v.Time < req.TimeRange.StartTime {
			continue
		}
		if req.TimeRange.EndTime != 0 && v.Time >= req.TimeRange.EndTime {
			continue
		}
		if len(req.FilterName) != 0 && !slices.Contains(req.FilterName, v.Name) {
			continue
		}

		if req.CursorID != 0 && req.Sort == "desc" {
			if v.ID >= req.CursorID {
				continue
			}
		} else {
			if v.ID <= req.CursorID {
				continue
			}
		}

		if req.Match.Log != "" {
			match := true
			query := search.QueryStringAnalysis(req.Match.Log)
			for _, q := range query {
				switch q.Cond {
				case search.Match, search.WildCard:
					if !strings.Contains(strings.ToLower(v.Log), strings.ToLower(q.Content)) {
						match = false
					}
				case search.NotMatch, search.NotWildCard:
					if strings.Contains(strings.ToLower(v.Log), strings.ToLower(q.Content)) {
						match = false
					}
				}
			}
			if !match {
				continue
			}
		}

		result1 = append(result1, v)
	}

	resp := logic.Search(req)
	if len(result1) != len(resp.Data) {
		log.Logger.Errorw("data is not correct", "result1", result1, "resp", resp, "req", req)
		return false
	}
	for _, v := range resp.Data {
		if !slices.ContainsFunc(result1, func(vv model.ProcessLog) bool {
			return vv.ID == v.ID
		}) {
			log.Logger.Errorw("data is not correct", "result1", result1, "resp", resp, "req", req)
			return false
		}
	}
	return true
}
