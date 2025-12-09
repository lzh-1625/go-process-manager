package repository

import (
	"errors"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/query"
)

type configRepository struct{}

var ConfigRepository = new(configRepository)

func (c *configRepository) GetConfigValue(key string) (string, error) {
	data, err := query.Config.Select(query.Config.Value).Where(query.Config.Key.Eq(key)).First()
	if data == nil || err != nil {
		return "", errors.ErrUnsupported
	}
	return *data.Value, err
}

func (c *configRepository) GetAllConfig() ([]*model.Config, error) {
	data, err := query.Config.Select(query.Config.Key, query.Config.Value).Find()
	return data, err
}

func (c *configRepository) SetConfigValue(key, value string) error {
	config := model.Config{Key: key}
	updateData := model.Config{Value: &value}
	err := db.Model(&config).Where(&config).Assign(updateData).FirstOrCreate(&config).Error
	if err != nil {
		return err
	}
	return nil
}
