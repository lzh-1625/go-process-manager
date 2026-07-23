package repository

import (
	"errors"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/query"
	"github.com/lzh-1625/go_process_manager/internal/app/types"
	"gorm.io/gorm"
)

func NewPermissionRepository(query *query.Query) *PermissionRepository {
	return &PermissionRepository{
		query: query,
	}
}

type PermissionRepository struct {
	query *query.Query
}

func (p *PermissionRepository) GetPermssionList(account string) []model.PermissionPo {
	result := []model.PermissionPo{}
	proc := p.query.Process
	perm := p.query.Permission.As("p2")
	proc.Select(
		proc.Name,
		proc.UUID.As("pid"),
		perm.Owned, perm.Start, perm.Stop,
		perm.Terminal, perm.Log, perm.Write,
	).LeftJoin(perm,
		perm.Account.Eq(account),
		perm.Pid.EqCol(proc.UUID),
	).Scan(&result)
	return result
}

func (p *PermissionRepository) EditPermssion(data model.Permission) error {
	per := p.query.Permission
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

func (p *PermissionRepository) GetPermission(user string, pid int) (result *model.Permission) {
	result, _ = p.query.Permission.Where(p.query.Permission.Account.Eq(user), p.query.Permission.Pid.Eq(int32(pid))).First()
	return
}

func (p *PermissionRepository) GetProcessNameByPermission(user string, op types.OprPermission) (result []string) {
	tx := p.query.Permission.Select(p.query.Process.Name).RightJoin(p.query.Process, p.query.Process.UUID.EqCol(p.query.Permission.Pid)).Where(p.query.Permission.Account.Eq(user)).Where(p.query.Permission.Owned.Is(true))
	switch op {
	case types.OperationLog:
		tx = tx.Where(query.Permission.Log.Is(true))
	case types.OperationStart:
		tx = tx.Where(query.Permission.Start.Is(true))
	case types.OperationStop:
		tx = tx.Where(query.Permission.Stop.Is(true))
	case types.OperationTerminal:
		tx = tx.Where(query.Permission.Terminal.Is(true))
	case types.OperationTerminalWrite:
		tx = tx.Where(query.Permission.Write.Is(true))
	default:
		panic("unknown operation")
	}
	tx.Scan(&result)
	return
}
