//go:build !slim

package bleve

import (
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
	_ "github.com/blevesearch/bleve/v2/search/highlight/highlighter/ansi"
	"github.com/google/uuid"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	sr "github.com/lzh-1625/go_process_manager/internal/app/repository/search"
	logger "github.com/lzh-1625/go_process_manager/log"
	gse "github.com/vcaesar/gse-bleve"
	// gse "github.com/lzh-1625/gse-bleve"
)

func init() {
	sr.Register("bleve", new(bleveSearch))
}

type bleveSearch struct {
	index bleve.Index
}

func (b *bleveSearch) Init() error {
	opt := gse.Option{
		Dicts: "embed, zh_s",
		Stop:  "",
		Opt:   "search-hmm",
		Trim:  "trim",
	}
	indexMapping, err := gse.NewMapping(opt)
	if err != nil {
		logger.Logger.Errorw("bleve init fail", "err", err)
		return err
	}
	mapping := bleve.NewDocumentMapping()
	log := bleve.NewTextFieldMapping()
	log.Index = true
	time := bleve.NewNumericFieldMapping()
	time.Index = true
	name := bleve.NewKeywordFieldMapping()
	name.Index = true
	using := bleve.NewKeywordFieldMapping()
	using.Index = true
	mapping.AddFieldMappingsAt("log", log)
	mapping.AddFieldMappingsAt("time", time)
	mapping.AddFieldMappingsAt("name", name)
	mapping.AddFieldMappingsAt("using", using)
	indexMapping.AddDocumentMapping("server_log_v1", mapping)
	index, err := bleve.Open("log.bleve")
	if err != nil {
		index, err = bleve.New("log.bleve", indexMapping)
		if err != nil {
			logger.Logger.Errorw("bleve init error", "err", err)
			return err
		}
	}
	b.index = index
	return nil
}

func (b *bleveSearch) Insert(logContent string, processName string, using string, ts int64) {
	if err := b.index.Index(uuid.NewString(), model.ProcessLog{
		Log:   logContent,
		Name:  processName,
		Using: using,
		Time:  ts,
	}); err != nil {
		logger.Logger.Warnw("bleve log insert failed", "err", err)
	}
}

func (b *bleveSearch) Search(req model.GetLogReq, filterProcessName ...string) (result model.LogResp) {
	buildQuery := bleve.NewBooleanQuery()
	for _, v := range sr.QueryStringAnalysis(req.Match.Log) {
		switch v.Cond {
		case sr.Match, sr.WildCard:
			logQuery := bleve.NewMatchQuery(v.Content)
			logQuery.SetField("log")
			buildQuery.AddMust(logQuery)
		case sr.NotMatch, sr.NotWildCard:
			logQuery := bleve.NewMatchQuery(v.Content)
			logQuery.SetField("log")
			buildQuery.AddMustNot(logQuery)
			// case sr.WildCard:
			// 	logQuery := bleve.NewWildcardQuery("*" + v.Content + "*")
			// 	logQuery.SetField("log")
			// 	buildQuery.AddMust(logQuery)
			// case sr.NotWildCard:
			// 	logQuery := bleve.NewWildcardQuery("*" + v.Content + "*")
			// 	logQuery.SetField("log")
			// 	buildQuery.AddMustNot(logQuery)
		}
	}
	if req.Match.Name != "" {
		nameQuery := bleve.NewTermQuery(req.Match.Name)
		nameQuery.SetField("name")
		buildQuery.AddMust(nameQuery)
	}
	if req.Match.Using != "" {
		usingQuery := bleve.NewWildcardQuery("*" + req.Match.Using + "*")
		usingQuery.SetField("using")
		buildQuery.AddMust(usingQuery)
	}
	if req.TimeRange.EndTime != 0 || req.TimeRange.StartTime != 0 {
		st := float64(req.TimeRange.StartTime) - 1
		et := float64(req.TimeRange.EndTime)
		timeQuery := bleve.NewNumericRangeQuery(&st, &et)
		buildQuery.AddMust(timeQuery)
	} else {
		st := float64(0)
		et := float64(time.Now().UnixMilli())
		timeQuery := bleve.NewNumericRangeQuery(&st, &et)
		buildQuery.AddMust(timeQuery)
	}
	if len(filterProcessName) != 0 {
		for _, v := range filterProcessName {
			filterQuery := bleve.NewTermQuery(v)
			filterQuery.SetField("name")
			buildQuery.AddShould(filterQuery)
		}
	}
	sortArgs := ([]string{"-_score"})
	if req.Sort == "desc" {

		sortArgs = ([]string{"-time"})
	}
	if req.Sort == "asc" {
		sortArgs = ([]string{"time"})
	}
	hl := bleve.HighlightRequest{}
	hl.AddField("log")
	style := "ansi"
	res, err := b.index.Search(&bleve.SearchRequest{
		Query:  buildQuery,
		Fields: []string{"log", "name", "using", "time"},
		From:   req.Page.From,
		Size:   req.Page.Size,
		Sort:   search.ParseSortOrderStrings(sortArgs),
		Highlight: &bleve.HighlightRequest{
			Style:  &style,
			Fields: []string{"log"},
		},
	})
	if err != nil {
		logger.Logger.Warnw("bleve search failed", "err", err)
		return
	}
	data := []*model.ProcessLog{}
	for _, v := range res.Hits {
		var log string
		if _, ok := v.Fragments["log"]; ok && len(v.Fragments["log"]) > 0 {
			log = v.Fragments["log"][0]
		} else {
			log = v.Fields["log"].(string)
		}
		data = append(data, &model.ProcessLog{
			Log:   log,
			Time:  int64(v.Fields["time"].(float64)),
			Using: v.Fields["using"].(string),
			Name:  v.Fields["name"].(string),
		})
	}

	result.Data = data
	result.Total = int64(res.Total)
	return
}
