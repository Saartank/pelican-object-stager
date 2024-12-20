package logger

import (
	"fmt"
	"os"
	"sync"

	"github.com/natefinch/lumberjack"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	baseLogger     *zap.Logger
	initLoggerOnce sync.Once
)

func InitLogger(serviceName, env string) {
	// Retrieve log level from config
	viper.SetDefault("log_level", "debug")
	logLevelStr := viper.GetString("log_level")

	// Convert config string to a zapcore.Level; default to Info if unknown
	var stdoutLevel zapcore.Level
	switch logLevelStr {
	case "debug":
		stdoutLevel = zap.DebugLevel
	case "info":
		stdoutLevel = zap.InfoLevel
	case "warn", "warning":
		stdoutLevel = zap.WarnLevel
	case "error":
		stdoutLevel = zap.ErrorLevel
	case "fatal":
		stdoutLevel = zap.FatalLevel
	case "panic":
		stdoutLevel = zap.PanicLevel
	default:
		// If the level is invalid, fallback to Info
		stdoutLevel = zap.InfoLevel
		fmt.Fprintf(os.Stderr, "Invalid log level '%s', defaulting to 'info'\n", logLevelStr)
	}

	// Encoder config for console logs (human-readable and color-coded)
	consoleEncoderCfg := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder, // Color-coded levels
		EncodeTime:     zapcore.ISO8601TimeEncoder,       // Human-readable time format
		EncodeDuration: zapcore.StringDurationEncoder,    // Human-readable durations
		EncodeCaller:   zapcore.ShortCallerEncoder,       // Shortened file paths
	}

	// Encoder config for JSON logs
	jsonEncoderCfg := zap.NewProductionEncoderConfig()
	jsonEncoderCfg.TimeKey = "timestamp"
	jsonEncoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	// Lumberjack for file rotation
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "/var/log/pelican-object-stager/pelican-object-stager.log",
		MaxSize:    100, // megabytes before rotation
		MaxBackups: 10,  // number of old files to keep
		MaxAge:     30,  // days to keep old logs
		Compress:   true,
	})

	// Create cores: one for stdout with dynamic level, one for file at Debug level
	stdoutCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEncoderCfg), // Console encoder for human-readable logs
		zapcore.AddSync(os.Stdout),
		stdoutLevel, // Dynamic level
	)

	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(jsonEncoderCfg), // JSON encoder for structured logs
		fileWriter,
		zap.DebugLevel, // Fixed level for file logs
	)

	// Tee the cores together
	combinedCore := zapcore.NewTee(stdoutCore, fileCore)

	// Create a base logger with common fields
	baseLogger = zap.New(combinedCore).With(
		zap.String("service", serviceName),
		zap.String("env", env),
	)

	// Now you can log a warning about the invalid log level if needed, as logger is ready:
	if logLevelStr != "" && logLevelStr != "debug" && logLevelStr != "info" &&
		logLevelStr != "warn" && logLevelStr != "warning" && logLevelStr != "error" &&
		logLevelStr != "fatal" && logLevelStr != "panic" {
		baseLogger.Warn("Invalid log level in config, defaulted to info", zap.String("provided_level", logLevelStr))
	}
	fmt.Println("Log init complete!")
}

func Base() *zap.Logger {
	initLoggerOnce.Do(func() {
		fmt.Println("Base logger initializing...")
		InitLogger("default-service", "development")
	})
	if baseLogger == nil {
		// Fallback to a no-op logger (unlikely to happen if InitLogger works)
		fmt.Fprintln(os.Stderr, "Base logger is nil. Falling back to no-op logger.")
		return zap.NewNop()
	}
	return baseLogger
}

// With returns a logger that includes additional fields.
func With(fields ...zap.Field) *zap.Logger {
	return Base().With(fields...)
}
