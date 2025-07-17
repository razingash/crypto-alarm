package appmetrics

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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
func AvailabilityMetricEvent(serviceType int, isAvailable int, message string) { // не остается открытым
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
	DefaultLogging(1, message, []zap.Field{zap.Int("type", serviceType), zap.Int("isAvailable", isAvailable)}, logger)

	// очистка буферов и закрытие дескриптора файла
	_ = logger.Sync()
	_ = file.Close()
}

// добавить пуш уведомление об ошибке | если и прикручивать телегу, то сюда
func AnalyticsServiceLogging(debugLevel int, message string, err error) { // не остается открытым
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.LevelKey = "level"
	encoderCfg.MessageKey = "message"
	encoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoder := zapcore.NewJSONEncoder(encoderCfg)

	path := filepath.Join("logs", "AnalyticsServiceErrors.log")
	file, err1 := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err1 != nil {
		panic("Failed to open metrics log: " + err.Error())
	}

	core := zapcore.NewCore(encoder, zapcore.AddSync(file), zapcore.InfoLevel)
	logger := zap.New(core)

	DefaultLogging(debugLevel, message, []zap.Field{zap.String("error", fmt.Sprint(err))}, logger)

	_ = logger.Sync()
	_ = file.Close()
}

// логи ошибок самого приложения, связанные с выполнением команд и резком выключении/перезапуске приложения из-за ошибок
func ApplicationErrorsLogging(debugLevel int, message string, err error) { // не остается открытым
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.LevelKey = "level"
	encoderCfg.MessageKey = "message"
	encoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoder := zapcore.NewJSONEncoder(encoderCfg)

	path := filepath.Join("logs", "ApplicationErrors.log")
	file, err1 := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err1 != nil {
		panic("Failed to open metrics log: " + err.Error())
	}

	core := zapcore.NewCore(encoder, zapcore.AddSync(file), zapcore.InfoLevel)
	logger := zap.New(core)

	DefaultLogging(debugLevel, message, []zap.Field{zap.String("error", fmt.Sprint(err))}, logger)

	_ = logger.Sync()
	_ = file.Close()
}

// ошибки связанные с Binance
func BinanceErrorsLogging(debugLevel int, message string, err error) { // не остается открытым
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.LevelKey = "level"
	encoderCfg.MessageKey = "message"
	encoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoder := zapcore.NewJSONEncoder(encoderCfg)

	path := filepath.Join("logs", "BinanceErrors.log")
	file, err1 := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err1 != nil {
		panic("Failed to open metrics log: " + err.Error())
	}

	core := zapcore.NewCore(encoder, zapcore.AddSync(file), zapcore.InfoLevel)
	logger := zap.New(core)

	DefaultLogging(debugLevel, message, []zap.Field{zap.String("error", fmt.Sprint(err))}, logger)

	_ = logger.Sync()
	_ = file.Close()
}

/*
// Все ошибки связаные с запросами в базу данных
func DatabaseErrorsLogging(debugLevel int, message string, err error) { // не остается открытым
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.LevelKey = "level"
	encoderCfg.MessageKey = "message"
	encoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoder := zapcore.NewJSONEncoder(encoderCfg)

	path := filepath.Join("logs", "DatabaseErrorsLogging.log")
	file, err1 := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err1 != nil {
		panic("Failed to open metrics log: " + err.Error())
	}

	core := zapcore.NewCore(encoder, zapcore.AddSync(file), zapcore.InfoLevel)
	logger := zap.New(core)

	DefaultLogging(debugLevel, message, []zap.Field{zap.String("error", fmt.Sprint(err))}, logger)

	_ = logger.Sync()
	_ = file.Close()
}
*/

// loggingType:
//
// 1 - INFO
//
// 2 - WARN
//
// 3 - ERROR
//
// 4 - FATAL
func DefaultLogging(loggingType int, message string, logs []zap.Field, logger *zap.Logger) {
	switch loggingType {
	case 1:
		logger.Info(message, logs...)
	case 2:
		logger.Warn(message, logs...)
	case 3:
		logger.Error(message, logs...)
	case 4:
		logger.Fatal(message, logs...)
	default:
		logger.Info(message, logs...)
	}
}
