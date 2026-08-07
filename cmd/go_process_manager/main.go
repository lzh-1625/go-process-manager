package main

import (
	"github.com/lzh-1625/go_process_manager/internal/app"
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
----------------------------------------------------------------------------

   _______________        \::::\         /::::/     ___________________
  /:::::::::::::::\        \::::\       /::::/     /::::::::::::::::::/
 /:::::::::::::::::\        \::::\     /::::/      |:::::_____________/
 |:::::________\::::\        \::::\   /::::/       |::::|
 |::::|         |::::|        \::::\ /::::/        |::::|
 |::::|         |::::|         \:::::::::/         |::::|
 |::::|________/::::/           \:::::::/          |::::|____________
 |:::::::::::::::::/             \:::::/           |::::::::::::::::/
 |:::::________\::::\             |::::|           |:::::___________/
 |::::|         |::::|            |::::|           |::::|
 |::::|         |::::|            |::::|           |::::|
 |::::|         |::::|            |::::|           |::::|
 |::::|________/::::/             |::::|           |::::|_____________
 |:::::::::::::::::/              |::::|           |::::::::::::::::::/
 |::::::::::::::::/               |::::|           |:::::::::::::::::/
 |_______________/                |____|           |________________/

----------------------------------------------------------------------------
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
		print(startTitle)
		app.NewApp().Run()
		print(stopTitle)
	},
}
