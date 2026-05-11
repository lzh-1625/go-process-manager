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

var wsshareCmd = &cobra.Command{
	Use:   "wsshare",
	Short: "WebSocket share management",
	Long:  `Manage WebSocket terminal share tokens.`,
}

func init() {
	rootCmd.AddCommand(wsshareCmd)
	wsshareCmd.AddCommand(
		wsshareListCmd,
		wsshareDeleteCmd,
	)
}

var wsshareListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all WebSocket shares",
	Long:  `List all active WebSocket terminal share tokens.`,
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(
			fx.NopLogger,
			app.Module,
			fx.Invoke(func(c *cli.WSShareCli) {
				err := c.GetList()
				if err != nil {
					log.Panic(err)
				}
				os.Exit(0)
			}),
		).Run()
	},
}

var wsshareDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete WebSocket share by ID",
	Long:  `Delete a WebSocket terminal share token by its ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(
			fx.NopLogger,
			app.Module,
			fx.Invoke(func(c *cli.WSShareCli) {
				err := c.Delete(utils.Unwarp(strconv.Atoi(args[0])))
				if err != nil {
					log.Panic(err)
				}
				fmt.Println("WebSocket share deleted successfully")
				os.Exit(0)
			}),
		).Run()
	},
}
