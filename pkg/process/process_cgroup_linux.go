//go:build linux

package process

import (
	"fmt"

	"github.com/containerd/cgroups/v3"
	"github.com/containerd/cgroups/v3/cgroup1"
	"github.com/containerd/cgroups/v3/cgroup2"

	"github.com/opencontainers/runtime-spec/specs-go"
)

func (p *Process) initCgroup() {
	if !p.Config.CgroupEnable {
		logger.Debug("cgroup not enabled")
		return
	}
	switch cgroups.Mode() {
	case cgroups.Unavailable:
		logger.Warn("cgroup not supported by current system")
	case cgroups.Legacy, cgroups.Hybrid:
		logger.Debug("cgroupv1 enabled")
		p.initCgroupV1()
	case cgroups.Unified:
		logger.Debug("cgroupv2 enabled")
		p.initCgroupV2()
	}
}

func (p *Process) initCgroupV1() {
	resources := &specs.LinuxResources{}
	if p.Config.CpuLimit != nil {
		resources.CPU = &specs.LinuxCPU{
			Period: new(uint64(p.Config.CgroupPeriod)),
			Quota:  new(int64(float32(p.Config.CgroupPeriod) * *p.Config.CpuLimit * 0.01)),
		}
	}
	if p.Config.MemoryLimit != nil {
		limit := int64(*p.Config.MemoryLimit * 1024 * 1024)
		memResources := &specs.LinuxMemory{
			Limit: &limit,
		}
		if p.Config.CgroupSwapLimit {
			memResources.Swap = &limit
		}
		resources.Memory = memResources
	}
	control, err := cgroup1.New(cgroup1.StaticPath(fmt.Sprintf("/GPM%d", p.Pid)), resources)
	if err != nil {
		logger.Error("enable cgroup failed", "err", err, "name", p.Name)
		return
	}
	if err := control.AddProc(uint64(p.Pid)); err != nil {
		logger.Error("add process to cgroup failed", "err", err, "name", p.Name)
		return
	}
	p.cgroup.delete = control.Delete
	p.cgroup.enable = true
}

func (p *Process) initCgroupV2() {
	resources := &cgroup2.Resources{}
	if p.Config.CpuLimit != nil {
		period := uint64(p.Config.CgroupPeriod)
		quota := int64(float32(p.Config.CgroupPeriod) * *p.Config.CpuLimit * 0.01)
		resources.CPU = &cgroup2.CPU{
			Max: cgroup2.NewCPUMax(&quota, &period),
		}
	}
	if p.Config.MemoryLimit != nil {
		limit := int64(*p.Config.MemoryLimit * 1024 * 1024)
		memResources := &cgroup2.Memory{
			Max: &limit,
		}
		if p.Config.CgroupSwapLimit {
			memResources.Swap = &limit
		}
		resources.Memory = memResources
	}
	control, err := cgroup2.NewSystemd("/", fmt.Sprintf("GPM%d.slice", p.Pid), -1, resources)
	if err != nil {
		logger.Error("enable cgroup failed", "err", err, "name", p.Name)
		return
	}
	if err := control.AddProc(uint64(p.Pid)); err != nil {
		logger.Error("add process to cgroup failed", "err", err, "name", p.Name)
		return
	}
	p.cgroup.delete = control.DeleteSystemd
	p.cgroup.enable = true
}
