package repository

import (
	"errors"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/query"
	"gorm.io/gorm/clause"
)

type ConfigRepository struct {
	query *query.Query
}

func NewConfigRepository() *ConfigRepository {
	return &ConfigRepository{
		query: query.Q,
	}
}

func (c *ConfigRepository) GetConfigValue(key string) (string, error) {
	data, err := c.query.Config.Select(query.Config.Value).Where(query.Config.Key.Eq(key)).First()
	if data == nil || err != nil {
		return "", errors.ErrUnsupported
	}
	return *data.Value, err
}

func (c *ConfigRepository) GetAllConfig() ([]*model.Config, error) {
	data, err := c.query.Config.Select(query.Config.Key, query.Config.Value).Find()
	return data, err
}

func (c *ConfigRepository) SetConfigValue(key, value string) error {
	config := model.Config{Key: key, Value: &value}
	return c.query.Config.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: query.Config.Key.ColumnName().String()}},
		DoUpdates: clause.AssignmentColumns([]string{query.Config.Value.ColumnName().String()}),
	}).Create(&config)
}
