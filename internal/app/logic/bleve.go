package logic

import (
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
	_ "github.com/blevesearch/bleve/v2/search/highlight/highlighter/ansi"
	"github.com/google/uuid"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	logger "github.com/lzh-1625/go_process_manager/log"
	gse "github.com/vcaesar/gse-bleve"
)

type bleveLogic struct {
	index bleve.Index
}

var BleveLogic = new(bleveLogic)

func (b *bleveLogic) InitBleve() {
	opt := gse.Option{
		Dicts: "embed, zh",
		Stop:  "",
		Opt:   "search-hmm",
		Trim:  "trim",
	}
	indexMapping, err := gse.NewMapping(opt)
	if err != nil {
		logger.Logger.Errorw("bleve init fail", "err", err)
		return
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
			return
		}
	}
	b.index = index
}

func (b *bleveLogic) Insert(logContent string, processName string, using string, ts int64) {
	if err := b.index.Index(uuid.NewString(), model.ProcessLog{
		Log:   logContent,
		Name:  processName,
		Using: using,
		Time:  ts,
	}); err != nil {
		logger.Logger.Warnw("bleve log insert failed", "err", err)
	}
}

func (b *bleveLogic) Search(req model.GetLogReq, filterProcessName ...string) (result model.LogResp) {
	buildQuery := bleve.NewBooleanQuery()
	if req.Match.Log != "" {
		logQuery := bleve.NewMatchQuery(req.Match.Log)
		logQuery.SetField("log")
		buildQuery.AddMust(logQuery)
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
		st := float64(req.TimeRange.StartTime)
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
		sortArgs = append(sortArgs, "-time")
	}
	if req.Sort == "asc" {
		sortArgs = append(sortArgs, "time")
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
		data = append(data, &model.ProcessLog{
			Log:   v.Fragments["log"][0],
			Time:  int64(v.Fields["time"].(float64)),
			Using: v.Fields["using"].(string),
			Name:  v.Fields["name"].(string),
		})
	}

	result.Data = data
	result.Total = int64(res.Total)
	return
}
