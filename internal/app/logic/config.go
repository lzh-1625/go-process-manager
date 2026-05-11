package logic

import (
	"log"
	"reflect"
	"strconv"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/utils"
)

func NewConfigLogic(configRepository *repository.ConfigRepository) *ConfigLogic {
	defer func() {
		if err := recover(); err != nil {
			log.Panic("config init fail", err)
		}
	}()
	configKvMap := map[string]string{}

	data, err := configRepository.GetAllConfig()
	if err != nil {
		panic(err)
	}
	for _, v := range data {
		configKvMap[v.Key] = *v.Value
	}

	typeElem := reflect.TypeFor[config.Configuration]()
	valueElem := reflect.ValueOf(config.CF).Elem()
	for i := 0; i < typeElem.NumField(); i++ {
		typeField := typeElem.Field(i)
		valueField := valueElem.Field(i)
		value, ok := configKvMap[typeField.Name]
		if !ok {
			value = typeField.Tag.Get("default")
		}
		if value == "-" {
			continue
		}
		switch typeField.Type.Kind() {
		case reflect.String:
			valueField.SetString(value)
		case reflect.Bool:
			valueField.SetBool(utils.UnwarpIgnore(strconv.ParseBool(value)))
		case reflect.Float64:
			valueField.SetFloat(utils.UnwarpIgnore(strconv.ParseFloat(value, 64)))
		case reflect.Int64, reflect.Int:
			valueField.SetInt(utils.UnwarpIgnore(strconv.ParseInt(value, 10, 64)))
		default:
			continue
		}
	}
	return &ConfigLogic{
		configRepository: configRepository,
	}
}

type ConfigLogic struct {
	configRepository *repository.ConfigRepository
}

func (c *ConfigLogic) GetSystemConfiguration() []model.SystemConfigurationVo {
	result := []model.SystemConfigurationVo{}
	typeElem := reflect.TypeFor[config.Configuration]()
	valueElem := reflect.ValueOf(config.CF).Elem()
	for i := 0; i < typeElem.NumField(); i++ {
		typeField := typeElem.Field(i)
		valueField := valueElem.Field(i)
		defaultValue := typeField.Tag.Get("default")
		if defaultValue == "-" {
			continue
		}
		var value any
		switch typeField.Type.Kind() {
		case reflect.Int64, reflect.Int:
			value = valueField.Int()
		case reflect.String:
			value = valueField.String()
		case reflect.Bool:
			value = valueField.Bool()
		case reflect.Float64:
			value = valueField.Float()
		default:
			continue
		}
		result = append(result, model.SystemConfigurationVo{
			Key:      typeField.Name,
			Value:    value,
			Default:  defaultValue,
			Describe: typeField.Tag.Get("describe"),
		})
	}
	return result
}

func (c *ConfigLogic) SetSystemConfiguration(kv map[string]string) error {
	typeElem := reflect.TypeFor[config.Configuration]()
	valueElem := reflect.ValueOf(config.CF).Elem()
	for i := 0; i < typeElem.NumField(); i++ {
		typeField := typeElem.Field(i)
		valueField := valueElem.Field(i)
		for k, v := range kv {
			if typeField.Name == k {
				var err error
				switch typeField.Type.Kind() {
				case reflect.String:
					valueField.SetString(v)
				case reflect.Bool:
					value, errV := strconv.ParseBool(v)
					err = errV
					if err == nil {
						valueField.SetBool(value)
					}
				case reflect.Float64:
					value, errV := strconv.ParseFloat(v, 64)
					err = errV
					if err == nil {
						valueField.SetFloat(value)
					}
				case reflect.Int64, reflect.Int:
					value, errV := strconv.ParseInt(v, 10, 64)
					err = errV
					if err == nil {
						valueField.SetInt(value)
					}
				default:
					continue
				}
				if err != nil {
					return err
				}
				err = c.configRepository.SetConfigValue(k, v)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// reset system config to default
func (c *ConfigLogic) ResetSystemConfiguration() error {
	typeElem := reflect.TypeFor[config.Configuration]()
	valueElem := reflect.ValueOf(config.CF).Elem()
	for i := 0; i < typeElem.NumField(); i++ {
		typeField := typeElem.Field(i)
		valueField := valueElem.Field(i)
		var err error
		defaultValue := typeField.Tag.Get("default")
		if defaultValue == "-" {
			continue
		}
		switch typeField.Type.Kind() {
		case reflect.String:
			valueField.SetString(defaultValue)
		case reflect.Bool:
			value, errV := strconv.ParseBool(defaultValue)
			err = errV
			if err == nil {
				valueField.SetBool(value)
			}
		case reflect.Float64:
			value, errV := strconv.ParseFloat(defaultValue, 64)
			err = errV
			if err == nil {
				valueField.SetFloat(value)
			}
		case reflect.Int64, reflect.Int:
			value, errV := strconv.ParseInt(defaultValue, 10, 64)
			err = errV
			if err == nil {
				valueField.SetInt(value)
			}
		default:
			continue
		}
		if err != nil {
			return err
		}
		err = c.configRepository.SetConfigValue(typeField.Name, defaultValue)
		if err != nil {
			return err
		}

	}
	return nil
}
