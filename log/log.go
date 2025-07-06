package log

import (
	"log"

	"github.com/lzh-1625/go_process_manager/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.SugaredLogger

func InitLog() {
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
		log.Printf("日志等级错误！不存在“%v”日志等级", config.CF.LogLevel)
		level = zap.DebugLevel
	}
	atom := zap.NewAtomicLevelAt(level)
	zap.NewDevelopmentConfig()
	var outputPaths []string = []string{"info.log"}
	if !config.CF.Tui { // 不使用tui则打印日志到stdout
		outputPaths = append(outputPaths, "stdout")
	}
	config := zap.Config{
		Level:            atom,
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    encoderConfig,
		OutputPaths:      outputPaths,
		ErrorOutputPaths: []string{"stderr"},
	}
	log, _ := config.Build()
	Logger = log.Sugar()
}
