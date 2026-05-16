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
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/process"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
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
	eventBus             *EventBus
}

func NewProcessCtlLogic(
	processRepository *repository.ProcessRepository,
	permissionRepository *repository.PermissionRepository,
	eventLogic *EventLogic,
	pushLogic *PushLogic,
	// taskLogic *TaskLogic,
	logHandler *LogHandler,
	eventBus *EventBus,
) *ProcessCtlLogic {
	return &ProcessCtlLogic{
		processMap:           sync.Map{},
		processRepository:    processRepository,
		permissionRepository: permissionRepository,
		eventLogic:           eventLogic,
		pushLogic:            pushLogic,
		logHandler:           logHandler,
		eventBus:             eventBus,
	}
}
func (p *ProcessCtlLogic) AddProcess(uuid int, proc *process.ProcessPty) {
	p.processMap.Store(uuid, proc)
}

func (p *ProcessCtlLogic) KillProcess(uuid int) error {
	value, ok := p.processMap.Load(uuid)
	if !ok {
		return errors.New("process not exist")
	}
	result, ok := value.(*process.ProcessPty)
	if !ok {
		return errors.New("process type error")
	}
	return result.Kill()
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
	stopPermissionProcess := p.permissionRepository.GetProcessNameByPermission(userName, eum.OperationStop)
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
	p.KillProcess(uuid)
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

		// 使用 Info() 方法获取进程信息快照
		pi.State.Info = process.State.Info
		pi.State.State = process.State.State
		pi.StartTime = process.State.StartTime.Format(time.DateTime)
		pi.User = process.GetUserString()
		pi.Usage.Cpu = process.PerformanceStatus.Cpu
		pi.Usage.Mem = process.PerformanceStatus.Mem
		pi.Usage.CpuCapacity = float64(runtime.NumCPU()) * 100.0
		pi.Usage.MemCapacity = float64(utils.UnwarpIgnore(mem.VirtualMemory()).Total >> 10)
		pi.Usage.Time = process.PerformanceStatus.Time
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
	startPermissionProcess := p.permissionRepository.GetProcessNameByPermission(userName, eum.OperationStart)
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
		process.SetAddCoonHook(func(p *process.ProcessBase, user string, c io.WriteCloser) {
			ProcessWaitCond.Trigger()
		}),
		process.SetDelCoonHook(func(p *process.ProcessBase, user string) {
			ProcessWaitCond.Trigger()
		}),
		process.SetLogHandler(func(proc *process.ProcessBase) process.IProcessLogHandler {
			if config.CF.LogReportOptimization && cf.LogReport {
				return process.NewProcessLogHandlerByPipe(func(log []byte) {
					p.logHandler.AddLog(model.ProcessLog{
						Using: proc.GetUserString(),
						Name:  proc.Name,
						Log:   string(log),
						Time:  time.Now().UnixMilli(),
					})
				})
			} else {
				return process.NewProcessLogHandler(func(log []byte) {
					p.logHandler.AddLog(model.ProcessLog{
						Using: proc.GetUserString(),
						Name:  proc.Name,
						Log:   string(log),
						Time:  time.Now().UnixMilli(),
					})
				})
			}
		}),
		process.SetPushHandle(func(proc *process.ProcessBase, pushIDs []int64, messagePlaceholders map[string]string) {
			p.pushLogic.Push(pushIDs, messagePlaceholders)
		}),
		process.SetStateHook(func(proc *process.ProcessBase, state eum.ProcessState) {
			ProcessWaitCond.Trigger()
			p.createEvent(proc, state)
			p.eventBus.Publish(Event{
				Proc:  proc,
				State: state,
			})
		}),
	)
}

func (p *ProcessCtlLogic) createEvent(proc *process.ProcessBase, state eum.ProcessState) {
	var eventType eum.EventType
	kv := []string{}
	switch state {
	case eum.ProcessStateRunning:
		eventType = eum.EventProcessStart
		kv = append(kv, "restartTimes", strconv.Itoa(proc.State.RestartTimes))
	case eum.ProcessStateStop:
		eventType = eum.EventProcessStop
		kv = append(kv, "startTime", proc.State.StartTime.Format(time.DateTime))
	case eum.ProcessStateWarnning:
		eventType = eum.EventProcessWarning
		kv = append(kv, "reason", proc.State.Info, "startTime", proc.State.StartTime.Format(time.DateTime))
	default:
		return
	}
	kv = append(kv, "operator", proc.GetOpertor())
	p.eventLogic.Create(proc.Name, eventType, kv...)
}
