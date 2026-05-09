package logic

import (
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
)

type permissionLogic struct{}

var PermissionLogic = new(permissionLogic)

func (p *permissionLogic) GetPermission(user string, pid int) (result *model.Permission) {
	return repository.PermissionRepository.GetPermission(user, pid)
}

func (p *permissionLogic) EditPermssion(data model.Permission) error {
	return repository.PermissionRepository.EditPermssion(data)
}

func (p *permissionLogic) GetPermssionList(account string) []model.PermissionPo {
	return repository.PermissionRepository.GetPermssionList(account)
}

func (p *permissionLogic) GetProcessNameByPermission(user string, op eum.OprPermission) (result []string) {
	return repository.PermissionRepository.GetProcessNameByPermission(user, op)
}
