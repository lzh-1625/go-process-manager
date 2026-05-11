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

type ProcApi struct {
	processCtlLogic *logic.ProcessCtlLogic
	permissionLogic *logic.PermissionLogic
	wsShareLogic    *logic.WsShareLogic
}

func NewProcApi(processCtlLogic *logic.ProcessCtlLogic, permissionLogic *logic.PermissionLogic, wsShareLogic *logic.WsShareLogic) *ProcApi {
	return &ProcApi{
		processCtlLogic: processCtlLogic,
		permissionLogic: permissionLogic,
		wsShareLogic:    wsShareLogic,
	}
}

func (p *ProcApi) CreateProcess(ctx *echo.Context) error {
	var req model.Process
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	proc := p.processCtlLogic.NewProcess(req)
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

func (p *ProcApi) DeleteProcess(ctx *echo.Context) error {
	var req struct {
		UUID int `query:"uuid"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	if err := p.processCtlLogic.DeleteProcess(req.UUID); err != nil {
		return err
	}
	return nil
}

func (p *ProcApi) KillProcess(ctx *echo.Context) error {
	var req struct {
		UUID int `query:"uuid"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	if !p.permissionLogic.GetPermission(getUserName(ctx), req.UUID).Stop == false {
		return errors.New("not permission")
	}
	proc, err := p.processCtlLogic.GetProcess(req.UUID)
	if err != nil {
		return err
	}
	proc.SetOpertor(getUserName(ctx))
	return proc.Kill()
}

func (p *ProcApi) StartProcess(ctx *echo.Context) error {
	var req struct {
		UUID int `json:"uuid"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	if !p.permissionLogic.GetPermission(getUserName(ctx), req.UUID).Start == false {
		return errors.New("not permission")
	}
	prod, err := p.processCtlLogic.GetProcess(req.UUID)
	if err != nil {
		proConfig, err := p.processCtlLogic.GetProcessConfigByID(req.UUID)
		if err != nil {
			return err
		}
		proc, err := p.processCtlLogic.RunProcess(*proConfig)
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

func (p *ProcApi) StartAllProcess(ctx *echo.Context) error {
	if isAdmin(ctx) {
		p.processCtlLogic.ProcessStartAll()
	} else {
		p.processCtlLogic.ProcesStartAllByUsername(getUserName(ctx))
	}
	return nil
}

func (p *ProcApi) KillAllProcess(ctx *echo.Context) error {
	if isAdmin(ctx) {
		p.processCtlLogic.KillAllProcess()
	} else {
		p.processCtlLogic.KillAllProcessByUserName(getUserName(ctx))
	}
	return nil
}

func (p *ProcApi) GetProcessList(ctx *echo.Context) error {
	if isAdmin(ctx) {
		return ctx.JSON(http.StatusOK, model.Response[[]model.ProcessInfo]{
			Data:    p.processCtlLogic.GetProcessList(),
			Message: "success",
			Code:    0,
		})
	} else {
		return ctx.JSON(http.StatusOK, model.Response[[]model.ProcessInfo]{
			Data:    p.processCtlLogic.GetProcessListByUser(getUserName(ctx)),
			Message: "success",
			Code:    0,
		})
	}
}

func (p *ProcApi) UpdateProcessConfig(ctx *echo.Context) error {
	var req model.Process
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return p.processCtlLogic.UpdateProcessConfig(req)
}

func (p *ProcApi) GetProcessConfig(ctx *echo.Context) error {
	var req struct {
		UUID int `query:"uuid"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	data, err := p.processCtlLogic.GetProcessConfigByID(req.UUID)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, model.Response[*model.Process]{
		Data:    data,
		Message: "success",
		Code:    0,
	})
}

func (p *ProcApi) ProcessControl(ctx *echo.Context) error {
	var req struct {
		UUID int `query:"uuid"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	user := getUserName(ctx)
	proc, err := p.processCtlLogic.GetProcess(req.UUID)
	if err != nil {
		return err
	}
	proc.ProcessControl(user)
	return nil
}

func (p *ProcApi) ProcessCreateShare(ctx *echo.Context) error {
	var req model.ProcessShare
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	token := utils.UnwarpIgnore(uuid.NewRandom()).String()
	if err := p.wsShareLogic.AddShareData(model.WsShare{
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
