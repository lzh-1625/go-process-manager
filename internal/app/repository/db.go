package repository

import (
	"context"
	"os"
	"path"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/query"
	"github.com/lzh-1625/go_process_manager/log"
	"go.uber.org/zap"
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

type GormLogger struct {
	logger *zap.SugaredLogger
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...any) {
	l.logger.Infow(msg, data...)
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...any) {
	l.logger.Warnw(msg, data...)
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...any) {
	l.logger.Errorw(msg, data...)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, rows := fc()
	l.logger.Debugw(sql, "rows", rows, "time", time.Since(begin), "err", err)
}

func NewDB() *gorm.DB {
	home, _ := os.UserHomeDir()

	gdb, err := gorm.Open(sqlite.Open(path.Join(home, ".gpm", "data.db")), &gorm.Config{
		Logger: &GormLogger{logger: log.Logger},
	})
	if err != nil {
		log.Logger.Panicf("sqlite database init failed! \nerror: %v", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		log.Logger.Panicf("sqlite database init failed! \nerror: %v", err)
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
