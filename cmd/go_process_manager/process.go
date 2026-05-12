package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/lzh-1625/go_process_manager/internal/app"
	"github.com/lzh-1625/go_process_manager/internal/app/cli"
	"github.com/lzh-1625/go_process_manager/utils"
	"github.com/spf13/cobra"

	"go.uber.org/fx"
)

var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Manage managed processes",
	Long: `Inspect and control processes managed by gpm.

Sub-commands let you list all managed processes, run a one-shot execution,
or start/stop a specific process by its ID.`,
	Example: `  gpm process list         # show all managed processes
  gpm process start 1      # start the process with ID 1
  gpm process stop 1       # stop the process with ID 1`,
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
		fx.New(
			fx.NopLogger,
			app.Module,
			// register sqlite implement search engine
			fx.Invoke(func(cli *cli.ProcessCli) {
				err := cli.GetList()
				if err != nil {
					log.Panic(err)
				}
				os.Exit(0)
			}),
		).Run()
	},
}

var processExecCmd = &cobra.Command{
	Use:   "exec [id]",
	Short: "Execute a process once in the foreground",
	Long: `Run the process with the given ID once in the foreground and stream
its output to the current terminal. The process is not kept under supervision.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(
			fx.NopLogger,
			app.Module,
			fx.Invoke(func(cli *cli.ProcessCli) {
				err := cli.Exec(utils.Unwarp(strconv.Atoi(args[0])))
				if err != nil {
					fmt.Print(err)
				}
				os.Exit(0)
			}),
		).Run()
	},
}

var processStartCmd = &cobra.Command{
	Use:   "start [id]",
	Short: "Start a managed process by ID",
	Long:  `Start the managed process with the given ID and keep it under gpm supervision.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(
			fx.NopLogger,
			app.Module,
			fx.Invoke(func(cli *cli.ProcessCli) {
				err := cli.Start(utils.Unwarp(strconv.Atoi(args[0])))
				if err != nil {
					fmt.Print(err)
				}
				os.Exit(0)
			}),
		).Run()
	},
}

var processStopCmd = &cobra.Command{
	Use:   "stop [id]",
	Short: "Stop a managed process by ID",
	Long:  `Gracefully stop the managed process with the given ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(
			fx.NopLogger,
			app.Module,
			fx.Invoke(func(cli *cli.ProcessCli) {
				err := cli.Stop(utils.Unwarp(strconv.Atoi(args[0])))
				if err != nil {
					fmt.Print(err)
				}
				os.Exit(0)
			}),
		).Run()
	},
}
