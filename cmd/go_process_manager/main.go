package main

import (
	"context"
	"log"

	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/search"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/search/sqlite"
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
			// register sqlite implement search engine
			fx.Invoke(func(logRepository *repository.LogRepository) {
				search.Register("sqlite", sqlite.NewSqliteSearch(logRepository))
			}),

			fx.Invoke(func(r *echo.Echo, lc fx.Lifecycle, processCtlLogic *logic.ProcessCtlLogic, taskLogic *logic.TaskLogic) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						processCtlLogic.ProcessInit()
						taskLogic.InitTaskJob()
						go r.Start(config.CF.Listen)
						return nil
					},
					OnStop: func(ctx context.Context) error {
						log.Println("stop all process")
						processCtlLogic.KillAllProcess()
						return nil
					},
				})
			}),
		).Run()
	},
}
