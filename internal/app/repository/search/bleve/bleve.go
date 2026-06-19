//go:build !slim

package bleve

import (
	"path"
	"strconv"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
	_ "github.com/blevesearch/bleve/v2/search/highlight/highlighter/ansi"
	"github.com/blevesearch/bleve/v2/search/query"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	sr "github.com/lzh-1625/go_process_manager/internal/app/repository/search"
	logger "github.com/lzh-1625/go_process_manager/log"
	"github.com/lzh-1625/go_process_manager/utils"
	gse "github.com/vcaesar/gse-bleve"
)

func NewBleveSearch() *bleveSearch {
	b := &bleveSearch{}
	b.init()
	return b
}

type bleveSearch struct {
	index bleve.Index
}

func (b *bleveSearch) Reload() error {
	return b.init()
}

func (b *bleveSearch) init() error {
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
	id := bleve.NewNumericFieldMapping()
	id.Index = true
	logKeyword := bleve.NewKeywordFieldMapping()
	logKeyword.Index = true
	time := bleve.NewNumericFieldMapping()
	time.Index = true
	name := bleve.NewKeywordFieldMapping()
	name.Index = true
	using := bleve.NewKeywordFieldMapping()
	using.Index = true
	mapping.AddFieldMappingsAt("log", log)
	mapping.AddFieldMappingsAt("id", id)
	mapping.AddFieldMappingsAt("log_keyword", logKeyword)
	mapping.AddFieldMappingsAt("time", time)
	mapping.AddFieldMappingsAt("name", name)
	mapping.AddFieldMappingsAt("using", using)
	indexMapping.AddDocumentMapping("server_log_v1", mapping)
	path := path.Join(config.CF.ConfigDir, "log.bleve")
	index, err := bleve.Open(path)
	if err != nil {
		index, err = bleve.New(path, indexMapping)
		if err != nil {
			logger.Logger.Errorw("bleve init error", "err", err)
			return err
		}
	}
	b.index = index
	return nil
}

func (b *bleveSearch) Insert(logs ...model.ProcessLog) {
	batch := b.index.NewBatch()
	for _, v := range logs {
		batch.Index(strconv.FormatInt(v.ID, 10), v)
	}
	if err := b.index.Batch(batch); err != nil {
		logger.Logger.Warnw("bleve log insert failed", "err", err)
	}
}

func (b *bleveSearch) Search(req model.GetLogReq) (result model.LogResp) {
	buildQuery := bleve.NewBooleanQuery()

	logQuery := sr.QueryStringAnalysis(req.Match.Log)
	for _, v := range logQuery {
		switch v.Cond {
		case sr.Match:
			logQuery := bleve.NewMatchQuery(v.Content)
			logQuery.SetField("log")
			buildQuery.AddMust(logQuery)
		case sr.NotMatch:
			logQuery := bleve.NewMatchQuery(v.Content)
			logQuery.SetField("log")
			buildQuery.AddMustNot(logQuery)
		case sr.WildCard:
			logQuery := bleve.NewWildcardQuery("*" + v.Content + "*")
			logQuery.SetField("log_keyword")
			buildQuery.AddMust(logQuery)
		case sr.NotWildCard:
			logQuery := bleve.NewWildcardQuery("*" + v.Content + "*")
			logQuery.SetField("log_keyword")
			buildQuery.AddMustNot(logQuery)
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

	if req.CursorID != 0 {
		var query *query.NumericRangeQuery
		if req.Sort == "desc" {
			query = bleve.NewNumericRangeQuery(nil, new(float64(req.CursorID)))
		} else {
			query = bleve.NewNumericRangeQuery(new(float64(req.CursorID+1)), nil)
		}
		query.SetField("id")
		buildQuery.AddMust(query)
	}

	st := new(0.1)
	ed := new(float64(time.Now().UnixMilli()))

	if req.TimeRange.StartTime != 0 {
		st = new(float64(req.TimeRange.StartTime))
	}
	if req.TimeRange.EndTime != 0 {
		ed = new(float64(req.TimeRange.EndTime))
	}
	timeQuery := bleve.NewNumericRangeQuery(st, ed)

	// at least one of the time range must be specified
	buildQuery.AddMust(timeQuery)

	if len(req.FilterName) != 0 {
		shouldQueryList := []query.Query{}
		for _, v := range req.FilterName {
			shouldQueryList = append(shouldQueryList, bleve.NewTermQuery(v))
		}
		if len(shouldQueryList) > 0 {
			shouldQuery := bleve.NewBooleanQuery()
			shouldQuery.AddShould(shouldQueryList...)
			buildQuery.AddMust(shouldQuery)
		}
	}
	sortArgs := ([]string{"-_score"})
	if req.Sort == "desc" {

		sortArgs = ([]string{"-time", "-id"})
	}
	if req.Sort == "asc" {
		sortArgs = ([]string{"time", "id"})
	}
	var hl *bleve.HighlightRequest
	if req.Match.HighLight {
		hl = &bleve.HighlightRequest{
			Style:  new("ansi"),
			Fields: []string{"log"},
		}
	}

	res, err := b.index.Search(&bleve.SearchRequest{
		Query:     buildQuery,
		Fields:    []string{"log", "name", "using", "time", "id"},
		From:      req.Page.From,
		Size:      req.Page.Size,
		Sort:      search.ParseSortOrderStrings(sortArgs),
		Highlight: hl,
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
			ID:    utils.UnwarpIgnore(strconv.ParseInt(v.ID, 10, 64)),
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
