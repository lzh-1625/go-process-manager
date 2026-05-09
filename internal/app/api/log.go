package api

import (
	"errors"
	"net/http"
	"slices"

	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
)

type logApi struct{}

var LogApi = new(logApi)

func (a *logApi) GetLog(ctx *echo.Context) error {
	var req model.GetLogReq
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	if isAdmin(ctx) {
		return ctx.JSON(http.StatusOK, model.Response[model.LogResp]{
			Data:    logic.LogLogicImpl.Search(req, req.FilterName...),
			Message: "success",
			Code:    0,
		})
	} else {
		processNameList := logic.PermissionLogic.GetProcessNameByPermission(getUserName(ctx), eum.OperationLog)
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
			Data:    logic.LogLogicImpl.Search(req, filterName...),
			Message: "success",
			Code:    0,
		})
	}
}

func (a *logApi) GetRunningLog(ctx *echo.Context) error {
	return ctx.JSON(http.StatusOK, logic.Loghandler.GetRunning())
}
