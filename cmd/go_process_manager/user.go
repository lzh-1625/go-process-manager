package main

import (
	"fmt"
	"log"
	"os"

	"github.com/lzh-1625/go_process_manager/internal/app"
	"github.com/lzh-1625/go_process_manager/internal/app/cli"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "User management",
	Long:  `Manage users including list, delete.`,
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(
		userListCmd,
		userDeleteCmd,
	)
}

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	Long:  `List all users with details.`,
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(
			fx.NopLogger,
			app.Module,
			fx.Invoke(func(c *cli.UserCli) {
				err := c.GetList()
				if err != nil {
					log.Panic(err)
				}
				os.Exit(0)
			}),
		).Run()
	},
}

var userDeleteCmd = &cobra.Command{
	Use:   "delete [account]",
	Short: "Delete user by account",
	Long:  `Delete a user by account name.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(
			fx.NopLogger,
			app.Module,
			fx.Invoke(func(c *cli.UserCli) {
				err := c.Delete(args[0])
				if err != nil {
					log.Panic(err)
				}
				fmt.Println("User deleted successfully")
				os.Exit(0)
			}),
		).Run()
	},
}
