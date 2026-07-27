//go:build !linux

package process

func (p *Process) initCgroup() {
	logger.Debug("cgroup not supported")
}
