package api

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/utils"

	"github.com/gin-gonic/gin"
)

type procApi struct{}

var ProcApi = new(procApi)

func (p *procApi) CreateNewProcess(ctx *gin.Context, req model.Process) (err error) {
	index, err := repository.ProcessRepository.AddProcessConfig(req)
	if err != nil {
		return err
	}
	req.Uuid = index
	proc, err := logic.ProcessCtlLogic.NewProcess(req)
	if err != nil {
		return err
	}
	logic.ProcessCtlLogic.AddProcess(req.Uuid, proc)
	rOk(ctx, "Operation successful!", gin.H{
		"id": req.Uuid,
	})
	return
}

func (p *procApi) DeleteNewProcess(ctx *gin.Context, req model.ProcessUuidReq) (err error) {
	logic.ProcessCtlLogic.KillProcess(req.Uuid)
	logic.ProcessCtlLogic.DeleteProcess(req.Uuid)
	return repository.ProcessRepository.DeleteProcessConfig(req.Uuid)
}

func (p *procApi) KillProcess(ctx *gin.Context, req model.ProcessUuidReq) (err error) {
	return logic.ProcessCtlLogic.KillProcess(req.Uuid)
}

func (p *procApi) StartProcess(ctx *gin.Context, req model.ProcessUuidReq) (err error) {
	prod, err := logic.ProcessCtlLogic.GetProcess(req.Uuid)
	if err != nil { // 进程不存在则创建
		proc, err1 := logic.ProcessCtlLogic.RunNewProcess(repository.ProcessRepository.GetProcessConfigById(req.Uuid))
		if err1 != nil {
			return err1
		}
		logic.ProcessCtlLogic.AddProcess(req.Uuid, proc)
		return nil
	}
	if prod.State.State == 1 {
		return errors.New("process is currently running")
	}
	prod.ResetRestartTimes()
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

func (p *procApi) GetProcessList(ctx *gin.Context, _ any) (err error) {
	if isAdmin(ctx) {
		rOk(ctx, "Query successful!", logic.ProcessCtlLogic.GetProcessList())
	} else {
		rOk(ctx, "Query successful!", logic.ProcessCtlLogic.GetProcessListByUser(getUserName(ctx)))
	}
	return
}

func (p *procApi) UpdateProcessConfig(ctx *gin.Context, req model.Process) (err error) {
	logic.ProcessCtlLogic.UpdateProcessConfig(req)
	err = repository.ProcessRepository.UpdateProcessConfig(req)
	return
}

func (p *procApi) GetProcessConfig(ctx *gin.Context, req model.ProcessUuidReq) (err error) {
	data := repository.ProcessRepository.GetProcessConfigById(req.Uuid)
	if data.Uuid == 0 {
		return errors.New("no information found")
	}
	rOk(ctx, "success", data)
	return
}

func (p *procApi) ProcessControl(ctx *gin.Context, req model.ProcessUuidReq) (err error) {
	user := getUserName(ctx)
	proc, err := logic.ProcessCtlLogic.GetProcess(req.Uuid)
	if err != nil {
		return err
	}
	proc.ProcessControl(user)
	return
}

func (p *procApi) ProcessCreateShare(ctx *gin.Context, req model.ProcessShare) (err error) {
	token := utils.UnwarpIgnore(uuid.NewRandom()).String()
	if err = repository.WsShare.AddShareData(model.WsShare{
		ExpireTime: time.Now().Add(time.Minute * time.Duration(req.Minutes)),
		Write:      req.Write,
		Token:      token,
		Pid:        req.Pid,
		CreateBy:   getUserName(ctx),
	}); err != nil {
		return err
	}
	rOk(ctx, "Operation successful!", gin.H{
		"token": token,
	})
	return
}
