package logic

import (
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
)

type PermissionLogic struct {
	permissionRepository *repository.PermissionRepository
}

func NewPermissionLogic(permissionRepository *repository.PermissionRepository) *PermissionLogic {
	return &PermissionLogic{
		permissionRepository: permissionRepository,
	}
}

func (p *PermissionLogic) GetPermission(user string, pid int) (result *model.Permission) {
	return p.permissionRepository.GetPermission(user, pid)
}

func (p *PermissionLogic) EditPermssion(data model.Permission) error {
	return p.permissionRepository.EditPermssion(data)
}

func (p *PermissionLogic) GetPermssionList(account string) []model.PermissionPo {
	return p.permissionRepository.GetPermssionList(account)
}

func (p *PermissionLogic) GetProcessNameByPermission(user string, op eum.OprPermission) (result []string) {
	return p.permissionRepository.GetProcessNameByPermission(user, op)
}
