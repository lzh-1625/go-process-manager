package log

import (
	"log"
	"path"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/pkg/process"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.SugaredLogger

func init() {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	level, err := zapcore.ParseLevel(config.CF.LogLevel)
	if err != nil {
		log.Printf("log level error! level [%v] not exist", config.CF.LogLevel)
		level = zap.DebugLevel
	}
	atom := zap.NewAtomicLevelAt(level)
	config := zap.Config{
		Level:            atom,
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout", path.Join(config.CF.ConfigDir, "info.log")},
		ErrorOutputPaths: []string{"stderr"},
	}
	log, _ := config.Build()
	Logger = log.Sugar()
	process.SetLogger(processLoggerImpl{Logger: Logger})
}

type processLoggerImpl struct {
	Logger *zap.SugaredLogger
}

func (p processLoggerImpl) Debug(msg string, keysAndValues ...any) {
	p.Logger.Debugw(msg, keysAndValues...)
}
func (p processLoggerImpl) Info(msg string, keysAndValues ...any) {
	p.Logger.Infow(msg, keysAndValues...)
}
func (p processLoggerImpl) Warn(msg string, keysAndValues ...any) {
	p.Logger.Warnw(msg, keysAndValues...)
}
func (p processLoggerImpl) Error(msg string, keysAndValues ...any) {
	p.Logger.Errorw(msg, keysAndValues...)
}
