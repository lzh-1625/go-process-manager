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
	Short: "Push configuration management",
	Long:  `Manage push configurations including list, delete.`,
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
	Long:  `List all push configurations.`,
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
	Short: "Delete push config by ID",
	Long:  `Delete a push configuration by its ID.`,
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
