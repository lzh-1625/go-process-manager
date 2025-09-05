package repository

import (
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/query"
	"github.com/lzh-1625/go_process_manager/log"
)

type processRepository struct{}

var ProcessRepository = new(processRepository)

func (p *processRepository) GetAllProcessConfig() []model.Process {
	result := []model.Process{}
	tx := db.Find(&result)
	if tx.Error != nil {
		log.Logger.Error(tx.Error)
		return nil
	}
	return result
}

func (p *processRepository) GetProcessConfigByUser(username string) []model.Process {
	result := []model.Process{}
	err := query.Process.LeftJoin(query.Permission, query.Process.Uuid.EqCol(query.Permission.Pid)).
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
	return db.Save(&process).Error
}

func (p *processRepository) AddProcessConfig(process model.Process) (id int, err error) {
	err = db.Create(&process).Error
	id = process.Uuid
	return
}

func (p *processRepository) DeleteProcessConfig(uuid int) error {
	_, err := query.Process.Where(query.Process.Uuid.Eq(uuid)).Delete()
	return err
}

func (p *processRepository) GetProcessConfigById(uuid int) (data model.Process, err error) {
	err = db.Model(&model.Process{}).Where(&model.Process{Uuid: uuid}).First(&data).Error
	return
}
