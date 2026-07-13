package main

import (
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"gorm.io/gen"
)

func main() {
	db := repository.NewDB()
	g := gen.NewGenerator(gen.Config{
		OutPath: "internal/app/repository/query",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
	})
	g.UseDB(db)
	g.ApplyBasic(repository.Tables...)
	g.Execute()
}
