package appmetrics

import (
	"fmt"
	"os"
	"path/filepath"

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

	logFile, err := os.OpenFile(filepath.Join("logs", "test.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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

// metrics for Webserver and Binance availability
//
// serviceType:
//
//	1 - WebServer
//	2 - Binance
//
// isAvailable:
//
//	1 - available
//	0 - unavailable
//
// event - additional info
func AvailabilityMetricEvent(serviceType int, isAvailable int, event string) { // не остается открытым
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.LevelKey = "level"
	encoderCfg.MessageKey = "message"
	encoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoder := zapcore.NewJSONEncoder(encoderCfg)

	path := filepath.Join("logs", "AvailabilityMetrics.log")
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic("Failed to open metrics log: " + err.Error())
	}

	core := zapcore.NewCore(encoder, zapcore.AddSync(file), zapcore.InfoLevel)
	logger := zap.New(core)

	logger.Info("Metric event",
		zap.Int("type", serviceType),
		zap.String("event", event),
		zap.Int("isAvailable", isAvailable),
	)

	// очистка буферов и закрытие дескриптора файла
	_ = logger.Sync()
	_ = file.Close()
}

// добавить пуш уведомление об ошибке | если и прикручивать телегу, то сюда
func ApplicationCriticalErrorsLogging(event string, err error) { // не остается открытым
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.LevelKey = "level"
	encoderCfg.MessageKey = "message"
	encoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoder := zapcore.NewJSONEncoder(encoderCfg)

	path := filepath.Join("logs", "ApplicationCriticalErrors.log")
	file, err1 := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err1 != nil {
		panic("Failed to open metrics log: " + err.Error())
	}

	core := zapcore.NewCore(encoder, zapcore.AddSync(file), zapcore.InfoLevel)
	logger := zap.New(core)

	logger.Info("Metric event",
		zap.String("event", event),
		zap.String("error", fmt.Sprint(err)),
	)

	// очистка буферов и закрытие дескриптора файла
	_ = logger.Sync()
	_ = file.Close()
}
