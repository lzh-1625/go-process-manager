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

var rootCmd = &cobra.Command{
	Use:   "gpm",
	Short: "Go Process Manager",
	Long:  `Go Process Manager is a tool for managing processes.`,
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(
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
								taskLogic.RunTaskByTriggerEvent(event.Proc.Name, event.State)
							}
						}()
						go r.Start(config.CF.Listen)
						return nil
					},
					OnStop: func(ctx context.Context) error {
						c.Stop()
						log.Logger.Infow("stop all process")
						processCtlLogic.KillAllProcess()
						eventBus.Close()
						return nil
					},
				})
			}),
		).Run()
	},
}
