package es

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	sr "github.com/lzh-1625/go_process_manager/internal/app/search"
	"github.com/lzh-1625/go_process_manager/log"

	"github.com/olivere/elastic/v7"
)

func init() {
	sr.Register("es", new(esSearch))
}

type esSearch struct {
	esClient *elastic.Client
}

func (e *esSearch) Init() error {
	var err error
	e.esClient, err = elastic.NewClient(
		elastic.SetURL(config.CF.EsUrl),
		elastic.SetBasicAuth(config.CF.EsUsername, config.CF.EsPassword),
		elastic.SetSniff(false),
		elastic.SetHttpClient(&http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: config.CF.LogHandlerPoolSize,
				IdleConnTimeout:     90 * time.Second,
			},
		}),
	)
	if err != nil {
		log.Logger.Warnw("Failed to connect to es", "err", err)
		return err
	}
	return nil
}

func (e *esSearch) Insert(logContent string, processName string, using string, ts int64) {
	data := model.ProcessLog{
		Log:   logContent,
		Name:  processName,
		Using: using,
		Time:  ts,
	}
	_, err := e.esClient.Index().Index(config.CF.EsIndex).BodyJson(data).Do(context.TODO())
	if err != nil {
		log.Logger.Errorw("es数据插入失败", "err", err)
	}
}

func (e *esSearch) Search(req model.GetLogReq, filterProcessName ...string) model.LogResp {
	// 检查 req 是否为 nil
	if req.Page.From < 0 || req.Page.Size <= 0 {
		log.Logger.Error("无效的分页请求参数")
		return model.LogResp{Total: 0, Data: []*model.ProcessLog{}}
	}

	search := e.esClient.Search(config.CF.EsIndex).From(req.Page.From).Size(req.Page.Size).TrackScores(true)
	if req.Sort == "asc" {
		search.Sort("time", true)
	}
	if req.Sort == "desc" {
		search.Sort("time", false)
	}

	queryList := []elastic.Query{}
	timeRangeQuery := elastic.NewRangeQuery("time")
	if req.TimeRange.StartTime != 0 {
		queryList = append(queryList, timeRangeQuery.Gte(req.TimeRange.StartTime))
	}
	if req.TimeRange.EndTime != 0 {
		queryList = append(queryList, timeRangeQuery.Lte(req.TimeRange.EndTime))
	}
	notQuery := []elastic.Query{}

	for _, v := range sr.QueryStringAnalysis(req.Match.Log) {
		switch v.Cond {
		case sr.Match:
			queryList = append(queryList, elastic.NewMatchPhraseQuery("log", v.Content))
		case sr.NotMatch:
			notQuery = append(notQuery, elastic.NewMatchPhraseQuery("log", v.Content))
		case sr.WildCard:
			queryList = append(queryList, elastic.NewWildcardQuery("log.keyword", "*"+v.Content+"*"))
		case sr.NotWildCard:
			notQuery = append(notQuery, elastic.NewWildcardQuery("log.keyword", "*"+v.Content+"*"))
		}
		fmt.Printf("v.Cond: %v\n", v.Cond)
		fmt.Printf("v.Content: %v\n", v.Content)
	}

	if req.Match.Name != "" {
		queryList = append(queryList, elastic.NewMatchQuery("name", req.Match.Name))
	}
	if req.Match.Using != "" {
		queryList = append(queryList, elastic.NewMatchQuery("using", req.Match.Using))
	}
	if len(filterProcessName) != 0 { // 过滤进程名
		shouldQueryList := []elastic.Query{}
		for _, fpn := range filterProcessName {
			shouldQueryList = append(shouldQueryList, elastic.NewMatchQuery("name", fpn))
		}
		if len(shouldQueryList) > 0 {
			shouldQuery := elastic.NewBoolQuery().Should(shouldQueryList...)
			queryList = append(queryList, shouldQuery)
		}
	}

	result := model.LogResp{}
	resp, err := search.Query(elastic.NewBoolQuery().Must(queryList...).MustNot(notQuery...)).Highlight(elastic.NewHighlight().Field("log").PreTags("\033[43m").PostTags("\033[0m")).Do(context.TODO())
	if err != nil {
		log.Logger.Errorw("es查询失败", "err", err, "reason", resp.Error.Reason)
		return result
	}

	// 遍历响应结果
	for _, v := range resp.Hits.Hits {
		if v.Source != nil {
			var data model.ProcessLog
			if err := json.Unmarshal(v.Source, &data); err == nil {
				if len(v.Highlight) > 0 && len(v.Highlight["log"]) > 0 {
					data.Log = v.Highlight["log"][0]
				}
				result.Data = append(result.Data, &data)
			} else {
				log.Logger.Errorw("JSON 解码失败", "err", err)
			}
		}
	}

	result.Total = resp.TotalHits()
	return result
}
