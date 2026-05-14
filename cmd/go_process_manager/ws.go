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
	Short: "Manage WebSocket terminal share tokens",
	Long: `Manage share tokens for the WebSocket-based remote terminal.

Each token grants temporary, browser-accessible terminal access; use these
sub-commands to audit or revoke them.`,
	Example: `  gpm wsshare list        # list all active share tokens
  gpm wsshare delete 1    # revoke the share token with ID 1`,
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
	Short: "List all active share tokens",
	Long:  `Print a table of all active WebSocket terminal share tokens, including their ID and expiry.`,
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
	Short: "Revoke a share token by ID",
	Long:  `Immediately revoke the WebSocket terminal share token with the given ID.`,
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
