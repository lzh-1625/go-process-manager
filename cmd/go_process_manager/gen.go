//go:build gen

package main

import (
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func init() {
	rootCmd.AddCommand(genCmd)
}

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate the repository code",
	Long:  `Generate the repository code`,
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(
			Module,
			fx.Invoke(func(db *gorm.DB) {
				g := gen.NewGenerator(gen.Config{
					OutPath: "internal/app/repository/query",
					Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
				})
				g.UseDB(db)
				g.ApplyBasic(repository.Tables...)
				g.Execute()
			}),
		).Run()
	},
}
