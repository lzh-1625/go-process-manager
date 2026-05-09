//go:build linux

package logic

import (
	"github.com/containerd/cgroups/v3"
	"github.com/containerd/cgroups/v3/cgroup1"
	"github.com/containerd/cgroups/v3/cgroup2"
	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/log"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func (p *ProcessBase) initCgroup() {
	if !p.Config.cgroupEnable {
		log.Logger.Debugw("cgroup not enabled")
		return
	}
	switch cgroups.Mode() {
	case cgroups.Unavailable:
		log.Logger.Warnw("cgroup not supported by current system")
	case cgroups.Legacy, cgroups.Hybrid:
		log.Logger.Debugw("cgroupv1 enabled")
		p.initCgroupV1()
	case cgroups.Unified:
		log.Logger.Debugw("cgroupv2 enabled")
		p.initCgroupV2()
	}
}

func (p *ProcessBase) initCgroupV1() {
	resources := &specs.LinuxResources{}
	if p.Config.cpuLimit != nil {
		period := uint64(config.CF.CgroupPeriod)
		quota := int64(float32(config.CF.CgroupPeriod) * *p.Config.cpuLimit * 0.01)
		cpuResources := &specs.LinuxCPU{
			Period: &period,
			Quota:  &quota,
		}
		resources.CPU = cpuResources
	}
	if p.Config.memoryLimit != nil {
		limit := int64(*p.Config.memoryLimit * 1024 * 1024)
		memResources := &specs.LinuxMemory{
			Limit: &limit,
		}
		if config.CF.CgroupSwapLimit {
			memResources.Swap = &limit
		}
		resources.Memory = memResources
	}
	control, err := cgroup1.New(cgroup1.StaticPath("/"+p.Name), resources)
	if err != nil {
		log.Logger.Errorw("enable cgroup failed", "err", err, "name", p.Name)
		return
	}
	control.AddProc(uint64(p.Pid))
	p.cgroup.delete = control.Delete
	p.cgroup.enable = true
}

func (p *ProcessBase) initCgroupV2() {
	resources := &cgroup2.Resources{}
	if p.Config.cpuLimit != nil {
		period := uint64(config.CF.CgroupPeriod)
		quota := int64(float32(config.CF.CgroupPeriod) * *p.Config.cpuLimit * 0.01)
		resources.CPU = &cgroup2.CPU{
			Max: cgroup2.NewCPUMax(&quota, &period),
		}
	}
	if p.Config.memoryLimit != nil {
		limit := int64(*p.Config.memoryLimit * 1024 * 1024)
		memResources := &cgroup2.Memory{
			Max: &limit,
		}
		if config.CF.CgroupSwapLimit {
			memResources.Swap = &limit
		}
		resources.Memory = memResources
	}
	control, err := cgroup2.NewSystemd("/", p.Name+".slice", -1, resources)
	if err != nil {
		log.Logger.Errorw("enable cgroup failed", "err", err, "name", p.Name)
		return
	}
	control.AddProc(uint64(p.Pid))
	p.cgroup.delete = control.DeleteSystemd
	p.cgroup.enable = true
}
