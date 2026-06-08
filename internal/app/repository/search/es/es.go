package es

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/search"
	sr "github.com/lzh-1625/go_process_manager/internal/app/repository/search"
	"github.com/lzh-1625/go_process_manager/log"

	"github.com/olivere/elastic/v7"
)

func NewEsSearch() search.ILogLogic {
	e := &esSearch{}
	e.init()
	return e
}

type esSearch struct {
	esClient *elastic.Client
}

func (e *esSearch) Reload() error {
	return e.init()
}

func (e *esSearch) init() error {
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

func (e *esSearch) Insert(logs ...model.ProcessLog) {
	reqs := make([]elastic.BulkableRequest, 0, len(logs))
	for _, v := range logs {
		req := elastic.NewBulkCreateRequest()
		req.Index(config.CF.EsIndex)
		req.Doc(v)
		reqs = append(reqs, req)
	}
	_, err := e.esClient.Bulk().Add(reqs...).Do(context.TODO())
	if err != nil {
		log.Logger.Errorw("es insert failed", "err", err)
	}
}

func (e *esSearch) Search(req model.GetLogReq) model.LogResp {
	search := e.esClient.Search(config.CF.EsIndex).From(req.Page.From).Size(req.Page.Size).TrackScores(true)
	if !config.CF.EsWindowLimit {
		search = search.TrackTotalHits(true)
	}
	if req.Sort == "asc" {
		search.Sort("id", true)
	}
	if req.Sort == "desc" {
		search.Sort("id", false)
	}

	queryList := []elastic.Query{}
	timeRangeQuery := elastic.NewRangeQuery("time")
	if req.TimeRange.StartTime != 0 {
		queryList = append(queryList, timeRangeQuery.Gte(req.TimeRange.StartTime))
	}
	if req.TimeRange.EndTime != 0 {
		queryList = append(queryList, timeRangeQuery.Lt(req.TimeRange.EndTime))
	}
	notQuery := []elastic.Query{}

	// analyze query string
	for _, v := range sr.QueryStringAnalysis(req.Match.Log) {
		switch v.Cond {
		case sr.Match:
			queryList = append(queryList, elastic.NewMatchQuery("log", v.Content).Boost(2))
			queryList = append(queryList, elastic.NewMatchPhraseQuery("log", v.Content))
		case sr.NotMatch:
			notQuery = append(notQuery, elastic.NewMatchPhraseQuery("log", v.Content))
		case sr.WildCard:
			queryList = append(queryList, elastic.NewWildcardQuery("log.keyword", "*"+v.Content+"*"))
		case sr.NotWildCard:
			notQuery = append(notQuery, elastic.NewWildcardQuery("log.keyword", "*"+v.Content+"*"))
		}
	}

	if req.Match.Name != "" {
		queryList = append(queryList, elastic.NewMatchQuery("name", req.Match.Name))
	}
	if req.Match.Using != "" {
		queryList = append(queryList, elastic.NewMatchQuery("using", req.Match.Using))
	}

	if req.CursorID != 0 {
		idRange := elastic.NewRangeQuery("id")
		if req.Sort == "desc" {
			idRange = idRange.Lt(req.CursorID)
		} else {
			idRange = idRange.Gt(req.CursorID)
		}
		queryList = append(queryList, idRange)
	}

	if len(req.FilterName) != 0 { // filter process name
		shouldQueryList := []elastic.Query{}
		for _, fpn := range req.FilterName {
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
		log.Logger.Warnw("es search failed", "err", err)
		return result
	}
	// iterate response hits
	for _, v := range resp.Hits.Hits {
		if v.Source != nil {
			var data model.ProcessLog
			if err := json.Unmarshal(v.Source, &data); err == nil {
				if len(v.Highlight) > 0 && len(v.Highlight["log"]) > 0 {
					data.Log = v.Highlight["log"][0]
				}
				result.Data = append(result.Data, &data)
			} else {
				log.Logger.Warnw("json decode failed", "err", err)
			}
		}
	}

	result.Total = resp.TotalHits()
	return result
}
