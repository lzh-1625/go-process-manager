package log

import (
	"log"

	"github.com/lzh-1625/go_process_manager/config"

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
		OutputPaths:      []string{"stdout", "info.log"},
		ErrorOutputPaths: []string{"stderr"},
	}
	log, _ := config.Build()
	Logger = log.Sugar()
}
