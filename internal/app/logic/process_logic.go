package logic

import (
	"errors"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/shlex"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
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

func (p *processCtlLogic) AddProcess(uuid int, process *ProcessPty) {
	p.processMap.Store(uuid, process)
}

func (p *processCtlLogic) KillProcess(uuid int) error {
	value, ok := p.processMap.Load(uuid)
	if !ok {
		return errors.New("process not exist")
	}
	result, ok := value.(*ProcessPty)
	if !ok {
		return errors.New("process type error")
	}
	return result.Kill()
}

func (p *processCtlLogic) GetProcess(uuid int) (*ProcessPty, error) {
	process, ok := p.processMap.Load(uuid)
	if !ok {
		return nil, errors.New("process not exist")
	}
	result, ok := process.(*ProcessPty)
	if !ok {
		return nil, errors.New("process type error")

	}
	return result, nil
}

func (p *processCtlLogic) KillAllProcess() {
	wg := sync.WaitGroup{}
	p.processMap.Range(func(key, value any) bool {
		process := value.(*ProcessPty)
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
		process := value.(*ProcessPty)
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
	return nil
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
		pi.StartTime = process.State.startTime.Format(time.DateTime)
		pi.User = process.GetUserString()
		pi.Usage.Cpu = process.performanceStatus.cpu
		pi.Usage.Mem = process.performanceStatus.mem
		pi.Usage.CpuCapacity = float64(runtime.NumCPU()) * 100.0
		pi.Usage.MemCapacity = float64(utils.UnwarpIgnore(mem.VirtualMemory()).Total >> 10)
		pi.Usage.Time = process.performanceStatus.time
		pi.CgroupEnable = process.Config.cgroupEnable
		pi.CpuLimit = process.Config.cpuLimit
		pi.MemoryLimit = process.Config.memoryLimit
		pi.Env = process.Env
		processInfoList = append(processInfoList, pi)
	}
	return processInfoList
}

func (p *processCtlLogic) ProcessStartAll() {
	p.processMap.Range(func(key, value any) bool {
		process := value.(*ProcessPty)
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
		proc := p.NewProcess(*v)
		if v.AutoRestart {
			err := proc.Start()
			if err != nil {
				log.Logger.Warnw("initialize process start failed", v.Name, "name", "err", err)
				continue
			}
		}
		p.AddProcess(v.UUID, proc)
	}
}

func (p *processCtlLogic) ProcesStartAllByUsername(userName string) {
	startPermissionProcess := repository.PermissionRepository.GetProcessNameByPermission(userName, eum.OperationStart)
	p.processMap.Range(func(key, value any) bool {
		process := value.(*ProcessPty)
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

func (p *processCtlLogic) UpdateProcessConfig(config model.Process) error {
	process, ok := p.processMap.Load(config.UUID)
	if !ok {
		return errors.New("process get failed")
	}
	result, ok := process.(*ProcessPty)
	if !ok {
		return errors.New("process type error")
	}
	if !result.Lock.TryLock() {
		return errors.New("process is being used")
	}
	defer result.Lock.Unlock()
	result.Config.logReport = config.LogReport
	result.Config.PushIds = utils.JsonStrToStruct[[]int64](config.PushIds)
	result.Config.cgroupEnable = config.CgroupEnable
	result.Config.memoryLimit = config.MemoryLimit
	result.Config.cpuLimit = config.CpuLimit
	result.Config.AutoRestart = config.AutoRestart
	result.Config.compulsoryRestart = config.CompulsoryRestart
	result.StartCommand = utils.UnwarpIgnore(shlex.Split(config.Cmd))
	result.WorkDir = config.Cwd
	result.Name = config.Name
	result.Env = strings.Split(config.Env, ";")
	return nil
}

func (p *processCtlLogic) NewProcess(config model.Process) (proc *ProcessPty) {
	proc = NewProcessPty(config)
	p.AddProcess(config.UUID, proc)
	return
}

func (p *processCtlLogic) RunNewProcess(config model.Process) (proc *ProcessPty, err error) {
	proc = p.NewProcess(config)
	err = proc.Start()
	return
}
