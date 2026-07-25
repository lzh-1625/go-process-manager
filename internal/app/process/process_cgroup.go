//go:build !linux

package process

import "github.com/lzh-1625/go_process_manager/log"

func (p *Process) initCgroup() {
	log.Logger.Debugw("cgroup not supported")
}
