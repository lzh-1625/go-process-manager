//go:build gen

package repository

import (
	"os"

	"gorm.io/gen"
	"gorm.io/gorm"
)

func gormGen(db *gorm.DB) {
	g := gen.NewGenerator(gen.Config{
		OutPath: "internal/app/repository/query",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
	})
	g.UseDB(db)
	g.ApplyBasic(tables...)
	g.Execute()
	os.Exit(0)
}
