package logic

import (
	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/search"

	"github.com/lzh-1625/go_process_manager/internal/app/repository/search/bleve"
	_ "github.com/lzh-1625/go_process_manager/internal/app/repository/search/bleve"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/search/es"
	_ "github.com/lzh-1625/go_process_manager/internal/app/repository/search/es"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/search/sqlite"
)

func NewILogLogic(logRepository *repository.LogRepository) search.ILogLogic {
	switch config.CF.StorgeType {
	case "es":
		return es.NewEsSearch()
	case "bleve":
		return bleve.NewBleveSearch()
	default:
		return sqlite.NewSqliteSearch(logRepository)
	}
}
