package logic

import (
	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/search"
	"github.com/lzh-1625/go_process_manager/log"

	"github.com/lzh-1625/go_process_manager/internal/app/repository/search/bleve"
	_ "github.com/lzh-1625/go_process_manager/internal/app/repository/search/bleve"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/search/es"
	_ "github.com/lzh-1625/go_process_manager/internal/app/repository/search/es"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/search/sqlite"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/search/victorialogs"
)

func NewILogLogic(logRepository *repository.LogRepository) search.ILogLogic {
	var impl search.ILogLogic
	var err error
	switch config.CF.StorgeType {
	case "es":
		impl, err = es.NewEsSearch()
	case "bleve":
		impl, err = bleve.NewBleveSearch()
	case "victorialogs":
		impl = victorialogs.NewVictoriaLogsSearch()
	default:
		impl = sqlite.NewSqliteSearch(logRepository)
	}
	if err != nil || impl == nil {
		log.Logger.Errorw("init log engine failed", "type", config.CF.StorgeType, "err", err)
		return &emptyLogLogic{err: err}
	}
	return impl
}

var _ search.ILogLogic = (*emptyLogLogic)(nil)

type emptyLogLogic struct {
	err error
}

func (e *emptyLogLogic) Search(req model.GetLogReq) (*model.LogResp, error) { return nil, e.err }
func (e *emptyLogLogic) Insert(...model.ProcessLog)                         {}
