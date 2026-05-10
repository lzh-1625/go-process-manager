package boot

import (
	"log"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"syscall"
	"time"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	logger "github.com/lzh-1625/go_process_manager/log"
	"github.com/lzh-1625/go_process_manager/utils"
	"github.com/robfig/cron/v3"
)

func Boot() {
	initDb()
	initResetConfig()
	initConfiguration()
	initLogger()
	initLogHandler()
	initProcess()
	initJwtSecret()
	InitTask()
	InitEventCleanCronJob()
	initListenKillSignal()
}

func initDb() {
	repository.InitDb()
}

func initConfiguration() {
	defer func() {
		if err := recover(); err != nil {
			log.Panic("config init fail", err)
		}
	}()
	configKvMap := map[string]string{}

	data, err := repository.ConfigRepository.GetAllConfig()
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
}

func initProcess() {
	logic.ProcessCtlLogic.ProcessInit()
}

func initJwtSecret() {
	if secret, err := repository.ConfigRepository.GetConfigValue(eum.SecretKey); err == nil {
		utils.SetSecret([]byte(secret))
		return
	}
	secret := utils.RandString(32)
	repository.ConfigRepository.SetConfigValue(eum.SecretKey, secret)
	utils.SetSecret([]byte(secret))
}

func initLogHandler() {
	logic.InitLog()
	logic.InitLogHandle()
}

func initLogger() {
	logger.InitLog()
}

func InitTask() {
	logic.TaskLogic.InitTaskJob()
}

func initResetConfig() {
	if len(os.Args) >= 2 && os.Args[1] == "reset" {
		err := logic.ConfigLogic.ResetSystemConfiguration()
		if err != nil {
			log.Panic(err)
		}
		log.Print("reset system config to deafult success!")
		os.Exit(0)
	}
}

func initListenKillSignal() {
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
		logger.Logger.Info("process is exiting, waiting for all processes to stop")
		logic.ProcessCtlLogic.KillAllProcess()
		log.Print("all processes have been stopped")
		os.Exit(0)
	}()
}

func InitEventCleanCronJob() {
	if config.CF.EventStorageTime <= 0 {
		return
	}
	c := cron.New()
	c.AddFunc("0 3 * * *", func() {
		logger.Logger.Infow("event cleaning execution")
		logic.EventLogic.Clean(time.Duration(config.CF.EventStorageTime) * time.Hour * 24)
	})
	c.Start()
}
