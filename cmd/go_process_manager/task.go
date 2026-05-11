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

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Task management",
	Long:  `Manage tasks including list, delete, start, stop.`,
}

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(
		taskListCmd,
		taskDeleteCmd,
		taskStartCmd,
		taskStopCmd,
	)
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	Long:  `List all tasks with details.`,
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(
			fx.NopLogger,
			app.Module,
			fx.Invoke(func(c *cli.TaskCli) {
				err := c.GetList()
				if err != nil {
					log.Panic(err)
				}
				os.Exit(0)
			}),
		).Run()
	},
}

var taskDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete task by ID",
	Long:  `Delete a task by its ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(
			fx.NopLogger,
			app.Module,
			fx.Invoke(func(c *cli.TaskCli) {
				err := c.Delete(utils.Unwarp(strconv.Atoi(args[0])))
				if err != nil {
					log.Panic(err)
				}
				fmt.Println("Task deleted successfully")
				os.Exit(0)
			}),
		).Run()
	},
}

var taskStartCmd = &cobra.Command{
	Use:   "start [id]",
	Short: "Start task by ID",
	Long:  `Start a task manually by its ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(
			fx.NopLogger,
			app.Module,
			fx.Invoke(func(c *cli.TaskCli) {
				err := c.Start(utils.Unwarp(strconv.Atoi(args[0])))
				if err != nil {
					log.Panic(err)
				}
				fmt.Println("Task started successfully")
				os.Exit(0)
			}),
		).Run()
	},
}

var taskStopCmd = &cobra.Command{
	Use:   "stop [id]",
	Short: "Stop task by ID",
	Long:  `Stop a task by its ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(
			fx.NopLogger,
			app.Module,
			fx.Invoke(func(c *cli.TaskCli) {
				err := c.Stop(utils.Unwarp(strconv.Atoi(args[0])))
				if err != nil {
					log.Panic(err)
				}
				fmt.Println("Task stopped successfully")
				os.Exit(0)
			}),
		).Run()
	},
}
