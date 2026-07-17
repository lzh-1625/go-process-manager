package main

import (
	"github.com/lzh-1625/go_process_manager/internal/app/cli"
	"github.com/spf13/cobra"
)

var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Manage managed processes",
	Long: `Inspect and control processes managed by gpm.

Sub-commands let you list all managed processes, run a one-shot execution,
or start/stop a specific process by name.`,
Example: `  gpm process list             # show all managed processes
  gpm process start my-app     # start the process named my-app
  gpm process stop my-app      # stop the process named my-app`,
}

func init() {
	rootCmd.AddCommand(processCmd)
}

func init() {
	processCmd.AddCommand(processListCmd)
	processCmd.AddCommand(processExecCmd)
	processCmd.AddCommand(processStartCmd)
	processCmd.AddCommand(processStopCmd)
}

var processListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all managed processes",
	Long:  `Print a table of all processes managed by gpm, including their ID, name and current state.`,
	Run: func(cmd *cobra.Command, args []string) {
		cli.NewProcessCli().GetList()
	},
}

var processExecCmd = &cobra.Command{
	Use:   "exec [name]",
	Short: "Execute a process once in the foreground",
	Long: `Run the process with the given name once in the foreground and stream
its output to the current terminal. The process is not kept under supervision.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli.NewProcessCli().Exec(args[0])
	},
}

var processStartCmd = &cobra.Command{
	Use:   "start [name]",
	Short: "Start a managed process by name",
	Long:  `Start the managed process with the given name and keep it under gpm supervision.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli.NewProcessCli().Start(args[0])
	},
}

var processStopCmd = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stop a managed process by name",
	Long:  `Gracefully stop the managed process with the given name.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli.NewProcessCli().Stop(args[0])
	},
}
