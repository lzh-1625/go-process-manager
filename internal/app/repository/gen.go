//go:build gen
// +build gen

package repository

import (
	"os"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func gormGen(db *gorm.DB) {
	g := gen.NewGenerator(gen.Config{
		OutPath: "internal/app/repository/query",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
	})
	g.UseDB(db)
	g.ApplyBasic(
		&model.Process{},
		&model.User{},
		&model.Permission{},
		&model.Push{},
		&model.Config{},
		&model.ProcessLog{},
		&model.Task{},
		&model.WsShare{},
		&model.Event{},
	)
	g.Execute()
	os.Exit(0)
}
