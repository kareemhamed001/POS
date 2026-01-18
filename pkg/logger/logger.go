package logger

import (
	"log"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	*zap.SugaredLogger
}

var (
	globalLogger *Logger
	once         sync.Once
)

func New(env string) *Logger {
	var (
		base *zap.Logger
	)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.ConsoleSeparator = " | "

	lumberjackLogger := &lumberjack.Logger{
		Filename:   "system.log",
		MaxSize:    2,    // Megabytes
		MaxBackups: 10,   // Keep 3 old files
		MaxAge:     28,   // Days
		Compress:   true, // Gzip old logs
	}

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.AddSync(lumberjackLogger), zap.InfoLevel),
		zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), zapcore.AddSync(os.Stdout), zap.DebugLevel),
	)

	base = zap.New(core)

	return &Logger{base.Sugar()}
}

// InitGlobal initializes the global logger instance (call this once in main)
func InitGlobal(env string) {
	once.Do(func() {
		globalLogger = New(env)
	})
}

// Get returns the global logger instance
func Get() *Logger {
	if globalLogger == nil {
		log.Println("Warning: global logger not initialized, creating default logger")
		globalLogger = New("development")
	}
	return globalLogger
}

// Info logs an info message using the global logger
func Info(args ...interface{}) {
	Get().Info(args...)
}

// Infof logs a formatted info message using the global logger
func Infof(template string, args ...interface{}) {
	Get().Infof(template, args...)
}

// Error logs an error message using the global logger
func Error(args ...interface{}) {
	Get().Error(args...)
}

// Errorf logs a formatted error message using the global logger
func Errorf(template string, args ...interface{}) {
	Get().Errorf(template, args...)
}

// Warn logs a warning message using the global logger
func Warn(args ...interface{}) {
	Get().Warn(args...)
}

// Warnf logs a formatted warning message using the global logger
func Warnf(template string, args ...interface{}) {
	Get().Warnf(template, args...)
}

// Debug logs a debug message using the global logger
func Debug(args ...interface{}) {
	Get().Debug(args...)
}

// Debugf logs a formatted debug message using the global logger
func Debugf(template string, args ...interface{}) {
	Get().Debugf(template, args...)
}

func (l *Logger) Sync() {
	_ = l.SugaredLogger.Sync()
}

// Sync syncs the global logger
func Sync() {
	if globalLogger != nil {
		globalLogger.Sync()
	}
}
