package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/utils"
)

type procApi struct{}

var ProcApi = new(procApi)

func (p *procApi) CreateProcess(ctx *echo.Context) error {
	var req model.Process
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	proc := logic.ProcessCtlLogic.NewProcess(req)
	if proc == nil {
		return errors.New("create process failed")
	}
	return ctx.JSON(http.StatusOK, model.Response[map[string]any]{
		Data: map[string]any{
			"id": proc.UUID,
		},
		Message: "success",
		Code:    0,
	})
}

func (p *procApi) DeleteProcess(ctx *echo.Context) error {
	var req struct {
		UUID int `query:"uuid"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	if err := logic.ProcessCtlLogic.DeleteProcess(req.UUID); err != nil {
		return err
	}
	return nil
}

func (p *procApi) KillProcess(ctx *echo.Context) error {
	var req struct {
		UUID int `query:"uuid"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	if !hasOprPermission(ctx, req.UUID, eum.OperationStop) {
		return errors.New("not permission")
	}
	proc, err := logic.ProcessCtlLogic.GetProcess(req.UUID)
	if err != nil {
		return err
	}
	proc.SetOpertor(getUserName(ctx))
	return proc.Kill()
}

func (p *procApi) StartProcess(ctx *echo.Context) error {
	var req struct {
		UUID int `json:"uuid"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	if !hasOprPermission(ctx, req.UUID, eum.OperationStart) {
		return errors.New("not permission")
	}
	prod, err := logic.ProcessCtlLogic.GetProcess(req.UUID)
	if err != nil {
		proConfig, err := logic.ProcessCtlLogic.GetProcessConfigByID(req.UUID)
		if err != nil {
			return err
		}
		proc, err := logic.ProcessCtlLogic.RunProcess(*proConfig)
		if err != nil {
			return err
		}
		proc.SetOpertor(getUserName(ctx))
		return nil
	}
	if prod.State.State == eum.ProcessStateStart || prod.State.State == eum.ProcessStateRunning {
		return errors.New("process is currently running")
	}
	prod.ResetRestartTimes()
	prod.SetOpertor(getUserName(ctx))
	return prod.Start()
}

func (p *procApi) StartAllProcess(ctx *echo.Context) error {
	if isAdmin(ctx) {
		logic.ProcessCtlLogic.ProcessStartAll()
	} else {
		logic.ProcessCtlLogic.ProcesStartAllByUsername(getUserName(ctx))
	}
	return nil
}

func (p *procApi) KillAllProcess(ctx *echo.Context) error {
	if isAdmin(ctx) {
		logic.ProcessCtlLogic.KillAllProcess()
	} else {
		logic.ProcessCtlLogic.KillAllProcessByUserName(getUserName(ctx))
	}
	return nil
}

func (p *procApi) GetProcessList(ctx *echo.Context) error {
	if isAdmin(ctx) {
		return ctx.JSON(http.StatusOK, model.Response[[]model.ProcessInfo]{
			Data:    logic.ProcessCtlLogic.GetProcessList(),
			Message: "success",
			Code:    0,
		})
	} else {
		return ctx.JSON(http.StatusOK, model.Response[[]model.ProcessInfo]{
			Data:    logic.ProcessCtlLogic.GetProcessListByUser(getUserName(ctx)),
			Message: "success",
			Code:    0,
		})
	}
}

func (p *procApi) UpdateProcessConfig(ctx *echo.Context) error {
	var req model.Process
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return logic.ProcessCtlLogic.UpdateProcessConfig(req)
}

func (p *procApi) GetProcessConfig(ctx *echo.Context) error {
	var req struct {
		UUID int `query:"uuid"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	data, err := logic.ProcessCtlLogic.GetProcessConfigByID(req.UUID)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, model.Response[*model.Process]{
		Data:    data,
		Message: "success",
		Code:    0,
	})
}

func (p *procApi) ProcessControl(ctx *echo.Context) error {
	var req struct {
		UUID int `query:"uuid"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	user := getUserName(ctx)
	proc, err := logic.ProcessCtlLogic.GetProcess(req.UUID)
	if err != nil {
		return err
	}
	proc.ProcessControl(user)
	return nil
}

func (p *procApi) ProcessCreateShare(ctx *echo.Context) error {
	var req model.ProcessShare
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	token := utils.UnwarpIgnore(uuid.NewRandom()).String()
	if err := logic.WsShareLogic.AddShareData(model.WsShare{
		ExpireTime: time.Now().Add(time.Minute * time.Duration(req.Minutes)),
		Write:      req.Write,
		Token:      token,
		Pid:        req.Pid,
		CreateBy:   getUserName(ctx),
	}); err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, model.Response[map[string]any]{
		Data: map[string]any{
			"token": token,
		},
		Message: "success",
		Code:    0,
	})
}
