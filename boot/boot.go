package boot

import (
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	logger "github.com/lzh-1625/go_process_manager/log"
	"github.com/lzh-1625/go_process_manager/utils"
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

}

func initProcess() {
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

func initLogger() {
	logger.InitLog()
}

func InitTask() {
}

func initResetConfig() {
	// if len(os.Args) >= 2 && os.Args[1] == "reset" {
	// 	err := logic.ConfigLogic.ResetSystemConfiguration()
	// 	if err != nil {
	// 		log.Panic(err)
	// 	}
	// 	log.Print("reset system config to deafult success!")
	// 	os.Exit(0)
	// }
}

func initListenKillSignal() {
	// go func() {
	// 	sigs := make(chan os.Signal, 1)
	// 	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// 	<-sigs
	// 	logger.Logger.Info("process is exiting, waiting for all processes to stop")
	// 	logic.ProcessCtlLogic.KillAllProcess()
	// 	log.Print("all processes have been stopped")
	// 	os.Exit(0)
	// }()
}

func InitEventCleanCronJob() {
	// if config.CF.EventStorageTime <= 0 {
	// 	return
	// }
	// c := cron.New()
	// c.AddFunc("0 3 * * *", func() {
	// 	logger.Logger.Infow("event cleaning execution")
	// 	logic.EventLogic.Clean(time.Duration(config.CF.EventStorageTime) * time.Hour * 24)
	// })
	// c.Start()
}
