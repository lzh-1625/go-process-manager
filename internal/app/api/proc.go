package api

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/utils"

	"github.com/gin-gonic/gin"
)

type procApi struct{}

var ProcApi = new(procApi)

func (p *procApi) CreateNewProcess(ctx *gin.Context, req model.Process) any {
	index, err := repository.ProcessRepository.AddProcessConfig(req)
	if err != nil {
		return err
	}
	req.UUID = index
	proc, err := logic.ProcessCtlLogic.NewProcess(req)
	if err != nil {
		return err
	}
	logic.ProcessCtlLogic.AddProcess(req.UUID, proc)
	return gin.H{
		"id": req.UUID,
	}
}

func (p *procApi) DeleteNewProcess(ctx *gin.Context, req struct {
	UUID int `form:"uuid" binding:"required"`
}) (err error) {
	logic.ProcessCtlLogic.KillProcess(req.UUID)
	logic.ProcessCtlLogic.DeleteProcess(req.UUID)
	return repository.ProcessRepository.DeleteProcessConfig(req.UUID)
}

func (p *procApi) KillProcess(ctx *gin.Context, req struct {
	UUID int `form:"uuid" binding:"required"`
}) (err error) {
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

func (p *procApi) StartProcess(ctx *gin.Context, req struct {
	UUID int `json:"uuid" binding:"required"`
}) (err error) {
	if !hasOprPermission(ctx, req.UUID, eum.OperationStart) {
		return errors.New("not permission")
	}
	prod, err := logic.ProcessCtlLogic.GetProcess(req.UUID)
	if err != nil { // 进程不存在则创建
		proConfig, err := repository.ProcessRepository.GetProcessConfigById(req.UUID)
		if err != nil {
			return err
		}
		proc, err := logic.ProcessCtlLogic.RunNewProcess(*proConfig)
		if err != nil {
			return err
		}
		proc.SetOpertor(getUserName(ctx))
		logic.ProcessCtlLogic.AddProcess(req.UUID, proc)
		return nil
	}
	if prod.State.State == eum.ProcessStateStart || prod.State.State == eum.ProcessStateRunning {
		return errors.New("process is currently running")
	}
	prod.ResetRestartTimes()
	prod.SetOpertor(getUserName(ctx))
	err = prod.Start()
	return
}

func (p *procApi) StartAllProcess(ctx *gin.Context, _ any) (err error) {
	if isAdmin(ctx) {
		logic.ProcessCtlLogic.ProcessStartAll()
	} else {
		logic.ProcessCtlLogic.ProcesStartAllByUsername(getUserName(ctx))
	}
	return
}

func (p *procApi) KillAllProcess(ctx *gin.Context, _ any) (err error) {
	if isAdmin(ctx) {
		logic.ProcessCtlLogic.KillAllProcess()
	} else {
		logic.ProcessCtlLogic.KillAllProcessByUserName(getUserName(ctx))
	}
	return
}

func (p *procApi) GetProcessList(ctx *gin.Context, _ any) any {
	if isAdmin(ctx) {
		return logic.ProcessCtlLogic.GetProcessList()
	} else {
		return logic.ProcessCtlLogic.GetProcessListByUser(getUserName(ctx))
	}
}

func (p *procApi) UpdateProcessConfig(ctx *gin.Context, req model.Process) (err error) {
	logic.ProcessCtlLogic.UpdateProcessConfig(req)
	err = repository.ProcessRepository.UpdateProcessConfig(req)
	return
}

func (p *procApi) GetProcessConfig(ctx *gin.Context, req struct {
	UUID int `form:"uuid" binding:"required"`
}) any {
	data, err := repository.ProcessRepository.GetProcessConfigById(req.UUID)
	if err != nil {
		return err
	}
	return data
}

func (p *procApi) ProcessControl(ctx *gin.Context, req struct {
	UUID int `form:"uuid" binding:"required"`
}) (err error) {
	user := getUserName(ctx)
	proc, err := logic.ProcessCtlLogic.GetProcess(req.UUID)
	if err != nil {
		return err
	}
	proc.ProcessControl(user)
	return
}

func (p *procApi) ProcessCreateShare(ctx *gin.Context, req model.ProcessShare) any {
	token := utils.UnwarpIgnore(uuid.NewRandom()).String()
	if err := repository.WsShare.AddShareData(model.WsShare{
		ExpireTime: time.Now().Add(time.Minute * time.Duration(req.Minutes)),
		Write:      req.Write,
		Token:      token,
		Pid:        req.Pid,
		CreateBy:   getUserName(ctx),
	}); err != nil {
		return err
	}
	return gin.H{
		"token": token,
	}
}
