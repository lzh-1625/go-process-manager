package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = ""

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the version of gpm",
	Long:  "Show the version of gpm",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	},
}
