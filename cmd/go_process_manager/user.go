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
	Short: "Manage gpm users",
	Long:  `Manage user accounts that can log in to the gpm web interface and API.`,
	Example: `  gpm user list                 # list all user accounts
  gpm user delete admin         # delete the user with account "admin"`,
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
	Short: "List all user accounts",
	Long:  `Print a table of all registered user accounts, including their account name and role.`,
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
	Short: "Delete a user by account name",
	Long:  `Permanently remove the user with the given account name from gpm.`,
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
