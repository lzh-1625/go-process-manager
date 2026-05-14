package main

import (
	"log"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configResetCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage gpm configuration",
	Long:  `Inspect and modify the gpm configuration file.`,
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to default values",
	Long: `Reset the gpm configuration file to its default values.

Warning: this overwrites any customisations you have made to the config file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.ResetConfig(); err != nil {
			log.Panic(err)
		}
		if err := config.DumpConfig(); err != nil {
			log.Panic(err)
		}
		log.Println("config reset success")
	},
}
