package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/api"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/middle"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/internal/app/route"
	"github.com/lzh-1625/go_process_manager/log"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd.AddCommand(runCmd)
	rootCmd.Execute()
}

var startTitle = `
----------------------------------------------------------------------------
          _____                    _____                    _____          
         /\    \                  /\    \                  /\    \         
        /::\    \                /::\    \                /::\____\        
       /::::\    \              /::::\    \              /::::|   |        
      /::::::\    \            /::::::\    \            /:::::|   |        
     /:::/\:::\    \          /:::/\:::\    \          /::::::|   |        
    /:::/  \:::\    \        /:::/__\:::\    \        /:::/|::|   |        
   /:::/    \:::\    \      /::::\   \:::\    \      /:::/ |::|   |        
  /:::/    / \:::\    \    /::::::\   \:::\    \    /:::/  |::|___|______  
 /:::/    /   \:::\ ___\  /:::/\:::\   \:::\____\  /:::/   |::::::::\    \ 
/:::/____/  ___\:::|    |/:::/  \:::\   \:::|    |/:::/    |:::::::::\____\
\:::\    \ /\  /:::|____|\::/    \:::\  /:::|____|\::/    / ~~~~~/:::/    /
 \:::\    /::\ \::/    /  \/_____/\:::\/:::/    /  \/____/      /:::/    / 
  \:::\   \:::\ \/____/            \::::::/    /               /:::/    /  
   \:::\   \:::\____\               \::::/    /               /:::/    /   
    \:::\  /:::/    /                \::/____/               /:::/    /    
     \:::\/:::/    /                  ~~                    /:::/    /     
      \::::::/    /                                        /:::/    /      
       \::::/    /                                        /:::/    /       
        \::/____/                                         \::/    /        
                                                           \/____/         
----------------------------------------------------------------------------  
`
var stopTitle = `
----------------------------------------
 ________      ___    ___ _______      
|\   __  \    |\  \  /  /|\  ___ \     
\ \  \|\ /_   \ \  \/  / | \   __/|    
 \ \   __  \   \ \    / / \ \  \_|/__  
  \ \  \|\  \   \/  /  /   \ \  \_|\ \ 
   \ \_______\__/  / /      \ \_______\
    \|_______|\___/ /        \|_______|
             \|___|/                   
                                       
                                       
----------------------------------------
`
var rootCmd = &cobra.Command{
	Use:   "gpm",
	Short: "Go Process Manager - a lightweight process supervisor",
	Long: `Go Process Manager (gpm) is a lightweight, cross-platform process supervisor.

It can start, stop, restart and monitor long-running processes, run scheduled
tasks, push events to external endpoints and expose a web API for remote
management. Run "gpm" with no sub-command to start the server in the foreground.`,
	Example: `  # Start the gpm server in the foreground
  gpm run

  # Install gpm as a system service and start it
  gpm service install
  gpm service start

  # List all managed processes
  gpm process list`,
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run gpm in the foreground",
	Long:  `Run gpm in the foreground.`,
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

func run() {
	db := repository.NewDB()
	query := repository.NewQuery(db)
	processRepository := repository.NewProcessRepository(query)
	userRepository := repository.NewUserRepository(query)
	pushRepository := repository.NewPushRepository(query)
	eventRepository := repository.NewEventRepository(query)
	permissionRepository := repository.NewPermissionRepository(query)
	logRepository := repository.NewLogRepository(query)
	wsShareRepository := repository.NewWsShareRepository(query)
	taskRepository := repository.NewTaskRepository(query)

	logLogic := logic.NewILogLogic(logRepository)
	eventBus := logic.NewEventBus()
	logHandler := logic.NewLogHandler(logLogic)
	eventLogic := logic.NewEventLogic(eventRepository)
	permissionLogic := logic.NewPermissionLogic(permissionRepository)
	pushLogic := logic.NewPushLogic(pushRepository)
	processCtlLogic := logic.NewProcessCtlLogic(processRepository, permissionRepository, eventLogic, pushLogic, logHandler, eventBus)
	taskLogic := logic.NewTaskLogic(taskRepository, eventLogic, processCtlLogic, eventBus)
	userLogic := logic.NewUserLogic(userRepository)
	wsShareLogic := logic.NewWsShareLogic(wsShareRepository)
	configLogic := logic.NewConfigLogic()
	metricLogic := logic.NewMetricLogic(processCtlLogic, logHandler, logLogic)

	permissionAPI := api.NewPermissionApi(permissionLogic)
	wsAPI := api.NewWsApi(processCtlLogic, wsShareLogic, eventLogic, permissionAPI)
	procAPI := api.NewProcApi(processCtlLogic, wsShareLogic, permissionAPI)
	taskAPI := api.NewTaskApi(taskLogic)
	userAPI := api.NewUserApi(userLogic)
	pushAPI := api.NewPushApi(pushLogic)
	eventAPI := api.NewEventApi(eventLogic)
	logAPI := api.NewLogApi(permissionLogic, logLogic)
	configAPI := api.NewConfigApi(configLogic, logLogic)
	metricAPI := api.NewMetricApi(metricLogic)
	loggerMiddleware := middle.NewEventLoggerMiddleware(eventLogic)
	authMiddleware := middle.NewAuthMiddleware(userLogic)
	r := route.NewRoute(wsAPI, procAPI, taskAPI, userAPI, pushAPI, eventAPI, permissionAPI, logAPI, configAPI, metricAPI, loggerMiddleware, authMiddleware)
	server := &http.Server{
		Addr:        config.CF.Listen,
		Handler:     r,
		ErrorLog:    slog.NewLogLogger(r.Logger.Handler(), slog.LevelError),
		ReadTimeout: 30 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	serverErr := make(chan error, 1)
	print(startTitle)
	go func() {
		log.Logger.Infow("starting echo server", "listen", config.CF.Listen)
		serverErr <- server.ListenAndServe()
	}()

	c := cron.New()
	// event cleaning cron job
	if config.CF.EventStorageTime >= 0 {
		c.AddFunc("0 3 * * *", func() {
			log.Logger.Infow("event cleaning execution")
			eventLogic.Clean(time.Duration(config.CF.EventStorageTime) * time.Hour * 24)
		})
		c.Start()
	}
	processCtlLogic.ProcessInit()
	taskLogic.InitTaskJob()
	go func() {
		// run task by trigger event
		for event := range eventBus.Subscribe() {
			go taskLogic.RunTaskByTriggerEvent(event.Proc.Name, event.State)
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Logger.Errorw("echo server stopped", "err", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second+time.Duration(config.CF.KillWaitTime)*time.Second)
	defer cancel()
	cleanupDone := make(chan struct{})
	go func() {
		defer close(cleanupDone)
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Logger.Errorw("shutdown echo server failed", "err", err)
		}
		c.Stop()
		log.Logger.Infow("waiting for all process to stop")
		processCtlLogic.KillAllProcess()
		logHandler.Close()
		eventBus.Close()
		print(stopTitle)
	}()

	select {
	case <-cleanupDone:
	case <-shutdownCtx.Done():
		log.Logger.Errorw("shutdown timed out", "err", shutdownCtx.Err())
		if err := server.Close(); err != nil {
			log.Logger.Errorw("force close echo server failed", "err", err)
		}
	}
}
