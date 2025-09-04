package search

import (
	"strings"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/log"
)

type LogLogic interface {
	Search(req model.GetLogReq, filterProcessName ...string) model.LogResp
	Insert(log string, processName string, using string, ts int64)
	Init() error
}

var searchImplMap map[string]LogLogic = map[string]LogLogic{}

func Register(name string, impl LogLogic) {
	searchImplMap[name] = impl
}

func GetSearchImpl(name string) LogLogic {
	v, ok := searchImplMap[name]
	if ok {
		return v
	}
	log.Logger.Warnw("未找到对应的存储引擎,使用默认[sqlite]", "name", name)
	return searchImplMap["sqlite"]
}

type Cond int

const (
	Match       Cond = iota // ^
	WildCard                // ~
	NotMatch                // !^
	NotWildCard             // !~
)

type Query struct {
	Cond    Cond
	Content string
}

func QueryStringAnalysis(s string) (query []Query) {
	if strings.TrimSpace(s) == "" {
		return
	}
	strList := strings.Fields(s)
	for _, v := range strList {
		switch {
		case strings.HasPrefix(v, "!^"):
			query = append(query, Query{NotMatch, v[2:]})
		case strings.HasPrefix(v, "!~"):
			query = append(query, Query{NotWildCard, v[2:]})
		case strings.HasPrefix(v, "!"):
			query = append(query, Query{NotMatch, v[1:]})
		case strings.HasPrefix(v, "^"):
			query = append(query, Query{Match, v[1:]})
		case strings.HasPrefix(v, "~"):
			query = append(query, Query{WildCard, v[1:]})
		default:
			query = append(query, Query{Match, v})
		}
	}
	return
}
