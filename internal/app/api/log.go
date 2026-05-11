package api

import (
	"errors"
	"net/http"
	"slices"

	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/search"
)

type LogApi struct {
	permissionLogic *logic.PermissionLogic
	logLogic        search.LogLogic
	logHandler      *logic.LogHandler
}

func NewLogApi(permissionLogic *logic.PermissionLogic, logLogic search.LogLogic) *LogApi {
	return &LogApi{
		permissionLogic: permissionLogic,
		logLogic:        logLogic,
	}
}

func (a *LogApi) GetLog(ctx *echo.Context) error {
	var req model.GetLogReq
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	if isAdmin(ctx) {
		return ctx.JSON(http.StatusOK, model.Response[model.LogResp]{
			Data:    a.logLogic.Search(req, req.FilterName...),
			Message: "success",
			Code:    0,
		})
	} else {
		processNameList := a.permissionLogic.GetProcessNameByPermission(getUserName(ctx), eum.OperationLog)
		filterName := slices.DeleteFunc(req.FilterName, func(s string) bool {
			return !slices.Contains(processNameList, s)
		})
		if len(filterName) == 0 {
			filterName = processNameList
		}
		if len(filterName) == 0 {
			return errors.New("no information found")
		}
		return ctx.JSON(http.StatusOK, model.Response[model.LogResp]{
			Data:    a.logLogic.Search(req, filterName...),
			Message: "success",
			Code:    0,
		})
	}
}

func (a *LogApi) GetRunningLog(ctx *echo.Context) error {
	return ctx.JSON(http.StatusOK, a.logHandler.GetRunning())
}
