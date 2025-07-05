package api

import (
	"errors"
	"slices"

	"github.com/lzh-1625/go_process_manager/internal/app/constants"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"

	"github.com/gin-gonic/gin"
)

type logApi struct{}

var LogApi = new(logApi)

func (a *logApi) GetLog(ctx *gin.Context, req model.GetLogReq) (err error) {
	if isAdmin(ctx) {
		rOk(ctx, "Query successful!", logic.LogLogicImpl.Search(req, req.FilterName...))
	} else {
		processNameList := repository.PermissionRepository.GetProcessNameByPermission(getUserName(ctx), constants.OPERATION_LOG)
		filterName := slices.DeleteFunc(req.FilterName, func(s string) bool {
			return !slices.Contains(processNameList, s)
		})
		if len(filterName) == 0 {
			filterName = processNameList
		}
		if len(filterName) == 0 {
			return errors.New("no information found")
		}
		rOk(ctx, "Query successful!", logic.LogLogicImpl.Search(req, filterName...))
	}
	return
}

func (a *logApi) GetRunningLog(ctx *gin.Context, _ any) error {
	rOk(ctx, "Query successful!", logic.Loghandler.GetRunning())
	return nil
}
