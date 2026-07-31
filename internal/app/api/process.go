package api

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/process"
	"github.com/lzh-1625/go_process_manager/internal/app/types"
	"github.com/lzh-1625/go_process_manager/utils"
)

type ProcApi struct {
	processCtlLogic *logic.ProcessCtlLogic
	wsShareLogic    *logic.WsShareLogic
	permissionApi   *PermissionApi
}

func NewProcApi(
	processCtlLogic *logic.ProcessCtlLogic,
	wsShareLogic *logic.WsShareLogic,
	permissionApi *PermissionApi) *ProcApi {
	return &ProcApi{
		processCtlLogic: processCtlLogic,
		wsShareLogic:    wsShareLogic,
		permissionApi:   permissionApi,
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
			"uuid": proc.UUID,
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
		UUID    int  `query:"uuid"`
		SIGKILL bool `query:"sigkill"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	if !p.permissionApi.hasOprPermission(ctx, req.UUID, types.OperationStop) {
		return errors.New("not permission")
	}
	proc, err := p.processCtlLogic.GetProcess(req.UUID)
	if err != nil {
		return err
	}
	return proc.Operate(getUserName(ctx), func() error {
		if req.SIGKILL {
			return proc.Kill9()
		}
		return proc.Kill()
	})

}

func (p *ProcApi) StartProcess(ctx *echo.Context) error {
	var req struct {
		UUID int `json:"uuid"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	if !p.permissionApi.hasOprPermission(ctx, req.UUID, types.OperationStart) {
		return errors.New("not permission")
	}
	proc, err := p.processCtlLogic.GetProcess(req.UUID)
	if err != nil {
		return err
	}
	if proc.State.State == types.ProcessStateStarting || proc.State.State == types.ProcessStateRunning {
		return errors.New("process is currently running")
	}
	proc.ResetRestartTimes() // Reset the current restart count when the process is started manually.
	return proc.Operate(getUserName(ctx), func() error {
		return proc.Start()
	})
}

func (p *ProcApi) StartAllProcess(ctx *echo.Context) error {
	user := getUserName(ctx)
	if isAdmin(ctx) {
		p.processCtlLogic.ForEach(func(proc *process.Process) {
			proc.Operate(user, func() error {
				return proc.Start()
			})
		})
	} else {
		p.processCtlLogic.ForEachByOwner(user, func(proc *process.Process) {
			proc.Operate(user, func() error {
				return proc.Start()
			})
		})
	}
	return nil
}

func (p *ProcApi) KillAllProcess(ctx *echo.Context) error {
	user := getUserName(ctx)
	wg := sync.WaitGroup{}
	if isAdmin(ctx) {
		p.processCtlLogic.ForEach(func(proc *process.Process) {
			wg.Go(func() {
				proc.Operate(user, func() error {
					return proc.Kill()
				})
			})
		})
	} else {
		p.processCtlLogic.ForEachByOwner(user, func(proc *process.Process) {
			wg.Go(func() {
				proc.Operate(user, func() error {
					return proc.Kill()
				})
			})
		})
	}
	wg.Wait()
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
		UUID int    `query:"uuid"`
		Name string `query:"name"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	var data *model.Process
	if req.UUID != 0 {
		pc, err := p.processCtlLogic.GetProcessConfigByID(req.UUID)
		if err != nil {
			return err
		}
		data = pc
	} else {
		pc, err := p.processCtlLogic.GetProcessConfigByName(req.Name)
		if err != nil {
			return err
		}
		data = pc
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
