package repository

import (
	"log"
	"os"
	"time"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/query"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db            *gorm.DB
	defaultConfig = gorm.Session{PrepareStmt: true, SkipDefaultTransaction: true}
	tables        = []any{
		&model.Process{},
		&model.User{},
		&model.Permission{},
		&model.Push{},
		&model.Config{},
		&model.ProcessLog{},
		&model.Task{},
	}
)

func InitDb() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  true,
		},
	)
	gdb, err := gorm.Open(sqlite.Open("data.db"), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Panicf("sqlite database init failed! \nerror: %v", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		log.Panicf("sqlite database init failed! \nerror: %v", err)
	}
	sqlDB.SetConnMaxLifetime(time.Hour)
	db = gdb.Session(&defaultConfig)
	// if config.CF.LogLevel == "debug" {
	db = db.Debug()
	// }
	err = db.AutoMigrate(tables...)
	if err != nil {
		log.Panicf("database migrate failed! \nerror: %v", err)
	}
	gormGen(db)
	query.SetDefault(db)
}
