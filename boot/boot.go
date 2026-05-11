package boot

func initConfiguration() {

}

func initProcess() {
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
