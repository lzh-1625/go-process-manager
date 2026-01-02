package repository

import (
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/query"
	"github.com/lzh-1625/go_process_manager/log"
)

type processRepository struct{}

var ProcessRepository = new(processRepository)

func (p *processRepository) GetAllProcessConfig() []*model.Process {
	processes, _ := query.Process.Find()
	return processes
}

func (p *processRepository) GetProcessConfigByUser(username string) []*model.Process {
	result := []*model.Process{}
	err := query.Process.LeftJoin(query.Permission, query.Process.UUID.EqCol(query.Permission.Pid)).
		Where(query.Permission.Owned.Is(true)).
		Where(query.Permission.Account.Eq(username)).
		Scan(&result)
	if err != nil {
		log.Logger.Error(err)
		return nil
	}
	return result
}

func (p *processRepository) UpdateProcessConfig(process model.Process) error {
	return query.Process.Save(&process)
}

func (p *processRepository) AddProcessConfig(process model.Process) (id int, err error) {
	err = query.Process.Create(&process)
	id = process.UUID
	return
}

func (p *processRepository) DeleteProcessConfig(uuid int) error {
	_, err := query.Process.Where(query.Process.UUID.Eq(uuid)).Delete()
	return err
}

func (p *processRepository) GetProcessConfigById(uuid int) (data *model.Process, err error) {
	data, err = query.Process.Where(query.Process.UUID.Eq(uuid)).First()
	return
}
