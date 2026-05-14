package main

import (
	"github.com/lzh-1625/go_process_manager/internal/app"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"go.uber.org/fx"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func main() {
	fx.New(
		app.Module,
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
}
