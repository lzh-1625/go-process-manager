package search

import (
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
