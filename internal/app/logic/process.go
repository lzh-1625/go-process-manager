package logic

import (
	"errors"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/shlex"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/process"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/log"
	"github.com/lzh-1625/go_process_manager/utils"
	"github.com/shirou/gopsutil/mem"
)

type processCtlLogic struct {
	processMap sync.Map
}

var (
	ProcessCtlLogic = new(processCtlLogic)
)

func (p *processCtlLogic) AddProcess(uuid int, proc *process.ProcessPty) {
	p.processMap.Store(uuid, proc)
}

func (p *processCtlLogic) KillProcess(uuid int) error {
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

func (p *processCtlLogic) GetProcess(uuid int) (*process.ProcessPty, error) {
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

func (p *processCtlLogic) KillAllProcess() {
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

func (p *processCtlLogic) KillAllProcessByUserName(userName string) {
	stopPermissionProcess := repository.PermissionRepository.GetProcessNameByPermission(userName, eum.OperationStop)
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

func (p *processCtlLogic) DeleteProcess(uuid int) error {
	p.KillProcess(uuid)
	p.processMap.Delete(uuid)
	return repository.ProcessRepository.DeleteProcessConfig(uuid)
}

func (p *processCtlLogic) GetProcessList() []model.ProcessInfo {
	processConfiglist := repository.ProcessRepository.GetAllProcessConfig()
	return p.getProcessInfoList(processConfiglist)
}

func (p *processCtlLogic) GetProcessListByUser(username string) []model.ProcessInfo {
	processConfiglist := repository.ProcessRepository.GetProcessConfigByUser(username)
	return p.getProcessInfoList(processConfiglist)
}

func (p *processCtlLogic) getProcessInfoList(processConfiglist []*model.Process) []model.ProcessInfo {
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

func (p *processCtlLogic) ProcessStartAll() {
	p.processMap.Range(func(key, value any) bool {
		process := value.(*process.ProcessPty)
		err := process.Start()
		if err != nil {
			log.Logger.Errorw("process start failed", "name", process.Name)
		}
		return true
	})
}

func (p *processCtlLogic) ProcessInit() {
	config := repository.ProcessRepository.GetAllProcessConfig()
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

func (p *processCtlLogic) ProcesStartAllByUsername(userName string) {
	startPermissionProcess := repository.PermissionRepository.GetProcessNameByPermission(userName, eum.OperationStart)
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

func (p *processCtlLogic) GetProcessConfigByID(uuid int) (*model.Process, error) {
	return repository.ProcessRepository.GetProcessConfigByID(uuid)
}

func (p *processCtlLogic) UpdateProcessConfig(config model.Process) error {
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
	return repository.ProcessRepository.UpdateProcessConfig(config)
}

func (p *processCtlLogic) NewProcess(config model.Process) (proc *process.ProcessPty) {
	index, err := repository.ProcessRepository.AddProcessConfig(config)
	if err != nil {
		return nil
	}
	config.UUID = index
	proc = p.createProcess(config)
	p.AddProcess(config.UUID, proc)
	return
}

func (p *processCtlLogic) RunProcess(config model.Process) (proc *process.ProcessPty, err error) {
	proc = p.createProcess(config)
	p.AddProcess(config.UUID, proc)
	err = proc.Start()
	return
}

func (p *processCtlLogic) createProcess(config model.Process) (proc *process.ProcessPty) {
	return process.NewProcessPty(config,
		process.SetAddCoonHook(func(p *process.ProcessBase, user string, c process.ConnectInstance) {
			ProcessWaitCond.Trigger()
		}),
		process.SetDelCoonHook(func(p *process.ProcessBase, user string) {
			ProcessWaitCond.Trigger()
		}),
		process.SetLogHandle(func(p *process.ProcessBase, log string) {
			Loghandler.AddLog(model.ProcessLog{
				Log:   log,
				Using: p.GetUserString(),
				Name:  p.Name,
				Time:  time.Now().UnixMilli(),
			})
		}),
		process.SetPushHandle(func(p *process.ProcessBase, pushIDs []int64, messagePlaceholders map[string]string) {
			PushLogic.Push(p.Config.PushIDs, messagePlaceholders)
		}),
		process.SetStateHook(func(proc *process.ProcessBase, state eum.ProcessState) {
			ProcessWaitCond.Trigger()
			p.createEvent(proc, state)
			go TaskLogic.RunTaskByTriggerEvent(proc.Name, state)
		}),
	)
}

func (p *processCtlLogic) createEvent(proc *process.ProcessBase, state eum.ProcessState) {
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
	EventLogic.Create(proc.Name, eventType, kv...)
}
