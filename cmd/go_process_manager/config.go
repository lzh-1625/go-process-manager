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
	Short: "Config the config",
	Long:  `Config the config`,
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the config",
	Long:  `Reset the config`,
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
