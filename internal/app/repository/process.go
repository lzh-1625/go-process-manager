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
		return []model.Process{}
	}
	return result
}

func (p *processRepository) GetProcessConfigByUser(username string) []model.Process {
	result := []model.Process{}
	err := query.Permission.LeftJoin(query.Process, query.Process.Uuid.EqCol(query.Permission.Pid)).Where(query.Permission.Owned.Is(true)).Where(query.Permission.Account.Eq(username)).Scan(&result)
	if err != nil {
		log.Logger.Error(err)
		return []model.Process{}
	}
	return result
}

func (p *processRepository) UpdateProcessConfig(process model.Process) error {
	tx := db.Save(&process)
	return tx.Error
}

func (p *processRepository) AddProcessConfig(process model.Process) (int, error) {
	tx := db.Create(&process)
	if tx.Error != nil {
		log.Logger.Error(tx.Error)
		return 0, tx.Error
	}
	return process.Uuid, nil
}

func (p *processRepository) DeleteProcessConfig(uuid int) error {
	return db.Delete(&model.Process{
		Uuid: uuid,
	}).Error
}

func (p *processRepository) GetProcessConfigById(uuid int) model.Process {
	result := model.Process{}
	tx := db.Model(&model.Process{}).Where(&model.Process{Uuid: uuid}).First(&result)
	if tx.Error != nil {
		log.Logger.Error(tx.Error)
		return model.Process{}
	}
	return result
}
