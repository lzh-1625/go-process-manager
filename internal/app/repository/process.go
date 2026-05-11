package repository

import (
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/query"
	"github.com/lzh-1625/go_process_manager/log"
)

func NewProcessRepository() *ProcessRepository {
	return &ProcessRepository{
		query: query.Q,
	}
}

type ProcessRepository struct {
	query *query.Query
}

func (p *ProcessRepository) GetAllProcessConfig() []*model.Process {
	processes, _ := p.query.Process.Find()
	return processes
}

func (p *ProcessRepository) GetProcessConfigByUser(username string) []*model.Process {
	result := []*model.Process{}
	err := p.query.Process.LeftJoin(p.query.Permission, p.query.Process.UUID.EqCol(p.query.Permission.Pid)).
		Where(p.query.Permission.Owned.Is(true)).
		Where(p.query.Permission.Account.Eq(username)).
		Scan(&result)
	if err != nil {
		log.Logger.Error(err)
		return nil
	}
	return result
}

func (p *ProcessRepository) UpdateProcessConfig(process model.Process) error {
	return p.query.Process.Save(&process)
}

func (p *ProcessRepository) AddProcessConfig(process model.Process) (id int, err error) {
	err = p.query.Process.Create(&process)
	id = process.UUID
	return
}

func (p *ProcessRepository) DeleteProcessConfig(uuid int) error {
	_, err := p.query.Process.Where(p.query.Process.UUID.Eq(uuid)).Delete()
	return err
}

func (p *ProcessRepository) GetProcessConfigByID(uuid int) (data *model.Process, err error) {
	data, err = p.query.Process.Where(p.query.Process.UUID.Eq(uuid)).First()
	return
}
