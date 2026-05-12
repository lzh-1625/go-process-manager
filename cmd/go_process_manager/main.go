package main

import (
	"context"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/log"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

func main() {
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
  gpm

  # Install gpm as a system service and start it
  gpm service install
  gpm service start

  # List all managed processes
  gpm process list`,
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(
			fx.NopLogger,
			app.Module,
			fx.Invoke(func(
				r *echo.Echo,
				lc fx.Lifecycle,
				processCtlLogic *logic.ProcessCtlLogic,
				taskLogic *logic.TaskLogic,
				eventLogic *logic.EventLogic,
				eventBus *logic.EventBus,
			) {
				c := cron.New()
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						go func() {
							log.Logger.Infow("starting echo server", "listen", config.CF.Listen)
							err := r.Start(config.CF.Listen)
							if err != nil {
								log.Logger.Panicw("start echo server failed", "err", err)
							}
						}()
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
						return nil
					},
					OnStop: func(ctx context.Context) error {
						c.Stop()
						log.Logger.Infow("waiting for all process to stop")
						processCtlLogic.KillAllProcess()
						eventBus.Close()
						print(stopTitle)
						return nil
					},
				})
			}),
		).Run()
	},
}
