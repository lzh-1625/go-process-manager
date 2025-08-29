package repository

import (
	"errors"

	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/query"
	"github.com/lzh-1625/go_process_manager/log"
	"gorm.io/gorm"
)

type permissionRepository struct{}

var PermissionRepository = new(permissionRepository)

func (p *permissionRepository) GetPermssionList(account string) []model.PermissionPo {
	result := []model.PermissionPo{}
	if err := db.Raw(`SELECT
	p.name ,
	p.uuid as pid,
	p2.owned ,
	p2."start" ,
	p2.stop ,
	p2.terminal,
	p2.log ,
	p2.write
FROM
	users u
full join process p
left join permission p2 on
	p2.account == u.account
	and p2.pid = p.uuid
WHERE
	u.account = ?
	or u.account ISNULL`, account).Find(&result); err.Error != nil {
		log.Logger.Warnw("权限查询失败", "err", err)
	}

	return result
}

func (p *permissionRepository) EditPermssion(data model.Permission) error {
	per := query.Permission
	if _, err := per.Where(per.Account.Eq(data.Account)).Where(per.Pid.Eq(data.Pid)).First(); errors.Is(err, gorm.ErrRecordNotFound) {
		per.Create(&model.Permission{
			Account: data.Account,
			Pid:     data.Pid,
		})
	}
	_, err := per.Where(per.Account.Eq(data.Account)).Where(per.Pid.Eq(data.Pid)).Updates(map[string]any{
		per.Owned.ColumnName().String():    data.Owned,
		per.Start.ColumnName().String():    data.Start,
		per.Stop.ColumnName().String():     data.Stop,
		per.Terminal.ColumnName().String(): data.Terminal,
		per.Log.ColumnName().String():      data.Log,
		per.Write.ColumnName().String():    data.Write,
	})
	return err
}

func (p *permissionRepository) GetPermission(user string, pid int) (result model.Permission) {
	db.Model(&model.Permission{}).Where(&model.Permission{Account: user, Pid: int32(pid)}).First(&result)
	return
}

func (p *permissionRepository) GetProcessNameByPermission(user string, op eum.OprPermission) (result []string) {
	tx := query.Permission.Select(query.Process.Name).RightJoin(query.Process, query.Process.Uuid.EqCol(query.Permission.Pid)).Where(query.Permission.Account.Eq(user)).Where(query.Permission.Owned.Is(true))
	switch op {
	case eum.OperationLog:
		tx = tx.Where(query.Permission.Log.Is(true))
	case eum.OperationStart:
		tx = tx.Where(query.Permission.Start.Is(true))
	case eum.OperationStop:
		tx = tx.Where(query.Permission.Stop.Is(true))
	case eum.OperationTerminal:
		tx = tx.Where(query.Permission.Terminal.Is(true))
	case eum.OperationTerminalWrite:
		tx = tx.Where(query.Permission.Write.Is(true))
	}
	tx.Scan(&result)
	return
}
