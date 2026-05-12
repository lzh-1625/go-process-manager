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

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Manage event push configurations",
	Long: `Manage configurations that push process or task events to external endpoints
(such as webhooks, IM bots, etc.).`,
	Example: `  gpm push list           # list all push configurations
  gpm push delete 1       # remove push configuration with ID 1`,
}

func init() {
	rootCmd.AddCommand(pushCmd)
	pushCmd.AddCommand(
		pushListCmd,
		pushDeleteCmd,
	)
}

var pushListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all push configurations",
	Long:  `Print a table of all event-push configurations, including their ID, type and target endpoint.`,
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(
			fx.NopLogger,
			app.Module,
			fx.Invoke(func(c *cli.PushCli) {
				err := c.GetList()
				if err != nil {
					log.Panic(err)
				}
				os.Exit(0)
			}),
		).Run()
	},
}

var pushDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a push configuration by ID",
	Long:  `Permanently remove the push configuration with the given ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(
			fx.NopLogger,
			app.Module,
			fx.Invoke(func(c *cli.PushCli) {
				err := c.Delete(utils.Unwarp(strconv.Atoi(args[0])))
				if err != nil {
					log.Panic(err)
				}
				fmt.Println("Push config deleted successfully")
				os.Exit(0)
			}),
		).Run()
	},
}
