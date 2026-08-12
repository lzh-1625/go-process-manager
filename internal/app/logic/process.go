package logic

import (
	"errors"
	"fmt"
	"io"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/shlex"
	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/process"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/internal/app/types"
	"github.com/lzh-1625/go_process_manager/log"
	"github.com/lzh-1625/go_process_manager/utils"
	"github.com/shirou/gopsutil/mem"
)

type ProcessCtlLogic struct {
	processMap           sync.Map
	processRepository    *repository.ProcessRepository
	permissionRepository *repository.PermissionRepository
	eventLogic           *EventLogic
	pushLogic            *PushLogic
	logHandler           *LogHandler
	processStateHandler  ProcessStateHandler
}

type ProcessStateHandler func(processName string, state types.ProcessState)

func NewProcessCtlLogic(
	processRepository *repository.ProcessRepository,
	permissionRepository *repository.PermissionRepository,
	eventLogic *EventLogic,
	pushLogic *PushLogic,
	logHandler *LogHandler,
) *ProcessCtlLogic {
	return &ProcessCtlLogic{
		processMap:           sync.Map{},
		processRepository:    processRepository,
		permissionRepository: permissionRepository,
		eventLogic:           eventLogic,
		pushLogic:            pushLogic,
		logHandler:           logHandler,
	}
}

func (p *ProcessCtlLogic) SetProcessStateHandler(handler ProcessStateHandler) {
	p.processStateHandler = handler
}

func (p *ProcessCtlLogic) KillProcess(uuid int) error {
	value, ok := p.processMap.Load(uuid)
	if !ok {
		return errors.New("process not exist")
	}
	result, ok := value.(*process.Process)
	if !ok {
		return errors.New("process type error")
	}
	return result.Kill()
}

func (p *ProcessCtlLogic) GetProcess(uuid int) (*process.Process, error) {
	proc, ok := p.processMap.Load(uuid)
	if !ok {
		return nil, errors.New("process not exist")
	}
	result, ok := proc.(*process.Process)
	if !ok {
		return nil, errors.New("process type error")

	}
	return result, nil
}

func (p *ProcessCtlLogic) DeleteProcess(uuid int) error {
	proc, err := p.GetProcess(uuid)
	if err != nil {
		return err
	}
	if proc.State.State != types.ProcessStateStopped && proc.State.State != types.ProcessStateWarning {
		return errors.New("stop the process before deleting it")
	}

	if err := p.processRepository.DeleteProcessConfig(uuid); err != nil {
		return err
	}
	p.processMap.Delete(uuid)
	return nil
}

func (p *ProcessCtlLogic) GetProcessList() []model.ProcessInfo {
	processConfiglist := p.processRepository.GetAllProcessConfig()
	return p.getProcessInfoList(processConfiglist)
}

func (p *ProcessCtlLogic) GetProcessListByUser(username string) []model.ProcessInfo {
	processConfiglist := p.processRepository.GetProcessConfigByUser(username)
	return p.getProcessInfoList(processConfiglist)
}

func (p *ProcessCtlLogic) getProcessInfoList(processConfiglist []*model.Process) []model.ProcessInfo {
	processInfoList := []model.ProcessInfo{}
	for _, v := range processConfiglist {
		pi := model.ProcessInfo{
			Name: v.Name,
			UUID: v.UUID,
		}
		process, err := p.GetProcess(v.UUID)
		if err != nil {
			continue
		}
		if !process.VerifyControl() {
			pi.Controller = process.Control.Controller
			pi.ControlExpiredTime = process.Control.ControlExpiredTime
		}
		pi.State.Info = process.State.Info
		pi.State.State = process.State.State
		pi.StartTime = process.State.StartTime.Format(time.DateTime)
		pi.User = process.GetUserString()
		pi.Usage.Cpu = process.PerformanceStatus.Cpu
		pi.Usage.Mem = process.PerformanceStatus.Mem
		pi.Usage.CpuCapacity = float64(runtime.NumCPU()) * 100.0
		pi.Usage.MemCapacity = float64(utils.UnwarpIgnore(mem.VirtualMemory()).Total >> 10)
		for _, v := range process.PerformanceStatus.Time {
			pi.Usage.Time = append(pi.Usage.Time, v.Format(time.DateTime))
		}

		// real-time performance information
		if c, m, err := process.GetPerformanceInfo(); err == nil {
			pi.Usage.Cpu = append(pi.Usage.Cpu, c)
			pi.Usage.Mem = append(pi.Usage.Mem, m)
			pi.Usage.Time = append(pi.Usage.Time, time.Now().Format(time.DateTime))
		}
		pi.CgroupEnable = process.Config.CgroupEnable
		pi.CpuLimit = process.Config.CpuLimit
		pi.MemoryLimit = process.Config.MemoryLimit
		pi.Env = process.Env
		processInfoList = append(processInfoList, pi)
	}
	return processInfoList
}

func (p *ProcessCtlLogic) ProcessInit() {
	config := p.processRepository.GetAllProcessConfig()
	for _, v := range config {
		proc, err := p.createProcess(*v)
		if err != nil {
			log.Logger.Error("initialize process start failed", "name", v.Name, "err", err)
			continue
		}
		if v.AutoRestart {
			err := proc.Start()
			if err != nil {
				log.Logger.Warnw("initialize process start failed", v.Name, "name", "err", err)
				continue
			}
		}
	}
}

func (p *ProcessCtlLogic) ForEach(fn func(proc *process.Process)) {
	p.processMap.Range(func(key, value any) bool {
		process := value.(*process.Process)
		fn(process)
		return true
	})
}

func (p *ProcessCtlLogic) ForEachByOwner(userName string, fn func(proc *process.Process)) {
	startPermissionProcess := p.permissionRepository.GetProcessNameByPermission(userName, types.OperationStart)
	p.processMap.Range(func(key, value any) bool {
		process := value.(*process.Process)
		if !slices.Contains(startPermissionProcess, process.Name) {
			return true
		}
		fn(process)
		return true
	})
}

func (p *ProcessCtlLogic) GetProcessConfigByID(uuid int) (*model.Process, error) {
	return p.processRepository.GetProcessConfigByID(uuid)
}

func (p *ProcessCtlLogic) GetProcessConfigByName(name string) (*model.Process, error) {
	return p.processRepository.GetProcessConfigByName(name)
}

func (p *ProcessCtlLogic) UpdateProcessConfig(config model.Process) error {
	proc, err := p.GetProcess(config.UUID)
	if err != nil {
		return err
	}
	if err := p.processRepository.UpdateProcessConfig(config); err != nil {
		return err
	}
	proc.Config.LogReport = config.LogReport
	proc.Config.CgroupEnable = config.CgroupEnable
	proc.Config.MemoryLimit = config.MemoryLimit
	proc.Config.CpuLimit = config.CpuLimit
	proc.Config.AutoRestart = config.AutoRestart
	proc.Config.CompulsoryRestart = config.CompulsoryRestart
	proc.StartCommand = utils.UnwarpIgnore(shlex.Split(config.Cmd))
	proc.WorkDir = config.Cwd
	proc.Name = config.Name
	proc.Env = strings.Split(config.Env, ";")
	return nil
}

func (p *ProcessCtlLogic) CreateProcess(config model.Process) (proc *process.Process, err error) {
	if args, err := shlex.Split(config.Cmd); len(args) == 0 || err != nil {
		return nil, errors.New("invalid command")
	}
	index, err := p.processRepository.AddProcessConfig(config)
	if err != nil {
		return nil, err
	}
	config.UUID = index
	return p.createProcess(config)
}

func (p *ProcessCtlLogic) createProcess(cf model.Process) (*process.Process, error) {
	proc := process.NewProcess(cf,
		process.SetAddWriterHook(func(proc *process.Process, user string, c io.WriteCloser) {
			p.eventLogic.Create(proc.Name, user, types.EventProcessConnect)
			ProcessWaitCond().Trigger()
		}),
		process.SetDelWriterHook(func(p *process.Process, user string) {
			ProcessWaitCond().Trigger()
		}),
		process.SetLogHandler(config.CF.LogReportOptimization, func(proc *process.Process, log []byte) {
			logStr := string(log)
			if strings.TrimSpace(utils.RemoveANSI(logStr)) == "" {
				return
			}
			p.logHandler.AddLog(model.ProcessLog{
				Using: proc.GetUserString(),
				Name:  proc.Name,
				Log:   logStr,
				Time:  time.Now().UnixMilli(),
			})
		}),
		process.SetStateHook(func(proc *process.Process, state types.ProcessState) {
			ProcessWaitCond().Trigger()
			p.push(proc, state)
			p.createEvent(proc, state)
			if p.processStateHandler != nil {
				go p.processStateHandler(proc.Name, state)
			}
		}),
	)
	if _, loaded := p.processMap.LoadOrStore(cf.UUID, proc); loaded {
		return nil, fmt.Errorf("process UUID %d already exists", cf.UUID)
	}
	return proc, nil
}

func (p *ProcessCtlLogic) createEvent(proc *process.Process, state types.ProcessState) {
	switch state {
	case types.ProcessStateRunning:
		p.eventLogic.Create(proc.Name, proc.GetOperator(), types.EventProcessStart, "restartTimes", strconv.Itoa(proc.State.RestartTimes))
	case types.ProcessStateStopping:
		stoppingTime := time.Now()
		operator := proc.GetOperator()
		go func() {
			<-proc.StopChan
			p.eventLogic.Create(proc.Name, operator, types.EventProcessStop, "startTime", proc.State.StartTime.Format(time.DateTime), "wait", time.Since(stoppingTime).String())
		}()
	case types.ProcessStateWarning:
		p.eventLogic.Create(proc.Name, proc.GetOperator(), types.EventProcessWarning, "reason", proc.State.Info, "startTime", proc.State.StartTime.Format(time.DateTime))
	default:
		return
	}
}

func (p *ProcessCtlLogic) push(proc *process.Process, state types.ProcessState) {
	if state == types.ProcessStateRunning || state == types.ProcessStateStopping || state == types.ProcessStateWaitingRestart {
		return
	}
	data, err := p.processRepository.GetProcessConfigByID(proc.UUID)
	if err != nil {
		return
	}
	pushIDs := utils.JsonStrToStruct[[]int64](data.PushIDs)
	messagePlaceholders := map[string]string{
		"{$name}":   proc.Name,
		"{$user}":   proc.GetOperator(),
		"{$status}": state.String(),
		"{$pid}":    strconv.Itoa(proc.Pid),
	}
	go p.pushLogic.Push(pushIDs, messagePlaceholders)
}
