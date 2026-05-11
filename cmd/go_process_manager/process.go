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
	Short: "Process the process",
	Long:  `Process the process`,
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
	Short: "List the process",
	Long:  `List the process`,
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
	Use:   "exec",
	Short: "Exec the process",
	Long:  `Exec the process`,
	Args:  cobra.ExactArgs(1),
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
	Use:   "start",
	Short: "Start the process",
	Long:  `Start the process`,
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
	Use:   "stop",
	Short: "Stop the process",
	Long:  `Stop the process`,
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
