package analytics

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func CustomLoggerSetup(debugMode bool) *zap.Logger {
	var logLevel zapcore.Level
	if debugMode {
		logLevel = zapcore.DebugLevel
	} else {
		logLevel = zapcore.InfoLevel
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder // читаемый timestamp
	encoderCfg.LevelKey = "level"
	encoderCfg.MessageKey = "message"
	encoderCfg.CallerKey = "caller"
	encoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder

	encoder := zapcore.NewJSONEncoder(encoderCfg)

	var cores []zapcore.Core

	logFile, err := os.OpenFile("../logs/test.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		fileCore := zapcore.NewCore(encoder, zapcore.AddSync(logFile), logLevel)
		cores = append(cores, fileCore)
	}

	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return logger
}

func DefaultLogging(loggingType int, message string) {
	logger := CustomLoggerSetup(true) // debugMode=true
	defer logger.Sync()               // очистка буфера

	switch loggingType {
	case 1:
		logger.Info("Приложение запущено",
			zap.String("message", message),
		)
	case 2:
		logger.Warn("Приложение запущено",
			zap.String("message", message),
		)
	case 3:
		logger.Error("Приложение запущено",
			zap.String("message", message),
		)
	case 4:
		logger.Fatal("Приложение запущено",
			zap.String("message", message),
		)
	default:
		panic("wrong integer in DefaultLogging(loggingType int), should be 1-4")
	}
}
