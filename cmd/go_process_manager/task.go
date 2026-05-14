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
	Short: "Manage scheduled tasks",
	Long: `Manage scheduled and event-triggered tasks.

Tasks can be listed, deleted, manually started or stopped from the command line.`,
	Example: `  gpm task list           # list all tasks
  gpm task start 1        # manually trigger task with ID 1
  gpm task delete 1       # remove task with ID 1`,
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
	Long:  `Print a table of all registered tasks, including their ID, name, schedule and last run status.`,
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
	Short: "Delete a task by ID",
	Long:  `Permanently remove the task with the given ID from gpm.`,
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
	Short: "Manually trigger a task by ID",
	Long:  `Manually run the task with the given ID, ignoring its schedule.`,
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
	Short: "Stop a running task by ID",
	Long:  `Stop the currently running task with the given ID.`,
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
