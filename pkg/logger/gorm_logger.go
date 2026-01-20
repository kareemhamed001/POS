package logger

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type GormLogger struct {
	ZapLogger                 *zap.Logger
	LogLevel                  gormlogger.LogLevel
	SlowThreshold             time.Duration
	SkipCallerLookup          bool
	IgnoreRecordNotFoundError bool
}

func NewGormLogger(zapLogger *zap.Logger) *GormLogger {
	return &GormLogger{
		ZapLogger:                 zapLogger,
		LogLevel:                  gormlogger.Info,
		SlowThreshold:             200 * time.Millisecond,
		SkipCallerLookup:          false,
		IgnoreRecordNotFoundError: true,
	}
}

func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		l.ZapLogger.Sugar().Infof(msg, data...)
	}
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		l.ZapLogger.Sugar().Warnf(msg, data...)
	}
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		l.ZapLogger.Sugar().Errorf(msg, data...)
	}
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	//if error occurred and log level is Error and it's not record not found error or we are not ignoring record not found errors
	case err != nil && l.LogLevel >= gormlogger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		l.ZapLogger.Error("database error",
			zap.Error(err),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormlogger.Warn:
		l.ZapLogger.Warn("slow query",
			zap.Duration("elapsed", elapsed),
			zap.Duration("threshold", l.SlowThreshold),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	case l.LogLevel == gormlogger.Info:
		l.ZapLogger.Info("database query",
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	}
}

// ParamsFilter can be used to filter sensitive data from SQL queries
func (l *GormLogger) ParamsFilter(ctx context.Context, sql string, params ...interface{}) (string, []interface{}) {
	if l.LogLevel == gormlogger.Info {
		return sql, params
	}
	return sql, nil
}

// GetZapLogger returns the underlying zap logger
func GetZapLogger() *zap.Logger {
	logger := Get()
	return logger.Desugar()
}

// NewGormLoggerFromGlobal creates a GORM logger from the global zap logger
func NewGormLoggerFromGlobal() *GormLogger {
	return NewGormLogger(GetZapLogger())
}

// SetLogLevel sets the log level for the GORM logger
func (l *GormLogger) SetLogLevel(level gormlogger.LogLevel) *GormLogger {
	l.LogLevel = level
	return l
}

// SetSlowThreshold sets the slow query threshold
func (l *GormLogger) SetSlowThreshold(threshold time.Duration) *GormLogger {
	l.SlowThreshold = threshold
	return l
}

// SetIgnoreRecordNotFoundError sets whether to ignore record not found errors
func (l *GormLogger) SetIgnoreRecordNotFoundError(ignore bool) *GormLogger {
	l.IgnoreRecordNotFoundError = ignore
	return l
}
