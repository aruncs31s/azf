package logger

import (
	"os"

	"github.com/aruncs31s/azf/constants"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func InitLogger() {
	if Log != nil {
		return
	}
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	// Ensure logs directory exists
	os.MkdirAll("logs", 0755)

	// Create or open log file
	logFile, err := os.OpenFile("logs/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic("Failed to open log file: " + err.Error())
	}

	var core zapcore.Core
	if os.Getenv("ENV") == constants.APP_PRODUCTION {
		// In production: write to both file (JSON) and stdout
		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(config),
			zapcore.AddSync(logFile),
			zap.InfoLevel,
		)
		stdoutCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(config),
			zapcore.AddSync(os.Stdout),
			zap.InfoLevel,
		)
		core = zapcore.NewTee(fileCore, stdoutCore)
	} else {
		// In development: write to both file (JSON) and stdout (console)
		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(config),
			zapcore.AddSync(logFile),
			zap.DebugLevel,
		)
		stdoutCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(config),
			zapcore.AddSync(os.Stdout),
			zap.DebugLevel,
		)
		core = zapcore.NewTee(fileCore, stdoutCore)
	}

	Log = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	Log.Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

func GetLogger() *zap.Logger {
	if Log == nil {
		InitLogger()
	}
	return Log
}
