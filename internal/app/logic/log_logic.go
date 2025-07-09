package logic

import (
	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/search"

	_ "github.com/lzh-1625/go_process_manager/internal/app/search/bleve"
	_ "github.com/lzh-1625/go_process_manager/internal/app/search/es"
	_ "github.com/lzh-1625/go_process_manager/internal/app/search/sqlite"
)

var LogLogicImpl search.LogLogic

func InitLog() {
	LogLogicImpl = search.GetSearchImpl(config.CF.StorgeType)
	LogLogicImpl.Init()
}
