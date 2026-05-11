package logic

import (
	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/search"

	_ "github.com/lzh-1625/go_process_manager/internal/app/repository/search/bleve"
	_ "github.com/lzh-1625/go_process_manager/internal/app/repository/search/es"
	_ "github.com/lzh-1625/go_process_manager/internal/app/repository/search/sqlite"
)

// var LogLogicImpl search.LogLogic

// func InitLog() {
// 	LogLogicImpl = search.GetSearchImpl(config.CF.StorgeType)
// 	LogLogicImpl.Init()
// }

func NewLogLogic() search.LogLogic {
	LogLogicImpl := search.GetSearchImpl(config.CF.StorgeType)
	LogLogicImpl.Init()
	return LogLogicImpl
}
