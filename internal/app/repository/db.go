package repository

import (
	"log"
	"os"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/query"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	defaultConfig = gorm.Session{PrepareStmt: true, SkipDefaultTransaction: true}
	Tables        = []any{
		&model.Process{},
		&model.User{},
		&model.Permission{},
		&model.Push{},
		&model.ProcessLog{},
		&model.Event{},
		&model.Task{},
		&model.WsShare{},
	}
)

func NewDB() *gorm.DB {
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
	db := gdb.Session(&defaultConfig)
	if config.CF.LogLevel == "debug" {
		db = db.Debug()
	}
	err = db.AutoMigrate(Tables...)
	return db
}

func NewQuery(db *gorm.DB) *query.Query {
	query.SetDefault(db)
	return query.Use(db)
}
