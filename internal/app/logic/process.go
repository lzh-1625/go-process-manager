package logic

import (
	"errors"
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

func (p *ProcessCtlLogic) AddProcess(uuid int, proc *process.ProcessPty) {
	p.processMap.Store(uuid, proc)
}

func (p *ProcessCtlLogic) KillProcess(uuid int, wait bool) error {
	value, ok := p.processMap.Load(uuid)
	if !ok {
		return errors.New("process not exist")
	}
	result, ok := value.(*process.ProcessPty)
	if !ok {
		return errors.New("process type error")
	}
	if wait {
		return result.Kill()
	}
	return result.Kill9()
}

func (p *ProcessCtlLogic) GetProcess(uuid int) (*process.ProcessPty, error) {
	proc, ok := p.processMap.Load(uuid)
	if !ok {
		return nil, errors.New("process not exist")
	}
	result, ok := proc.(*process.ProcessPty)
	if !ok {
		return nil, errors.New("process type error")

	}
	return result, nil
}

func (p *ProcessCtlLogic) KillAllProcess() {
	wg := sync.WaitGroup{}
	p.processMap.Range(func(key, value any) bool {
		process := value.(*process.ProcessPty)
		wg.Go(func() {
			process.Kill()
		})
		return true
	})
	wg.Wait()
}

func (p *ProcessCtlLogic) KillAllProcessByUserName(userName string) {
	stopPermissionProcess := p.permissionRepository.GetProcessNameByPermission(userName, types.OperationStop)
	wg := sync.WaitGroup{}
	p.processMap.Range(func(key, value any) bool {
		process := value.(*process.ProcessPty)
		if !slices.Contains(stopPermissionProcess, process.Name) {
			return true
		}
		wg.Go(func() {
			process.Kill()
		})
		return true
	})
	wg.Wait()
}

func (p *ProcessCtlLogic) DeleteProcess(uuid int) error {
	p.KillProcess(uuid, true)
	p.processMap.Delete(uuid)
	return p.processRepository.DeleteProcessConfig(uuid)
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

func (p *ProcessCtlLogic) ProcessStartAll() {
	p.processMap.Range(func(key, value any) bool {
		process := value.(*process.ProcessPty)
		err := process.Start()
		if err != nil {
			log.Logger.Errorw("process start failed", "name", process.Name)
		}
		return true
	})
}

func (p *ProcessCtlLogic) ProcessInit() {
	config := p.processRepository.GetAllProcessConfig()
	for _, v := range config {
		proc := p.createProcess(*v)
		p.AddProcess(v.UUID, proc)
		if v.AutoRestart {
			err := proc.Start()
			if err != nil {
				log.Logger.Warnw("initialize process start failed", v.Name, "name", "err", err)
				continue
			}
		}
	}
}

func (p *ProcessCtlLogic) ProcesStartAllByUsername(userName string) {
	startPermissionProcess := p.permissionRepository.GetProcessNameByPermission(userName, types.OperationStart)
	p.processMap.Range(func(key, value any) bool {
		process := value.(*process.ProcessPty)
		if !slices.Contains(startPermissionProcess, process.Name) {
			return true
		}
		err := process.Start()
		if err != nil {
			log.Logger.Errorw("process start failed", "name", process.Name)
		}
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
	proc, ok := p.processMap.Load(config.UUID)
	if !ok {
		return errors.New("process get failed")
	}
	result, ok := proc.(*process.ProcessPty)
	if !ok {
		return errors.New("process type error")
	}
	if !result.Lock.TryLock() {
		return errors.New("process is being used")
	}
	defer result.Lock.Unlock()
	result.Config.LogReport = config.LogReport
	result.Config.PushIDs = utils.JsonStrToStruct[[]int64](config.PushIDs)
	result.Config.CgroupEnable = config.CgroupEnable
	result.Config.MemoryLimit = config.MemoryLimit
	result.Config.CpuLimit = config.CpuLimit
	result.Config.AutoRestart = config.AutoRestart
	result.Config.CompulsoryRestart = config.CompulsoryRestart
	result.StartCommand = utils.UnwarpIgnore(shlex.Split(config.Cmd))
	result.WorkDir = config.Cwd
	result.Name = config.Name
	result.Env = strings.Split(config.Env, ";")
	return p.processRepository.UpdateProcessConfig(config)
}

func (p *ProcessCtlLogic) NewProcess(config model.Process) (proc *process.ProcessPty) {
	index, err := p.processRepository.AddProcessConfig(config)
	if err != nil {
		return nil
	}
	config.UUID = index
	proc = p.createProcess(config)
	p.AddProcess(config.UUID, proc)
	return
}

func (p *ProcessCtlLogic) RunProcess(config model.Process) (proc *process.ProcessPty, err error) {
	proc = p.createProcess(config)
	p.AddProcess(config.UUID, proc)
	err = proc.Start()
	return
}

func (p *ProcessCtlLogic) createProcess(cf model.Process) (proc *process.ProcessPty) {
	return process.NewProcessPty(cf,
		process.SetAddWriterHook(func(p *process.ProcessBase, user string, c io.WriteCloser) {
			ProcessWaitCond().Trigger()
		}),
		process.SetDelWriterHook(func(p *process.ProcessBase, user string) {
			ProcessWaitCond().Trigger()
		}),
		process.SetLogHandler(config.CF.LogReportOptimization, func(proc *process.ProcessBase, log []byte) {
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
		process.SetPushHandle(func(proc *process.ProcessBase, pushIDs []int64, messagePlaceholders map[string]string) {
			p.pushLogic.Push(pushIDs, messagePlaceholders)
		}),
		process.SetStateHook(func(proc *process.ProcessBase, state types.ProcessState) {
			ProcessWaitCond().Trigger()
			p.createEvent(proc, state)
			if p.processStateHandler != nil {
				go p.processStateHandler(proc.Name, state)
			}
		}),
	)
}

func (p *ProcessCtlLogic) createEvent(proc *process.ProcessBase, state types.ProcessState) {
	var eventType types.EventType
	kv := []string{}
	switch state {
	case types.ProcessStateRunning:
		eventType = types.EventProcessStart
		kv = append(kv, "restartTimes", strconv.Itoa(proc.State.RestartTimes))
	case types.ProcessStateStopped:
		eventType = types.EventProcessStop
		kv = append(kv, "startTime", proc.State.StartTime.Format(time.DateTime))
	case types.ProcessStateWarning:
		eventType = types.EventProcessWarning
		kv = append(kv, "reason", proc.State.Info, "startTime", proc.State.StartTime.Format(time.DateTime))
	default:
		return
	}
	kv = append(kv, "operator", proc.GetOpertor())
	p.eventLogic.Create(proc.Name, eventType, kv...)
}
